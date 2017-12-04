package scheduleworker

import (
	"time"

	"github.com/welldigital/callme/logger"
	"github.com/welldigital/callme/metrics"
	"github.com/welldigital/callme/repetitive"
	cron "gopkg.in/robfig/cron.v2"

	"github.com/welldigital/callme/data"
)

const pkg = "github.com/welldigital/callme/scheduleworker"

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
	scheduleGetStart := time.Now()
	sc, ok, err := scheduleGetter(workerName, lockExpiryMinutes)
	scheduleGetDuration := time.Since(scheduleGetStart) / time.Millisecond
	if err != nil {
		logger.For(pkg, "findAndExecuteWork").WithField("workerName", workerName).WithError(err).Error("failed to get schedule crontab")
		metrics.ScheduleLeaseCounts.WithLabelValues("error").Inc()
		metrics.ScheduleLeaseDurations.WithLabelValues("error").Observe(float64(scheduleGetDuration))
		return
	}
	if !ok {
		logger.For(pkg, "findAndExecuteWork").WithField("workerName", workerName).Info("no crontabs to update")
		metrics.ScheduleLeaseCounts.WithLabelValues("none_available").Inc()
		metrics.ScheduleLeaseDurations.WithLabelValues("none_available").Observe(float64(scheduleGetDuration))
		return
	}
	metrics.ScheduleLeaseCounts.WithLabelValues("success").Inc()
	metrics.ScheduleLeaseDurations.WithLabelValues("success").Observe(float64(scheduleGetDuration))

	c, err := cron.Parse(sc.Crontab.Crontab)
	if err != nil {
		logger.WithCrontab(pkg, "findAndExecuteWork", sc.Crontab).
			WithField("workerName", workerName).
			WithError(err).
			Error("failed to parse crontab, skipping")
		metrics.ScheduleExecutedCounts.WithLabelValues("error").Inc()
		return
	}
	metrics.ScheduleExecutedCounts.WithLabelValues("success").Inc()

	// Schedule a job to run immediately and update the cronjob.
	scheduleDelay := time.Now().UTC().Sub(sc.Crontab.Next) / time.Millisecond
	newNext := c.Next(sc.Crontab.Next)

	scheduledJobStartTime := time.Now()
	jobID, err := scheduledJobStarter(sc.Crontab.CrontabID, sc.Schedule.ScheduleID, sc.CrontabLeaseID, newNext)
	scheduledJobStartDuration := time.Since(scheduledJobStartTime) / time.Millisecond
	if err != nil || jobID == 0 {
		logger.For(pkg, "findAndExecuteWork").WithField("workerName", workerName).WithError(err).Error("failed to start job and update cron")
		metrics.ScheduleJobStartedCounts.WithLabelValues("error").Inc()
		metrics.ScheduleJobStartedDurations.WithLabelValues("error").Observe(float64(scheduledJobStartDuration))
		workDone = true
		return
	}
	metrics.ScheduleExecutedDelay.Observe(float64(scheduleDelay))
	metrics.ScheduleJobStartedCounts.WithLabelValues("success").Inc()
	metrics.ScheduleJobStartedDurations.WithLabelValues("success").Observe(float64(scheduledJobStartDuration))
	workDone = true
	return
}
