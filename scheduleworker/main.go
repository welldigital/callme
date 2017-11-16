package scheduleworker

import (
	"time"

	"github.com/a-h/callme/logger"
	"github.com/a-h/callme/repetitive"
	cron "gopkg.in/robfig/cron.v2"

	"github.com/a-h/callme/data"
)

// NewScheduleWorker creates a worker for the repetitive.Work function which processes schedules and queues any required jobs.
func NewScheduleWorker(now func() time.Time,
	scheduleLeaseAcquirer data.ScheduleLeaseAcquirer,
	lockedBy string,
	scheduleLeaseRescinder data.ScheduleLeaseRescinder,
	scheduleGetter data.ScheduleGetter,
	jobStarter data.JobStarter,
	cronUpdater data.CronUpdater) repetitive.Worker {
	return func() (workDone bool, err error) {
		return findAndExecuteWork(now, scheduleLeaseAcquirer, lockedBy, scheduleLeaseRescinder, scheduleGetter, jobStarter, cronUpdater)
	}
}

func findAndExecuteWork(clock func() time.Time,
	scheduleLeaseAcquirer data.ScheduleLeaseAcquirer,
	lockedBy string,
	scheduleLeaseRescinder data.ScheduleLeaseRescinder,
	scheduleGetter data.ScheduleGetter,
	jobStarter data.JobStarter,
	cronUpdater data.CronUpdater,
) (workDone bool, err error) {
	now := clock()
	scheduleLeaseID, until, err := scheduleLeaseAcquirer(now, lockedBy)
	if err != nil {
		logger.Errorf("scheduleworker: failed to acquire lease with error: %v", err)
		return
	}
	logger.Infof("scheduleworker: got lease %v until %v", scheduleLeaseID, until)
	defer scheduleLeaseRescinder(scheduleLeaseID)

	// See if there's some work to do.
	scheduleCrontabs, err := scheduleGetter(scheduleLeaseID, now)
	if err != nil {
		logger.Errorf("scheduleworker: failed to get schedules with error: %v", err)
		return
	}

	logger.Infof("scheduleworker: processing %v crontabs", len(scheduleCrontabs))

	for _, sc := range scheduleCrontabs {
		if !needsUpdating(sc.Crontab, now) {
			logger.WithCrontab(sc.Crontab).Debugf("scheduleworker: skipping crontab: it has not yet expired")
			continue
		}

		c, err := cron.Parse(sc.Crontab.Crontab)
		if err != nil {
			logger.WithCrontab(sc.Crontab).Errorf("scheduleworker: skipping crontab: failed to parse")
			continue
		}

		// Schedule a job to run immediately.
		job, err := jobStarter(now, sc.Schedule.ARN, sc.Schedule.Payload, &sc.Crontab.ScheduleID)
		if err != nil {
			logger.WithCrontab(sc.Crontab).Errorf("scheduleworker: failed to start: %v", err)
			continue
		}
		logger.WithJob(job).Infof("sceduler.Process: started job")

		logger.WithCrontab(sc.Crontab).Infof("sceduler.Process: updating crontab to run again in the future")
		newPrevious := sc.Crontab.Next
		newNext := c.Next(sc.Crontab.Next)
		newLastUpdated := now
		err = cronUpdater(sc.Crontab.CrontabID, newPrevious, newNext, newLastUpdated)
		if err != nil {
			logger.WithCrontab(sc.Crontab).Errorf("scheduleworker: failed to update cron: %v", err)
		}
	}
	return
}

func needsUpdating(crontab data.Crontab, now time.Time) bool {
	return crontab.Next.Before(now)
}
