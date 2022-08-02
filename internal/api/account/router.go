package account

import (
	"context"
	"github.com/cecobask/ocd-tracker-api/internal/db/postgres"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// NewRouter creates all routes associated with accounts
func NewRouter(ctx context.Context, pg *postgres.Client) http.Handler {
	h := NewHandler(ctx, pg)
	r := chi.NewRouter()
	r.Post("/", h.CreateAccount)
	r.Route("/me", func(r chi.Router) {
		r.Patch("/", h.UpdateAccount)
		r.Get("/", h.GetAccount)
		r.Delete("/", h.DeleteAccount)
	})
	return r
}
