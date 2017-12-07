package main

import (
	"net/http"

	"github.com/welldigital/callme/api/response"
	"github.com/welldigital/callme/logger"
)

// NotFoundHandler is the 404 handler for the API.
type NotFoundHandler struct {
}

// ServeHTTP serves up the 404 not found handler.
func (nfh NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.For(pkg, "NotFoundHandler").WithField("url", r.URL).Infof("not found")
	response.ErrorString("404: not found", w, http.StatusNotFound)
}
