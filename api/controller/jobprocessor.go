package controller

import (
	"net/http"

	"github.com/YOVO-LABS/workflow/api/model"
	"github.com/YOVO-LABS/workflow/internal/service"
)

// JobProcessorController ...
type JobProcessorController struct {
	BaseController
	JobProcessorService *service.JobProcessorService
}

// CreateJob ...
func (l *JobProcessorController) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req model.QueryParams

	err := l.decodeAndValidate(r, &req)
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	exec, err := l.JobProcessorService.CreateJob(r.Context(), &req)
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	l.WriteJSON(r, w, http.StatusOK, exec.ID)
}


// GetJob ...
func (l *JobProcessorController) GetJob(w http.ResponseWriter, r *http.Request) {
	var req model.Workflow

	err := l.decodeAndValidate(r, &req)
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	exec, err := l.JobProcessorService.GetJobInfo(r.Context(), &req)
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	l.WriteJSON(r, w, http.StatusOK, exec)

}

func (l *JobProcessorController) JobStatusCount(w http.ResponseWriter, r *http.Request) {

	duration, ok := r.URL.Query()["duration"]
	if !ok {
		l.WriteErrorWithMessage(r, w, nil, "Missing Query Param Duration")
	}

	exec, err := l.JobProcessorService.JobStatusCount(r.Context(), duration[0])
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	l.WriteJSON(r, w, http.StatusOK, exec)
}

func (l *JobProcessorController) GetLogs(w http.ResponseWriter, r *http.Request) {

	duration, ok := r.URL.Query()["duration"]
	if !ok {
		l.WriteErrorWithMessage(r, w, nil, "Missing Query Param Duration")
	}

	starttime, ok := r.URL.Query()["starttime"]
	if !ok {
		l.WriteErrorWithMessage(r, w, nil, "Missing Query Param Duration")
	}

	err := l.JobProcessorService.GetLogs(r.Context(), starttime[0], duration[0])
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	l.WriteJSON(r, w, http.StatusOK, "Saved to /tmp/cadence-logs.csv")

}

// CreateCron ...
func (l *JobProcessorController) CreateCron(w http.ResponseWriter, r *http.Request) {
	var req model.Cron

	err := l.decodeAndValidate(r, &req)
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	cron, err := l.JobProcessorService.CreateCron(r.Context(), req.Time)
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	l.WriteJSON(r, w, http.StatusOK, cron)

}

// CreateCron ...
func (l *JobProcessorController) GetData(w http.ResponseWriter, r *http.Request) {
	var req model.DataRange

	err := l.decodeAndValidate(r, &req)
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	cron, err := l.JobProcessorService.GetData(r.Context(), &req)
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	l.WriteJSON(r, w, http.StatusOK, cron)

}