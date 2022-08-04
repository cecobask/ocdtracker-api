package ocdlog

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

// NewRouter creates all routes associated with OCD logs
func NewRouter(h *handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/", h.CreateLog)
	r.Get("/", h.GetAllLogs)
	r.Delete("/", h.DeleteAllLogs)
	r.Route("/{id}", func(r chi.Router) {
		r.Patch("/", h.UpdateLog)
		r.Get("/", h.GetLog)
		r.Delete("/", h.DeleteLog)
	})
	return r
}
