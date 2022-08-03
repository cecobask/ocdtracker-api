package account

import (
	"context"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/cecobask/ocd-tracker-api/internal/db/postgres"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// NewRouter creates all routes associated with accounts
func NewRouter(ctx context.Context, accountRepo *postgres.AccountRepository, authClient *firebaseAuth.Client) http.Handler {
	h := NewHandler(ctx, accountRepo, authClient)
	r := chi.NewRouter()
	r.Route("/me", func(r chi.Router) {
		r.Patch("/", h.UpdateAccount)
		r.Get("/", h.GetAccount)
		r.Delete("/", h.DeleteAccount)
	})
	return r
}
