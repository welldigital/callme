package data

import "time"

// JobStarter schedules a job to start in the future.
type JobStarter func(when time.Time, arn string, payload string, scheduleID *int64) (Job, error)

// JobGetter retrieves a job that's ready to run from the queue.
type JobGetter func(jobLeaseID int64, now time.Time) (*Job, error)

// JobCompleter marks a job as complete.
type JobCompleter func(jobLeaseID, jobID int64, now time.Time, resp string, err error) error

// JobLeaseAcquirer gets a lease to process jobs.
type JobLeaseAcquirer func(now time.Time, lockedBy string) (jobLeaseID int64, until time.Time, err error)

// JobLeaseRescinder rescinds the right on a lease.
type JobLeaseRescinder func(jobLeaseID int64) (err error)
