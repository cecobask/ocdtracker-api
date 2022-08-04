package account

import (
	"context"
	"encoding/json"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/cecobask/ocd-tracker-api/internal/api"
	"github.com/cecobask/ocd-tracker-api/internal/api/middleware"
	"github.com/cecobask/ocd-tracker-api/internal/db/postgres"
	"github.com/cecobask/ocd-tracker-api/pkg/entity"
	"github.com/go-chi/render"
	"io"
	"net/http"
)

type handler struct {
	ctx         context.Context
	accountRepo *postgres.AccountRepository
	authClient  *firebaseAuth.Client
}

func NewHandler(ctx context.Context, accountRepo *postgres.AccountRepository, authClient *firebaseAuth.Client) *handler {
	return &handler{
		ctx:         ctx,
		accountRepo: accountRepo,
		authClient:  authClient,
	}
}

func (h *handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	requestBody := processRequestBody(w, r)
	if requestBody == nil {
		return
	}
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	_, err = h.accountRepo.GetAccount(r.Context(), account.ID)
	if err != nil {
		api.HandleRetrievalError(w, r, err)
		return
	}
	err = h.accountRepo.UpdateAccount(r.Context(), account.ID, *requestBody)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	_, err = h.authClient.UpdateUser(r.Context(), account.ID, buildAccountUpdateParams(requestBody))
	if err != nil {
		api.InternalServerError(w, r, "firebase-error", err)
		return
	}
	render.NoContent(w, r)
}

func buildAccountUpdateParams(account *entity.Account) *firebaseAuth.UserToUpdate {
	params := &firebaseAuth.UserToUpdate{}
	if account.Email != nil {
		params.Email(*account.Email)
	}
	if account.DisplayName != nil {
		params.DisplayName(*account.DisplayName)
	}
	if account.Password != nil {
		params.Password(*account.Password)
	}
	if account.PhotoURL != nil {
		params.PhotoURL(*account.PhotoURL)
	}
	return params
}

func (h *handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	result, err := h.accountRepo.GetAccount(r.Context(), account.ID)
	if err != nil {
		api.HandleRetrievalError(w, r, err)
		return
	}
	render.JSON(w, r, result)
}

func (h *handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	err = h.accountRepo.DeleteAccount(r.Context(), account.ID)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	err = h.authClient.DeleteUser(r.Context(), account.ID)
	if err != nil {
		api.InternalServerError(w, r, "firebase-error", err)
		return
	}
	render.NoContent(w, r)
}

func processRequestBody(w http.ResponseWriter, r *http.Request) *entity.Account {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		api.BadRequestError(w, r, "invalid-request-body", err)
		return nil
	}
	var account entity.Account
	err = json.Unmarshal(body, &account)
	if err != nil {
		api.BadRequestError(w, r, "invalid-request-body", err)
		return nil
	}
	if err := account.Validate(); err != nil {
		api.BadRequestError(w, r, "invalid-request-body", err)
		return nil
	}
	return &account
}
