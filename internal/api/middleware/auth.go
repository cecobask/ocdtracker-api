package middleware

import (
	"context"
	"database/sql"
	"errors"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/cecobask/ocd-tracker-api/internal/api"
	"github.com/cecobask/ocd-tracker-api/internal/db/postgres"
	"github.com/cecobask/ocd-tracker-api/pkg/entity"
	"github.com/cecobask/ocd-tracker-api/pkg/log"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

const (
	ctxKeyAccount string = "ctxKeyAccount"
)

var (
	ErrorNoAccountInContext = errors.New("no account in context")
)

type authMiddleware struct {
	ctx         context.Context
	authClient  *firebaseAuth.Client
	accountRepo *postgres.AccountRepository
}

func NewAuthMiddleware(ctx context.Context, authClient *firebaseAuth.Client, accountRepo *postgres.AccountRepository) *authMiddleware {
	return &authMiddleware{
		ctx:         ctx,
		authClient:  authClient,
		accountRepo: accountRepo,
	}
}

// Handle verifies the jwt in the request and injects user into the request context
func (a *authMiddleware) Handle(next http.Handler) http.Handler {
	handlerFn := func(w http.ResponseWriter, r *http.Request) {
		bearerToken := tokenFromHeader(r)
		if bearerToken == "" {
			api.UnauthorisedError(w, r, "invalid-jwt", nil)
			return
		}
		token, err := a.authClient.VerifyIDToken(a.ctx, bearerToken)
		if err != nil {
			api.UnauthorisedError(w, r, "invalid-jwt", err)
			return
		}
		user, err := a.authClient.GetUser(a.ctx, token.UID)
		if err != nil {
			api.NotFoundError(w, r, "firebase-account-not-found", err)
			return
		}
		logger := log.LoggerFromContext(r.Context()).With(zap.String("account", user.UID))
		ctx := log.ContextWithLogger(r.Context(), logger)
		r = r.WithContext(ctx)
		account, err := a.accountRepo.GetAccount(ctx, user.UID)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				account = &entity.Account{
					ID:          user.UID,
					Email:       &user.Email,
					DisplayName: &user.DisplayName,
					PhotoURL:    &user.PhotoURL,
				}
				err = a.accountRepo.CreateAccount(ctx, account)
				if err != nil {
					api.InternalServerError(w, r, "database-error", err)
					return
				}
			default:
				api.InternalServerError(w, r, "database-error", err)
				return
			}
		}
		ctx = ContextWithAccount(log.ContextWithLogger(r.Context(), logger), account)
		r = r.WithContext(ctx)
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

func ContextWithAccount(ctx context.Context, account *entity.Account) context.Context {
	return context.WithValue(ctx, ctxKeyAccount, account)
}

func AccountFromContext(ctx context.Context) (*entity.Account, error) {
	user, ok := ctx.Value(ctxKeyAccount).(*entity.Account)
	if ok {
		return user, nil
	}
	return nil, ErrorNoAccountInContext
}
