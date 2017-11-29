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
		logger.Errorf("cmd.main: missing connection string environment variable (CALLME_CONNECTION_STRING)")
		os.Exit(-1)
	}
	scheduleWorkerCount := getIntegerSetting("CALLME_SCHEDULE_WORKER_COUNT", 1)
	jobWorkerCount := getIntegerSetting("CALLME_JOB_WORKER_COUNT", 1)
	lockExpiryMinutes := getIntegerSetting("CALLME_LOCK_EXPIRY_MINUTES", 30)
	prometheusPort := getIntegerSetting("CALLME_PROMETHEUS_PORT", 6666)

	var executor jobworker.Executor
	switch os.Getenv("CALLME_MODE") {
	case "web":
		logger.Infof("cmd.main: using web execution mode")
		executor = web.Execute
	case "sns":
	default:
		logger.Infof("cmd.main: using SNS (default) execution mode")
		executor = sns.Execute
	}

	totalProcesses := scheduleWorkerCount + jobWorkerCount
	logger.Infof("cmd.main: starting %v processes - %v schedule workers and %v job workers", totalProcesses, scheduleWorkerCount, jobWorkerCount)

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
			logger.Errorf("cmd.main: failed to update schema, but will retry again, err: %v", err)
		}
		return err
	}

	logger.Infof("cmd.main: checking database version and upgrading")

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = time.Minute * 5
	executionError := backoff.Retry(schemaUpdate, bo)
	if executionError != nil {
		logger.Errorf("cmd.main: update schema retry timeout exceeded, logging error: %v", executionError)
		os.Exit(-1)
	}
	logger.Infof("cmd.main: updated schema, continuing")

	hostName, _ := os.Hostname()
	nodeName := fmt.Sprintf("callme_%v_%v", hostName, os.Getpid())

	waiter := make(chan bool, totalProcesses)

	logger.Infof("cmd.main: starting up schedulers and job workers")

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

	logger.Infof("cmd.main: all processes started")
	for i := 0; i < totalProcesses; i++ {
		<-waiter
		logger.Infof("cmd.main: shut down process %v of %v", i+1, totalProcesses)
	}
	logger.Infof("cmd.main: exiting application")
}

func getIntegerSetting(n string, def int) int {
	str := os.Getenv(n)
	if str == "" {
		logger.Infof("cmd.main: %v environment variable not found, defaulting to %v", n, def)
		return def
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		logger.Warnf("cmd.main: %v environment variable: '%v' could not be parsed, defaulting to %v", n, str, def)
		return def
	}
	return int(i)
}

func waitForUpTo(ms int) {
	jitter, _ := time.ParseDuration(strconv.Itoa(rand.Intn(ms)) + "ms")
	time.Sleep(jitter)
}
