# Call Me

Schedule SNS notifications to be sent on a repeating schedule, or in the future with database backed persistence.

# Features
 
 * Multiple agents can process jobs and schedules, by taking a lease on a schedule's individual cron or job.
 * JSON logging to allow CloudWatch to extract metrics from the logs.

# Dependenices

 * MySQL 5.6 database (e.g. Aurora)

# Usage

 * Create a database for `callme`.
   * `CREATE SCHEMA `callme` DEFAULT CHARACTER SET utf8 ;`
 * Set the `CALLME_CONNECTION_STRING` environment variable.
   * `export CALLME_CONNECTION_STRING='Server=localhost;Port=3306;Database=callme;Uid=root;Pwd=callme;Allow User Variables=true;multiStatements=true'`
 * Run the `callme` executable or Docker container.
 * Interact with the API to setup recurring notifications, or jobs at a specific point in time.

# Testing

The system is unit tested, and also has different types of integration test. The first is the `mysql` tests which test that the MySQL queries function as designed, while the second tests the system behaviour when loaded with synthetic data (see `./harness`).

# Key Concepts

 * Schedule
   * An operation that should happen on a periodic basis.
   * A schedule triggers the creation of jobs on a just-in-time basis, i.e. when the schedule is ready for the next item, a job is created at that moment.
 * Job
   * An operation that should happen at some point in the future.
   * Jobs will be created by Schedules, and will happen _near_ to the date and time of the next schedule. The exact date/time can't be guaranteed, since it depends on system load and the regularity of database polling to retrieve upcoming jobs.
 * Lease
   * A lock acquired by a system process which gives that process the right to process a schedule's cron or a single job for a time limit.

# Monitoring

## Prometheus Metrics

A Prometheus metric endpoint is opened at port 9090 by default, at the `/metrics` URL. To modify the listening port, set the `CALLME_PROMETHEUS_PORT` environment variable.

### Job Metrics

* job_leased_total
  * The number of calls to collect jobs, seperated by success, error or none_available.
* job_leased_duration_milliseconds
  * How long it took to execute a database claim and retrieve a job to work on.
* job_executed_total
  * The number of executions of jobs, split up by success.
* job_executed_duration_milliseconds
  * How long it took to execute jobs, split up by success.
* job_executed_delay_milliseconds
  * The amount of delay between a job's scheduled start time, and when it actually started.
* job_completed_total
  * The number of jobs marked as completed, split up by status.
* job_completed_duration_milliseconds
  * How long it took to mark jobs as completed.

### Schedule Metrics

* schedule_leased_total
  * The number of calls to collect a lease on a single schedule, seperated by success, error or none_available.
* schedule_leased_duration_milliseconds
  * How long it took to execute a database claim and retrieve a schedule to work on.
* schedule_executed_total
  * The number of times that the schedule's cron expression was parsed, separated by success or error.
* schedule_executed_delay_milliseconds
  * The amount of delay between a schedule's update time, and when it actually started.
* schedule_job_started_total
  * The count of jobs started by a schedule.
* schedule_job_started_duration_milliseconds
  * The amount of time taken to start jobs and mark the schedule as updated.

### Troubleshooting

Checklist for problems:

* Can the callme process access the database to acquire leases?
  * Check the `schedule_leased_total` and `job_leased_total` metrics for the error count.
  * Check the logs for database connection errors.
* Is there work to process?
  * The `schedule_leased_total` and `job_leased_total` metrics have a dimension of `none_available` if there aren't.
* Is the work being collected successfully?
  * The `schedule_leased_total` and `job_leased_total` metrics have a dimension of `error`, which should be zero.
  * There are also `schedule_leased_duration_milliseconds` and `job_leased_duration_milliseconds` metrics which track how long database operations are taking.
* Are jobs being processed?
  * The `job_executed_total` metric tells us how many jobs are processed. Check the `success` / `error` dimension though.
* Are jobs taking a long time to process?
  * The `job_executed_duration_milliseconds` metric records how long jobs are taking to execute.
* Are jobs starting later than they should?
  * The `job_executed_delay_milliseconds` metric records the delta between when jobs should have started, and when they did. If jobs are taking too long to start, you may wish to take action, e.g. increasing `CALLME_JOB_WORKER_COUNT` to process more on a single box, or increasing the performance of your database instance, depending on where the bottleneck is.
* How long is it taking to send the SNS notification?
  * The `job_executed_duration_milliseconds` metric tracks the duration.
* Is there a problem marking jobs as complete resulting in jobs being processed twice?
  * The `job_completed_total` metric tells us whether jobs are being completed in `success` or `error` states.
* Are database operations slow?
  * The `job_completed_duration_milliseconds` and `job_leased_duration_milliseconds` metrics record how long each database operation takes.
* Are schedules being processed?
  * The work should be completed succesfully (see `schedule_leased_total`), after that cron parsing counts are tracked by the `schedule_executed_total` metric.
* Are schedules being processed fast enough?
  * The `schedule_executed_delay_milliseconds` metric tracks how many milliseconds taken between when a schedule should have been executed, and when it did.
* Are jobs being started based on schedules?
  * The `schedule_job_started_total` metric tracks how many jobs were started from schedules.
* How long is it taking to schedule a new job to start and update the schedule?
  * The `schedule_job_started_duration_milliseconds` metric tracks how long it is taking to complete the schedule processing.
* Are any workers logging errors?
  * The `error_total` metric tracks the number of errors being logged by callme processes.

## Schedule scenarios

* Failed to get schedule
 * Impact: Possible delay to scheduling jobs, track the `schedule_executed_delay_milliseconds` metric to see the impact.
* Failed to parse cron
 * Impact: Cron will be retried forever every 30 minutes until database record is updated. 
* Failed to update schedule and start a job
 * Impact: The schedule update and job start is carried out within a transaction using the `sm_startjobandupdatecron` procedure, the operation will be retried by any schedule worker process when the lease expires (by default, 30 minutes), so the scheduled job will be deplayed.

## Job statuses

* Failed to get job
 * Impact: Possible delay to starting scheduled jobs, track the `job_executed_delay_milliseconds` metric to see the impact.

|                         | Sent SNS OK                | Sent SNS failed                                       |
| ----------------------- | -------------------------- | ----------------------------------------------------- |
| Marked complete OK      | A: No problem              | B: Job marked as errored, human intervention required |
| Marked complete failed  | C: SNS will be sent again  | D: Will be retried again                              |

### A: Sent the SNS notification and marked the job complete within the timeouts
This is the normal usage scenario. Jobs with this status do not require action.

### B: Sending the SNS notification failed, and the job was marked as complete with an error
Human intervention is required here, since the SNS notification may not have been sent for many reasons. Several errors could be the case:

  * The `callme` process lacked permission to write to the SNS topic.
    * Create an instance IAM role which allows the SNS topic to be written to and apply it to the instance running `callme`
  * The SNS topic didn't exist, so it can't be written to.
    * Update the database `job` record to change the SNS topic ARN and update any matching schedule which will create more jobs in the future
  * SNS was unavailable due to a lack of connectivity to SNS, or an AWS SNS outage.
    * Delete the `jobresponse` records associated with the outage, which will trigger the jobs to run again.

### C: Sent the SNS notification, but the job could not successfully be marked as complete within the completion timeout
These jobs may be sent multiple times, because the job worker was not able to mark them as complete, resulting in them being retried.

It's the responsibility of the receiver to reject duplicate messages if required, e.g. by creating a hash of the payload content and keeping a cache of previously received messages in Redis.

The most likely scenario here is the backing database has become unavailable. In this case, because jobs are requested from the database one at a time, then it's likely only jobs currently being processed will be sent more than once.

### D: Couldn't send the SNS notification or mark it as complete
This scenario is likely that after pulling a job from the database, all network connectivity was lost. As a result, the process won't be able to send notifications or mark it as complete. This is equivalent to a no-op.

# Configuration values

| Environment Variable          | Default               | Description                                           |
|-------------------------------|-----------------------|-------------------------------------------------------|
| CALLME_CONNECTION_STRING      | None, it's required   | The connection string to the database.                |
| CALLME_SCHEDULE_WORKER_COUNT  | 1                     | Number of routines processing schedules               |
| CALLME_JOB_WORKER_COUNT       | 1                     | Number of routines processing jobs.                   |
| CALLME_LOCK_EXPIRY_MINUTES    | 30                    | Minutes a routine has to process a job or schedule.   |
| CALLME_PROMETHEUS_PORT        | 9090                  | The port for the metrics HTTP endpoint                |

# Development

## Creating a database migration

 * Uses https://github.com/mattes/migrate
 * Use `date -u +"%Y%m%d%H%M%S"` to generate an ISO date for naming the migration.
