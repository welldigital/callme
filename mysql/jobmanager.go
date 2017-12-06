package mysql

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"       // Requires MySQL
	gomysql "github.com/go-sql-driver/mysql" // Requires MySQL
	"github.com/welldigital/callme/data"
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

	row := db.QueryRow("call jm_startjob(?, ?, ?, ?)", j.ARN, j.Payload, j.ScheduleID, j.When)
	err = row.Scan(&j.JobID)
	return j, err
}

// GetAvailableJobCount returns the number of jobs present in the DB to process, i.e. where they have no job response and
// they're ready to process.
func (m JobManager) GetAvailableJobCount() (int, error) {
	count := 0

	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return count, err
	}
	defer db.Close()

	rows, err := db.Query("call jm_getavailablejobcount()")
	if err != nil {
		return count, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&count)
	}
	return count, err
}

// GetJob retrieves a job that's ready to run from the queue.
func (m JobManager) GetJob(lockedBy string, lockExpiryMinutes int) (j data.Job, ok bool, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("call jm_getjob(?, ?)", lockedBy, lockExpiryMinutes)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&j.JobID, &j.ScheduleID, &j.When, &j.ARN, &j.Payload)
		ok = true
		break
	}
	return
}

// GetJobResponse retrieves a completed job's data.
func (m JobManager) GetJobResponse(jobID int64) (j data.Job, r data.JobResponse, jobOK, responseOK bool, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("call jm_getjobresponse(?)", jobID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		jobOK = true
		var jrID, jrJobID sql.NullInt64
		var jrTime gomysql.NullTime
		var jrResp, jrIsErrorStr, jrError sql.NullString

		err = rows.Scan(&j.JobID, &j.ScheduleID, &j.When, &j.ARN, &j.Payload,
			&jrID, &jrJobID, &jrTime, &jrResp, &jrIsErrorStr, &jrError)

		if jrID.Int64 > 0 {
			r.JobResponseID = jrID.Int64
			r.JobID = jrJobID.Int64
			r.Time = jrTime.Time
			r.Response = jrResp.String
			r.IsError = convertMySQLBoolean(jrIsErrorStr.String)
			r.Error = jrError.String
		}
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
func (m JobManager) CompleteJob(jobID int64, resp string, jobError error) error {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return err
	}
	defer db.Close()
	var isError bool
	if jobError != nil {
		isError = true
	}
	var errorString string
	if jobError != nil {
		errorString = jobError.Error()
	}
	_, err = db.Exec("call jm_completejob(?, ?, ?, ?)",
		jobID, resp, isError, errorString)
	return err
}

// DeleteJob deletes a job.
func (m JobManager) DeleteJob(jobID int64) (ok bool, err error) {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return
	}
	defer db.Close()
	result, err := db.Exec("call jm_deletejob(?)", jobID)
	if err != nil {
		return
	}
	affected, err := result.RowsAffected()
	ok = affected > 0
	return
}
