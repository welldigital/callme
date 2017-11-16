package data

import (
	"time"
)

// ScheduleCreator schedules a job to repeat.
type ScheduleCreator func(from time.Time, arn string, payload string, crontabs []string, externalID, by string) (scheduleID int64, err error)

// ScheduleDeactivator stops a schedule from functioning and deletes scheduled tasks belonging to it.
type ScheduleDeactivator func(scheduleID int64) error

// ScheduleGetter gets all schedules where Next is in the past, in order to schedule jobs.
type ScheduleGetter func(scheduleLeaseID int64, now time.Time) ([]ScheduleCrontab, error)

// CronUpdater updates a Crontab record so that it's not included in future updates.
type CronUpdater func(crontabID int64, newPrevious, newNext, newLastUpdated time.Time) error

// ScheduleLeaseAcquirer gets a lease to update the schedule and queue jobs.
type ScheduleLeaseAcquirer func(now time.Time, by string) (scheduleLeaseID int64, until time.Time, err error)

// ScheduleLeaseRescinder rescinds the right on a lease.
type ScheduleLeaseRescinder func(scheduleLeaseID int64) (err error)
