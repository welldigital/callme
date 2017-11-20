package mysql

import (
	"testing"
	"time"
)

func TestLeaseManager(t *testing.T) {
	if !testing.Short() {
		dsn, dbName, err := createTestDatabase()
		if err != nil {
			t.Errorf("failed to create test database with error: %v", err)
		}
		defer dropTestDatabase(dbName)

		expectedLockedBy := "TestThatLeasesCanBeAcquiredAndRescinded"
		leaseType := "testLeaseType"

		lm := NewLeaseManager(dsn)
		leaseID, until, ok, err := lm.Acquire(leaseType, expectedLockedBy)
		if err != nil || !ok {
			t.Fatalf("failed to acquire the lease with err: %v", err)
		}

		// Get the lease and check that it's valid until the 'until'.
		lease, found, err := lm.Get(leaseID)
		if err != nil {
			t.Fatalf("failed to get the lease with err: %v", err)
		}
		if !found {
			t.Fatal("was unable to find the newly created lease")
		}
		if lease.LockedBy != expectedLockedBy {
			t.Errorf("expected LockedBy to be '%v', but got '%v'", expectedLockedBy, lease.LockedBy)
		}
		if !dateIsWithinRange(lease.Until, until, time.Minute) {
			t.Errorf("expected Until to be '%v', but got '%v'", until, lease.Until)
		}

		// Attempting to get another lease should fail, because one is already in use.
		newLeaseID, until, ok, err := lm.Acquire(leaseType, expectedLockedBy)
		if err != nil {
			t.Fatalf("failed to acquire another lease with err: %v", err)
		}
		if newLeaseID != 0 {
			t.Errorf("expected acquiring another lease to fail by returning leaseID = 0, but got %v", newLeaseID)
		}

		// Rescind the lease.
		err = lm.Rescind(leaseID)
		if err != nil {
			t.Errorf("unexpected error while rescinding a lease: %v", err)
		}

		// Get the lease and check that it's now in the past.
		lease, found, err = lm.Get(leaseID)
		if err != nil {
			t.Fatalf("failed to get the lease with err: %v", err)
		}
		if !found {
			t.Fatal("was unable to find the newly created lease after it was rescinded")
		}
		if lease.LockedBy != expectedLockedBy {
			t.Errorf("expected LockedBy to be '%v', but got '%v'", expectedLockedBy, lease.LockedBy)
		}
		rightNow := time.Now().UTC()
		if !lease.Until.Before(rightNow) {
			t.Errorf("expected Until to be before '%v' but was '%v'", rightNow, lease.Until)
		}

		// Now getting a new lease, should be fine.
		newLeaseID, until, ok, err = lm.Acquire(leaseType, expectedLockedBy)
		if err != nil {
			t.Fatalf("failed to acquire another lease with err: %v", err)
		}
		if newLeaseID == 0 {
			t.Errorf("expected acquiring a new lease after rescinding our old lease should be fine, but got leaseID of %v", newLeaseID)
		}
	}
}
