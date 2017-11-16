package repetitive

import (
	"time"

	"github.com/a-h/callme/logger"
)

// Worker is a function which carries out some work.
type Worker func() (workDone bool, err error)

// WaitForFiveSeconds sleeps for 5 seconds.
var WaitForFiveSeconds = func() { time.Sleep(5 * time.Second) }

// Work runs until the stopper channel receives a stop signal.
func Work(name string, worker Worker,
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
				logger.Infof("%s: work complete, sleeping", name)
				sleep()
			}
		case <-stopper:
			logger.Infof("%s: stop signal received", name)
			return
		}
	}
}
