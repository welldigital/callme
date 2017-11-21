package data

import "time"

// A Lease is a record of a worker which has claimed the right to update some data until the lease expires.
// If a lease is in play, then no schedule processing of jobs by another agent can be done.
type Lease struct {
	LeaseID   int64
	Type      string
	LockedBy  string
	At        time.Time
	Until     time.Time
	Rescinded bool
}
