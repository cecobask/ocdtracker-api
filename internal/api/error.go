package api

import (
	"github.com/cecobask/ocd-tracker-api/internal/log"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"net/http"
)

func UnauthorisedError(w http.ResponseWriter, r *http.Request, slug string, err error) {
	httpRespondWithError(w, r, slug, err, "unauthorised", http.StatusUnauthorized)
}

func NotFoundError(w http.ResponseWriter, r *http.Request, slug string, err error) {
	httpRespondWithError(w, r, slug, err, "not-found", http.StatusNotFound)
}

func InternalServerError(w http.ResponseWriter, r *http.Request, slug string, err error) {
	httpRespondWithError(w, r, slug, err, "internal-server-error", http.StatusInternalServerError)
}

func BadRequestError(w http.ResponseWriter, r *http.Request, slug string, err error) {
	httpRespondWithError(w, r, slug, err, "bad-request", http.StatusBadRequest)
}

func httpRespondWithError(w http.ResponseWriter, r *http.Request, slug string, err error, message string, status int) {
	logger := log.LoggerFromContext(r.Context())
	logger.Warn(message, zap.String("error-slug", slug), zap.Int("status", status), zap.Error(err))
	resp := ErrorResponse{slug}
	render.Status(r, status)
	render.JSON(w, r, resp)
}

type ErrorResponse struct {
	Slug string `json:"slug"`
}
