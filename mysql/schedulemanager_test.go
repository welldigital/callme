package mysql

import (
	"testing"
	"time"
)

func TestThatLeasesCanBeAcquiredAndRescinded(t *testing.T) {
	if !testing.Short() {
		dsn, dbName, err := createTestDatabase()
		if err != nil {
			t.Errorf("failed to create test database with error: %v", err)
		}
		defer dropTestDatabase(dbName)

		expectedLockedBy := "TestThatLeasesCanBeAcquiredAndRescinded"
		leaseType := "testLeaseType"

		lm := NewLeaseManager(dsn)
		leaseID, until, ok, err := lm.Acquire(time.Now().UTC(), leaseType, expectedLockedBy)
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
		if lease.Until != until {
			t.Errorf("expected Until to be '%v', but got '%v'", until, lease.Until)
		}

		// Attempting to get another lease should fail, because one is already in use.
		newLeaseID, until, ok, err := lm.Acquire(time.Now().UTC(), leaseType, expectedLockedBy)
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
		if !lease.Until.Before(time.Now().UTC()) {
			t.Errorf("expected Until to now be in the past but was '%v'", lease.Until)
		}

		// Now getting a new lease, should be fine.
		newLeaseID, until, ok, err = lm.Acquire(time.Now().UTC(), leaseType, expectedLockedBy)
		if err != nil {
			t.Fatalf("failed to acquire another lease with err: %v", err)
		}
		if newLeaseID == 0 {
			t.Errorf("expected acquiring a lease after rescding our old lease should be fine, but got leaseID of %v", newLeaseID)
		}
	}
}
