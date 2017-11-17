# Call Me

Schedule SNS notifications in the future or recurring with database backed persistence.

# Features
 
 * As few as possible.
 * Active / passive design. Agents take a lease on processing.
 * JSON logging to allow CloudWatch to extract metrics from the logs.

# Dependenices

 * MySQL 5.6 database (e.g. Aurora)

# Usage

 * Create a database for `callme`.
   * `CREATE SCHEMA `callme` DEFAULT CHARACTER SET utf8 ;`
 * Set the `CALLME_CONNECTION_STRING` environment variable.
   * `export CALLME_CONNECTION_STRING='Server=localhost;Port=3309;Database=callme;Uid=root;Pwd=callme;Allow User Variables=true;multiStatements=true'`
 * Run the `callme` executable or Docker container.
 * Interact with the API to setup recurring notifications, or jobs at a specific point in time.

# Monitoring

## Can the callme process access the database to acquire leases?
## What leases are currently valid and what processes are they assigned to?
## How long was it since the lease was last used?
## How many schedules have expired and should have been turned into jobs?
## How many jobs are waiting to run?
## Have many jobs have executed in a particular time period?
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
