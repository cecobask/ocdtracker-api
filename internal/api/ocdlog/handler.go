package ocdlog

import (
	"context"
	"encoding/json"
	"github.com/cecobask/ocdtracker-api/internal/api"
	"github.com/cecobask/ocdtracker-api/internal/api/middleware"
	"github.com/cecobask/ocdtracker-api/internal/db/postgres"
	"github.com/cecobask/ocdtracker-api/pkg/entity"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"io"
	"net/http"
)

type handler struct {
	ctx        context.Context
	ocdLogRepo *postgres.OCDLogRepository
}

func NewHandler(ctx context.Context, ocdLogRepo *postgres.OCDLogRepository) *handler {
	return &handler{
		ctx:        ctx,
		ocdLogRepo: ocdLogRepo,
	}
}

func (h *handler) GetAllLogs(w http.ResponseWriter, r *http.Request) {
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	pagination := middleware.PaginationFromContext(r.Context())
	result, err := h.ocdLogRepo.GetAllLogs(r.Context(), account.ID, pagination.Limit, pagination.Offset)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.JSON(w, r, result)
}

func (h *handler) DeleteAllLogs(w http.ResponseWriter, r *http.Request) {
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	err = h.ocdLogRepo.DeleteAllLogs(r.Context(), account.ID)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.NoContent(w, r)
}

func (h *handler) GetLog(w http.ResponseWriter, r *http.Request) {
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
	result, err := h.ocdLogRepo.GetLog(r.Context(), account.ID, id)
	if err != nil {
		api.HandleRetrievalError(w, r, err)
		return
	}
	render.JSON(w, r, result)
}

func (h *handler) UpdateLog(w http.ResponseWriter, r *http.Request) {
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
	_, err = h.ocdLogRepo.GetLog(r.Context(), account.ID, id)
	if err != nil {
		api.HandleRetrievalError(w, r, err)
		return
	}
	err = h.ocdLogRepo.UpdateLog(r.Context(), account.ID, id, requestBody)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.NoContent(w, r)
}

func (h *handler) CreateLog(w http.ResponseWriter, r *http.Request) {
	requestBody := processRequestBody(w, r)
	if requestBody == nil {
		return
	}
	account, err := middleware.AccountFromContext(r.Context())
	if err != nil {
		api.InternalServerError(w, r, "invalid-account-ctx", err)
		return
	}
	err = h.ocdLogRepo.CreateLog(r.Context(), account.ID, requestBody)
	if err != nil {
		api.InternalServerError(w, r, "database-error", err)
		return
	}
	render.Status(r, http.StatusCreated)
}

func (h *handler) DeleteLog(w http.ResponseWriter, r *http.Request) {
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
	err = h.ocdLogRepo.DeleteLog(r.Context(), account.ID, id)
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
