package mysql

import (
	"errors"
	"testing"
	"time"

	"github.com/a-h/callme/data"
)

func TestThatJobsCanBeStartedWithAndWithoutBeingAssociatedWithASchedule(t *testing.T) {
	if !testing.Short() {
		dsn, dbName, err := createTestDatabase()
		if err != nil {
			t.Errorf("failed to create test database with error: %v", err)
		}
		defer dropTestDatabase(dbName)

		jm := NewJobManager(dsn)
		when := time.Now().UTC().Add(-5 * time.Second).Truncate(time.Second)

		job1 := data.Job{
			JobID:      1,
			ARN:        "testarn",
			When:       when,
			Payload:    "testpayload",
			ScheduleID: nil,
		}

		// Start job without a shecedule.
		actualJob1, err := jm.StartJob(job1.When, job1.ARN, job1.Payload, job1.ScheduleID)
		if err != nil {
			t.Fatalf("without schedule: error starting job: %v", err)
		}
		Assert(t, "without schedule", job1, actualJob1)

		// Start job with valid schedule.
		// Create a schedule.
		sm := NewScheduleManager(dsn)
		scheduleID, err := sm.Create(time.Now().UTC(), "testarn", `{ nonsense: "payload" }`, []string{"* * * *"}, "externalid", "jobmanager_test")
		if err != nil {
			t.Fatalf("failed to create schedule with error: %v", err)
		}

		job2 := data.Job{
			JobID:      2,
			ARN:        "testarn2",
			When:       when,
			Payload:    "testpayload2",
			ScheduleID: &scheduleID,
		}

		actualJob2, err := jm.StartJob(job2.When, job2.ARN, job2.Payload, job2.ScheduleID)
		if err != nil {
			t.Fatalf("with schedule: error starting job: %v", err)
		}
		Assert(t, "with schedule", job2, actualJob2)

		// Attemp to start a job with invalid schedule.
		invalidSchedule := int64(-1)
		_, err = jm.StartJob(when, "testarn", "testpayload", &invalidSchedule)
		if err == nil {
			t.Errorf("invalid schedule: expected error, because it's not possible to start a job associated with an invalid schedule ID")
		}

		// Grab a lease and pull the first job.
		// Acquire a lease.
		lm := NewLeaseManager(dsn)
		leaseID, _, ok, err := lm.Acquire(time.Now().UTC(), "job", "jobmanager_test")
		if err != nil {
			t.Fatalf("could not acquire lease with error: %v", err)
		}
		if !ok {
			t.Fatalf("could not acquire lease, even though there's only one process accessing the DB")
		}

		// Use the lease to pull the job.
		actualPtr, err := jm.GetJob(leaseID, time.Now().UTC())
		if err != nil {
			t.Fatalf("error getting job with valid lease: %v", err)
		}
		if actualPtr == nil {
			t.Fatalf("expected to get a job, but didn't")
		}
		actualJob1 = *actualPtr
		Assert(t, "get job 1", job1, actualJob1)

		// Complete the job.
		err = jm.CompleteJob(leaseID, job1.JobID, time.Now().UTC(), "response", errors.New("just a test"))
		if err != nil {
			t.Errorf("got an error completing the job: %v", err)
		}

		// Pull the second job.
		actualPtr, err = jm.GetJob(leaseID, time.Now().UTC())
		if err != nil {
			t.Fatalf("error getting job (after completion): %v", err)
		}
		if actualPtr == nil {
			t.Errorf("job 2 should be available, but no job was retrieved")
		}
		actualJob2 = *actualPtr
		Assert(t, "get job 2", job1, actualJob1)

		// Check that it's possible to get the job response for ID 1, but not 2.
		j1, r, jOK, rOK, err := jm.GetJobResponse(1)
		if err != nil {
			t.Errorf("failed to get job repsonse 1 with error: %v", err)
		}
		if !jOK {
			t.Errorf("failed to get job 1 from database without throwing an error")
		}
		if !rOK {
			t.Errorf("failed to get response 1 from database without throwing an error")
		}
		// Test that the job is correct.
		Assert(t, "get job response 1", job1, j1)
		if !r.IsError {
			t.Errorf("for job response 1, expected IsError=true, but was false")
		}
		if r.Error != "just a test" {
			t.Errorf("expected error message: 'just a test', but got '%v'", r.Error)
		}
		if r.JobID != 1 {
			t.Errorf("expected JobID=1, but got %v", r.JobID)
		}
		if r.JobResponseID != 1 {
			t.Errorf("expected JobResponseID=1, but got %v", r.JobResponseID)
		}
	}
}

func Assert(t *testing.T, testName string, expected, actual data.Job) {
	if expected.JobID != actual.JobID {
		t.Errorf("%v: expected JobID='%v', but was '%v'", testName, expected.JobID, actual.JobID)
	}
	if expected.ARN != actual.ARN {
		t.Errorf("%v: expected ARN='%v', but was '%v'", testName, expected.ARN, actual.ARN)
	}
	if expected.Payload != actual.Payload {
		t.Errorf("%v: expected Payload='%v', but was '%v'", testName, expected.Payload, actual.Payload)
	}
	if expected.ScheduleID == nil && actual.ScheduleID != nil {
		t.Errorf("%v: expected ScheduleID to be nil, but was '%v'", testName, *actual.ScheduleID)
	}
	if expected.ScheduleID != nil && actual.ScheduleID == nil {
		t.Errorf("%v: expected ScheduleID='%v', but was nil", testName, *expected.ScheduleID)
	}
	if expected.ScheduleID != nil && actual.ScheduleID != nil {
		if *expected.ScheduleID != *actual.ScheduleID {
			t.Errorf("%v: expected ScheduleID='%v', but was '%v'", testName, *expected.ScheduleID, *actual.ScheduleID)
		}
	}
	if expected.When != actual.When {
		t.Errorf("%v: expected When='%v', but was '%v'", testName, expected.When, actual.When)
	}
}
