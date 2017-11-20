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

	"github.com/a-h/callme/jobworker"
	"github.com/a-h/callme/scheduleworker"
	"github.com/a-h/callme/sns"

	"github.com/a-h/callme/logger"
	"github.com/a-h/callme/mysql"
)

func main() {
	sigs := make(chan os.Signal, 2)
	stopper := make(chan bool, 2)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = <-sigs
		stopper <- true
		stopper <- true
	}()

	logger.Infof("cmd.main: checking database version and upgrading")
	connectionString := os.Getenv("CALLME_CONNECTION_STRING")

	if connectionString == "" {
		logger.Errorf("cmd.main: missing connection string")
		os.Exit(-1)
	}

	mm := mysql.NewMigrationManager(connectionString)
	err := mm.UpdateSchema()
	if err != nil {
		logger.Errorf("cmd.main: failed to update schema with error: %v", err)
		os.Exit(-1)
	}

	hostName, _ := os.Hostname()
	nodeName := fmt.Sprintf("callme_%v_%v", hostName, os.Getpid())

	lm := mysql.NewLeaseManager(connectionString)
	jm := mysql.NewJobManager(connectionString)
	sm := mysql.NewScheduleManager(connectionString)

	waiter := make(chan bool, 2)

	logger.Infof("cmd.main: starting up scheduler and job worker")

	scheduleWorkerFunction := scheduleworker.NewScheduleWorker(lm.Acquire,
		nodeName,
		lm.Rescind,
		sm.GetSchedules,
		jm.StartJob,
		sm.UpdateCron)

	go func() {
		repetitive.Work("schedules", scheduleWorkerFunction, func() { time.Sleep(time.Minute) }, stopper)
		waiter <- true
	}()

	jobWorkerFunction := jobworker.NewJobWorker(lm.Acquire,
		nodeName,
		lm.Rescind,
		jm.GetJob,
		sns.Execute,
		jm.CompleteJob)

	go func() {
		jitter, _ := time.ParseDuration(strconv.Itoa(rand.Intn(1000)) + "ms")
		time.Sleep(jitter)
		repetitive.Work("jobs", jobWorkerFunction, repetitive.WaitForFiveSeconds, stopper)
		waiter <- true
	}()

	<-waiter
	<-waiter
	logger.Infof("cmd.main: exiting application")
}
