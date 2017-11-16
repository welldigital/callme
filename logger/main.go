package logger

import "github.com/Sirupsen/logrus"
import "github.com/a-h/callme/data"

var l = logrus.New()

func Infof(format string, args ...interface{}) {
	l.Infof(format, args...)
}

func Debugf(format string, args ...interface{}) {
	l.Debugf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	l.Errorf(format, args...)
}

func WithCrontab(ct data.Crontab) *logrus.Entry {
	return l.
		WithField("CrontabID", ct.CrontabID).
		WithField("Crontab", ct.Crontab).
		WithField("LastUpdated", ct.LastUpdated).
		WithField("Next", ct.Next).
		WithField("ScheduleID", ct.ScheduleID)
}

func WithJob(job data.Job) *logrus.Entry {
	return l.
		WithField("JobID", job.JobID).
		WithField("ARN", job.ARN).
		WithField("ScheduleID", job.ScheduleID).
		WithField("When", job.When)
}
