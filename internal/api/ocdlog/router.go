package ocdlog

import (
	"context"
	"github.com/cecobask/ocd-tracker-api/internal/db/postgres"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// NewRouter creates all routes associated with OCD logs
func NewRouter(ctx context.Context, chiRouter *chi.Mux, pg *postgres.Client) http.Handler {
	h := NewHandler(ctx, pg)
	chiRouter.Route("/ocdlog", func(rootRouter chi.Router) {
		rootRouter.Get("/", h.GetAllLogs)
		rootRouter.Delete("/", h.DeleteAllLogs)
		rootRouter.Route("/{log-id}", func(logRouter chi.Router) {
			logRouter.Get("/", h.GetLog)
			logRouter.Put("/", h.CreateOrUpdateLog)
			logRouter.Delete("/", h.DeleteLog)
		})
	})
	return chiRouter
}
