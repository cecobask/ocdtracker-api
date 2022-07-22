package ocdlog

import (
	"context"
	"github.com/cecobask/ocd-tracker-api/internal/db/postgres"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// NewRouter creates all routes associated with OCD logs
func NewRouter(ctx context.Context, pg *postgres.Client) http.Handler {
	h := NewHandler(ctx, pg)
	r := chi.NewRouter()
	r.Post("/", h.CreateLog)
	r.Get("/", h.GetAllLogs)
	r.Delete("/", h.DeleteAllLogs)
	r.Route("/{log-id}", func(r chi.Router) {
		r.Patch("/", h.UpdateLog)
		r.Get("/", h.GetLog)
		r.Delete("/", h.DeleteLog)
	})
	return r
}
