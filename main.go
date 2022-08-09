package main

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"fmt"
	"github.com/cecobask/ocd-tracker-api/internal/api/account"
	"github.com/cecobask/ocd-tracker-api/internal/api/middleware"
	"github.com/cecobask/ocd-tracker-api/internal/api/ocdlog"
	"github.com/cecobask/ocd-tracker-api/internal/db/postgres"
	"github.com/cecobask/ocd-tracker-api/pkg/log"
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
	conn, err := postgres.Connect(ctx)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Fatal("failed to close database connection", zap.Error(err))
		}
	}()
	err = postgres.Migrate(ctx, conn)
	if err != nil {
		logger.Fatal("failed to migrate database", zap.Error(err))
	}

	config := firebase.Config{ProjectID: os.Getenv("FIREBASE_PROJECT_ID")}
	firebaseApp, err := firebase.NewApp(ctx, &config)
	if err != nil {
		logger.Fatal("error initializing firebase app", zap.Error(err))
	}
	authClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		logger.Fatal("unable to create firebase auth client", zap.Error(err))
	}

	accountRepo := postgres.NewAccountRepository(conn)
	ocdLogRepo := postgres.NewOCDLogRepository(conn)
	accountHandler := account.NewHandler(ctx, accountRepo, authClient)
	ocdLogHandler := ocdlog.NewHandler(ctx, ocdLogRepo)

	chiRouter := chi.NewRouter()
	chiRouter.Use(
		chiMiddleware.Recoverer,
		middleware.NewRequestLoggerMiddleware(ctx).Handle,
		middleware.NewAuthMiddleware(ctx, authClient, accountRepo).Handle,
		middleware.NewPaginationMiddleware(ctx).Handle,
	)
	chiRouter.Mount("/ocdlog", ocdlog.NewRouter(ocdLogHandler))
	chiRouter.Mount("/account", account.NewRouter(accountHandler))
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
