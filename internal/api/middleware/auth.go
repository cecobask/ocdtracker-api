package middleware

import (
	"context"
	"errors"
	firebase "firebase.google.com/go/v4"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/cecobask/ocd-tracker-api/internal/api"
	"github.com/cecobask/ocd-tracker-api/internal/log"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strings"
)

const (
	ctxKeyUser string = "ctxKeyUser"
)

var (
	ErrorNoUserInContext = errors.New("no user in context")
)

type authMiddleware struct {
	ctx        context.Context
	authClient *firebaseAuth.Client
}

func NewAuthMiddleware(ctx context.Context) *authMiddleware {
	logger := log.LoggerFromContext(ctx)
	config := firebase.Config{ProjectID: os.Getenv("FIREBASE_PROJECT_ID")}
	firebaseApp, err := firebase.NewApp(ctx, &config)
	if err != nil {
		logger.Fatal("error initializing firebase app", zap.Error(err))
	}
	authClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		logger.Fatal("unable to create firebase auth client", zap.Error(err))
	}
	return &authMiddleware{
		ctx:        ctx,
		authClient: authClient,
	}
}

// Handle verifies the jwt in the request and injects user into the request context
func (a *authMiddleware) Handle(next http.Handler) http.Handler {
	handlerFn := func(w http.ResponseWriter, r *http.Request) {
		bearerToken := tokenFromHeader(r)
		if bearerToken == "" {
			api.UnauthorisedError(w, r, "empty-bearer-token", nil)
			return
		}
		token, err := a.authClient.VerifyIDToken(a.ctx, bearerToken)
		if err != nil {
			api.UnauthorisedError(w, r, "unable-to-verify-jwt", err)
			return
		}
		user, err := a.authClient.GetUser(a.ctx, token.UID)
		if err != nil {
			api.NotFoundError(w, r, "unable-to-find-user", err)
			return
		}
		// add user to the request context and as field to the logger
		logger := log.LoggerFromContext(r.Context()).With(zap.String("user", user.UID))
		r = r.WithContext(ContextWithUser(log.ContextWithLogger(r.Context(), logger), user))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handlerFn)
}

func tokenFromHeader(r *http.Request) string {
	headerValue := r.Header.Get("Authorization")
	if len(headerValue) > 7 && strings.ToLower(headerValue[0:6]) == "bearer" {
		return headerValue[7:]
	}
	return ""
}

func ContextWithUser(ctx context.Context, user *firebaseAuth.UserRecord) context.Context {
	return context.WithValue(ctx, ctxKeyUser, user)
}

func UserFromContext(ctx context.Context) (*firebaseAuth.UserRecord, error) {
	user, ok := ctx.Value(ctxKeyUser).(*firebaseAuth.UserRecord)
	if ok {
		return user, nil
	}
	return nil, ErrorNoUserInContext
}
