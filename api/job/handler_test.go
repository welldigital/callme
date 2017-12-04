package job

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/a-h/callme/data"
	"github.com/gorilla/mux"
)

const notFoundMessage = `404 page not found` + "\n"

func TestGetByID(t *testing.T) {
	tests := []struct {
		name           string
		g              data.JobAndResponseByIDGetter
		r              *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			g: func(jobID int64) (j data.Job, r data.JobResponse, jobOK, responseOK bool, err error) {
				j = data.Job{
					ARN:        "testarn",
					JobID:      1,
					Payload:    "testpayload",
					ScheduleID: nil,
					When:       time.Date(2000, time.January, 1, 1, 1, 0, 0, time.UTC),
				}
				jobOK = true
				return
			},
			r:              httptest.NewRequest("GET", "/job/1", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"Job":{"JobID":1,"ScheduleID":null,"When":"2000-01-01T01:01:00Z","ARN":"testarn","Payload":"testpayload"},"JobResponse":{"JobResponseID":0,"JobID":0,"Time":"0001-01-01T00:00:00Z","Response":"","IsError":false,"Error":""},"HasJobResponse":false}`,
		},
		{
			name:           "missing id",
			r:              httptest.NewRequest("GET", "/job/", nil),
			expectedStatus: http.StatusNotFound,
			expectedBody:   notFoundMessage,
		},
		{
			name:           "invalid id",
			r:              httptest.NewRequest("GET", "/job/sad", nil),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"err":"failed to parse jobID"}`,
		},
		{
			name: "failed to get job",
			g: func(jobID int64) (j data.Job, r data.JobResponse, jobOK, responseOK bool, err error) {
				responseOK = false
				err = errors.New("failed to get job")
				return
			},
			r:              httptest.NewRequest("GET", "/job/1", nil),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"err":"failed to retrieve job"}`,
		},
		{
			name: "job not found",
			g: func(jobID int64) (j data.Job, r data.JobResponse, jobOK, responseOK bool, err error) {
				responseOK = false
				return
			},
			r:              httptest.NewRequest("GET", "/job/1", nil),
			expectedStatus: http.StatusNotFound,
			expectedBody:   notFoundMessage,
		},
	}

	for _, test := range tests {
		router := mux.NewRouter()
		jh := New(test.g, nil, nil)
		router.Path("/job/{id}").Methods(http.MethodGet).HandlerFunc(jh.Get)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, test.r)

		if test.expectedStatus != w.Code {
			t.Errorf("%s: expected status %v, got %v", test.name, test.expectedStatus, w.Code)
		}
		actualBody, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Errorf("%s: unexpected error reading body: '%v'", test.name, err)
		}
		if test.expectedBody != string(actualBody) {
			t.Errorf("%s: expected body '%v', got '%v'", test.name, test.expectedBody, string(actualBody))
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
			name:           "id not present",
			d:              func(jobID int64) (ok bool, err error) { return true, nil },
			r:              httptest.NewRequest("POST", "/job/something_else/another/delete", nil),
			expectedStatus: http.StatusNotFound,
			expectedBody:   notFoundMessage,
		},
		{
			name:           "invalid job id",
			d:              func(jobID int64) (ok bool, err error) { return false, nil },
			r:              httptest.NewRequest("POST", "/job/_/delete", nil),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"err":"failed to parse jobID"}`,
		},
		{
			name:           "failure to access database",
			d:              func(jobID int64) (ok bool, err error) { return false, errors.New("failed to access database") },
			r:              httptest.NewRequest("POST", "/job/1/delete", nil),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"err":"failed to delete job"}`,
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
			name:           "malformed body",
			r:              httptest.NewRequest("POST", "/job/", ReaderFunc{F: func(p []byte) (n int, err error) { return 0, errors.New("general fault") }}),
			s:              nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"err":"failed to read body"}`,
		},
		{
			name:           "invalid JSON body",
			r:              httptest.NewRequest("POST", "/job/", strings.NewReader("_not_json_")),
			s:              nil,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"failed to parse request"}`,
		},
		{
			name: "failure to start job",
			r: httptest.NewRequest("POST", "/job/",
				strings.NewReader(`{ "when": "2000-01-01T00:00:00Z", "arn": "example.com", "payload": "test_payload" }`)),
			s: func(when time.Time, arn string, payload string, scheduleID *int64) (data.Job, error) {
				return data.Job{}, errors.New("failed to start job")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"err":"failed to start job"}`,
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

type ReaderFunc struct {
	F func(p []byte) (n int, err error)
}

func (rf ReaderFunc) Read(p []byte) (n int, err error) {
	return rf.F(p)
}

func longString(size int) string {
	s := make([]byte, size)
	for i := 0; i < len(s); i++ {
		s[i] = 'a'
	}
	return string(s)
}
