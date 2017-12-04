package job

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/a-h/callme/data"
	"github.com/gorilla/mux"
)

func TestGetByID(t *testing.T) {
	g := func(jobID int64) (j data.Job, r data.JobResponse, jobOK, responseOK bool, err error) {
		j = data.Job{
			ARN:        "testarn",
			JobID:      1,
			Payload:    "testpayload",
			ScheduleID: nil,
			When:       time.Date(2000, time.January, 1, 1, 1, 0, 0, time.UTC),
		}
		jobOK = true
		return
	}

	jh := New(g, nil, nil)

	router := mux.NewRouter()
	router.Path("/job/{id}").Methods(http.MethodGet).HandlerFunc(jh.Get)

	tests := []struct {
		name           string
		r              *http.Request
		expectedStatus int
		skipBodyCheck  bool
		expectedBody   string
	}{
		{
			name:           "success",
			r:              httptest.NewRequest("GET", "/job/1", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"Job":{"JobID":1,"ScheduleID":null,"When":"2000-01-01T01:01:00Z","ARN":"testarn","Payload":"testpayload"},"JobResponse":{"JobResponseID":0,"JobID":0,"Time":"0001-01-01T00:00:00Z","Response":"","IsError":false,"Error":""},"HasJobResponse":false}`,
		},
		{
			name:           "invalid id",
			r:              httptest.NewRequest("GET", "/job/sad", nil),
			expectedStatus: http.StatusBadRequest,
			skipBodyCheck:  true,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, test.r)

		if test.expectedStatus != w.Code {
			t.Errorf("%s: expected status %v, got %v", test.name, test.expectedStatus, w.Code)
		}
		actualBody, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Errorf("%s: unexpected error reading body: '%v'", test.name, err)
		}
		if !test.skipBodyCheck {
			if test.expectedBody != string(actualBody) {
				t.Errorf("%s: expected body '%v', got '%v'", test.name, test.expectedBody, string(actualBody))
			}
		}
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name           string
		d              data.JobDeleter
		r              *http.Request
		expectedStatus int
		skipBodyCheck  bool
		expectedBody   string
	}{
		{
			name:           "success",
			d:              func(jobID int64) (ok bool, err error) { return true, nil },
			r:              httptest.NewRequest("POST", "/job/1/delete", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"ok":true}`,
		},
		{
			name:           "invalid job id",
			d:              func(jobID int64) (ok bool, err error) { return false, nil },
			r:              httptest.NewRequest("POST", "/job/_/delete", nil),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"err":"failed to parse jobID"}`,
		},
		{
			name:           "job id not found",
			d:              func(jobID int64) (ok bool, err error) { return false, nil },
			r:              httptest.NewRequest("POST", "/job/1/delete", nil),
			expectedStatus: http.StatusNotModified,
			expectedBody:   `{"ok":false}`,
		},
	}

	for _, test := range tests {
		router := mux.NewRouter()
		jh := New(nil, nil, test.d)
		router.Path("/job/{id}/delete").Methods(http.MethodPost).HandlerFunc(jh.Delete)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, test.r)

		if test.expectedStatus != w.Code {
			t.Errorf("%s: expected status %v, got %v", test.name, test.expectedStatus, w.Code)
		}
		actualBody, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Errorf("%s: unexpected error reading body: '%v'", test.name, err)
		}
		if !test.skipBodyCheck {
			if test.expectedBody != string(actualBody) {
				t.Errorf("%s: expected body '%v', got '%v'", test.name, test.expectedBody, string(actualBody))
			}
		}
	}
}

func TestPost(t *testing.T) {
	tests := []struct {
		name           string
		s              data.JobStarter
		r              *http.Request
		expectedStatus int
		skipBodyCheck  bool
		expectedBody   string
	}{
		{
			name: "successful post",
			r: httptest.NewRequest("POST", "/job/",
				strings.NewReader(`{ "when": "2000-01-01T00:00:00Z", "arn": "example.com", "payload": "test_payload" }`)),
			s: func(when time.Time, arn string, payload string, scheduleID *int64) (data.Job, error) {
				return data.Job{
					JobID:      1,
					When:       when,
					ARN:        arn,
					Payload:    payload,
					ScheduleID: scheduleID}, nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"JobID":1,"ScheduleID":null,"When":"2000-01-01T00:00:00Z","ARN":"example.com","Payload":"test_payload"}`,
		},
		{
			name: "try and update an existing job fails",
			r: httptest.NewRequest("POST", "/job/",
				strings.NewReader(`{ "jobID": 1, "when": "2000-01-01T00:00:00Z", "arn": "example.com", "payload": "test_payload" }`)),
			s: func(when time.Time, arn string, payload string, scheduleID *int64) (data.Job, error) {
				return data.Job{
					JobID:      1,
					When:       when,
					ARN:        arn,
					Payload:    payload,
					ScheduleID: scheduleID}, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"cannot post to an existing job"}`,
		},
		{
			name: "try and update a schedule fails",
			r: httptest.NewRequest("POST", "/job/",
				strings.NewReader(`{ "scheduleID": 35, "when": "2000-01-01T00:00:00Z", "arn": "example.com", "payload": "test_payload" }`)),
			s: func(when time.Time, arn string, payload string, scheduleID *int64) (data.Job, error) {
				return data.Job{
					JobID:      1,
					When:       when,
					ARN:        arn,
					Payload:    payload,
					ScheduleID: scheduleID}, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"cannot post to an existing schedule"}`,
		},
		{
			name: "missing ARN fails",
			r: httptest.NewRequest("POST", "/job/",
				strings.NewReader(`{ "when": "2000-01-01T00:00:00Z", "payload": "test_payload" }`)),
			s: func(when time.Time, arn string, payload string, scheduleID *int64) (data.Job, error) {
				return data.Job{
					JobID:      1,
					When:       when,
					ARN:        arn,
					Payload:    payload,
					ScheduleID: scheduleID}, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"an ARN is required"}`,
		},
		{
			name: "ARN is too big",
			r: httptest.NewRequest("POST", "/job/",
				strings.NewReader(`{ "when": "2000-01-01T00:00:00Z", "payload": "test_payload", "arn": "`+longString(3000)+`" }`)),
			s: func(when time.Time, arn string, payload string, scheduleID *int64) (data.Job, error) {
				return data.Job{
					JobID:      1,
					When:       when,
					ARN:        arn,
					Payload:    payload,
					ScheduleID: scheduleID}, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"maximum length of the ARN is 2048 characters"}`,
		},
		{
			name: "Payload is too big",
			r: httptest.NewRequest("POST", "/job/",
				strings.NewReader(`{ "when": "2000-01-01T00:00:00Z", "arn": "test_arn", "payload": "`+longString(1024*1024*32)+`" }`)),
			s: func(when time.Time, arn string, payload string, scheduleID *int64) (data.Job, error) {
				return data.Job{
					JobID:      1,
					When:       when,
					ARN:        arn,
					Payload:    payload,
					ScheduleID: scheduleID}, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"exceeded maximum length of payload"}`,
		},
	}

	for _, test := range tests {
		router := mux.NewRouter()
		jh := New(nil, test.s, nil)
		router.Path("/job/").Methods(http.MethodPost).HandlerFunc(jh.Post)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, test.r)

		if test.expectedStatus != w.Code {
			t.Errorf("%s: expected status %v, got %v", test.name, test.expectedStatus, w.Code)
		}
		actualBody, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Errorf("%s: unexpected error reading body: '%v'", test.name, err)
		}
		if !test.skipBodyCheck {
			if test.expectedBody != string(actualBody) {
				t.Errorf("%s: expected body '%v', got '%v'", test.name, test.expectedBody, string(actualBody))
			}
		}
	}
}

func longString(size int) string {
	s := make([]byte, size)
	for i := 0; i < len(s); i++ {
		s[i] = 'a'
	}
	return string(s)
}
