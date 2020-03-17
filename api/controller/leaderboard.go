package controller

import (
	"jobprocessor/api/model"
	"jobprocessor/api/service"
	"net/http"
)

// LeaderboardController ...
type LeaderboardController struct {
	BaseController
	LeaderboardService service.LeaderboardInterface
}

// CreateCron ...
func (l *LeaderboardController) CreateCron(w http.ResponseWriter, r *http.Request) {
	var req model.Cron

	err := l.decodeAndValidate(r, &req)
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	cron, err := l.LeaderboardService.CreateCron(r.Context(), req.Time)
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	l.WriteJSON(r, w, http.StatusOK, cron)

}

// TerminateCron ...
func (l *LeaderboardController) TerminateCron(w http.ResponseWriter, r *http.Request) {
	var req model.Workflow

	err := l.decodeAndValidate(r, &req)
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	err = l.LeaderboardService.TerminateCron(r.Context(), req.WfID)
	if err != nil {
		l.WriteError(r, w, err)
		return
	}

	l.WriteJSON(r, w, http.StatusOK, req.WfID)

}
