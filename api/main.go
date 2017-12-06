package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/welldigital/callme/api/response"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/welldigital/callme/api/job"
	"github.com/welldigital/callme/logger"
	"github.com/welldigital/callme/mysql"
)

const pkg = "github.com/welldigital/callme/api"

func main() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	connectionString := os.Getenv("CALLME_CONNECTION_STRING")

	if connectionString == "" {
		logger.For(pkg, "main").Error("missing connection string environment variable (CALLME_CONNECTION_STRING)")
		os.Exit(-1)
	}
	apiPort := getIntegerSetting("CALLME_API_PORT", 8080)
	prometheusPort := getIntegerSetting("CALLME_PROMETHEUS_PORT", 7777)

	go func() {
		logger.For(pkg, "main").Info("starting prometheus listener")
		r := http.NewServeMux()
		r.Handle("/metrics", prometheus.Handler())
		http.ListenAndServe(fmt.Sprintf(":%v", prometheusPort), r)
	}()

	logger.For(pkg, "main").Info("creating job handler and router")
	jm := mysql.NewJobManager(connectionString)

	jh := job.New(jm.GetJobResponse, jm.StartJob, jm.DeleteJob)
	r := createRouter(jh)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%v", apiPort),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start serving
	go func() {
		logger.For(pkg, "main").Info("starting web server")
		log.Fatal(s.ListenAndServe())
	}()

	logger.For(pkg, "main").WithField("port", apiPort).Infof("listening on port %v", apiPort)

	<-sigs
	logger.For(pkg, "main").Info("shutting down")
	err := s.Close()
	if err != nil {
		logger.For(pkg, "main").WithError(err).Error("failed to shut down server")
	}
	logger.For(pkg, "main").Info("complete")
}

func getIntegerSetting(n string, def int) int {
	str := os.Getenv(n)
	if str == "" {
		logger.For(pkg, "getIntegerSetting").WithField("env", n).Info("environment variable not found, defaulting to %v", def)
		return def
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		logger.For(pkg, "getIntegerSetting").WithField("env", n).WithField("val", str).Warn("environment variable could not be pasrsed, defaulting to %v", def)
		return def
	}
	return int(i)
}

func createRouter(jh *job.Handler) *mux.Router {
	r := mux.NewRouter()
	r.NotFoundHandler = NotFoundHandler{}
	r.Path("/job").Methods(http.MethodPost).HandlerFunc(jh.Post)
	r.Path("/job/{id}").Methods(http.MethodGet).HandlerFunc(jh.Get)
	r.Path("/job/{id}/delete").Methods(http.MethodPost).HandlerFunc(jh.Delete)
	return r
}

// NotFoundHandler is the 404 handler for the API.
type NotFoundHandler struct {
}

func (nfh NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.For(pkg, "NotFoundHandler").WithField("url", r.URL).Infof("not found")
	response.ErrorString("404: not found", w, http.StatusNotFound)
}
