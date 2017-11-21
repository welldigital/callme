package data

import (
	"time"
)

// ScheduleCreator schedules a job to repeat.
type ScheduleCreator func(from time.Time, arn string, payload string, crontabs []string, externalID, by string) (scheduleID int64, err error)

// ScheduleDeactivator stops a schedule from functioning and deletes scheduled tasks belonging to it.
type ScheduleDeactivator func(scheduleID int64) error

// ScheduleGetter gets all schedules where Next is in the past, in order to schedule jobs.
type ScheduleGetter func() ([]ScheduleCrontab, error)

// ScheduledJobStarter starts a new job and updates a Crontab record in a transaction so that it's not included in future updates.
type ScheduledJobStarter func(leaseID int64, crontabID int64, scheduleID int64, newNext time.Time) (jobID int64, err error)
