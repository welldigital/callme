package jobworker

import (
	"errors"
	"testing"
	"time"

	"github.com/a-h/callme/data"
)

const nodeName = "testnode"

var n = time.Now().UTC()
var now = func() time.Time { return n }

func TestThatNoWorkIsDoneIfALeaseIsNotAcquired(t *testing.T) {
	// Don't acquire a lease, because one is in use.
	actual := Values{}
	leaseAcquirer := func(now time.Time, leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 0, time.Time{}, false, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64, now time.Time) (*data.Job, error) {
		actual.JobRetrieved = true
		return nil, nil
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return `{ "response": "ok" }`, nil
	}

	jobCompleter := func(leaseID, jobID int64, now time.Time, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	w := NewJobWorker(now, leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred:  false,
		LeaseRescinded: false,
		JobRetrieved:   false,
		JobExecuted:    false,
		JobCompleted:   false,
		WorkDone:       false,
	}

	expected.Assert(t, actual)
}

func TestThatErrorsAcquiringTheLeaseQuitWork(t *testing.T) {
	// Don't acquire a lease, because one is in use.
	actual := Values{}
	leaseAcquirer := func(now time.Time, leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 0, time.Time{}, false, errors.New("something bad happened")
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64, now time.Time) (*data.Job, error) {
		actual.JobRetrieved = true
		return nil, nil
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return `{ "response": "ok" }`, nil
	}

	jobCompleter := func(leaseID, jobID int64, now time.Time, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	w := NewJobWorker(now, leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred:  true,
		LeaseRescinded: false,
		JobRetrieved:   false,
		JobExecuted:    false,
		JobCompleted:   false,
		WorkDone:       false,
	}

	expected.Assert(t, actual)
}

func TestThatAcquiringALeaseResultsInTryingToGetAJob(t *testing.T) {
	// Don't acquire a lease, because one is in use.
	actual := Values{}
	leaseAcquirer := func(now time.Time, leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 1, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64, now time.Time) (*data.Job, error) {
		actual.JobRetrieved = true
		return nil, nil
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return `{ "response": "ok" }`, nil
	}

	jobCompleter := func(leaseID, jobID int64, now time.Time, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	w := NewJobWorker(now, leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred:  false,
		LeaseRescinded: true,
		JobRetrieved:   true,
		JobExecuted:    false,
		JobCompleted:   false,
		WorkDone:       false,
	}

	expected.Assert(t, actual)
}

func TestThatGettingAJobResultsInWorkBeingExecuted(t *testing.T) {
	// Don't acquire a lease, because one is in use.
	actual := Values{}
	leaseAcquirer := func(now time.Time, leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 1, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64, now time.Time) (*data.Job, error) {
		actual.JobRetrieved = true
		scheduleID := int64(1)

		return &data.Job{
			JobID:      1,
			ARN:        "arn",
			Payload:    "payload",
			ScheduleID: &scheduleID,
			When:       now,
		}, nil
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return `{ "response": "ok" }`, nil
	}

	jobCompleter := func(leaseID, jobID int64, now time.Time, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	w := NewJobWorker(now, leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred:  false,
		LeaseRescinded: true,
		JobRetrieved:   true,
		JobExecuted:    true,
		JobCompleted:   true,
		WorkDone:       true,
	}

	expected.Assert(t, actual)
}

func TestThatErrorsGettingAJobResultsInNoWorkBeingExecuted(t *testing.T) {
	// Don't acquire a lease, because one is in use.
	actual := Values{}
	leaseAcquirer := func(now time.Time, leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 1, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64, now time.Time) (*data.Job, error) {
		actual.JobRetrieved = true
		return nil, errors.New("failed to get job")
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return `{ "response": "ok" }`, nil
	}

	jobCompleter := func(leaseID, jobID int64, now time.Time, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	w := NewJobWorker(now, leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred:  true,
		LeaseRescinded: true,
		JobRetrieved:   true,
		JobExecuted:    false,
		JobCompleted:   false,
		WorkDone:       false,
	}

	expected.Assert(t, actual)
}

func TestThatWhenJobsAndCompletionsFailTheyGetRetried(t *testing.T) {
	// Don't acquire a lease, because one is in use.
	actual := Values{}
	leaseAcquirer := func(now time.Time, leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 1, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64, now time.Time) (*data.Job, error) {
		actual.JobRetrieved = true
		scheduleID := int64(1)

		return &data.Job{
			JobID:      1,
			ARN:        "arn",
			Payload:    "payload",
			ScheduleID: &scheduleID,
			When:       now,
		}, nil
	}

	executions := 0
	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		executions++
		if executions == 1 {
			return "", errors.New("failed for no reason whatsoever")
		}
		return `{ "response": "ok" }`, nil
	}

	completions := 0
	jobCompleter := func(leaseID, jobID int64, now time.Time, resp string, err error) error {
		actual.JobCompleted = true
		completions++
		if completions == 1 {
			return errors.New("failed for no reason whatsoever")
		}
		return nil
	}

	w := NewJobWorker(now, leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred:  false,
		LeaseRescinded: true,
		JobRetrieved:   true,
		JobExecuted:    true,
		JobCompleted:   true,
		WorkDone:       true,
	}

	expected.Assert(t, actual)
	if executions != 2 {
		t.Errorf("expected the work to be retried, but it wasn't")
	}
	if completions != 2 {
		t.Errorf("expected job completion to be retried, but it wasn't")
	}
}

type Values struct {
	WorkDone       bool
	LeaseRescinded bool
	JobRetrieved   bool
	JobExecuted    bool
	JobCompleted   bool
	ErrorOccurred  bool
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
	if expected.JobRetrieved != actual.JobRetrieved {
		t.Errorf("expected job retrieved=%v, but got %v", expected.JobRetrieved, actual.JobRetrieved)
	}
	if expected.JobExecuted != actual.JobExecuted {
		t.Errorf("expected job executed=%v, but got %v", expected.JobExecuted, actual.JobExecuted)
	}
	if expected.JobCompleted != actual.JobCompleted {
		t.Errorf("expected job completed=%v, but got %v", expected.JobCompleted, actual.JobCompleted)
	}
}
