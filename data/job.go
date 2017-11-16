package data

import "time"

// A Job is some work for an executor to do.
//
// A job can be delayed by setting the When field to be in the future.
// A worker attempts to put an entry into a JobLease to grab a job.
type Job struct {
	JobID      int64
	ScheduleID *int64
	When       time.Time
	ARN        string
	Payload    string
}

// A JobLease is a record of a worker which has claimed jobs. If a lease is in play, then
// no processing of jobs by another agent can be done.
//TODO: In future, consider adding sharding to jobs so that multiple workers can grab jobs.
type JobLease struct {
	JobLeaseID int64
	LockedBy   string
	At         time.Time
	Until      time.Time
}

// A JobResponse records an execution of the Job.
type JobResponse struct {
	JobResponseID int64
	JobLeaseID    int64
	JobID         int64
	Time          time.Time
	Response      string
	IsError       bool
	Error         string
}
