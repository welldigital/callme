package mysql

import (
	"database/sql"
	"time"

	"github.com/a-h/callme/data"
	_ "github.com/go-sql-driver/mysql" // Requires MySQL

	_ "github.com/mattes/migrate/source/file" // Be able to migrate from files.
)

// LeaseDuration is the amount of time that the lease will be locked to a processor.
const LeaseDuration = time.Hour

// LeaseManager provides features to manage leases using MySQL.
type LeaseManager struct {
	ConnectionString string
}

// NewLeaseManager creates a new LeaseManager.
func NewLeaseManager(connectionString string) LeaseManager {
	return LeaseManager{
		ConnectionString: connectionString,
	}
}

// Acquire gets a lease.
func (m LeaseManager) Acquire(now time.Time, leaseType string, lockedBy string) (leaseID int64, until time.Time, ok bool, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	until = now.Add(LeaseDuration).Truncate(time.Second)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO `lease`(`type`, lockedby, `at`, `until`) " +
		"SELECT ?, ?, ?, ? FROM dual WHERE NOT EXISTS (SELECT * FROM `lease` WHERE `type` = ? HAVING MAX(`until`) > ?);")
	if err != nil {
		return
	}
	res, err := stmt.Exec(leaseType, lockedBy, now, until, leaseType, now)
	if err != nil {
		return
	}
	leaseID, err = res.LastInsertId()
	if err == nil && leaseID > 0 {
		ok = true
	}
	return
}

// Get gets a lease record from a table.
func (m LeaseManager) Get(leaseID int64) (lease data.Lease, ok bool, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT idlease, `type`, lockedby, `at`, `until` FROM `lease` WHERE idlease = ? limit 1;", leaseID)
	if err != nil {
		return
	}

	for rows.Next() {
		err = rows.Scan(&lease.LeaseID, &lease.Type, &lease.LockedBy, &lease.At, &lease.Until)
		if err != nil {
			return
		}
		ok = true
	}

	return
}

// Rescind rescinds the right on a lease.
func (m LeaseManager) Rescind(leaseID int64, now time.Time) (err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return
	}
	defer db.Close()

	// The date precision is quite low, so set it to the past.
	now = now.Add(-1 * time.Minute)

	stmt, err := db.Prepare("UPDATE `lease` SET `until`=? WHERE idlease=?")
	if err != nil {
		return
	}
	_, err = stmt.Exec(now, leaseID)
	return
}
