package mysql

import (
	"database/sql"
	"time"

	"github.com/a-h/callme/data"
	_ "github.com/go-sql-driver/mysql" // Requires MySQL

	_ "github.com/mattes/migrate/source/file" // Be able to migrate from files.
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
		DeactivatedDate: time.Time{},
	}

	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	scheduleInsertSQL := "INSERT INTO `schedule` " +
		"(`externalid`,`by`,`arn`,`payload`,`created`,`from`,`active`) " +
		"VALUES (?, ?, ?, ?, ?, ?, ?)"

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
	res, err := scheduleInsert.Exec(s.ExternalID, s.By, s.ARN, s.Payload,
		s.Created, s.From, s.Active)
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
func (m ScheduleManager) GetSchedules() ([]data.ScheduleCrontab, error) {
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
		"WHERE ct.next < utc_timestamp()"

	rows, err := db.Query(query)
	if err != nil {
		return sc, err
	}

	var isActiveStr string
	var deactivatedDate *time.Time
	for rows.Next() {
		r := data.ScheduleCrontab{}
		err = rows.Scan(&r.Schedule.ScheduleID,
			&r.Schedule.ExternalID,
			&r.Schedule.By,
			&r.Schedule.ARN,
			&r.Schedule.Payload,
			&r.Schedule.Created,
			&r.Schedule.From,
			&isActiveStr,
			&deactivatedDate,
			&r.Crontab.CrontabID,
			&r.Crontab.ScheduleID,
			&r.Crontab.Crontab,
			&r.Crontab.Previous,
			&r.Crontab.Next,
			&r.Crontab.LastUpdated)

		r.Schedule.Active = convertMySQLBoolean(isActiveStr)
		if deactivatedDate != nil {
			r.Schedule.DeactivatedDate = *deactivatedDate
		}

		if err != nil {
			return sc, err
		}

		sc = append(sc, r)
	}
	return sc, err
}

// StartJobAndUpdateCron starts a new job based on the schedule record's arn and payload and updates the existing crontab to the new date.
// All dates should be UTC.
func (m ScheduleManager) StartJobAndUpdateCron(crontabID int64, scheduleID int64, newNext time.Time) (jobID int64, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	scheduleJobStmt, err := tx.Prepare("INSERT INTO `job`(arn, payload, idschedule, `when`) " +
		"SELECT arn, payload, ?, utc_timestamp() FROM schedule;")
	if err != nil {
		return 0, nil
	}
	r, err := scheduleJobStmt.Exec(scheduleID)
	if err != nil {
		return 0, err
	}

	updateCrontabStmt, err := tx.Prepare("UPDATE crontab SET " +
		"`previous`=`next`, " +
		"`next`=?, " +
		"`lastupdated`=utc_timestamp() " +
		"WHERE idcrontab = ?")
	if err != nil {
		return 0, err
	}

	_, err = updateCrontabStmt.Exec(newNext, crontabID)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return r.LastInsertId()
}
