package data

import "time"

// Schedule represents a recurring schedule for a job.
// The Minute, Hour, Weekday and MonthDay schedules are distinct, meaning that
// if there's MinuteSchedules of 0 and 15, and HourlySchedules of
type Schedule struct {
	// ScheduleID is the unique database id for the record.
	ScheduleID int64
	// ExternalID links this record to external systems, max length 256.
	ExternalID string
	// By tracks which system made this record, max length 256.
	By string
	// ARN is the Amazon Resource Name that will be initiated, e.g. a Lambda ARN or SNS queue ARN.
	ARN string
	// Payload is the payload sent to the ARN (resource).
	Payload string
	// Created is the date that the record was created.
	Created time.Time
	// From is the time the the schedule starts from.
	From time.Time
	// Active stores whether the schedule is active or not.
	Active bool
	// DeactivatedDate returns the date that the schedule was disabled.
	DeactivatedDate time.Time
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
	CrontabID   int64
	ScheduleID  int64
	Crontab     string
	Previous    time.Time
	Next        time.Time
	LastUpdated time.Time
}

// ScheduleCrontab is the Crontab data with its matching schedule attached.
type ScheduleCrontab struct {
	Schedule Schedule
	Crontab  Crontab
}
