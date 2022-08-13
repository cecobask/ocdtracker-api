package api

import (
	"database/sql"
	"errors"
	"github.com/cecobask/ocdtracker-api/pkg/log"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"net/http"
)

func UnauthorisedError(w http.ResponseWriter, r *http.Request, message string, err error) {
	httpRespondWithError(w, r, "unauthorised", err, message, http.StatusUnauthorized)
}

func NotFoundError(w http.ResponseWriter, r *http.Request, message string, err error) {
	httpRespondWithError(w, r, "not-found", err, message, http.StatusNotFound)
}

func InternalServerError(w http.ResponseWriter, r *http.Request, message string, err error) {
	httpRespondWithError(w, r, "internal-server-error", err, message, http.StatusInternalServerError)
}

func BadRequestError(w http.ResponseWriter, r *http.Request, message string, err error) {
	httpRespondWithError(w, r, "bad-request", err, message, http.StatusBadRequest)
}

func httpRespondWithError(w http.ResponseWriter, r *http.Request, slug string, err error, message string, status int) {
	logger := log.LoggerFromContext(r.Context())
	logger.Warn(message, zap.String("error-slug", slug), zap.Int("status", status), zap.Error(err))
	resp := ErrorResponse{
		Slug:    slug,
		Message: message,
	}
	render.Status(r, status)
	render.JSON(w, r, resp)
}

type ErrorResponse struct {
	Slug    string `json:"slug"`
	Message string `json:"message"`
}

func HandleRetrievalError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		NotFoundError(w, r, "database-error", sql.ErrNoRows)
	default:
		InternalServerError(w, r, "database-error", err)
	}
}
