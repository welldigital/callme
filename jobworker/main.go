package jobworker

import (
	"fmt"
	"time"

	"github.com/a-h/callme/logger"
	"github.com/a-h/callme/repetitive"

	"github.com/a-h/callme/data"

	"github.com/cenkalti/backoff"
)

const leaseName = "job"

const defaultTimeout = time.Minute * 5

// An Executor executes work.
type Executor func(arn string, payload string) (resp string, err error)

// NewJobWorker creates a worker for the repetitive.Work function which processes pending jobs.
func NewJobWorker(leaseAcquirer data.LeaseAcquirer,
	nodeName string,
	leaseRescinder data.LeaseRescinder,
	jobGetter data.JobGetter,
	e Executor,
	jobCompleter data.JobCompleter) repetitive.Worker {
	return func() (workDone bool, err error) {
		return findAndExecuteWork(leaseAcquirer, nodeName, leaseRescinder, jobGetter, e, jobCompleter, defaultTimeout)
	}
}

func findAndExecuteWork(leaseAcquirer data.LeaseAcquirer,
	nodeName string,
	leaseRescinder data.LeaseRescinder,
	jobGetter data.JobGetter,
	e Executor,
	jobCompleter data.JobCompleter,
	timeout time.Duration) (workDone bool, err error) {
	leaseID, until, ok, err := leaseAcquirer(leaseName, nodeName)
	if err != nil {
		logger.Errorf("jobworker: failed to acquire lease with error: %v", err)
		return
	}
	if !ok {
		logger.Infof("jobworker: no work to do, another process has the lease")
		return
	}
	logger.Infof("jobworker: got lease %v on %v until %v", leaseID, leaseName, until)
	defer leaseRescinder(leaseID)

	// See if there's some work to do.
	j, err := jobGetter(leaseID)
	if err != nil {
		logger.Errorf("jobworker: error getting jobs: %v", err)
		return
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
	bo.MaxElapsedTime = timeout
	executionError := backoff.Retry(execute, bo)
	if executionError != nil {
		logger.WithJob(job).Errorf("jobworker: retries exceeded, logging error: %v", executionError)
	} else {
		workDone = true
	}

	// Attempt to complete the work.
	complete := func() error {
		jce := jobCompleter(leaseID, job.JobID, resp, executionError)
		if jce == nil {
			logger.WithJob(job).Infof("jobworker: job marked as complete successfully")
		} else {
			logger.WithJob(job).Warnf("jobworker: job complete failure, but may retry")
		}
		return jce
	}

	bo = backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = timeout
	completionError := backoff.Retry(complete, bo)
	if completionError != nil {
		logger.WithJob(job).Errorf("jobworker: job complete retries exceeded, error: %v", completionError)
	}

	err = mergeErrors(executionError, completionError)
	return
}

func mergeErrors(execution, completion error) error {
	if execution == nil && completion == nil {
		return nil
	}
	if execution == nil && completion != nil {
		return fmt.Errorf("completion: %v", completion)
	}
	if execution != nil && completion == nil {
		return fmt.Errorf("execution: %v", execution)
	}
	return fmt.Errorf("execution: %v, completion: %v", execution, completion)
}
