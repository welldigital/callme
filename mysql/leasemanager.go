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
func (m LeaseManager) Acquire(leaseType string, lockedBy string) (leaseID int64, until time.Time, ok bool, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)

	if err != nil {
		return
	}
	defer db.Close()

	until = time.Now().UTC().Add(time.Hour)

	stmt, err := db.Prepare("INSERT INTO `lease`(`type`, lockedby, `at`, `until`) " +
		"SELECT ?, ?, utc_timestamp(), ? FROM dual WHERE NOT EXISTS (SELECT idlease FROM `lease` WHERE `type` = ? HAVING MAX(`until`) >= utc_timestamp());")
	if err != nil {
		return
	}
	res, err := stmt.Exec(leaseType, lockedBy, until, leaseType)
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
func (m LeaseManager) Rescind(leaseID int64) (err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE `lease` SET `until`=DATE_SUB(utc_timestamp(), INTERVAL 1 HOUR) WHERE idlease=?")
	if err != nil {
		return
	}
	_, err = stmt.Exec(leaseID)
	return
}
