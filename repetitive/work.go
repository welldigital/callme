package repetitive

import (
	"time"

	"github.com/welldigital/callme/logger"
)

// Worker is a function which carries out some work.
type Worker func() (workDone bool, err error)

// Work runs until the stopper channel receives a stop signal.
func Work(name string, worker Worker,
	sleep time.Duration,
	stopper <-chan bool) {
	work(name, worker, sleep, func() { time.Sleep(sleep) }, stopper)
}

const pkg = "github.com/welldigital/callme/repetitive"

func work(name string, worker Worker,
	sleepFor time.Duration,
	sleep func(),
	stopper <-chan bool) {
	for {
		select {
		default:
			logger.For(pkg, "work").WithField("workerName", name).Info("executing worker")
			workDone, err := worker()
			if err != nil {
				logger.For(pkg, "work").WithField("workerName", name).WithError(err).Error("worker returned error")
				sleep()
				continue
			}
			if !workDone {
				logger.For(pkg, "work").WithField("workerName", name).Infof("work complete, sleeping for %v, next work at %v", sleepFor, time.Now().UTC().Add(sleepFor))
				sleep()
			}
		case <-stopper:
			logger.For(pkg, "work").WithField("workerName", name).Info("stop signal received")
			return
		}
	}
}
