package response

import (
	"encoding/json"
	"net/http"
)

// JSON writes the value v as JSON to the ResponseWriter.
func JSON(v interface{}, w http.ResponseWriter, status int) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "error creating JSON", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}
