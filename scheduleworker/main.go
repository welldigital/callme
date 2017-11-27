package scheduleworker

import (
	"github.com/a-h/callme/logger"
	"github.com/a-h/callme/repetitive"
	cron "gopkg.in/robfig/cron.v2"

	"github.com/a-h/callme/data"
)

// NewScheduleWorker creates a worker for the repetitive.Work function which processes schedules and queues any required jobs.
func NewScheduleWorker(workerName string,
	lockExpiryMinutes int,
	scheduleGetter data.ScheduleGetter,
	scheduledJobStarter data.ScheduledJobStarter) repetitive.Worker {
	return func() (workDone bool, err error) {
		return findAndExecuteWork(workerName, lockExpiryMinutes, scheduleGetter, scheduledJobStarter)
	}
}

func findAndExecuteWork(workerName string,
	lockExpiryMinutes int,
	scheduleGetter data.ScheduleGetter,
	scheduledJobStarter data.ScheduledJobStarter,
) (workDone bool, err error) {
	// See if there's some work to do.
	sc, ok, err := scheduleGetter(workerName, lockExpiryMinutes)
	if err != nil {
		logger.Errorf("%v: failed to get schedule crontab with error: %v", workerName, err)
		return
	}
	if !ok {
		logger.Infof("%v: no crontabs to update", workerName)
		return
	}

	c, err := cron.Parse(sc.Crontab.Crontab)
	if err != nil {
		logger.WithCrontab(sc.Crontab).Errorf("%v: skipping crontab: failed to parse: '%v'", workerName, sc.Crontab.Crontab)
		return
	}

	// Schedule a job to run immediately and update the cronjob.
	newNext := c.Next(sc.Crontab.Next)
	jobID, err := scheduledJobStarter(sc.Crontab.CrontabID, sc.Schedule.ScheduleID, sc.CrontabLeaseID, newNext)
	if err != nil || jobID == 0 {
		logger.WithCrontab(sc.Crontab).Errorf("%v: failed to start job and update cron: %v", workerName, err)
	}
	workDone = true
	return
}
