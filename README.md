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


## Can the callme process access the database to acquire leases?
## What leases are currently valid and what processes are they assigned to?
## How long was it since the lease was last used?
## How many schedules have expired and should have been turned into jobs?
## How long is it taking from a schedule expiring, to a job being created?
## How long is it taking to execute jobs?
## How long are jobs waiting in a queue?
## How long is it taking from a schedule expiring, to a job being executed?
## How many jobs are waiting to run?
## Have many jobs have executed in a particular time period?
## How many faults have occurred?
## Schedule statuses

* Failed to parse cron
  * The schedule lease lasts an hour, so in an hour it will be retried forever. Best to fix it. The API shouldn't have allowed an invalid cron expression, so there's likely been some human updates on the database.
* Failed to update schedule and start a job
  * The schedule update and job start is carried out within a transaction using the `sm_startjobandupdatecron` procedure, the operation will be retried when the lease expires (by default, an hour).

|                         | Got lease                  | Failed to parse cron    |  Failed to mark complete    |
| ----------------------- | -------------------------- | ----------------------------------------------------- |
| Failed to parse         | A: No problem              | B: Job marked as errored, human intervention required |
| Marked complete failed  | C: SNS will be sent again  | D: Will be retried again                              |

## Job statuses

|                         | Sent SNS OK                | Sent SNS failed                                       |
| ----------------------- | -------------------------- | ----------------------------------------------------- |
| Marked complete OK      | A: No problem              | B: Job marked as errored, human intervention required |
| Marked complete failed  | C: SNS will be sent again  | D: Will be retried again                              |

### A: Sent the SNS notification and marked the job complete within the timeouts
This is the normal usage scenario. Jobs with this status do not require action.

### B: Sending the SNS notification failed, and the job was marked as complete with an error
Human intervention is required here. Several errors could be the case.

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

# Development

## Creating a database migration

 * Uses https://github.com/mattes/migrate
 * Use `date -u +"%Y%m%d%H%M%S"` to generate an ISO date for naming the migration.
