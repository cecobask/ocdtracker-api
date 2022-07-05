package main

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"fmt"
	"github.com/cecobask/ocd-tracker-api/internal/auth"
	"github.com/cecobask/ocd-tracker-api/internal/db/postgres"
	"github.com/cecobask/ocd-tracker-api/internal/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func main() {
	ctx := context.Background()
	logger := log.LoggerFromContext(ctx)
	ctx = log.ContextWithLogger(ctx, logger)
	pg, err := postgres.Connect(ctx)
	if err != nil {
		logger.Fatal("failed to connect postgres", zap.Error(err))
	}
	defer pg.Connection.Close(context.Background())
	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", os.Getenv("SERVER_PORT")),
		Handler: router(ctx),
	}
	logger.Info("starting http server", zap.String("url", server.Addr))
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatal("failed to start http server", zap.Error(err))
	}
}

func router(ctx context.Context) http.Handler {
	chiRouter := chi.NewRouter()
	chiRouter.Use(middleware.Recoverer)
	addAuthMiddleware(ctx, chiRouter)
	chiRouter.Get("/test", func(writer http.ResponseWriter, request *http.Request) {})
	return chiRouter
}

func addAuthMiddleware(ctx context.Context, chiRouter *chi.Mux) {
	logger := log.LoggerFromContext(ctx)
	config := &firebase.Config{ProjectID: os.Getenv("FIREBASE_PROJECT_ID")}
	firebaseApp, err := firebase.NewApp(ctx, config)
	if err != nil {
		logger.Fatal("error initializing firebase app", zap.Error(err))
	}
	authClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		logger.Fatal("unable to create firebase auth client", zap.Error(err))
	}
	chiRouter.Use(auth.FirebaseHttpMiddleware{AuthClient: authClient}.Middleware)
	logger.Info("added firebase auth middleware to the router")
}
