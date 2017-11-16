package data

import "time"

// LeaseAcquirer gets a lease to update data held by the lease type.
type LeaseAcquirer func(now time.Time, leaseType string, by string) (leaseID int64, until time.Time, ok bool, err error)

// LeaseRescinder rescinds the right on a lease.
type LeaseRescinder func(leaseID int64) (err error)
