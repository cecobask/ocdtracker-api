package log

import (
	"context"
	"go.uber.org/zap"
)

type ctxKey int

const (
	ctxKeyLogger ctxKey = iota
)

func ContextWithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, logger)
}

func LoggerFromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(ctxKeyLogger).(*zap.Logger); ok {
		return logger
	}
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	return logger
}
