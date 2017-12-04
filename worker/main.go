package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/a-h/callme/metrics"
	"github.com/a-h/callme/repetitive"
	"github.com/a-h/callme/web"
	"github.com/cenkalti/backoff"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/a-h/callme/jobworker"
	"github.com/a-h/callme/scheduleworker"
	"github.com/a-h/callme/sns"

	"github.com/a-h/callme/logger"
	"github.com/a-h/callme/mysql"
)

const pkg = "github.com/a-h/callme/worker"

func init() {
	prometheus.MustRegister(metrics.JobCompletedCounts)
	prometheus.MustRegister(metrics.JobCompletedDurations)
	prometheus.MustRegister(metrics.JobExecutedCounts)
	prometheus.MustRegister(metrics.JobExecutedDelay)
	prometheus.MustRegister(metrics.JobExecutedDurations)
	prometheus.MustRegister(metrics.JobLeaseCounts)
	prometheus.MustRegister(metrics.JobLeaseDurations)

	prometheus.MustRegister(metrics.ScheduleExecutedCounts)
	prometheus.MustRegister(metrics.ScheduleExecutedDelay)
	prometheus.MustRegister(metrics.ScheduleJobStartedCounts)
	prometheus.MustRegister(metrics.ScheduleJobStartedDurations)
	prometheus.MustRegister(metrics.ScheduleLeaseCounts)
	prometheus.MustRegister(metrics.ScheduleLeaseDurations)
}

func main() {
	connectionString := os.Getenv("CALLME_CONNECTION_STRING")
	if connectionString == "" {
		logger.For(pkg, "main").Error("missing connection string environment variable (CALLME_CONNECTION_STRING)")
		os.Exit(-1)
	}
	scheduleWorkerCount := getIntegerSetting("CALLME_SCHEDULE_WORKER_COUNT", 1)
	jobWorkerCount := getIntegerSetting("CALLME_JOB_WORKER_COUNT", 1)
	lockExpiryMinutes := getIntegerSetting("CALLME_LOCK_EXPIRY_MINUTES", 30)
	prometheusPort := getIntegerSetting("CALLME_PROMETHEUS_PORT", 6666)

	var executor jobworker.Executor
	switch os.Getenv("CALLME_MODE") {
	case "web":
		logger.For(pkg, "main").Info("using web execution mode")
		executor = web.Execute
	case "sns":
	default:
		logger.For(pkg, "main").Info("using SNS (default) execution mode")
		executor = sns.Execute
	}

	totalProcesses := scheduleWorkerCount + jobWorkerCount
	logger.For(pkg, "main").
		WithField("totalProcessCount", totalProcesses).
		WithField("scheduleWorkerCount", scheduleWorkerCount).
		WithField("jobWorkerCount", jobWorkerCount).
		Info("starting processes")

	// Start serving metrics.
	go func() {
		http.Handle("/metrics", prometheus.Handler())
		http.ListenAndServe(fmt.Sprintf(":%v", prometheusPort), nil)
	}()

	sigs := make(chan os.Signal, 2)
	stopper := make(chan bool, totalProcesses)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = <-sigs
		for i := 0; i < totalProcesses; i++ {
			stopper <- true
		}
	}()

	schemaUpdate := func() error {
		mm := mysql.NewMigrationManager(connectionString)
		err := mm.UpdateSchema()
		if err != nil {
			logger.For(pkg, "main").WithError(err).Warn("failed to update schema, but will retry again")
		}
		return err
	}

	logger.For(pkg, "main").Info("checking database version and upgrading")

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = time.Minute * 5
	executionError := backoff.Retry(schemaUpdate, bo)
	if executionError != nil {
		logger.For(pkg, "main").WithError(executionError).Warn("update schema retry timeout exceeded")
		os.Exit(-1)
	}
	logger.For(pkg, "main").Info("updated schema, continuing")

	hostName, _ := os.Hostname()
	nodeName := fmt.Sprintf("callme_%v_%v", hostName, os.Getpid())

	waiter := make(chan bool, totalProcesses)

	logger.For(pkg, "main").Info("starting up schedulers and job workers")

	for i := 0; i < jobWorkerCount; i++ {
		go func(j int) {
			sm := mysql.NewScheduleManager(connectionString)
			scheduleWorkerFunction := scheduleworker.NewScheduleWorker(nodeName,
				lockExpiryMinutes,
				sm.GetSchedule,
				sm.StartJobAndUpdateCron)
			repetitive.Work(nodeName+"_schedules_"+strconv.Itoa(j), scheduleWorkerFunction, time.Second*5, stopper)
			waiter <- true
		}(i)
		waitForUpTo(50)
	}

	for i := 0; i < scheduleWorkerCount; i++ {
		jm := mysql.NewJobManager(connectionString)
		jobWorkerFunction := jobworker.NewJobWorker(nodeName,
			lockExpiryMinutes,
			jm.GetJob,
			executor,
			jm.CompleteJob)
		go func(j int) {
			repetitive.Work(nodeName+"_jobs_"+strconv.Itoa(j), jobWorkerFunction, time.Second*5, stopper)
			waiter <- true
		}(i)
		waitForUpTo(50)
	}

	logger.For(pkg, "main").Info("all processes started")
	for i := 0; i < totalProcesses; i++ {
		<-waiter
		logger.For(pkg, "main").Infof("shut down process %v of %v", i+1, totalProcesses)
	}
	logger.For(pkg, "main").Info("exiting application")
}

func getIntegerSetting(n string, def int) int {
	str := os.Getenv(n)
	if str == "" {
		logger.For(pkg, "getIntegerSetting").WithField("env", n).Info("environment variable not found, defaulting to %v", def)
		return def
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		logger.For(pkg, "getIntegerSetting").WithField("env", n).WithField("val", str).Warn("environment variable could not be pasrsed, defaulting to %v", def)
		return def
	}
	return int(i)
}

func waitForUpTo(ms int) {
	jitter, _ := time.ParseDuration(strconv.Itoa(rand.Intn(ms)) + "ms")
	time.Sleep(jitter)
}
