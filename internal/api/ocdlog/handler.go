package ocdlog

import (
	"context"
	"github.com/cecobask/ocd-tracker-api/internal/api"
	"github.com/cecobask/ocd-tracker-api/internal/db/postgres"
	"github.com/go-chi/render"
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
	result, err := h.pg.GetAllLogs(h.ctx)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.JSON(w, r, result)
}

func (h handler) DeleteAllLogs(w http.ResponseWriter, r *http.Request) {
	err := h.pg.DeleteAllLogs(h.ctx)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.Status(r, http.StatusOK)
}

func (h handler) GetLog(w http.ResponseWriter, r *http.Request) {
	result, err := h.pg.GetLog(h.ctx)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.JSON(w, r, result)
}

func (h handler) CreateOrUpdateLog(w http.ResponseWriter, r *http.Request) {
	err := h.pg.CreateOrUpdateLog(h.ctx)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.Status(r, http.StatusCreated)
}

func (h handler) DeleteLog(w http.ResponseWriter, r *http.Request) {
	err := h.pg.DeleteLog(h.ctx)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.Status(r, http.StatusOK)
}
