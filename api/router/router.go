package router

import (
	"github.com/YOVO-LABS/workflow/api/dicontainer"
	"github.com/YOVO-LABS/workflow/config"
	"github.com/YOVO-LABS/workflow/internal/handler"

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
		r.Post("/workflow/cron/create", container.LeaderboardController.CreateCron)
		r.Post("/workflow/terminate", container.LeaderboardController.TerminateCron)

		r.Post("/workflow/job/create", container.JobProcessorController.CreateJob)
		r.Get("/workflow/job/start", handler.StartJobHandler)
		r.Post("/workflow/job/register", handler.CallbackHandler)
		r.Get("/workflow/job/action", container.JobProcessorController.ActionHandler)
		r.Get("/workflow/job/list", handler.ListHandler)
	})

	h.mux.NotFound(container.HTTPErrorController.ResourceNotFound)
}
