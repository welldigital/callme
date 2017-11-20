package jobworker

import (
	"errors"
	"testing"
	"time"

	"github.com/a-h/callme/data"
)

const nodeName = "testnode"

func TestThatNoWorkIsDoneIfALeaseIsNotAcquired(t *testing.T) {
	// Don't acquire a lease, because one is in use.
	actual := Values{}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 0, time.Time{}, false, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64) (*data.Job, error) {
		actual.JobRetrieved = true
		return nil, nil
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return `{ "response": "ok" }`, nil
	}

	jobCompleter := func(leaseID, jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	w := NewJobWorker(leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter)

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
	actual := Values{}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 0, time.Time{}, false, errors.New("something bad happened")
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64) (*data.Job, error) {
		actual.JobRetrieved = true
		return nil, nil
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return `{ "response": "ok" }`, nil
	}

	jobCompleter := func(leaseID, jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	w := NewJobWorker(leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter)

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
	actual := Values{}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 1, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64) (*data.Job, error) {
		actual.JobRetrieved = true
		return nil, nil
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return `{ "response": "ok" }`, nil
	}

	jobCompleter := func(leaseID, jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	w := NewJobWorker(leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter)

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
	actual := Values{}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 1, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64) (*data.Job, error) {
		actual.JobRetrieved = true
		scheduleID := int64(1)

		return &data.Job{
			JobID:      1,
			ARN:        "arn",
			Payload:    "payload",
			ScheduleID: &scheduleID,
			When:       time.Now().UTC(),
		}, nil
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return `{ "response": "ok" }`, nil
	}

	jobCompleter := func(leaseID, jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	w := NewJobWorker(leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter)

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
	actual := Values{}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 1, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64) (*data.Job, error) {
		actual.JobRetrieved = true
		return nil, errors.New("failed to get job")
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return `{ "response": "ok" }`, nil
	}

	jobCompleter := func(leaseID, jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	w := NewJobWorker(leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter)

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
	actual := Values{}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 1, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64) (*data.Job, error) {
		actual.JobRetrieved = true
		scheduleID := int64(1)

		return &data.Job{
			JobID:      1,
			ARN:        "arn",
			Payload:    "payload",
			ScheduleID: &scheduleID,
			When:       time.Now().UTC(),
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
	jobCompleter := func(leaseID, jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		completions++
		if completions == 1 {
			return errors.New("failed for no reason whatsoever")
		}
		return nil
	}

	w := NewJobWorker(leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter)

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

func TestThatJobExecutionRetriesAreTimeLimited(t *testing.T) {
	actual := Values{}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 1, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64) (*data.Job, error) {
		actual.JobRetrieved = true
		scheduleID := int64(1)

		return &data.Job{
			JobID:      1,
			ARN:        "arn",
			Payload:    "payload",
			ScheduleID: &scheduleID,
			When:       time.Now().UTC(),
		}, nil
	}

	executions := 0
	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		executions++
		return "", errors.New("failed for no reason whatsoever")
	}

	jobCompleter := func(leaseID, jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	var err error
	timeout := time.Second * 1
	actual.WorkDone, err = findAndExecuteWork(leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter, timeout)
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred:  true,
		LeaseRescinded: true,
		JobRetrieved:   true,
		JobExecuted:    true,  // The SNS was never sent, because of errors, but the function was called.
		JobCompleted:   true,  // It was completed, but marked as errored.
		WorkDone:       false, // The SNS was never sent, because of errors.
	}

	expected.Assert(t, actual)
	if executions < 1 {
		t.Errorf("expected the work to be retried multiple times (up to 5 seconds), but it wasn't")
	}
	actualErrMsg := err.Error()
	expectedErrMsg := "execution: failed for no reason whatsoever"
	if actualErrMsg != expectedErrMsg {
		t.Errorf("expected error message: '%v', got: '%v'", expectedErrMsg, actualErrMsg)
	}
}

func TestThatMarkCompleteRetriesAreTimeLimited(t *testing.T) {
	actual := Values{}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 1, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64) (*data.Job, error) {
		actual.JobRetrieved = true
		scheduleID := int64(1)

		return &data.Job{
			JobID:      1,
			ARN:        "arn",
			Payload:    "payload",
			ScheduleID: &scheduleID,
			When:       time.Now().UTC(),
		}, nil
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return "", nil
	}

	completions := 0
	jobCompleter := func(leaseID, jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		completions++
		return errors.New("failed for no reason whatsoever")
	}

	var err error
	timeout := time.Second * 1
	actual.WorkDone, err = findAndExecuteWork(leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter, timeout)
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred:  true,
		LeaseRescinded: true,
		JobRetrieved:   true,
		JobExecuted:    true, // The SNS executor was called, but it returned an error.
		JobCompleted:   true, // The job completer was called, but it returned an error.
		WorkDone:       true,
	}

	expected.Assert(t, actual)
	if completions < 1 {
		t.Errorf("expected the completion to be retried multiple times (up to 5 seconds), but it wasn't")
	}
	actualErrMsg := err.Error()
	expectedErrMsg := "completion: failed for no reason whatsoever"
	if actualErrMsg != expectedErrMsg {
		t.Errorf("expected error message: '%v', got: '%v'", expectedErrMsg, actualErrMsg)
	}
}

func TestThatBothExecutionAndCompletionErrorsAreTracked(t *testing.T) {
	actual := Values{}
	leaseAcquirer := func(leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error) {
		return 1, time.Time{}, true, nil
	}

	leaseRescinder := func(leaseID int64) (err error) {
		actual.LeaseRescinded = true
		return nil
	}

	jobGetter := func(leaseID int64) (*data.Job, error) {
		actual.JobRetrieved = true
		scheduleID := int64(1)

		return &data.Job{
			JobID:      1,
			ARN:        "arn",
			Payload:    "payload",
			ScheduleID: &scheduleID,
			When:       time.Now().UTC(),
		}, nil
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return "", errors.New("execution error")
	}

	jobCompleter := func(leaseID, jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		return errors.New("completion error")
	}

	var err error
	timeout := time.Millisecond * 100
	actual.WorkDone, err = findAndExecuteWork(leaseAcquirer, nodeName, leaseRescinder, jobGetter, executor, jobCompleter, timeout)
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred:  true,
		LeaseRescinded: true,
		JobRetrieved:   true,
		JobExecuted:    true,  // The SNS executor was called, but it returned an error.
		JobCompleted:   true,  // The job completer was called, but it returned an error.
		WorkDone:       false, // No work was done.
	}

	expected.Assert(t, actual)
	actualErrMsg := err.Error()
	expectedErrMsg := "execution: execution error, completion: completion error"
	if actualErrMsg != expectedErrMsg {
		t.Errorf("expected error message: '%v', got: '%v'", expectedErrMsg, actualErrMsg)
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
