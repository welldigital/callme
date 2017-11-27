package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/a-h/callme/repetitive"
	"github.com/a-h/callme/web"

	"github.com/a-h/callme/jobworker"
	"github.com/a-h/callme/scheduleworker"
	"github.com/a-h/callme/sns"

	"github.com/a-h/callme/logger"
	"github.com/a-h/callme/mysql"
)

func main() {
	scheduleWorkerCount := getIntegerSetting("CALLME_SCHEDULE_WORKER_COUNT", 1)
	jobWorkerCount := getIntegerSetting("CALLME_JOB_WORKER_COUNT", 1)

	totalProcesses := scheduleWorkerCount + jobWorkerCount
	logger.Infof("cmd.main: starting %v processes - %v schedule workers and %v job workers", totalProcesses, scheduleWorkerCount, jobWorkerCount)

	lockExpiryMinutes := getIntegerSetting("CALLME_LOCK_EXPIRY_MINUTES", 30)

	sigs := make(chan os.Signal, 2)
	stopper := make(chan bool, totalProcesses)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = <-sigs
		for i := 0; i < totalProcesses; i++ {
			stopper <- true
		}
	}()

	logger.Infof("cmd.main: checking database version and upgrading")
	connectionString := os.Getenv("CALLME_CONNECTION_STRING")

	if connectionString == "" {
		logger.Errorf("cmd.main: missing connection string")
		os.Exit(-1)
	}

	var exectutor jobworker.Executor
	switch os.Getenv("CALLME_MODE") {
	case "web":
		logger.Infof("cmd.main: using web execution mode")
		exectutor = web.Execute
	case "sns":
	default:
		logger.Infof("cmd.main: using SNS execution mode")
		exectutor = sns.Execute
	}

	mm := mysql.NewMigrationManager(connectionString)
	err := mm.UpdateSchema()
	if err != nil {
		logger.Errorf("cmd.main: failed to update schema with error: %v", err)
		os.Exit(-1)
	}

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
			exectutor,
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
