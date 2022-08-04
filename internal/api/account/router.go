package account

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

// NewRouter creates all routes associated with accounts
func NewRouter(h *handler) http.Handler {
	r := chi.NewRouter()
	r.Route("/me", func(r chi.Router) {
		r.Patch("/", h.UpdateAccount)
		r.Get("/", h.GetAccount)
		r.Delete("/", h.DeleteAccount)
	})
	return r
}
