package scheduleworker

import (
	"errors"
	"testing"
	"time"

	"github.com/welldigital/callme/data"
)

const nodeName = "scheduleworker_test"

const lockExpiryMins = 5

func TestThatErrorsRetrievingSchedulesAreReturned(t *testing.T) {
	actual := Values{}

	scheduleGetter := func(lockedBy string, lockExpiryMinutes int) (sc data.ScheduleCrontab, ok bool, err error) {
		actual.ScheduleRetrieved = true
		err = errors.New("error getting schedules")
		return
	}

	scheduledJobStarter := func(crontabID, scheduleID, crontabLeaseID int64, newNext time.Time) (jobID int64, err error) {
		actual.JobStarted = true
		actual.NextTime = newNext
		return 1, nil
	}

	w := NewScheduleWorker(nodeName, lockExpiryMins, scheduleGetter, scheduledJobStarter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		WorkDone:          false,
		ScheduleRetrieved: true,
		JobStarted:        false,
		NextTime:          time.Time{},
		ErrorOccurred:     true,
	}

	expected.Assert(t, actual)
}

func TestThatIfNoJobsAreFoundNoUpdatesAreMade(t *testing.T) {
	actual := Values{}

	scheduleGetter := func(lockedBy string, lockExpiryMinutes int) (sc data.ScheduleCrontab, ok bool, err error) {
		actual.ScheduleRetrieved = true
		ok = false
		return
	}

	scheduledJobStarter := func(crontabID, scheduleID, crontabLeaseID int64, newNext time.Time) (jobID int64, err error) {
		actual.JobStarted = true
		actual.NextTime = newNext
		return 1, nil
	}

	w := NewScheduleWorker(nodeName, lockExpiryMins, scheduleGetter, scheduledJobStarter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		WorkDone:          false,
		ScheduleRetrieved: true,
		JobStarted:        false,
		NextTime:          time.Time{},
		ErrorOccurred:     false,
	}

	expected.Assert(t, actual)
}

func TestThatExpiredSchedulesStartNewJobs(t *testing.T) {
	actual := Values{}

	now := time.Now().UTC()

	scheduleGetter := func(lockedBy string, lockExpiryMinutes int) (sc data.ScheduleCrontab, ok bool, err error) {
		actual.ScheduleRetrieved = true
		sc = data.ScheduleCrontab{
			Schedule: data.Schedule{
				ScheduleID:      1,
				Active:          true,
				ARN:             "testarn",
				By:              "scheduleworker.main_test",
				Created:         time.Now().UTC(),
				DeactivatedDate: time.Time{},
				ExternalID:      "externalid",
				Payload:         "testpayload",
			},
			Crontab: data.Crontab{
				Crontab:     "0 * * * *", // once per hour
				CrontabID:   1,
				LastUpdated: now,
				Next:        now,
				Previous:    now.Add(-1 * time.Hour).UTC(),
				ScheduleID:  1,
			},
			CrontabLeaseID: 1,
		}
		ok = true
		return
	}

	scheduledJobStarter := func(crontabID, scheduleID, crontabLeaseID int64, newNext time.Time) (jobID int64, err error) {
		actual.JobStarted = true
		actual.NextTime = newNext
		return 1, nil
	}

	w := NewScheduleWorker(nodeName, lockExpiryMins, scheduleGetter, scheduledJobStarter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		WorkDone:          true,
		ScheduleRetrieved: true,
		JobStarted:        true,
		NextTime:          time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location()),
		ErrorOccurred:     false,
	}

	expected.Assert(t, actual)
}

func TestThatErrorsParsingCronStatementsAreTracked(t *testing.T) {
	actual := Values{}

	now := time.Now().UTC()

	scheduleGetter := func(lockedBy string, lockExpiryMinutes int) (sc data.ScheduleCrontab, ok bool, err error) {
		actual.ScheduleRetrieved = true
		sc = data.ScheduleCrontab{
			Schedule: data.Schedule{
				ScheduleID:      1,
				Active:          true,
				ARN:             "testarn",
				By:              "scheduleworker.main_test",
				Created:         time.Now().UTC(),
				DeactivatedDate: time.Time{},
				ExternalID:      "externalid",
				Payload:         "testpayload",
			},
			Crontab: data.Crontab{
				Crontab:     "absolute nonsense",
				CrontabID:   1,
				LastUpdated: now,
				Next:        now,
				Previous:    now.Add(-1 * time.Hour).UTC(),
				ScheduleID:  1,
			},
			CrontabLeaseID: 1,
		}
		ok = true
		return
	}

	scheduledJobStarter := func(crontabID, scheduleID, crontabLeaseID int64, newNext time.Time) (jobID int64, err error) {
		actual.JobStarted = true
		actual.NextTime = newNext
		return 1, nil
	}

	w := NewScheduleWorker(nodeName, lockExpiryMins, scheduleGetter, scheduledJobStarter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		WorkDone:          false,
		ScheduleRetrieved: true,
		JobStarted:        false,
		NextTime:          time.Time{},
		ErrorOccurred:     true,
	}

	expected.Assert(t, actual)
}

func TestThatErrorsStartingWorkDoNotBlock(t *testing.T) {
	actual := Values{}

	now := time.Now().UTC()

	scheduleGetter := func(lockedBy string, lockExpiryMinutes int) (sc data.ScheduleCrontab, ok bool, err error) {
		actual.ScheduleRetrieved = true
		sc = data.ScheduleCrontab{
			Schedule: data.Schedule{
				ScheduleID:      1,
				Active:          true,
				ARN:             "testarn",
				By:              "scheduleworker.main_test",
				Created:         time.Now().UTC(),
				DeactivatedDate: time.Time{},
				ExternalID:      "externalid",
				Payload:         "testpayload",
			},
			Crontab: data.Crontab{
				Crontab:     "0 * * * *", // once per hour
				CrontabID:   1,
				LastUpdated: now,
				Next:        now,
				Previous:    now.Add(-1 * time.Hour).UTC(),
				ScheduleID:  1,
			},
			CrontabLeaseID: 1,
		}
		ok = true
		return
	}

	scheduledJobStarter := func(crontabID, scheduleID, crontabLeaseID int64, newNext time.Time) (jobID int64, err error) {
		return 0, errors.New("this is a failure")
	}

	w := NewScheduleWorker(nodeName, lockExpiryMins, scheduleGetter, scheduledJobStarter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		WorkDone:          true,
		ScheduleRetrieved: true,
		JobStarted:        false,
		NextTime:          time.Time{},
		ErrorOccurred:     true,
	}

	expected.Assert(t, actual)
}

type Values struct {
	WorkDone          bool
	ScheduleRetrieved bool
	JobStarted        bool
	NextTime          time.Time
	ErrorOccurred     bool
}

func (expected Values) Assert(t *testing.T, actual Values) {
	if expected.ErrorOccurred != actual.ErrorOccurred {
		t.Errorf("expected error=%v, but got %v", expected.ErrorOccurred, actual.ErrorOccurred)
	}
	if expected.WorkDone != actual.WorkDone {
		t.Errorf("expected work done=%v, but got %v", expected.WorkDone, actual.WorkDone)
	}
	if expected.ScheduleRetrieved != actual.ScheduleRetrieved {
		t.Errorf("expected schedule retrieved=%v, but got %v", expected.ScheduleRetrieved, actual.ScheduleRetrieved)
	}
	if expected.JobStarted != actual.JobStarted {
		t.Errorf("expected job to have been started=%v, but %v were started", expected.JobStarted, actual.JobStarted)
	}
	if !expected.NextTime.Equal(actual.NextTime) {
		t.Errorf("expected next time of crontab to be %v, but was %v", expected.NextTime, actual.NextTime)
	}
}
