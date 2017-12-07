package data

import (
	"time"
)

// A Job is some work for an executor to do.
//
// A job can be delayed by setting the When field to be in the future.
// A worker must acquire a lease before it should grab a job.
type Job struct {
	JobID      int64     `json:"jobId"`
	ScheduleID *int64    `json:"scheduleId"`
	When       time.Time `json:"when"`
	ARN        string    `json:"arn"`
	Payload    string    `json:"payload"`
}

// A JobResponse records an execution of the Job.
type JobResponse struct {
	JobResponseID int64     `json:"jobResponseId"`
	JobID         int64     `json:"jobId"`
	Time          time.Time `json:"time"`
	Response      string    `json:"response"`
	IsError       bool      `json:"isError"`
	Error         string    `json:"error"`
}

// JobAndResponse contains a job, and an optional response.
type JobAndResponse struct {
	Job            Job         `json:"job"`
	JobResponse    JobResponse `json:"response"`
	HasJobResponse bool        `json:"hasJobResponse"`
}
