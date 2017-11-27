package repetitive

import (
	"time"

	"github.com/a-h/callme/logger"
)

// Worker is a function which carries out some work.
type Worker func() (workDone bool, err error)

// Work runs until the stopper channel receives a stop signal.
func Work(name string, worker Worker,
	sleep time.Duration,
	stopper <-chan bool) {
	work(name, worker, sleep, func() { time.Sleep(sleep) }, stopper)
}

func work(name string, worker Worker,
	sleepFor time.Duration,
	sleep func(),
	stopper <-chan bool) {
	for {
		select {
		default:
			logger.Infof("%s: executing worker", name)
			workDone, err := worker()
			if err != nil {
				logger.Errorf("%s: worker returned error: %v", name, err)
				sleep()
				continue
			}
			if !workDone {
				logger.Infof("%s: work complete, sleeping for %v, next work at %v", name, sleepFor, time.Now().UTC().Add(sleepFor))
				sleep()
			}
		case <-stopper:
			logger.Infof("%s: stop signal received", name)
			return
		}
	}
}
