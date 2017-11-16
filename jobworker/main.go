package jobworker

import (
	"time"

	"github.com/a-h/callme/logger"
	"github.com/a-h/callme/repetitive"

	"github.com/a-h/callme/data"

	"github.com/cenkalti/backoff"
)

// An Executor executes work.
type Executor func(arn string, payload string) (resp string, err error)

// NewJobWorker creates a worker for the repetitive.Work function which processes pending jobs.
func NewJobWorker(now func() time.Time,
	leaseAcquirer data.JobLeaseAcquirer,
	nodeName string,
	leaseRescinder data.JobLeaseRescinder,
	jobGetter data.JobGetter,
	e Executor,
	jobCompleter data.JobCompleter) repetitive.Worker {
	return func() (workDone bool, err error) {
		return findAndExecuteWork(now, leaseAcquirer, nodeName, leaseRescinder, jobGetter, e, jobCompleter)
	}
}

func findAndExecuteWork(now func() time.Time,
	leaseAcquirer data.JobLeaseAcquirer,
	nodeName string,
	leaseRescinder data.JobLeaseRescinder,
	jobGetter data.JobGetter,
	e Executor,
	jobCompleter data.JobCompleter) (workDone bool, err error) {
	jobLeaseID, until, err := leaseAcquirer(now(), nodeName)
	if err != nil {
		logger.Errorf("jobworker: failed to acquire lease with error: %v", err)
		return
	}
	logger.Infof("jobworker: got lease %v until %v", jobLeaseID, until)
	defer leaseRescinder(jobLeaseID)

	// See if there's some work to do.
	j, err := jobGetter(jobLeaseID, now())
	if err != nil {
		logger.Errorf("jobworker: error getting jobs: %v", err)
	}
	if j == nil {
		logger.Infof("jobworker: no jobs available")
		return
	}
	job := *j

	logger.WithJob(job).Infof("jobworker: executing")

	// Attempt to execute the work.
	var resp string
	var ee error
	execute := func() error {
		resp, ee = e(j.ARN, j.Payload)
		if ee == nil {
			logger.WithJob(job).Infof("jobworker: success")
		} else {
			logger.WithJob(job).Warnf("jobworker: failure, but may retry")
		}
		return ee
	}

	bo := backoff.NewExponentialBackOff()
	err = backoff.Retry(execute, bo)
	if err != nil {
		logger.WithJob(job).Errorf("jobworker: retries exceeded, logging error: %v", err)
	}

	workDone = true

	// Attempt to complete the work.
	complete := func() error {
		jce := jobCompleter(jobLeaseID, job.JobID, now(), resp, err)
		if jce == nil {
			logger.WithJob(job).Infof("jobworker: job complete success")
		} else {
			logger.WithJob(job).Warnf("jobworker: job complete failure, but may retry")
		}
		return jce
	}

	bo = backoff.NewExponentialBackOff()
	err = backoff.Retry(complete, bo)
	if err != nil {
		logger.WithJob(job).Error("jobworker: job complete retries exceeded")
		return
	}

	return
}
