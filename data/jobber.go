package data

import "time"

// JobStarter schedules a job to start in the future.
type JobStarter func(when time.Time, arn string, payload string, scheduleID *int64) (Job, error)

// JobGetter retrieves a job that's ready to run from the queue.
type JobGetter func(leaseID int64, now time.Time) (*Job, error)

// JobCompleter marks a job as complete.
type JobCompleter func(leaseID, jobID int64, now time.Time, resp string, err error) error
