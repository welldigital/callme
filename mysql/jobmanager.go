package mysql

import (
	"database/sql"
	"time"

	"github.com/a-h/callme/data"
	_ "github.com/go-sql-driver/mysql" // Requires MySQL

	_ "github.com/mattes/migrate/source/file" // Be able to migrate from files.
)

// JobManager provides features to manage jobs using MySQL.
type JobManager struct {
	ConnectionString string
}

// NewJobManager creates a new JobManager.
func NewJobManager(connectionString string) JobManager {
	return JobManager{
		ConnectionString: connectionString,
	}
}

// StartJob schedules a job to start in the future.
func (m JobManager) StartJob(when time.Time, arn string, payload string, scheduleID *int64) (data.Job, error) {
	j := data.Job{
		ARN:        arn,
		Payload:    payload,
		ScheduleID: scheduleID,
		When:       when,
	}

	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return j, err
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT `job` SET arn=?,payload=?,idschedule=?,`when`=?")
	if err != nil {
		return j, err
	}
	res, err := stmt.Exec(j.ARN, j.Payload, j.ScheduleID, j.When)
	if err != nil {
		return j, err
	}
	id, err := res.LastInsertId()
	j.JobID = id
	return j, err
}

// GetAvailableJobCount returns the number of jobs present in the DB to process.
func (m JobManager) GetAvailableJobCount() (int, error) {
	count := 0

	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return count, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT COUNT(*) FROM `job` j " +
		"LEFT JOIN `jobresponse` jr on jr.idjobid = j.idjob " +
		"WHERE " +
		"jr.idjobid IS NULL")

	if err != nil {
		return count, err
	}

	for rows.Next() {
		err = rows.Scan(&count)
	}

	return count, err
}

// GetJob retrieves a job that's ready to run from the queue.
func (m JobManager) GetJob(leaseID int64, now time.Time) (*data.Job, error) {
	var j *data.Job

	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return j, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT "+
		"j.idjob, "+
		"j.idschedule, "+
		"j.`when`, "+
		"j.arn, "+
		"j.payload "+
		"FROM `job` j "+
		"LEFT JOIN jobresponse jr on jr.idjobid = j.idjob "+
		"INNER JOIN lease l ON l.idlease = ? AND l.`type`='job' "+
		"WHERE "+
		"jr.idjobid IS NULL AND "+
		"l.`until` > ?", leaseID, now)

	if err != nil {
		return j, err
	}

	for rows.Next() {
		j = &data.Job{}
		err = rows.Scan(&j.JobID, &j.ScheduleID, &j.When, &j.ARN, &j.Payload)
		break
	}
	return j, err
}

// CompleteJob marks a job as complete.
func (m JobManager) CompleteJob(leaseID, jobID int64, now time.Time, resp string, jobError error) error {
	statement := "INSERT INTO `jobresponse` " +
		"(`idlease`, `idjobid`, `time`, `response`, `iserror`, `error`) " +
		"VALUES " +
		"(?, ?, ?, ?, ?, ?);"

	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}

	var isError bool
	if jobError != nil {
		isError = true
	}
	_, err = stmt.Exec(leaseID, jobID, now, resp, isError, jobError.Error())
	return err
}
