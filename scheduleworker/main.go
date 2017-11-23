package scheduleworker

import (
	"github.com/a-h/callme/logger"
	"github.com/a-h/callme/repetitive"
	cron "gopkg.in/robfig/cron.v2"

	"github.com/a-h/callme/data"
)

const leaseName = "schedule"

// NewScheduleWorker creates a worker for the repetitive.Work function which processes schedules and queues any required jobs.
func NewScheduleWorker(lockedBy string,
	scheduleGetter data.ScheduleGetter,
	scheduledJobStarter data.ScheduledJobStarter) repetitive.Worker {
	return func() (workDone bool, err error) {
		return findAndExecuteWork(lockedBy, scheduleGetter, scheduledJobStarter)
	}
}

func findAndExecuteWork(lockedBy string,
	scheduleGetter data.ScheduleGetter,
	scheduledJobStarter data.ScheduledJobStarter,
) (workDone bool, err error) {
	// See if there's some work to do.
	sc, ok, err := scheduleGetter(lockedBy)
	if err != nil {
		logger.Errorf("scheduleworker: failed to get schedule crontab with error: %v", err)
		return
	}
	if !ok {
		logger.Infof("scheduleworker: no crontabs to update")
		return
	}

	c, err := cron.Parse(sc.Crontab.Crontab)
	if err != nil {
		logger.WithCrontab(sc.Crontab).Errorf("scheduleworker: skipping crontab: failed to parse: '%v'", sc.Crontab.Crontab)
		return
	}

	// Schedule a job to run immediately and update the cronjob.
	newNext := c.Next(sc.Crontab.Next)
	jobID, err := scheduledJobStarter(sc.Crontab.CrontabID, sc.Schedule.ScheduleID, sc.CrontabLeaseID, newNext)
	if err != nil || jobID == 0 {
		logger.WithCrontab(sc.Crontab).Errorf("scheduleworker: failed to start job and update cron: %v", err)
	}
	workDone = true
	return
}
