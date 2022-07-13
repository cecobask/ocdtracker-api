package ocdlog

import (
	"context"
	"encoding/json"
	"github.com/cecobask/ocd-tracker-api/internal/api"
	"github.com/cecobask/ocd-tracker-api/internal/db/postgres"
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
	result, err := h.pg.GetAllLogs(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.JSON(w, r, result)
}

func (h handler) DeleteAllLogs(w http.ResponseWriter, r *http.Request) {
	err := h.pg.DeleteAllLogs(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.NoContent(w, r)
}

func (h handler) GetLog(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "log-id"))
	if err != nil {
		api.BadRequestError(w, r, "invalid-id", err)
		return
	}
	result, err := h.pg.GetLog(r.Context(), id)
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

func (h handler) CreateOrUpdateLog(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		api.BadRequestError(w, r, "invalid-body", err)
		return
	}
	var log postgres.Log
	err = json.Unmarshal(body, &log)
	if err != nil {
		api.BadRequestError(w, r, "invalid-body", err)
		return
	}
	err = h.pg.CreateOrUpdateLog(r.Context(), log)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.Status(r, http.StatusCreated)
}

func (h handler) DeleteLog(w http.ResponseWriter, r *http.Request) {
	err := h.pg.DeleteLog(r.Context(), chi.URLParam(r, "log-id"))
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.NoContent(w, r)
}
