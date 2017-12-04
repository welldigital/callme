package job

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/a-h/callme/api/response"
	"github.com/a-h/callme/data"
	"github.com/a-h/callme/logger"
	"github.com/gorilla/mux"
)

// Handler is the HTTP handler for the /job path of the API.
type Handler struct {
	JobAndResponseByIDGetter data.JobAndResponseByIDGetter
	JobStarter               data.JobStarter
	JobDeleter               data.JobDeleter
}

// New creates a new handler.
func New(getter data.JobAndResponseByIDGetter, starter data.JobStarter, deleter data.JobDeleter) *Handler {
	return &Handler{
		JobAndResponseByIDGetter: getter,
		JobStarter:               starter,
		JobDeleter:               deleter,
	}
}

const pkg = "github.com/a-h/callme/api/job/handler"

// Post handles the creation of new jobs.
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	logger.For(pkg, "Post").WithField("url", r.URL).Info("start")
	// Read the new job
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.For(pkg, "Post").WithError(err).Error("failed to read body")
		response.ErrorString("failed to read body", w, http.StatusBadRequest)
		return
	}
	var j data.Job
	if err := json.Unmarshal(body, &j); err != nil {
		logger.For(pkg, "Post").WithError(err).Errorf("failed to parse request")
		response.ErrorString("failed to parse request", w, http.StatusUnprocessableEntity)
		return
	}
	// Validate the job.
	if err = j.Validate(); err != nil {
		logger.WithJob(pkg, "Post", j).WithError(err).Errorf("failed to validate job")
		response.Error(err, w, http.StatusUnprocessableEntity)
	}
	// Start it.
	j, err = h.JobStarter(j.When, j.ARN, j.Payload, nil)
	if err != nil {
		logger.WithJob(pkg, "Post", j).WithError(err).Errorf("failed to start job")
		response.ErrorString("failed to start job", w, http.StatusInternalServerError)
		return
	}
	response.JSON(j, w, http.StatusCreated)
}

// Get gets a job by its id.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	logger.For(pkg, "Get").WithField("url", r.URL).Info("start")
	vars := mux.Vars(r)
	id, hasID := vars["id"]
	if !hasID {
		http.NotFound(w, r)
		return
	}
	jobID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		logger.For(pkg, "Get").WithError(err).WithField("jobID", id).Errorf("failed to parse jobID")
		response.ErrorString("could not parse jobID", w, http.StatusBadRequest)
		return
	}
	job, jobResp, jobOK, responseOK, err := h.JobAndResponseByIDGetter(jobID)
	if err != nil {
		logger.For(pkg, "Get").WithError(err).WithField("jobID", jobID).Error("failed to retrieve job")
		response.ErrorString("failed to retrieve job", w, http.StatusInternalServerError)
		return
	}
	if !jobOK {
		logger.For(pkg, "Get").WithField("jobID", jobID).Warn("job not found")
		http.NotFound(w, r)
		return
	}
	jr := data.JobAndResponse{
		Job:            job,
		JobResponse:    jobResp,
		HasJobResponse: responseOK,
	}
	response.JSON(jr, w, http.StatusOK)
}

// Delete deletes a job by its id.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	logger.For(pkg, "Delete").WithField("url", r.URL).Info("start")
	vars := mux.Vars(r)
	id, hasID := vars["id"]
	if !hasID {
		logger.For(pkg, "Delete").WithField("url", r.URL).Warn("id not found")
		http.NotFound(w, r)
		return
	}
	jobID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		logger.For(pkg, "Delete").WithError(err).WithField("jobID", id).Errorf("failed to parse jobID")
		response.ErrorString("failed to parse jobID", w, http.StatusBadRequest)
		return
	}
	ok, err := h.JobDeleter(jobID)
	if err != nil {
		logger.For(pkg, "Delete").WithError(err).WithField("jobID", jobID).Errorf("failed to delete job")
		http.Error(w, "failed to delete job", http.StatusInternalServerError)
		return
	}
	if !ok {
		logger.For(pkg, "Delete").WithField("jobID", jobID).Errorf("could not delete job, it's probably been processed already")
		response.OK(false, w, http.StatusNotModified)
		return
	}
	response.OK(true, w, http.StatusOK)
}
