package data

import "time"

// JobStarter schedules a job to start in the future.
type JobStarter func(when time.Time, arn string, payload string, scheduleID *int64) (Job, error)

// JobGetter retrieves a job that's ready to run from the queue.
type JobGetter func(lockedBy string) (j Job, ok bool, err error)

// JobCompleter marks a job as complete.
type JobCompleter func(jobID int64, resp string, err error) error
