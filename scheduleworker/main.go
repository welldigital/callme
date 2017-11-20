package scheduleworker

import (
	"github.com/a-h/callme/logger"
	"github.com/a-h/callme/repetitive"
	cron "gopkg.in/robfig/cron.v2"

	"github.com/a-h/callme/data"
)

const leaseName = "schedule"

// NewScheduleWorker creates a worker for the repetitive.Work function which processes schedules and queues any required jobs.
func NewScheduleWorker(leaseAcquirer data.LeaseAcquirer,
	lockedBy string,
	leaseRescinder data.LeaseRescinder,
	scheduleGetter data.ScheduleGetter,
	scheduledJobStarter data.ScheduledJobStarter) repetitive.Worker {
	return func() (workDone bool, err error) {
		return findAndExecuteWork(leaseAcquirer, lockedBy, leaseRescinder, scheduleGetter, scheduledJobStarter)
	}
}

func findAndExecuteWork(leaseAcquirer data.LeaseAcquirer,
	lockedBy string,
	leaseRescinder data.LeaseRescinder,
	scheduleGetter data.ScheduleGetter,
	scheduledJobStarter data.ScheduledJobStarter,
) (workDone bool, err error) {
	leaseID, until, ok, err := leaseAcquirer(leaseName, lockedBy)
	if err != nil {
		logger.Errorf("scheduleworker: failed to acquire lease with error: %v", err)
		return
	}
	if !ok {
		logger.Infof("scheduleworker: no work to do, another process has the lease")
		return
	}
	logger.Infof("scheduleworker: got lease %v on %v until %v", leaseID, leaseName, until)
	defer leaseRescinder(leaseID)

	// See if there's some work to do.
	scheduleCrontabs, err := scheduleGetter()
	if err != nil {
		logger.Errorf("scheduleworker: failed to get schedules with error: %v", err)
		return
	}

	logger.Infof("scheduleworker: processing %v crontabs", len(scheduleCrontabs))

	for _, sc := range scheduleCrontabs {
		c, err := cron.Parse(sc.Crontab.Crontab)
		if err != nil {
			logger.WithCrontab(sc.Crontab).Errorf("scheduleworker: skipping crontab: failed to parse")
			continue
		}

		// Schedule a job to run immediately and update the cronjob.
		newNext := c.Next(sc.Crontab.Next)
		jobID, err := scheduledJobStarter(sc.Crontab.CrontabID, sc.Schedule.ScheduleID, newNext)
		if err != nil {
			logger.WithCrontab(sc.Crontab).Errorf("scheduleworker: failed to start job and update cron: %v", err)
			continue
		}
		if jobID == 0 {
			logger.WithCrontab(sc.Crontab).Error("scheduleworker: received a zero jobID when starting a job")
		}
	}
	return
}
