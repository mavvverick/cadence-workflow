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
	var req model.Query

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

	l.WriteJSON(r, w, http.StatusOK, exec)

}

//ActionHandler ....
func (l *JobProcessorController) ActionHandler(w http.ResponseWriter, r *http.Request) {
	_ = l.JobProcessorService.NotifyJobStateChange(w, r)
	// l.WriteJSON(r, w, http.StatusOK, nil)

}
