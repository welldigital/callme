package mysql

import (
	"database/sql"
	"time"

	"github.com/a-h/callme/data"
	_ "github.com/go-sql-driver/mysql" // Requires MySQL

	_ "github.com/mattes/migrate/source/file" // Be able to migrate from files.
)

// ScheduleLeaseDuration is the amount of time that the lease will be locked down.
const ScheduleLeaseDuration = time.Hour

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

// Create creates a repeating schedule.
func (m ScheduleManager) Create(from time.Time, arn string, payload string, crontabs []string, externalID string, by string) (id int64, err error) {
	s := data.Schedule{
		ExternalID:      externalID,
		By:              by,
		ARN:             arn,
		Payload:         payload,
		Created:         time.Now().UTC(),
		From:            from,
		Active:          true,
		DeactivatedDate: nil,
	}

	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	scheduleInsertSQL := "INSERT INTO `callme`.`schedule` " +
		"(`externalid`,`by`,`arn`,`payload`,`created`,`from`,`active`,`deactivateddate`) " +
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?)"

	crontabInsertSQL := "INSERT INTO `callme`.`crontab` " +
		"(`ScheduleID`, `Crontab`, `Previous`, `Next`, `LastUpdated`)" +
		"VALUES (?, ?, ?, ?, ?)"

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	scheduleInsert, err := db.Prepare(scheduleInsertSQL)
	if err != nil {
		return 0, err
	}
	res, err := scheduleInsert.Exec(s.ExternalID, s.By, s.ARN, s.Payload,
		s.Created, s.From, s.Active, s.DeactivatedDate)
	if err != nil {
		return 0, err
	}

	crontabInsert, err := db.Prepare(crontabInsertSQL)
	var emptyTime time.Time
	for _, crontab := range crontabs {
		_, err := crontabInsert.Exec(res.LastInsertId, crontab, emptyTime, from, emptyTime)
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
func (m ScheduleManager) Deactivate(scheduleID int64) error {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("UPDATE `schedule` SET active = 0 WHERE `schedule`.`idschedule` = ?",
		scheduleID)

	return err
}

// GetSchedules is a ScheduleGetter which gets all schedules where Next is in the past, in order to schedule jobs.
func (m ScheduleManager) GetSchedules(scheduleLeaseID int64, now time.Time) ([]data.ScheduleCrontab, error) {
	sc := make([]data.ScheduleCrontab, 0)

	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return sc, err
	}
	defer db.Close()

	query := "SELECT " +
		"sc.`idschedule`, `externalid`, `by`, `arn`, `payload`, `created`, `from`, `active`, `deactivateddate`, `idcrontab`, ct.`idschedule`, `crontab`, `previous`, `next`, `lastupdated` " +
		"FROM " +
		"`schedule` sc " +
		"INNER JOIN `crontab` ct ON sc.idschedule = ct.idschedule " +
		"INNER JOIN `schedulelease` sl ON sl.idschedulelease = ? " +
		"WHERE ct.next < utc_date() AND sl.until > utc_date()"

	rows, err := db.Query(query, scheduleLeaseID)
	if err != nil {
		return sc, err
	}

	for rows.Next() {
		r := data.ScheduleCrontab{}
		// previous`, `next`, `lastupdated` " +
		err = rows.Scan(&r.Schedule.ScheduleID,
			&r.Schedule.ExternalID,
			&r.Schedule.By,
			&r.Schedule.ARN,
			&r.Schedule.Payload,
			&r.Schedule.Created,
			&r.Schedule.From,
			&r.Schedule.Active,
			&r.Schedule.DeactivatedDate,
			&r.Crontab.CrontabID,
			&r.Crontab.ScheduleID,
			&r.Crontab.Crontab,
			&r.Crontab.Previous,
			&r.Crontab.Next,
			&r.Crontab.LastUpdated)

		if err != nil {
			return sc, err
		}

		sc = append(sc, r)
	}
	return sc, err
}

// UpdateCron is a cron updater which updates a Crontab record so that it's not included in future updates.
func (m ScheduleManager) UpdateCron(crontabID int64, newPrevious, newNext, newLastUpdated time.Time) error {
	statement := "UPDATE crontab SET " +
		"(`previous`, `next`, `lastupdated`) " +
		"VALUES " +
		"(?, ?, ?) " +
		"WHERE idcrontab = ?"

	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(newPrevious, newNext, newLastUpdated, crontabID)
	return err
}

// AcquireScheduleLease gets a lease to process schedules.
func (m ScheduleManager) AcquireScheduleLease(now time.Time, lockedBy string) (scheduleLeaseID int64, until time.Time, ok bool, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	until = now.Add(ScheduleLeaseDuration).Truncate(time.Second)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO `schedulelease`(lockedby, `at`, `until`) " +
		"SELECT ?, ?, ? FROM dual WHERE NOT EXISTS (SELECT * FROM `schedulelease` HAVING MAX(`until`) > ?);")
	if err != nil {
		return
	}
	res, err := stmt.Exec(lockedBy, now, until, now)
	if err != nil {
		return
	}
	scheduleLeaseID, err = res.LastInsertId()
	if err == nil && scheduleLeaseID > 0 {
		ok = true
	}
	return
}

// GetScheduleLease gets a lease record from a table.
func (m ScheduleManager) GetScheduleLease(scheduleLeaseID int64) (lease data.ScheduleLease, ok bool, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT idschedulelease, lockedby, `at`, `until` FROM `schedulelease` WHERE idschedulelease = ? limit 1;", scheduleLeaseID)
	if err != nil {
		return
	}

	for rows.Next() {
		err = rows.Scan(&lease.ScheduleLeaseID, &lease.LockedBy, &lease.At, &lease.Until)
		if err != nil {
			return
		}
		ok = true
	}

	return
}

// RescindScheduleLease rescinds the right on a lease.
func (m ScheduleManager) RescindScheduleLease(scheduleLeaseID int64) (err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE `schedulelease` SET `until`=UTC_DATE() WHERE idschedulelease=?")
	if err != nil {
		return
	}
	_, err = stmt.Exec(scheduleLeaseID)
	return
}
