package main

import (
	"context"
	"fmt"
	"github.com/cecobask/ocd-tracker-api/internal/api/middleware"
	"github.com/cecobask/ocd-tracker-api/internal/api/ocdlog"
	"github.com/cecobask/ocd-tracker-api/internal/db/postgres"
	"github.com/cecobask/ocd-tracker-api/internal/log"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func main() {
	logger := log.NewLogger()
	ctx := log.ContextWithLogger(context.Background(), logger)
	pg, err := postgres.Connect(ctx)
	if err != nil {
		logger.Fatal("failed to connect postgres", zap.Error(err))
	}
	defer pg.Connection.Close(context.Background())

	chiRouter := chi.NewRouter()
	chiRouter.Use(
		chiMiddleware.Recoverer,
		middleware.NewRequestLoggerMiddleware(ctx).Handle,
		middleware.NewAuthMiddleware(ctx).Handle,
		middleware.NewPaginationMiddleware(ctx).Handle,
	)
	chiRouter.Mount("/ocdlog", ocdlog.NewRouter(ctx, pg))
	server := http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")),
		Handler: chiRouter,
	}
	logger.Info("starting http server", zap.String("url", server.Addr))
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatal("failed to start http server", zap.Error(err))
	}
}
