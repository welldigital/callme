package mysql

import (
	"database/sql"

	"github.com/a-h/callme/logger"
	"github.com/a-h/callme/mysql/migrations"
	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/mysql"
	"github.com/mattes/migrate/source"

	_ "github.com/go-sql-driver/mysql" // Requires MySQL
	bindata "github.com/mattes/migrate/source/go-bindata"
)

// MigrationManager provides features to manage jobs using MySQL.
type MigrationManager struct {
	ConnectionString string
}

// NewMigrationManager creates a new MigrationManager.
func NewMigrationManager(connectionString string) MigrationManager {
	return MigrationManager{
		ConnectionString: connectionString,
	}
}

// UpdateSchema updates the database schema to match.
func (m MigrationManager) UpdateSchema() error {
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return err
	}

	migrationSource, err := LoadMigrations()
	if err != nil {
		return err
	}

	r, err := migrate.NewWithInstance("go-bindata", migrationSource, "mysql", driver)
	if err != nil {
		return err
	}
	r.Log = MigrationLogger{}
	err = r.Up()
	// Don't error on no changes.
	if err == migrate.ErrNoChange {
		err = nil
	}
	return err
}

// MigrationLogger provides a logger which the migration system can use.
type MigrationLogger struct {
}

// Printf provides a printf function.
func (ml MigrationLogger) Printf(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Verbose always returns true.
func (ml MigrationLogger) Verbose() bool {
	return true
}

// LoadMigrations loads database migrations from the bindata included with the program.
func LoadMigrations() (source.Driver, error) {
	s := bindata.Resource(migrations.AssetNames(),
		func(name string) ([]byte, error) {
			return migrations.Asset(name)
		})

	return bindata.WithInstance(s)
}
