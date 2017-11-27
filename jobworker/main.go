package jobworker

import (
	"fmt"
	"time"

	"github.com/a-h/callme/logger"
	"github.com/a-h/callme/metrics"
	"github.com/a-h/callme/repetitive"

	"github.com/a-h/callme/data"

	"github.com/cenkalti/backoff"
)

const leaseName = "job"
const defaultTimeout = time.Minute * 5

// An Executor executes work.
type Executor func(arn string, payload string) (resp string, err error)

// NewJobWorker creates a worker for the repetitive.Work function which processes pending jobs.
func NewJobWorker(workerName string,
	lockExpiryMinutes int,
	jobGetter data.JobGetter,
	e Executor,
	jobCompleter data.JobCompleter) repetitive.Worker {
	return func() (workDone bool, err error) {
		return findAndExecuteWork(workerName, lockExpiryMinutes, jobGetter, e, jobCompleter, defaultTimeout)
	}
}

func findAndExecuteWork(workerName string,
	lockExpiryMinutes int,
	jobGetter data.JobGetter,
	e Executor,
	jobCompleter data.JobCompleter,
	timeout time.Duration) (workDone bool, err error) {
	// See if there's some work to do.
	jobGetStart := time.Now()
	job, ok, err := jobGetter(workerName, lockExpiryMinutes)
	jobGetDuration := time.Since(jobGetStart) / time.Millisecond
	if err != nil {
		logger.Errorf("%v: error getting job: %v", workerName, err)
		metrics.JobLeaseCounts.WithLabelValues("error").Inc()
		metrics.JobLeaseDurations.WithLabelValues("error").Observe(float64(jobGetDuration))
		return
	}
	if !ok {
		logger.Infof("%v: no job available", workerName)
		metrics.JobLeaseDurations.WithLabelValues("none_available").Observe(float64(jobGetDuration))
		metrics.JobLeaseCounts.WithLabelValues("none_available").Inc()
		return
	}
	metrics.JobLeaseDurations.WithLabelValues("success").Observe(float64(jobGetDuration))
	metrics.JobLeaseCounts.WithLabelValues("success").Inc()

	logger.WithJob(job).Infof("%v: executing", workerName)

	// Attempt to execute the work.
	var resp string
	var ee error
	execute := func() error {
		jobDelay := time.Now().UTC().Sub(job.When)
		jobExecuteStart := time.Now()
		resp, ee = e(job.ARN, job.Payload)
		jobExecuteDuration := time.Since(jobExecuteStart) / time.Millisecond
		if ee == nil {
			logger.WithJob(job).Infof("%v: success", workerName)
			metrics.JobExecutedCounts.WithLabelValues("success").Inc()
			metrics.JobExecutedDurations.WithLabelValues("success").Observe(float64(jobExecuteDuration))
			metrics.JobExecutedDelay.Observe(float64(jobDelay))
		} else {
			logger.WithJob(job).Warnf("%v: failure, but may retry, err: %v", workerName, ee)
			metrics.JobExecutedCounts.WithLabelValues("error").Inc()
			metrics.JobExecutedDurations.WithLabelValues("error").Observe(float64(jobExecuteDuration))
		}
		return ee
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = timeout
	executionError := backoff.Retry(execute, bo)
	if executionError != nil {
		logger.WithJob(job).Errorf("%v: retries exceeded, logging error: %v", workerName, executionError)
	} else {
		workDone = true
	}

	// Attempt to complete the work.
	complete := func() error {
		jobCompleteStart := time.Now()
		jce := jobCompleter(job.JobID, resp, executionError)
		jobCompleteDuration := time.Since(jobCompleteStart) / time.Millisecond
		if jce == nil {
			logger.WithJob(job).Infof("%v: job marked as complete successfully", workerName)
			metrics.JobCompletedCounts.WithLabelValues("success").Inc()
			metrics.JobCompletedDurations.WithLabelValues("success").Observe(float64(jobCompleteDuration))
		} else {
			logger.WithJob(job).Warnf("%v: job complete failure, but may retry", workerName)
			metrics.JobCompletedCounts.WithLabelValues("error").Inc()
			metrics.JobCompletedDurations.WithLabelValues("error").Observe(float64(jobCompleteDuration))
		}
		return jce
	}

	bo = backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = timeout
	completionError := backoff.Retry(complete, bo)
	if completionError != nil {
		logger.WithJob(job).Errorf("%v: job complete retries exceeded, error: %v", completionError, workerName)
	}

	err = mergeErrors(workerName, executionError, completionError)
	return
}

func mergeErrors(workerName string, execution, completion error) error {
	if execution == nil && completion == nil {
		return nil
	}
	if execution == nil && completion != nil {
		return fmt.Errorf("completion: %v", completion)
	}
	if execution != nil && completion == nil {
		return fmt.Errorf("execution: %v", execution)
	}
	return fmt.Errorf("%v: execution: %v, completion: %v", workerName, execution, completion)
}
