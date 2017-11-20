package scheduleworker

import (
	"errors"
	"testing"
	"time"

	"github.com/a-h/callme/data"
)

var nodeName = "scheduleworker_test"

func TestThatNoWorkIsDoneIfALeaseIsNotAcquired(t *testing.T) {
	// Don't acquire a lease.
	actual := Values{}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 0, time.Time{}, false, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	scheduleGetter := func() ([]data.ScheduleCrontab, error) {
		actual.SchedulesRetrieved = true
		return []data.ScheduleCrontab{}, nil
	}

	scheduledJobStarter := func(crontabID int64, scheduleID int64, newNext time.Time) (jobID int64, err error) {
		actual.JobsStarted++
		actual.CrontabsUpdated[crontabID] = newNext
		return 1, nil
	}

	w := NewScheduleWorker(leaseAcquirer, nodeName, leaseRescinder, scheduleGetter, scheduledJobStarter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		WorkDone:           false,
		LeaseRescinded:     false,
		SchedulesRetrieved: false,
		JobsStarted:        0,
		CrontabsUpdated:    map[int64]time.Time{},
		ErrorOccurred:      false,
	}

	expected.Assert(t, actual)
}

func TestThatErrorsAcquiringALeaseAreReturned(t *testing.T) {
	// Don't acquire a lease.
	actual := Values{}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 0, time.Time{}, false, errors.New("failed for unknown reason")
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	scheduleGetter := func() ([]data.ScheduleCrontab, error) {
		actual.SchedulesRetrieved = true
		return []data.ScheduleCrontab{}, nil
	}

	scheduledJobStarter := func(crontabID int64, scheduleID int64, newNext time.Time) (jobID int64, err error) {
		actual.JobsStarted++
		actual.CrontabsUpdated[crontabID] = newNext
		return 1, nil
	}

	w := NewScheduleWorker(leaseAcquirer, nodeName, leaseRescinder, scheduleGetter, scheduledJobStarter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		WorkDone:           false,
		LeaseRescinded:     false,
		SchedulesRetrieved: false,
		JobsStarted:        0,
		CrontabsUpdated:    map[int64]time.Time{},
		ErrorOccurred:      true,
	}

	expected.Assert(t, actual)
}

func TestThatIfALeaseIsAcquiredSchedulesAreQueried(t *testing.T) {
	// Don't acquire a lease.
	actual := Values{}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 0, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	scheduleGetter := func() ([]data.ScheduleCrontab, error) {
		actual.SchedulesRetrieved = true
		return []data.ScheduleCrontab{}, nil
	}

	scheduledJobStarter := func(crontabID int64, scheduleID int64, newNext time.Time) (jobID int64, err error) {
		actual.JobsStarted++
		actual.CrontabsUpdated[crontabID] = newNext
		return 1, nil
	}

	w := NewScheduleWorker(leaseAcquirer, nodeName, leaseRescinder, scheduleGetter, scheduledJobStarter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		WorkDone:           false,
		LeaseRescinded:     true,
		SchedulesRetrieved: true,
		JobsStarted:        0,
		CrontabsUpdated:    map[int64]time.Time{},
		ErrorOccurred:      false,
	}

	expected.Assert(t, actual)
}

func TestThatErrorsRetrievingSchedulesAreReturned(t *testing.T) {
	// Don't acquire a lease.
	actual := Values{}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 0, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	scheduleGetter := func() ([]data.ScheduleCrontab, error) {
		actual.SchedulesRetrieved = true
		return []data.ScheduleCrontab{}, errors.New("error getting schedules")
	}

	scheduledJobStarter := func(crontabID int64, scheduleID int64, newNext time.Time) (jobID int64, err error) {
		actual.JobsStarted++
		actual.CrontabsUpdated[crontabID] = newNext
		return 1, nil
	}

	w := NewScheduleWorker(leaseAcquirer, nodeName, leaseRescinder, scheduleGetter, scheduledJobStarter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		WorkDone:           false,
		LeaseRescinded:     true,
		SchedulesRetrieved: true,
		JobsStarted:        0,
		CrontabsUpdated:    map[int64]time.Time{},
		ErrorOccurred:      true,
	}

	expected.Assert(t, actual)
}

func TestThatExpiredSchedulesStartNewJobs(t *testing.T) {
	// Don't acquire a lease.
	actual := Values{
		CrontabsUpdated: make(map[int64]time.Time),
	}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 0, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	now := time.Now().UTC()

	scheduleGetter := func() ([]data.ScheduleCrontab, error) {
		actual.SchedulesRetrieved = true
		return []data.ScheduleCrontab{
			{
				Schedule: data.Schedule{
					ScheduleID:      1,
					Active:          true,
					ARN:             "testarn",
					By:              "scheduleworker.main_test",
					Created:         time.Now().UTC(),
					DeactivatedDate: time.Time{},
					ExternalID:      "externalid",
					From:            time.Now().Add(time.Hour * -1).UTC(),
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
			},
		}, nil
	}

	scheduledJobStarter := func(crontabID int64, scheduleID int64, newNext time.Time) (jobID int64, err error) {
		actual.JobsStarted++
		actual.CrontabsUpdated[crontabID] = newNext
		return 1, nil
	}

	w := NewScheduleWorker(leaseAcquirer, nodeName, leaseRescinder, scheduleGetter, scheduledJobStarter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		WorkDone:           false,
		LeaseRescinded:     true,
		SchedulesRetrieved: true,
		JobsStarted:        1,
		CrontabsUpdated: map[int64]time.Time{
			// The next execution will be the start of the next hour, because the crontab specifies that.
			1: time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location()),
		},
		ErrorOccurred: false,
	}

	expected.Assert(t, actual)
}

type Values struct {
	WorkDone           bool
	LeaseRescinded     bool
	SchedulesRetrieved bool
	JobsStarted        int
	// CrontabsUpdated is a map of crontabIDs to the new time.
	CrontabsUpdated map[int64]time.Time
	ErrorOccurred   bool
}

func (expected Values) Assert(t *testing.T, actual Values) {
	if expected.ErrorOccurred != actual.ErrorOccurred {
		t.Errorf("expected error=%v, but got %v", expected.ErrorOccurred, actual.ErrorOccurred)
	}
	if expected.WorkDone != actual.WorkDone {
		t.Errorf("expected work done=%v, but got %v", expected.WorkDone, actual.WorkDone)
	}
	if expected.LeaseRescinded != actual.LeaseRescinded {
		t.Errorf("expected lease rescinded=%v, but got %v", expected.LeaseRescinded, actual.LeaseRescinded)
	}
	if expected.SchedulesRetrieved != actual.SchedulesRetrieved {
		t.Errorf("expected schedules retrieved=%v, but got %v", expected.SchedulesRetrieved, actual.SchedulesRetrieved)
	}
	if expected.JobsStarted != actual.JobsStarted {
		t.Errorf("expected %v jobs to have been started, but %v were started", expected.JobsStarted, actual.JobsStarted)
	}
	if !crontabsAreEqual(expected.CrontabsUpdated, actual.CrontabsUpdated) {
		t.Errorf("expected crontabs to be %v, but was %v", expected.CrontabsUpdated, actual.CrontabsUpdated)
	}
}

func crontabsAreEqual(a, b map[int64]time.Time) bool {
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		var bv time.Time
		var ok bool
		if bv, ok = b[k]; !ok {
			return false
		}
		if !av.Equal(bv) {
			return false
		}
	}
	return true
}
