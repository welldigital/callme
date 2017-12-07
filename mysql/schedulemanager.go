package mysql

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql" // Requires MySQL
	"github.com/welldigital/callme/data"
)

// ScheduleManager provides features to manage schedules using MySQL.
type ScheduleManager struct {
	ConnectionString string
}

// NewScheduleManager creates a new ScheduleManager.
func NewScheduleManager(connectionString string) ScheduleManager {
	return ScheduleManager{
		ConnectionString: connectionString,
	}
}

// Create creates a repeating schedule. Doesn't require a lease, any process can do this.
func (m ScheduleManager) Create(from time.Time, arn string, payload string, crontabs []string, externalID string, by string) (id int64, err error) {
	s := data.Schedule{
		ExternalID:      externalID,
		By:              by,
		ARN:             arn,
		Payload:         payload,
		Created:         time.Now().UTC(),
		Active:          true,
		DeactivatedDate: time.Time{},
	}

	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	scheduleInsertSQL := "INSERT INTO `schedule` " +
		"(`externalid`,`by`,`arn`,`payload`,`created`,`active`) " +
		"VALUES (?, ?, ?, ?, ?, ?)"

	crontabInsertSQL := "INSERT INTO `crontab` " +
		"(`idschedule`, `crontab`, `previous`, `next`, `lastupdated`)" +
		"VALUES (?, ?, ?, ?, ?)"

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	scheduleInsert, err := tx.Prepare(scheduleInsertSQL)
	if err != nil {
		return 0, err
	}
	res, err := scheduleInsert.Exec(s.ExternalID, s.By, s.ARN, s.Payload, s.Created, s.Active)
	if err != nil {
		return 0, err
	}
	scheduleID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	crontabInsert, err := tx.Prepare(crontabInsertSQL)
	if err != nil {
		return 0, err
	}
	var emptyTime time.Time
	for _, crontab := range crontabs {
		_, err := crontabInsert.Exec(scheduleID, crontab, emptyTime, from, emptyTime)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Deactivate deactivates a schedule.
func (m ScheduleManager) Deactivate(scheduleID int64) (ok bool, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return false, err
	}
	defer db.Close()

	res, err := db.Exec("UPDATE `schedule` SET active = 0, deactivateddate = utc_timestamp() WHERE `schedule`.`idschedule` = ?",
		scheduleID)
	if err != nil {
		return
	}

	affectedRows, err := res.RowsAffected()
	ok = affectedRows > 0

	return
}

// GetScheduleByID gets a schedule's information by its ID.
func (m ScheduleManager) GetScheduleByID(scheduleID int64) (sc data.ScheduleCrontabs, ok bool, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("call sm_getschedulebyid(?)", scheduleID)
	if err != nil {
		return
	}
	defer rows.Close()

	sc.Crontabs = make([]data.Crontab, 0)
	var isActiveStr string
	var deactivatedDate *time.Time
	for rows.Next() {
		var ct data.Crontab
		err = rows.Scan(&sc.Schedule.ScheduleID,
			&sc.Schedule.ExternalID,
			&sc.Schedule.By,
			&sc.Schedule.ARN,
			&sc.Schedule.Payload,
			&sc.Schedule.Created,
			&isActiveStr,
			&deactivatedDate,
			&ct.CrontabID,
			&ct.ScheduleID,
			&ct.Crontab,
			&ct.Previous,
			&ct.Next,
			&ct.LastUpdated)
		if err != nil {
			return
		}
		sc.Schedule.Active = convertMySQLBoolean(isActiveStr)
		if deactivatedDate != nil {
			sc.Schedule.DeactivatedDate = *deactivatedDate
		}
		sc.Crontabs = append(sc.Crontabs, ct)
		ok = true
	}
	return
}

// GetSchedule is a ScheduleGetter which locks a schedule where Next is in the past, in order to schedule jobs.
func (m ScheduleManager) GetSchedule(lockedBy string, lockExpiryMinutes int) (sc data.ScheduleCrontab, ok bool, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("call sm_getschedule(?, ?)", lockedBy, lockExpiryMinutes)
	if err != nil {
		return
	}
	defer rows.Close()

	var isActiveStr string
	var deactivatedDate *time.Time
	for rows.Next() {
		err = rows.Scan(&sc.CrontabLeaseID,
			&sc.Schedule.ScheduleID,
			&sc.Schedule.ExternalID,
			&sc.Schedule.By,
			&sc.Schedule.ARN,
			&sc.Schedule.Payload,
			&sc.Schedule.Created,
			&isActiveStr,
			&deactivatedDate,
			&sc.Crontab.CrontabID,
			&sc.Crontab.ScheduleID,
			&sc.Crontab.Crontab,
			&sc.Crontab.Previous,
			&sc.Crontab.Next,
			&sc.Crontab.LastUpdated)
		sc.Schedule.Active = convertMySQLBoolean(isActiveStr)
		if deactivatedDate != nil {
			sc.Schedule.DeactivatedDate = *deactivatedDate
		}
		if err != nil {
			return
		}
		ok = true
	}
	return
}

// StartJobAndUpdateCron starts a new job based on the schedule record's arn and payload and updates the existing crontab to the new date.
// It requires a crontabLeaseID so that it can be cancelled, allowing crontab refreshes at a rate faster than the lease timeout.
func (m ScheduleManager) StartJobAndUpdateCron(crontabID, scheduleID, crontabLeaseID int64, newNext time.Time) (jobID int64, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	rows, err := db.Query("call sm_startjobandupdatecron(?, ?, ?, ?)", crontabID, scheduleID, crontabLeaseID, newNext)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&jobID)
	}

	return
}
