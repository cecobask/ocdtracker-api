package auth

import (
	"context"
	"errors"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/cecobask/ocd-tracker-api/internal/httperr"
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

func (fhm FirebaseHttpMiddleware) Middleware(next http.Handler) http.Handler {
	handlerFn := func(w http.ResponseWriter, r *http.Request) {
		logger := log.LoggerFromContext(r.Context())
		ctx := log.ContextWithLogger(r.Context(), logger)
		bearerToken := fhm.tokenFromHeader(r)
		if bearerToken == "" {
			httperr.Unauthorised(w, r, "empty-bearer-token", nil)
			return
		}
		token, err := fhm.AuthClient.VerifyIDToken(ctx, bearerToken)
		if err != nil {
			httperr.Unauthorised(w, r, "unable-to-verify-jwt", err)
			return
		}
		ctx = context.WithValue(ctx, ctxKeyUser, User{
			UUID:        token.UID,
			Email:       token.Claims["email"].(string),
			DisplayName: token.Claims["name"].(string),
		})
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handlerFn)
}

type User struct {
	UUID        string
	Email       string
	DisplayName string
}

func (fhm FirebaseHttpMiddleware) tokenFromHeader(r *http.Request) string {
	headerValue := r.Header.Get("Authorization")
	if len(headerValue) > 7 && strings.ToLower(headerValue[0:6]) == "bearer" {
		return headerValue[7:]
	}
	return ""
}

func UserFromContext(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(ctxKeyUser).(*User)
	if ok {
		return user, nil
	}
	return nil, ErrorNoUserInContext
}
