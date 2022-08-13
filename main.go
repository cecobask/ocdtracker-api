package main

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cecobask/ocdtracker-api/internal/api/account"
	"github.com/cecobask/ocdtracker-api/internal/api/middleware"
	"github.com/cecobask/ocdtracker-api/internal/api/ocdlog"
	"github.com/cecobask/ocdtracker-api/internal/aws"
	"github.com/cecobask/ocdtracker-api/internal/db/postgres"
	"github.com/cecobask/ocdtracker-api/pkg/log"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"net/http"
)

func main() {
	logger := log.NewLogger()
	ctx := log.ContextWithLogger(context.Background(), logger)

	sess := session.Must(session.NewSession())
	secretsManager := aws.NewSecretsManager(sess)

	postgresCreds, err := secretsManager.GetPostgresCreds()
	if err != nil {
		logger.Fatal("failed to get postgres credentials", zap.Error(err))
	}
	db, err := postgres.Connect(ctx, postgresCreds)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatal("failed to close database connection", zap.Error(err))
		}
	}()
	err = postgres.Migrate(db)
	if err != nil {
		logger.Fatal("failed to migrate database", zap.Error(err))
	}

	googleAppCreds, err := secretsManager.GetGoogleAppCreds(ctx)
	if err != nil {
		logger.Fatal("failed to get google application credentials", zap.Error(err))
	}
	config := firebase.Config{ProjectID: googleAppCreds.ProjectID}
	firebaseApp, err := firebase.NewApp(ctx, &config, option.WithCredentials(googleAppCreds))
	if err != nil {
		logger.Fatal("error initialising firebase app", zap.Error(err))
	}
	authClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		logger.Fatal("unable to create firebase auth client", zap.Error(err))
	}

	accountRepo := postgres.NewAccountRepository(db)
	ocdLogRepo := postgres.NewOCDLogRepository(db)
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
		Addr:    fmt.Sprintf(":%s", "8080"),
		Handler: chiRouter,
	}
	logger.Info("starting http server", zap.String("url", server.Addr))
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatal("failed to start http server", zap.Error(err))
	}
}
