package data

import "time"

// Schedule represents a recurring schedule for a job.
// The Minute, Hour, Weekday and MonthDay schedules are distinct, meaning that
// if there's MinuteSchedules of 0 and 15, and HourlySchedules of
type Schedule struct {
	// ScheduleID is the unique database id for the record.
	ScheduleID int64 `json:"scheduleId"`
	// ExternalID links this record to external systems, max length 256.
	ExternalID string `json:"externalId"`
	// By tracks which system made this record, max length 256.
	By string `json:"by"`
	// ARN is the Amazon Resource Name that will be initiated, e.g. a Lambda ARN or SNS queue ARN.
	ARN string `json:"arn"`
	// Payload is the payload sent to the ARN (resource).
	Payload string `json:"payload"`
	// Created is the date that the record was created.
	Created time.Time `json:"created"`
	// Active stores whether the schedule is active or not.
	Active bool `json:"active"`
	// DeactivatedDate returns the date that the schedule was disabled.
	DeactivatedDate time.Time `json:"deactivatedDate"`
}

// Crontab contains the schedule data and when it was last executed.
//
// When a crontab record is created, the Next value is set to the value after the
// From value of the Schedule, and the LastUpdated value is set to the current time.
//
// Periodically, a process grabs records which need to be scheduled, i.e. where
// the Next is in the past. The process then:
//   * Calculates the "new Next value"
//   * Schedules a Job to start immediately.
//   * Places the "old Next value" into the "LastUpdated" field.
//   * Places the "new Next value" into the "Next" field to schedule the next refresh of the cron schedule.
type Crontab struct {
	CrontabID   int64     `json:"crontabId"`
	ScheduleID  int64     `json:"scheduleId"`
	Crontab     string    `json:"crontab"`
	Previous    time.Time `json:"previous"`
	Next        time.Time `json:"next"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// ScheduleCrontab is the Crontab data with its matching schedule attached.
// It's used by the ScheduleGetter to get a lease on a schedule, and isn't
// designed to be marshalled across an API.
type ScheduleCrontab struct {
	Schedule       Schedule
	Crontab        Crontab
	CrontabLeaseID int64
}

// ScheduleCrontabs is the schedule data with its matching crontabs attached.
type ScheduleCrontabs struct {
	Schedule Schedule  `json:"schedule"`
	Crontabs []Crontab `json:"crontabs"`
}
