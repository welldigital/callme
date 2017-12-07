package data

import (
	"time"
)

// ScheduleCreator schedules a job to repeat.
type ScheduleCreator func(from time.Time, arn string, payload string, crontabs []string, externalID, by string) (scheduleID int64, err error)

// ScheduleDeactivator stops a schedule from functioning and deletes scheduled tasks belonging to it.
type ScheduleDeactivator func(scheduleID int64) (ok bool, err error)

// ScheduleByIDGetter gets the schedule specified in the ID.
type ScheduleByIDGetter func(scheduleID int64) (sc ScheduleCrontabs, ok bool, err error)

// ScheduleGetter gets a schedule and the next crontab which is due to start, in order to schedule jobs.
type ScheduleGetter func(lockedBy string, lockExpiryMinutes int) (sc ScheduleCrontab, ok bool, err error)

// ScheduledJobStarter starts a new job and updates a Crontab record in a transaction so that it's not included in future updates.
type ScheduledJobStarter func(crontabID, scheduleID, crontabLeaseID int64, newNext time.Time) (jobID int64, err error)
