package data

import (
	"time"
)

// JobStarter schedules a job to start in the future.
type JobStarter func(when time.Time, arn string, payload string, scheduleID *int64) (Job, error)

// JobGetter retrieves a job that's ready to run from the queue.
type JobGetter func(lockedBy string, lockExpiryMinutes int) (j Job, ok bool, err error)

// JobCompleter marks a job as complete.
type JobCompleter func(jobID int64, resp string, err error) error

// JobAndResponseByIDGetter gets a job and its response by its Job ID.
type JobAndResponseByIDGetter func(jobID int64) (j Job, r JobResponse, jobOK, responseOK bool, err error)

// JobDeleter deletes a job that hasn't yet been completed or locked for processing.
type JobDeleter func(jobID int64) (ok bool, err error)
