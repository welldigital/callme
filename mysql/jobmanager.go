package mysql

import (
	"database/sql"
	"time"

	"github.com/a-h/callme/data"
	_ "github.com/go-sql-driver/mysql" // Requires MySQL
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

// GetAvailableJobCount returns the number of jobs present in the DB to process, i.e. where they have no job response.
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
func (m JobManager) GetJob(leaseID int64) (*data.Job, error) {
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
		"l.`until` >= utc_timestamp() AND "+
		"l.rescinded = 0", leaseID)

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

// GetJobResponse retrieves a completed job's data.
func (m JobManager) GetJobResponse(jobID int64) (j data.Job, r data.JobResponse, jobOK, responseOK bool, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT "+
		"j.idjob, j.idschedule, j.`when`, j.arn, j.payload, "+
		"jr.idjobresponse, jr.idlease, jr.idjobid, jr.`time`, jr.response, jr.iserror, jr.`error` "+
		"FROM `job` j "+
		"LEFT JOIN jobresponse jr on jr.idjobid = j.idjob "+
		"WHERE "+
		"jr.idjobid = ?", jobID)

	if err != nil {
		return
	}

	for rows.Next() {
		jobOK = true
		var isErrorStr string
		err = rows.Scan(&j.JobID, &j.ScheduleID, &j.When, &j.ARN, &j.Payload,
			&r.JobResponseID, &r.LeaseID, &r.JobID, &r.Time, &r.Response, &isErrorStr, &r.Error)
		r.IsError = convertMySQLBoolean(isErrorStr)
		break
	}
	return j, r, jobOK, r.JobResponseID > 0, err
}

func convertMySQLBoolean(s string) bool {
	if len(s) == 1 && s[0] == 1 {
		return true
	}
	return false
}

// CompleteJob marks a job as complete.
func (m JobManager) CompleteJob(leaseID, jobID int64, resp string, jobError error) error {
	statement := "INSERT INTO `jobresponse` " +
		"(`idlease`, `idjobid`, `time`, `response`, `iserror`, `error`) " +
		"VALUES " +
		"(?, ?, utc_timestamp(), ?, ?, ?);"

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
	_, err = stmt.Exec(leaseID, jobID, resp, isError, jobError.Error())
	return err
}
