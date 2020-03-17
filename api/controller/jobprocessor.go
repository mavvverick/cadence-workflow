package controller

import (
	"jobprocessor/api/model"
	"jobprocessor/api/service"
	"net/http"
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
