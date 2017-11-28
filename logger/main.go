package logger

import (
	"github.com/a-h/callme/data"
	"github.com/a-h/callme/metrics"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	metrics.ErrorCounts.Inc()
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func WithCrontab(ct data.Crontab) *log.Entry {
	return log.
		WithField("CrontabID", ct.CrontabID).
		WithField("Crontab", ct.Crontab).
		WithField("LastUpdated", ct.LastUpdated).
		WithField("Next", ct.Next).
		WithField("ScheduleID", ct.ScheduleID)
}

func WithJob(job data.Job) *log.Entry {
	return log.
		WithField("JobID", job.JobID).
		WithField("ARN", job.ARN).
		WithField("ScheduleID", job.ScheduleID).
		WithField("When", job.When)
}
