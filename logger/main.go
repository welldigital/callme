package logger

import (
	"github.com/a-h/callme/data"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

func For(pkg string, fn string) *log.Entry {
	return log.
		WithField("pkg", pkg).
		WithField("fn", fn)
}

func WithCrontab(pkg string, fn string, ct data.Crontab) *log.Entry {
	return For(pkg, fn).
		WithField("CrontabID", ct.CrontabID).
		WithField("Crontab", ct.Crontab).
		WithField("LastUpdated", ct.LastUpdated).
		WithField("Next", ct.Next).
		WithField("ScheduleID", ct.ScheduleID)
}

func WithJob(pkg string, fn string, job data.Job) *log.Entry {
	return For(pkg, fn).
		WithField("JobID", job.JobID).
		WithField("ARN", limit(job.ARN, 100, 10)).
		WithField("ScheduleID", job.ScheduleID).
		WithField("When", job.When)
}

func limit(s string, ifBiggerThan, max int) string {
	if max > ifBiggerThan {
		max = ifBiggerThan
	}
	if len(s) > ifBiggerThan {
		return s[:max] + "..."
	}
	return s
}
