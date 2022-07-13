package middleware

import (
	"context"
	"github.com/cecobask/ocd-tracker-api/internal/log"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type requestLoggerMiddleware struct {
	ctx context.Context
}

func NewRequestLoggerMiddleware(ctx context.Context) *requestLoggerMiddleware {
	return &requestLoggerMiddleware{
		ctx: ctx,
	}
}

func (rlm *requestLoggerMiddleware) Handle(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		requestTime := time.Now()
		logger := log.LoggerFromContext(r.Context()).With(zap.String("request_id", uuid.New().String()))
		r = r.WithContext(log.ContextWithLogger(r.Context(), logger))
		wrw := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		defer func() {
			logger.Info(
				"request info",
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
				zap.Int("status", wrw.Status()),
				zap.Duration("duration", time.Since(requestTime)),
			)
		}()
		next.ServeHTTP(wrw, r)
	}
	return http.HandlerFunc(fn)
}
