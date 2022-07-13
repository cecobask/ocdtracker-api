package log

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

const (
	ctxKeyLogger string = "ctxKeyLogger"
)

func NewLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	return logger
}

func ContextWithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, logger)
}

func LoggerFromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(ctxKeyLogger).(*zap.Logger); ok {
		return logger
	}
	return NewLogger()
}
