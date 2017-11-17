package mysql

import (
	"testing"
	"time"
)

func TestThatJobsCanBeStartedWithoutAScheduleID(t *testing.T) {
	if !testing.Short() {
		dsn, dbName, err := createTestDatabase()
		if err != nil {
			t.Errorf("failed to create test database with error: %v", err)
		}
		defer dropTestDatabase(dbName)

		jm := NewJobManager(dsn)
		when := time.Now().UTC().Add(-5 * time.Second).Truncate(time.Second)
		job, err := jm.StartJob(when, "testarn", "testpayload", nil)
		if err != nil {
			t.Fatalf("error starting job: %v", err)
		}
		if job.ARN != "testarn" {
			t.Errorf("expected ARN of 'testarn', but got '%v'", job.ARN)
		}
		if job.JobID == 0 {
			t.Error("expected JobID > 0, but was zero")
		}
		if job.Payload != "testpayload" {
			t.Errorf("expected payload of 'testpayload', but got '%v'", job.Payload)
		}
		if job.ScheduleID != nil {
			t.Errorf("expected schedule ID of 'nil', but was %v", job.ScheduleID)
		}
		if job.When != when {
			t.Errorf("expected job to be scheduled for %v, but got %v", when, job.When)
		}
	}
}
func TestThatJobsCanBeStartedAndRelatedToASchedule(t *testing.T) {
	if !testing.Short() {
		dsn, dbName, err := createTestDatabase()
		if err != nil {
			t.Errorf("failed to create test database with error: %v", err)
		}
		defer dropTestDatabase(dbName)

		// Create a schedule.
		sm := NewScheduleManager(dsn)
		scheduleID, err := sm.Create(time.Now().UTC(), "testarn", `{ nonsense: "payload" }`, []string{"* * * *"}, "externalid", "jobmanager_test")
		if err != nil {
			t.Fatalf("failed to create schedule with error: %v", err)
		}

		jm := NewJobManager(dsn)
		when := time.Now().UTC().Add(-5 * time.Second).Truncate(time.Second)

		// Start job with valid schedule.
		job, err := jm.StartJob(when, "testarn", "testpayload", &scheduleID)
		if err != nil {
			t.Fatalf("error starting job with valid schedule: %v", err)
		}
		if job.ARN != "testarn" {
			t.Errorf("expected ARN of 'testarn', but got '%v'", job.ARN)
		}
		if job.JobID == 0 {
			t.Error("expected JobID > 0, but was zero")
		}
		if job.Payload != "testpayload" {
			t.Errorf("expected payload of 'testpayload', but got '%v'", job.Payload)
		}
		if *job.ScheduleID != scheduleID {
			t.Errorf("expected schedule ID of %v, but was %v", scheduleID, *job.ScheduleID)
		}
		if job.When != when {
			t.Errorf("expected job to be scheduled for %v, but got %v", when, job.When)
		}

		// Start job with invalid schedule.
		invalidSchedule := int64(-1)
		_, err = jm.StartJob(when, "testarn", "testpayload", &invalidSchedule)
		if err == nil {
			t.Errorf("expected error, because it's not possible to start a job associated with an invalid schedule ID")
		}
	}
}

func TestThatAJobCanBeRetrievedAfterItIsScheduledWithAValidLease(t *testing.T) {
	if !testing.Short() {
		dsn, dbName, err := createTestDatabase()
		if err != nil {
			t.Errorf("failed to create test database with error: %v", err)
		}
		defer dropTestDatabase(dbName)

		jm := NewJobManager(dsn)
		// Schedule in the past.
		when := time.Now().UTC().Add(-1 * time.Minute).Truncate(time.Second)

		// Start job not associated with a schedule.
		_, err = jm.StartJob(when, "testarn", "testpayload", nil)
		if err != nil {
			t.Fatalf("error starting job with valid schedule: %v", err)
		}

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
		job := *actualPtr

		if job.ARN != "testarn" {
			t.Errorf("expected ARN of 'testarn', but got '%v'", job.ARN)
		}
		if job.JobID == 0 {
			t.Error("expected JobID > 0, but was zero")
		}
		if job.Payload != "testpayload" {
			t.Errorf("expected payload of 'testpayload', but got '%v'", job.Payload)
		}
		if job.ScheduleID != nil {
			t.Errorf("expected nil scheduleID, but got %v", *job.ScheduleID)
		}
		if job.When != when {
			t.Errorf("expected job to be scheduled for %v, but got %v", when, job.When)
		}
	}
}
