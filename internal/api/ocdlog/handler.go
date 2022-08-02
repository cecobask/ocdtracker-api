package ocdlog

import (
	"context"
	"encoding/json"
	"github.com/cecobask/ocd-tracker-api/internal/api"
	"github.com/cecobask/ocd-tracker-api/internal/api/middleware"
	"github.com/cecobask/ocd-tracker-api/internal/db/postgres"
	"github.com/cecobask/ocd-tracker-api/internal/entity"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
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

func (h handler) GetAllLogs(w http.ResponseWriter, r *http.Request) {
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	pagination := middleware.PaginationFromContext(r.Context())
	result, err := h.pg.GetAllLogs(r.Context(), account.ID, pagination.Limit, pagination.Offset)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.JSON(w, r, result)
}

func (h handler) DeleteAllLogs(w http.ResponseWriter, r *http.Request) {
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	err = h.pg.DeleteAllLogs(r.Context(), account.ID)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.NoContent(w, r)
}

func (h handler) GetLog(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequestError(w, r, "invalid-id", err)
		return
	}
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	result, err := h.pg.GetLog(r.Context(), account.ID, id)
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

func (h handler) UpdateLog(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequestError(w, r, "invalid-id", err)
		return
	}
	requestBody := processRequestBody(w, r)
	if requestBody == nil {
		return
	}
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	exists, _, err := h.pg.LogExists(r.Context(), account.ID, id)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	if !exists {
		api.NotFoundError(w, r, "not-found", nil)
		return
	}
	err = h.pg.UpdateLog(r.Context(), account.ID, id, *requestBody)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.NoContent(w, r)
}

func (h handler) CreateLog(w http.ResponseWriter, r *http.Request) {
	requestBody := processRequestBody(w, r)
	if requestBody == nil {
		return
	}
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	requestBody.AccountID = account.ID
	err = h.pg.CreateLog(r.Context(), *requestBody)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.Status(r, http.StatusCreated)
}

func (h handler) DeleteLog(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequestError(w, r, "invalid-id", err)
		return
	}
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	err = h.pg.DeleteLog(r.Context(), account.ID, id)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.NoContent(w, r)
}

func processRequestBody(w http.ResponseWriter, r *http.Request) *entity.OCDLog {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		api.BadRequestError(w, r, "invalid-request-body", err)
		return nil
	}
	var ocdLog entity.OCDLog
	err = json.Unmarshal(body, &ocdLog)
	if err != nil {
		api.BadRequestError(w, r, "invalid-request-body", err)
		return nil
	}
	if err := ocdLog.Validate(); err != nil {
		api.BadRequestError(w, r, "invalid-request-body", err)
		return nil
	}
	return &ocdLog
}
