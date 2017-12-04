package response

import (
	"encoding/json"
	"net/http"

	"github.com/welldigital/callme/logger"
)

// JSON writes the value v as JSON to the ResponseWriter.
func JSON(v interface{}, w http.ResponseWriter, status int) {
	data, err := json.Marshal(v)
	if err != nil {
		logger.For("github.com/welldigital/callme/response", "JSON").WithError(err).Error("failed to marshal JSON")
		ErrorString("failed to marshal JSON", w, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}
