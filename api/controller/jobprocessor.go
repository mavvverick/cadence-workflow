package controller

import (
	"net/http"

	"github.com/YOVO-LABS/workflow/api/model"
	"github.com/YOVO-LABS/workflow/internal/service"
)

// JobProcessorController ...
type JobProcessorController struct {
	BaseController
	JobProcessorService service.JobProcessorInterface
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

//ActionHandler ....
func (l *JobProcessorController) ActionHandler(w http.ResponseWriter, r *http.Request) {
	_ = l.JobProcessorService.NotifyJobStateChange(w, r)
	// l.WriteJSON(r, w, http.StatusOK, nil)

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

func (l *JobProcessorController) ListJob(w http.ResponseWriter, r *http.Request) {

	duration, ok := r.URL.Query()["duration"]
	if !ok {
		l.WriteErrorWithMessage(r, w, nil, "Missing Query Param Duration")
	}

	exec, err := l.JobProcessorService.ListJob(r.Context(), duration[0])
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	l.WriteJSON(r, w, http.StatusOK, exec)
}