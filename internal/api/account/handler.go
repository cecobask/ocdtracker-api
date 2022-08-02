package account

import (
	"context"
	"encoding/json"
	"github.com/cecobask/ocd-tracker-api/internal/api"
	"github.com/cecobask/ocd-tracker-api/internal/api/middleware"
	"github.com/cecobask/ocd-tracker-api/internal/db/postgres"
	"github.com/cecobask/ocd-tracker-api/internal/entity"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v4"
	"io"
	"net/http"
)

type handler struct {
	ctx context.Context
	pg  *postgres.Client
}

func NewHandler(ctx context.Context, pg *postgres.Client) *handler {
	return &handler{
		ctx: ctx,
		pg:  pg,
	}
}

func (h handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	requestBody := processRequestBody(w, r)
	if requestBody == nil {
		return
	}
	err := h.pg.CreateAccount(r.Context(), *requestBody)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.Status(r, http.StatusCreated)
}

func (h handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	requestBody := processRequestBody(w, r)
	if requestBody == nil {
		return
	}
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	exists, _, err := h.pg.AccountExists(r.Context(), account.ID)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	if !exists {
		api.NotFoundError(w, r, "not-found", nil)
		return
	}
	err = h.pg.UpdateAccount(r.Context(), account.ID, *requestBody)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.NoContent(w, r)
}

func (h handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	result, err := h.pg.GetAccount(r.Context(), account.ID)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			api.NotFoundError(w, r, "not-found", err)
			return
		default:
			api.InternalServerError(w, r, "database-error", err)
			return
		}
	}
	render.JSON(w, r, result)
}

func (h handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	err = h.pg.DeleteAccount(r.Context(), account.ID)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
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
