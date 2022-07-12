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

type loggerMiddleware struct {
	ctx context.Context
}

func NewLoggerMiddleware(ctx context.Context) *loggerMiddleware {
	return &loggerMiddleware{
		ctx: ctx,
	}
}

func (l *loggerMiddleware) Handle(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		requestTime := time.Now()
		wrw := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		defer func() {
			log.LoggerFromContext(l.ctx).Info(
				"request info",
				zap.String("id", uuid.New().String()),
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
