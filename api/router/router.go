package router

import (
	"github.com/YOVO-LABS/workflow/api/dicontainer"
	"github.com/YOVO-LABS/workflow/config"
	"github.com/go-chi/cors"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

//RoutingInterface ...
type RoutingInterface interface {
	Routes(serviceContainer *dicontainer.ServiceContainer)
	RouteMultiplexer() *chi.Mux
}

type router struct {
	config config.AppConfig
	mux    *chi.Mux
}

//NewRouter ...
func NewRouter(generalConfig config.AppConfig) RoutingInterface {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.RealIP)
	// mux.Use(SetJSON)
	// mux.Use(logger.NewStructuredLogger())
	mux.Use(cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
	}).Handler)
	return &router{
		mux:    mux,
		config: generalConfig,
		// logger: logger,
	}
}

func (h *router) RouteMultiplexer() *chi.Mux {
	return h.mux
}

func (h *router) Routes(container *dicontainer.ServiceContainer) {
	h.mux.Group(func(r chi.Router) {
		r.Post("/workflow/lb/cron/create", container.LeaderboardController.CreateCron)
		r.Post("/workflow/job/cron/create", container.JobProcessorController.CreateCron)
		r.Post("/workflow/cron/terminate", container.LeaderboardController.TerminateCron)

		r.Post("/v1/start_encode2", container.JobProcessorController.CreateJob)

		r.Post("/workflow/job/info", container.JobProcessorController.GetJob)
		r.Get("/workflow/job/count", container.JobProcessorController.JobStatusCount)
		r.Get("/workflow/job/logs", container.JobProcessorController.GetLogs)

		r.Get("/workflow/job/data", container.JobProcessorController.GetData)
	})

	h.mux.NotFound(container.HTTPErrorController.ResourceNotFound)
}