package mysql

import (
	"database/sql"
	"math/rand"
	"os"
	"time"

	mysql "github.com/go-sql-driver/mysql"
)

func getTestConnectionString() string {
	dsn := os.Getenv("CALLME_CONNECTION_STRING")
	if dsn == "" {
		dsn = "root:callme@tcp(localhost:3309)/callme?parseTime=true&multiStatements=true"
	}
	return dsn
}

// CreateTestDatabase creates a test database for use with integration tests.
func CreateTestDatabase() (dsn, dbName string, err error) {
	parsedDSN, err := mysql.ParseDSN(getTestConnectionString())
	if err != nil {
		return
	}

	dbName = randomDatabaseName()

	db, err := sql.Open("mysql", getTestConnectionString())
	if err != nil {
		return
	}
	defer db.Close()

	_, err = db.Exec("CREATE DATABASE " + dbName)

	// Update the DSN to point at the new database.
	parsedDSN.DBName = dbName
	dsn = parsedDSN.FormatDSN()

	// Fill it with schema.
	mm := NewMigrationManager(dsn)
	err = mm.UpdateSchema()

	return
}

var randomSource = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomDatabaseName() string {
	r := rand.New(randomSource)
	b := make([]byte, 10)
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return "callme_test_" + string(b)
}

// DropTestDatabase drops a database - for use with integration tests.
func DropTestDatabase(name string) (err error) {
	db, err := sql.Open("mysql", getTestConnectionString())
	if err != nil {
		return
	}
	defer db.Close()

	_, err = db.Exec("DROP DATABASE " + name)
	return
}
