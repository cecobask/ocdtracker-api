package middleware

import (
	"context"
	"errors"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/cecobask/ocd-tracker-api/internal/api"
	"github.com/cecobask/ocd-tracker-api/internal/log"
	"net/http"
	"strings"
)

type FirebaseHttpMiddleware struct {
	AuthClient *firebaseAuth.Client
}

type ctxKey int

const (
	ctxKeyUser ctxKey = iota
)

var (
	ErrorNoUserInContext = errors.New("no user in context")
)

// Middleware verifies the jwt in the request and injects user into the request context
func (fhm FirebaseHttpMiddleware) Middleware(next http.Handler) http.Handler {
	handlerFn := func(w http.ResponseWriter, r *http.Request) {
		logger := log.LoggerFromContext(r.Context())
		ctx := log.ContextWithLogger(r.Context(), logger)
		bearerToken := fhm.tokenFromHeader(r)
		if bearerToken == "" {
			api.UnauthorisedError(w, r, "empty-bearer-token", nil)
			return
		}
		token, err := fhm.AuthClient.VerifyIDToken(ctx, bearerToken)
		if err != nil {
			api.UnauthorisedError(w, r, "unable-to-verify-jwt", err)
			return
		}
		user, err := fhm.AuthClient.GetUser(ctx, token.UID)
		if err != nil {
			api.NotFoundError(w, r, "unable-to-find-user", err)
			return
		}
		ctx = context.WithValue(ctx, ctxKeyUser, user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handlerFn)
}

func (fhm FirebaseHttpMiddleware) tokenFromHeader(r *http.Request) string {
	headerValue := r.Header.Get("Authorization")
	if len(headerValue) > 7 && strings.ToLower(headerValue[0:6]) == "bearer" {
		return headerValue[7:]
	}
	return ""
}

func UserFromContext(ctx context.Context) (*firebaseAuth.UserRecord, error) {
	user, ok := ctx.Value(ctxKeyUser).(*firebaseAuth.UserRecord)
	if ok {
		return user, nil
	}
	return nil, ErrorNoUserInContext
}
