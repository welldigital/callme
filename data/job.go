package data

import (
	"errors"
	"math"
	"time"
)

// A Job is some work for an executor to do.
//
// A job can be delayed by setting the When field to be in the future.
// A worker must acquire a lease before it should grab a job.
type Job struct {
	JobID      int64
	ScheduleID *int64
	When       time.Time
	ARN        string
	Payload    string
}

// Validate validates that the job could be stored in the backing store.
func (j Job) Validate() error {
	if j.ARN == "" {
		return errors.New("an ARN is required")
	}
	if len(j.ARN) > 2048 {
		return errors.New("maximum length of the ARN is 2048 characters")
	}
	if len(j.Payload) > int(math.Pow(2, 24)-1) {
		return errors.New("exceeded maximum length of payload")
	}
	return nil
}

// A JobResponse records an execution of the Job.
type JobResponse struct {
	JobResponseID int64
	JobID         int64
	Time          time.Time
	Response      string
	IsError       bool
	Error         string
}

// JobAndResponse contains a job, and an optional response.
type JobAndResponse struct {
	Job            Job
	JobResponse    JobResponse
	HasJobResponse bool
}
