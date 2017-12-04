package response

import "net/http"

// ErrorResult is for creating a JSON response of `{ err: "message" }`.
type ErrorResult struct {
	Error string `json:"err"`
}

// Error returns a JSON ErrorResult.
func Error(err error, w http.ResponseWriter, status int) {
	JSON(ErrorResult{Error: err.Error()}, w, status)
}

// ErrorString returns a JSON ErrorResult.
func ErrorString(err string, w http.ResponseWriter, status int) {
	JSON(ErrorResult{Error: err}, w, status)
}
