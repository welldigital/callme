package data

import (
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
