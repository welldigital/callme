package mysql

import (
	"testing"
	"time"
)

func TestThatSchedulesCanBeCreated(t *testing.T) {
	if !testing.Short() {
		dsn, dbName, err := createTestDatabase()
		if err != nil {
			t.Errorf("failed to create test database with error: %v", err)
		}
		defer dropTestDatabase(dbName)

		// Create a schedule.
		sm := NewScheduleManager(dsn)
		scheduleID, err := sm.Create(time.Now().UTC(),
			"testarn",
			`{ nonsense: "payload" }`,
			[]string{"* * * *"},
			"externalid",
			"jobmanager_test")
		if err != nil {
			t.Fatalf("failed to create schedule with error: %v", err)
		}
		if scheduleID == 0 {
			t.Error("expected scheduleID > 0, but got 0")
		}
	}
}
