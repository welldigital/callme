package mysql

import (
	"testing"
	"time"

	"github.com/a-h/callme/data"
)

func TestScheduleManager(t *testing.T) {
	if !testing.Short() {
		dsn, dbName, err := createTestDatabase()
		if err != nil {
			t.Errorf("failed to create test database with error: %v", err)
		}
		defer dropTestDatabase(dbName)

		// Create a schedule.
		var emptyTime time.Time
		// Start in the past in case of clock skew.
		from := time.Now().UTC().Add(time.Minute * -5)
		expected := data.ScheduleCrontab{
			Schedule: data.Schedule{
				Active:          true,
				ARN:             "testarn",
				By:              "jobmanager_test",
				Created:         time.Now().UTC(),
				DeactivatedDate: emptyTime,
				ExternalID:      "externalid",
				From:            from,
				Payload:         `{ nonsense: "payload" }`,
				ScheduleID:      1,
			},
			Crontab: data.Crontab{
				Crontab:     "* * * *",
				CrontabID:   1,
				LastUpdated: emptyTime,
				Next:        from,
				Previous:    emptyTime,
				ScheduleID:  1,
			},
		}

		sm := NewScheduleManager(dsn)
		scheduleID, err := sm.Create(expected.Schedule.From,
			expected.Schedule.ARN,
			expected.Schedule.Payload,
			[]string{expected.Crontab.Crontab},
			expected.Schedule.ExternalID,
			expected.Schedule.By)
		if err != nil {
			t.Fatalf("failed to create schedule with error: %v", err)
		}
		if scheduleID == 0 {
			t.Error("expected scheduleID > 0, but got 0")
		}

		// Start processing schedules, newly processed ones should appear in the list.
		scs, err := sm.GetSchedules()
		if err != nil {
			t.Fatalf("faied to get schedules with error: %v", err)
		}
		if len(scs) != 1 {
			t.Fatalf("expected to retrieve one schedule crontab, but got %v", len(scs))
		}
		actual := scs[0]
		AssertSchedule(t, "get schedule", expected.Schedule, actual.Schedule)
		AssertCrontab(t, "get schedule", expected.Crontab, actual.Crontab)

		// Update crontab to be checked again in the future.
		newPrevious := actual.Crontab.Previous
		newNext := time.Now().Add(time.Hour * 24)
		err = sm.UpdateCron(actual.Crontab.CrontabID, newPrevious, newNext, newNext)
		if err != nil {
			t.Errorf("failed to update the cron with error: %v", err)
		}

		// Check it's gone from the list.
		scs, err = sm.GetSchedules()
		if err != nil {
			t.Fatalf("faied to get schedules with error: %v", err)
		}
		if len(scs) != 0 {
			t.Errorf("expected not to retrieve any scheduled crontabs, but got %v", len(scs))
		}
	}
}

func AssertCrontab(t *testing.T, testName string, expected, actual data.Crontab) {
	if expected.Crontab != actual.Crontab {
		t.Errorf("%v: expected crontab Crontab='%v', but was '%v'", testName, expected.Crontab, actual.Crontab)
	}
	if expected.CrontabID != actual.CrontabID {
		t.Errorf("%v: expected crontab CrontabID='%v', but was '%v'", testName, expected.CrontabID, actual.CrontabID)
	}
	if !dateIsWithinRange(expected.LastUpdated, actual.LastUpdated, time.Minute*5) {
		t.Errorf("%v: expected crontab LastUpdated='%v', but was '%v'", testName, expected.LastUpdated, actual.LastUpdated)
	}
	if !dateIsWithinRange(expected.Next, actual.Next, time.Minute*5) {
		t.Errorf("%v: expected crontab Next='%v', but was '%v'", testName, expected.Next, actual.Next)
	}
	if !dateIsWithinRange(expected.Previous, actual.Previous, time.Minute*5) {
		t.Errorf("%v: expected crontab Previous='%v', but was '%v'", testName, expected.Previous, actual.Previous)
	}
	if expected.ScheduleID != actual.ScheduleID {
		t.Errorf("%v: expected crontab ScheduleID='%v', but was '%v'", testName, expected.ScheduleID, actual.ScheduleID)
	}
}

func AssertSchedule(t *testing.T, testName string, expected, actual data.Schedule) {
	if expected.Active != actual.Active {
		t.Errorf("%v: expected active Active='%v', but was '%v'", testName, expected.Active, actual.Active)
	}
	if expected.ARN != actual.ARN {
		t.Errorf("%v: expected schedule ARN='%v', but was '%v'", testName, expected.ARN, actual.ARN)
	}
	if expected.By != actual.By {
		t.Errorf("%v: expected schedule By='%v', but was '%v'", testName, expected.By, actual.By)
	}
	if !dateIsWithinRange(expected.Created, actual.Created, time.Minute*5) {
		t.Errorf("%v: expected schedule Created='%v', but was '%v'", testName, expected.Created, actual.Created)
	}
	if !dateIsWithinRange(expected.DeactivatedDate, actual.DeactivatedDate, time.Minute*5) {
		t.Errorf("%v: expected schedule DeactivatedDate='%v', but was '%v'", testName, expected.DeactivatedDate, actual.DeactivatedDate)
	}
	if expected.ExternalID != actual.ExternalID {
		t.Errorf("%v: expected schedule ExternalID='%v', but was '%v'", testName, expected.ExternalID, actual.ExternalID)
	}
	if !dateIsWithinRange(expected.From, actual.From, time.Minute*5) {
		t.Errorf("%v: expected schedule From='%v', but was '%v'", testName, expected.From, actual.From)
	}
	if expected.Payload != actual.Payload {
		t.Errorf("%v: expected schedule Payload='%v', but was '%v'", testName, expected.Payload, actual.Payload)
	}
	if expected.ScheduleID != actual.ScheduleID {
		t.Errorf("%v: expected schedule ID='%v', but was '%v'", testName, expected.ScheduleID, actual.ScheduleID)
	}
}

func dateIsWithinRange(a, b time.Time, r time.Duration) bool {
	if a.Equal(b) {
		return true
	}
	if a.Before(b) {
		if a.Add(r).After(b) {
			return true
		}
	}
	if b.Before(a) {
		if b.Add(r).After(a) {
			return true
		}
	}
	return false
}
