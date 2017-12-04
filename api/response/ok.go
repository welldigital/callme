package response

import "net/http"

// OKResult is for creating a JSON response of `{ ok: true }` or `{ ok: false }`.
type OKResult struct {
	OK bool `json:"ok"`
}

// OK returns a JSON OKResult.
func OK(ok bool, w http.ResponseWriter, status int) {
	JSON(OKResult{OK: ok}, w, status)
}
