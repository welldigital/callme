package jobworker

import (
	"errors"
	"testing"
	"time"

	"github.com/welldigital/callme/data"
)

const nodeName = "jobworker_test"
const lockExpiryMins = 5

func TestThatNoWorkIsDoneIfAJobIsNotRetrieved(t *testing.T) {
	actual := Values{}

	jobGetter := func(lockedBy string, lockExpiryMinutes int) (j data.Job, ok bool, err error) {
		actual.JobRetrieved = true
		ok = false
		return
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return `{ "response": "ok" }`, nil
	}

	jobCompleter := func(jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	w := NewJobWorker(nodeName, lockExpiryMins, jobGetter, executor, jobCompleter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred: false,
		JobRetrieved:  true,
		JobExecuted:   false,
		JobCompleted:  false,
		WorkDone:      false,
	}

	expected.Assert(t, actual)
}

func TestThatGettingAJobResultsInWorkBeingExecuted(t *testing.T) {
	actual := Values{}

	jobGetter := func(lockedBy string, lockExpiryMinutes int) (j data.Job, ok bool, err error) {
		actual.JobRetrieved = true
		scheduleID := int64(1)

		j = data.Job{
			JobID:      1,
			ARN:        "arn",
			Payload:    "payload",
			ScheduleID: &scheduleID,
			When:       time.Now().UTC(),
		}
		ok = true
		return
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return `{ "response": "ok" }`, nil
	}

	jobCompleter := func(jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	w := NewJobWorker(nodeName, lockExpiryMins, jobGetter, executor, jobCompleter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred: false,
		JobRetrieved:  true,
		JobExecuted:   true,
		JobCompleted:  true,
		WorkDone:      true,
	}

	expected.Assert(t, actual)
}

func TestThatErrorsGettingAJobResultsInNoWorkBeingExecuted(t *testing.T) {
	actual := Values{}

	jobGetter := func(lockedBy string, lockExpiryMinutes int) (j data.Job, ok bool, err error) {
		actual.JobRetrieved = true
		err = errors.New("failed to get job")
		return
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return `{ "response": "ok" }`, nil
	}

	jobCompleter := func(jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	w := NewJobWorker(nodeName, lockExpiryMins, jobGetter, executor, jobCompleter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred: true,
		JobRetrieved:  true,
		JobExecuted:   false,
		JobCompleted:  false,
		WorkDone:      false,
	}

	expected.Assert(t, actual)
}

func TestThatWhenJobsAndCompletionsFailTheyGetRetried(t *testing.T) {
	actual := Values{}

	jobGetter := func(lockedBy string, lockExpiryMinutes int) (j data.Job, ok bool, err error) {
		actual.JobRetrieved = true
		scheduleID := int64(1)

		j = data.Job{
			JobID:      1,
			ARN:        "arn",
			Payload:    "payload",
			ScheduleID: &scheduleID,
			When:       time.Now().UTC(),
		}
		ok = true
		return
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
	jobCompleter := func(jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		completions++
		if completions == 1 {
			return errors.New("failed for no reason whatsoever")
		}
		return nil
	}

	w := NewJobWorker(nodeName, lockExpiryMins, jobGetter, executor, jobCompleter)

	var err error
	actual.WorkDone, err = w()
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred: false,
		JobRetrieved:  true,
		JobExecuted:   true,
		JobCompleted:  true,
		WorkDone:      true,
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

	jobGetter := func(lockedBy string, lockExpiryMinutes int) (j data.Job, ok bool, err error) {
		actual.JobRetrieved = true
		scheduleID := int64(1)

		j = data.Job{
			JobID:      1,
			ARN:        "arn",
			Payload:    "payload",
			ScheduleID: &scheduleID,
			When:       time.Now().UTC(),
		}
		ok = true
		return
	}

	executions := 0
	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		executions++
		return "", errors.New("failed for no reason whatsoever")
	}

	jobCompleter := func(jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		return nil
	}

	var err error
	timeout := time.Second * 1
	actual.WorkDone, err = findAndExecuteWork(nodeName, lockExpiryMins, jobGetter, executor, jobCompleter, timeout)
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred: true,
		JobRetrieved:  true,
		JobExecuted:   true,  // The SNS was never sent, because of errors, but the function was called.
		JobCompleted:  true,  // It was completed, but marked as errored.
		WorkDone:      false, // The SNS was never sent, because of errors.
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

	jobGetter := func(lockedBy string, lockExpiryMinutes int) (j data.Job, ok bool, err error) {
		actual.JobRetrieved = true
		scheduleID := int64(1)

		j = data.Job{
			JobID:      1,
			ARN:        "arn",
			Payload:    "payload",
			ScheduleID: &scheduleID,
			When:       time.Now().UTC(),
		}
		ok = true
		return
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return "", nil
	}

	completions := 0
	jobCompleter := func(jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		completions++
		return errors.New("failed for no reason whatsoever")
	}

	var err error
	timeout := time.Second * 1
	actual.WorkDone, err = findAndExecuteWork(nodeName, lockExpiryMins, jobGetter, executor, jobCompleter, timeout)
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred: true,
		JobRetrieved:  true,
		JobExecuted:   true, // The SNS executor was called, but it returned an error.
		JobCompleted:  true, // The job completer was called, but it returned an error.
		WorkDone:      true,
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

	jobGetter := func(lockedBy string, lockExpiryMinutes int) (j data.Job, ok bool, err error) {
		actual.JobRetrieved = true
		scheduleID := int64(1)

		j = data.Job{
			JobID:      1,
			ARN:        "arn",
			Payload:    "payload",
			ScheduleID: &scheduleID,
			When:       time.Now().UTC(),
		}
		ok = true
		return
	}

	executor := func(arn string, payload string) (resp string, err error) {
		actual.JobExecuted = true
		return "", errors.New("execution error")
	}

	jobCompleter := func(jobID int64, resp string, err error) error {
		actual.JobCompleted = true
		return errors.New("completion error")
	}

	var err error
	timeout := time.Millisecond * 100
	actual.WorkDone, err = findAndExecuteWork(nodeName, lockExpiryMins, jobGetter, executor, jobCompleter, timeout)
	actual.ErrorOccurred = err != nil

	expected := Values{
		ErrorOccurred: true,
		JobRetrieved:  true,
		JobExecuted:   true,  // The SNS executor was called, but it returned an error.
		JobCompleted:  true,  // The job completer was called, but it returned an error.
		WorkDone:      false, // No work was done.
	}

	expected.Assert(t, actual)
	actualErrMsg := err.Error()
	expectedErrMsg := "jobworker_test: execution: execution error, completion: completion error"
	if actualErrMsg != expectedErrMsg {
		t.Errorf("expected error message: '%v', got: '%v'", expectedErrMsg, actualErrMsg)
	}
}

type Values struct {
	WorkDone      bool
	JobRetrieved  bool
	JobExecuted   bool
	JobCompleted  bool
	ErrorOccurred bool
}

func (expected Values) Assert(t *testing.T, actual Values) {
	if expected.ErrorOccurred != actual.ErrorOccurred {
		t.Errorf("expected error=%v, but got %v", expected.ErrorOccurred, actual.ErrorOccurred)
	}
	if expected.WorkDone != actual.WorkDone {
		t.Errorf("expected work done=%v, but got %v", expected.WorkDone, actual.WorkDone)
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
