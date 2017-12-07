package schedule

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/welldigital/callme/data"
)

const notFoundMessage = `404 page not found` + "\n"

func TestGetByID(t *testing.T) {
	tests := []struct {
		name           string
		g              data.ScheduleByIDGetter
		r              *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			g: func(scheduleID int64) (sc data.ScheduleCrontabs, ok bool, err error) {
				sc = data.ScheduleCrontabs{
					Schedule: data.Schedule{
						ScheduleID:      1,
						Active:          true,
						ARN:             "testarn",
						By:              "testby",
						Created:         time.Date(2010, time.January, 1, 1, 0, 0, 0, time.UTC),
						DeactivatedDate: time.Date(2000, time.January, 1, 1, 0, 0, 0, time.UTC),
						ExternalID:      "testexternalid",
						Payload:         "testpayload",
					},
				}
				ok = true
				return
			},
			r:              httptest.NewRequest("GET", "/schedule/1", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"schedule":{"scheduleId":1,"externalId":"testexternalid","by":"testby","arn":"testarn","payload":"testpayload","created":"2010-01-01T01:00:00Z","active":true,"deactivatedDate":"2000-01-01T01:00:00Z"},"crontabs":null}`,
		},
		{
			name:           "missing id",
			r:              httptest.NewRequest("GET", "/schedule", nil),
			expectedStatus: http.StatusNotFound,
			expectedBody:   notFoundMessage,
		},
		{
			name:           "invalid id",
			r:              httptest.NewRequest("GET", "/schedule/sad", nil),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"err":"failed to parse scheduleID"}`,
		},
		{
			name: "failed to retrieve schedule",
			g: func(scheduleID int64) (sc data.ScheduleCrontabs, ok bool, err error) {
				err = errors.New("some internal error")
				ok = false
				return
			},
			r:              httptest.NewRequest("GET", "/schedule/1", nil),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"err":"failed to retrieve schedule"}`,
		},
		{
			name: "schedule not found",
			g: func(scheduleID int64) (sc data.ScheduleCrontabs, ok bool, err error) {
				ok = false
				return
			},
			r:              httptest.NewRequest("GET", "/schedule/1", nil),
			expectedStatus: http.StatusNotFound,
			expectedBody:   notFoundMessage,
		},
	}

	for _, test := range tests {
		router := mux.NewRouter()
		sh := New(nil, test.g, nil)
		router.Path("/schedule/{id}").Methods(http.MethodGet).HandlerFunc(sh.Get)

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

func TestDeactivate(t *testing.T) {
	tests := []struct {
		name           string
		d              data.ScheduleDeactivator
		r              *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success",
			d:              func(scheduleID int64) (ok bool, err error) { return true, nil },
			r:              httptest.NewRequest("POST", "/schedule/1/deactivate", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"ok":true}`,
		},
		{
			name:           "id not present",
			d:              nil,
			r:              httptest.NewRequest("POST", "/schedule/something_else/another/deactivate", nil),
			expectedStatus: http.StatusNotFound,
			expectedBody:   notFoundMessage,
		},
		{
			name:           "invalid schedule id",
			d:              nil,
			r:              httptest.NewRequest("POST", "/schedule/_/deactivate", nil),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"err":"failed to parse scheduleID"}`,
		},
		{
			name:           "failure to access database",
			d:              func(scheduleID int64) (ok bool, err error) { return false, errors.New("failed to access database") },
			r:              httptest.NewRequest("POST", "/schedule/1/deactivate", nil),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"err":"failed to deactivate schedule"}`,
		},
		{
			name:           "schedule id not found",
			d:              func(scheduleID int64) (ok bool, err error) { return false, nil },
			r:              httptest.NewRequest("POST", "/schedule/1/deactivate", nil),
			expectedStatus: http.StatusNotModified,
			expectedBody:   `{"ok":false}`,
		},
	}

	for _, test := range tests {
		router := mux.NewRouter()
		sh := New(nil, nil, test.d)
		router.Path("/schedule/{id}/deactivate").Methods(http.MethodPost).HandlerFunc(sh.Deactivate)

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

func TestPost(t *testing.T) {
	tests := []struct {
		name           string
		s              data.ScheduleCreator
		r              *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful post",
			r: httptest.NewRequest("POST", "/schedule",
				strings.NewReader(`{"arn":"testarn","payload":"testpayload","crontabs":["* * * * *"],"externalId":"testexternalid","by":"testby"}`)),
			s: func(from time.Time, arn string, payload string, crontabs []string, externalID, by string) (scheduleID int64, err error) {
				return 1, nil
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"scheduleId":1,"from":"0001-01-01T00:00:00Z","arn":"testarn","payload":"testpayload","crontabs":["* * * * *"],"externalId":"testexternalid","by":"testby"}`,
		},
		{
			name:           "malformed body",
			r:              httptest.NewRequest("POST", "/schedule", ReaderFunc{F: func(p []byte) (n int, err error) { return 0, errors.New("general fault") }}),
			s:              nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"err":"failed to read body"}`,
		},
		{
			name:           "invalid JSON body",
			r:              httptest.NewRequest("POST", "/schedule", strings.NewReader("_not_json_")),
			s:              nil,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"failed to parse request"}`,
		},
		{
			name: "failure to create schedule",
			r: httptest.NewRequest("POST", "/schedule",
				strings.NewReader(`{"arn":"testarn","payload":"testpayload","crontabs":["* * * * *"],"externalId":"testexternalid","by":"testby"}`)),
			s: func(from time.Time, arn string, payload string, crontabs []string, externalID, by string) (scheduleID int64, err error) {
				return 0, errors.New("failed to create schedule")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"err":"failed to create schedule"}`,
		},
		{
			name: "try and update an existing schedule fails",
			r: httptest.NewRequest("POST", "/schedule",
				strings.NewReader(`{"scheduleId":1,"arn":"testarn","payload":"testpayload","crontabs":["* * * * *"],"externalId":"testexternalid","by":"testby"}`)),
			s:              nil,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"cannot post to an existing schedule"}`,
		},
		{
			name: "missing ARN fails",
			r: httptest.NewRequest("POST", "/schedule",
				strings.NewReader(`{ "when": "2000-01-01T00:00:00Z", "payload": "test_payload" }`)),
			s:              nil,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"ARN is required"}`,
		},
		{
			name: "ARN is too big",
			r: httptest.NewRequest("POST", "/schedule",
				strings.NewReader(`{"arn":"`+longString(3000)+`","payload":"testpayload","crontabs":["* * * * *"],"externalId":"testexternalid","by":"testby"}`)),
			s:              nil,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"maximum length of the ARN is 2048 characters"}`,
		},
		{
			name: "ExternalID is too big",
			r: httptest.NewRequest("POST", "/schedule",
				strings.NewReader(`{"arn":"testarn","payload":"testpayload","crontabs":["* * * * *"],"externalId":"`+longString(257)+`","by":"testby"}`)),
			s:              nil,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"maximum length of the ExternalID is 256 characters"}`,
		},
		{
			name: "By is too big",
			r: httptest.NewRequest("POST", "/schedule",
				strings.NewReader(`{"arn":"testarn","payload":"testpayload","crontabs":["* * * * *"],"externalId":"testexternalid","by":"`+longString(257)+`"}`)),
			s:              nil,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"maximum length of the By field is 256 characters"}`,
		},
		{
			name: "Payload is too big",
			r: httptest.NewRequest("POST", "/schedule",
				strings.NewReader(`{"arn":"testarn","payload":"`+longString(1024*1024*32)+`","crontabs":["* * * * *"],"externalId":"testexternalid","by":"testby"}`)),
			s:              nil,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"exceeded maximum length of payload"}`,
		},
		{
			name: "Missing crontabs field",
			r: httptest.NewRequest("POST", "/schedule",
				strings.NewReader(`{"arn":"testarn","payload":"testpayload","externalId":"testexternalid","by":"testby"}`)),
			s:              nil,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"crontabs field must be present"}`,
		},
		{
			name: "No crontabs in the array",
			r: httptest.NewRequest("POST", "/schedule",
				strings.NewReader(`{"arn":"testarn","payload":"testpayload","crontabs":[],"externalId":"testexternalid","by":"testby"}`)),
			s:              nil,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"at least one crontab must be provided"}`,
		},
		{
			name: "Junk crontabs in the array",
			r: httptest.NewRequest("POST", "/schedule",
				strings.NewReader(`{"arn":"testarn","payload":"testpayload","crontabs":["* * * * *", "nonsense"],"externalId":"testexternalid","by":"testby"}`)),
			s:              nil,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"err":"failed to parse crontab with error 'Expected 5 or 6 fields, found 1: nonsense'"}`,
		},
	}

	for _, test := range tests {
		router := mux.NewRouter()
		sh := New(test.s, nil, nil)
		router.Path("/schedule").Methods(http.MethodPost).HandlerFunc(sh.Post)

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
