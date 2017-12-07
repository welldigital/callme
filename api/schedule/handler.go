package schedule

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/welldigital/callme/api/response"
	"github.com/welldigital/callme/data"
	cron "gopkg.in/robfig/cron.v2"

	"github.com/gorilla/mux"
	"github.com/welldigital/callme/logger"
)

// Handler is the HTTP handler for the /schedule path of the API.
type Handler struct {
	ScheduleCreator     data.ScheduleCreator
	ScheduleByIDGetter  data.ScheduleByIDGetter
	ScheduleDeactivator data.ScheduleDeactivator
}

// PostRequest is the request that must be passed to create a schedule.
type PostRequest struct {
	ScheduleID int64     `json:"scheduleId"`
	From       time.Time `json:"from"`
	ARN        string    `json:"arn"`
	Payload    string    `json:"payload"`
	Crontabs   []string  `json:"crontabs"`
	ExternalID string    `json:"externalId"`
	By         string    `json:"by"`
}

// Validate that the SchedulePostRequest is valid.
func (spr PostRequest) Validate() error {
	if spr.ScheduleID > 0 {
		return errors.New("cannot post to an existing schedule")
	}
	if spr.ARN == "" {
		return errors.New("ARN is required")
	}
	if len(spr.ARN) > 2048 {
		return errors.New("maximum length of the ARN is 2048 characters")
	}
	if len(spr.Payload) > int(math.Pow(2, 24)-1) {
		return errors.New("exceeded maximum length of payload")
	}
	if len(spr.ExternalID) > 256 {
		return errors.New("maximum length of the ExternalID is 256 characters")
	}
	if len(spr.By) > 256 {
		return errors.New("maximum length of the By field is 256 characters")
	}
	if spr.Crontabs == nil {
		return errors.New("crontabs field must be present")
	}
	if len(spr.Crontabs) == 0 {
		return errors.New("at least one crontab must be provided")
	}
	for _, ct := range spr.Crontabs {
		_, err := cron.Parse(ct)
		if err != nil {
			logger.For(pkg, "PostRequest.Validate").
				WithField("crontab", ct).
				WithError(err).
				Error("failed to parse")
			return fmt.Errorf("failed to parse crontab with error '%v'", err)
		}
	}
	return nil
}

// New creates a new handler.
func New(creator data.ScheduleCreator, getter data.ScheduleByIDGetter, deactivator data.ScheduleDeactivator) *Handler {
	return &Handler{
		ScheduleCreator:     creator,
		ScheduleByIDGetter:  getter,
		ScheduleDeactivator: deactivator,
	}
}

const pkg = "github.com/welldigital/callme/api/schedule/handler"

// Post handles the creation of new schedules.
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	logger.For(pkg, "Post").WithField("url", r.URL).Info("start")
	// Read the new schedule
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.For(pkg, "Post").WithError(err).Error("failed to read body")
		response.ErrorString("failed to read body", w, http.StatusBadRequest)
		return
	}
	var s PostRequest
	if err := json.Unmarshal(body, &s); err != nil {
		logger.For(pkg, "Post").WithError(err).Error("failed to parse request")
		response.ErrorString("failed to parse request", w, http.StatusUnprocessableEntity)
		return
	}
	// Validate the schedule.
	if err = s.Validate(); err != nil {
		logger.For(pkg, "Post").WithError(err).Error("failed to validate schedule request")
		response.Error(err, w, http.StatusUnprocessableEntity)
		return
	}
	// Create it.
	s.ScheduleID, err = h.ScheduleCreator(s.From, s.ARN, s.Payload, s.Crontabs, s.ExternalID, s.By)
	if err != nil {
		logger.For(pkg, "Post").WithError(err).Error("failed to create schedule")
		response.ErrorString("failed to create schedule", w, http.StatusInternalServerError)
		return
	}
	response.JSON(s, w, http.StatusCreated)
}

// Get responds to the GET /schedule/{id} route.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	logger.For(pkg, "Get").WithField("url", r.URL).Info("start")
	vars := mux.Vars(r)
	id, hasID := vars["id"]
	if !hasID {
		logger.For(pkg, "Get").WithField("url", r.URL).Info("id not found")
		http.NotFound(w, r)
		return
	}
	scheduleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		logger.For(pkg, "Get").WithError(err).WithField("scheduleID", id).Error("failed to parse scheduleID")
		response.ErrorString("failed to parse scheduleID", w, http.StatusBadRequest)
		return
	}
	sc, ok, err := h.ScheduleByIDGetter(scheduleID)
	if err != nil {
		logger.For(pkg, "Get").WithError(err).WithField("scheduleID", scheduleID).Error("failed to retrieve schedule")
		response.ErrorString("failed to retrieve schedule", w, http.StatusInternalServerError)
		return
	}
	if !ok {
		logger.For(pkg, "Get").WithField("scheduleID", scheduleID).Warn("schedule not found")
		http.NotFound(w, r)
		return
	}
	response.JSON(sc, w, http.StatusOK)
}

// Deactivate deactivates a schedule by its id.
func (h *Handler) Deactivate(w http.ResponseWriter, r *http.Request) {
	logger.For(pkg, "Deactivate").WithField("url", r.URL).Info("start")
	vars := mux.Vars(r)
	id, hasID := vars["id"]
	if !hasID {
		logger.For(pkg, "Deactivate").WithField("url", r.URL).Warn("id not found")
		http.NotFound(w, r)
		return
	}
	scheduleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		logger.For(pkg, "Deactivate").WithError(err).WithField("scheduleID", id).Error("failed to parse scheduleID")
		response.ErrorString("failed to parse scheduleID", w, http.StatusBadRequest)
		return
	}
	ok, err := h.ScheduleDeactivator(scheduleID)
	if err != nil {
		logger.For(pkg, "Deactivate").WithError(err).WithField("scheduleID", scheduleID).Error("failed to deactivate schedule")
		response.ErrorString("failed to deactivate schedule", w, http.StatusInternalServerError)
		return
	}
	if !ok {
		logger.For(pkg, "Deactivate").WithField("scheduleID", scheduleID).Warn("could not deactivate schedule, it could not be found")
		response.OK(false, w, http.StatusNotModified)
		return
	}
	response.OK(true, w, http.StatusOK)
}
