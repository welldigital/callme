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

# Development

## Creating a database migration

 * Uses https://github.com/mattes/migrate
 * Use `date -u +"%Y%m%d%H%M%S"` to generate an ISO date for naming the migration.
