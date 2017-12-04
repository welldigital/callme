package response

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name         string
		v            interface{}
		status       int
		expectedBody string
	}{
		{
			name: "json",
			v: struct {
				Name string `json:"name"`
			}{Name: "value"},
			status:       http.StatusOK,
			expectedBody: `{"name":"value"}`,
		},
		{
			name:         "json: unsupported",
			v:            unmarshallable{Name: "value"},
			status:       http.StatusInternalServerError,
			expectedBody: `{"err":"failed to marshal JSON"}`,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		JSON(test.v, w, test.status)
		if w.Code != test.status {
			t.Errorf("%s: expected status %v, got %v", test.name, test.status, w.Code)
		}
		if w.Body.String() != test.expectedBody {
			t.Errorf("%s: expected body '%v', got '%v'", test.name, test.expectedBody, w.Body.String())
		}
	}
}

type unmarshallable struct {
	Name string
}

func (u unmarshallable) MarshalJSON() ([]byte, error) {
	return []byte{}, errors.New("failed to marshal")
}

func TestOK(t *testing.T) {
	tests := []struct {
		name         string
		ok           bool
		status       int
		expectedBody string
	}{
		{
			name:         "true",
			ok:           true,
			status:       http.StatusOK,
			expectedBody: `{"ok":true}`,
		},
		{
			name:         "true",
			ok:           false,
			status:       http.StatusBadGateway,
			expectedBody: `{"ok":false}`,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		OK(test.ok, w, test.status)
		if w.Code != test.status {
			t.Errorf("%s: expected status %v, got %v", test.name, test.status, w.Code)
		}
		if w.Body.String() != test.expectedBody {
			t.Errorf("%s: expected body '%v', got '%v'", test.name, test.expectedBody, w.Body.String())
		}
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		status       int
		expectedBody string
	}{
		{
			name:         "test1",
			err:          errors.New("test1"),
			status:       http.StatusInternalServerError,
			expectedBody: `{"err":"test1"}`,
		},
		{
			name:         "test1",
			err:          errors.New("test1"),
			status:       http.StatusBadGateway,
			expectedBody: `{"err":"test1"}`,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		Error(test.err, w, test.status)
		if w.Code != test.status {
			t.Errorf("%s: expected status %v, got %v", test.name, test.status, w.Code)
		}
		if w.Body.String() != test.expectedBody {
			t.Errorf("%s: expected body '%v', got '%v'", test.name, test.expectedBody, w.Body.String())
		}
	}
}
