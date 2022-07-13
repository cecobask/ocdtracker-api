package middleware

import (
	"context"
	"net/http"
	"strconv"
)

const (
	ctxKeyPagination string = "ctxKeyPagination"
)

type paginationMiddleware struct {
	ctx context.Context
}

type PaginationDetails struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Total  int `json:"total"`
}

func NewPaginationMiddleware(ctx context.Context) *paginationMiddleware {
	return &paginationMiddleware{
		ctx: ctx,
	}
}

func (p paginationMiddleware) Handle(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		limitParam := r.URL.Query().Get("limit")
		offsetParam := r.URL.Query().Get("offset")
		limit, err := strconv.Atoi(limitParam)
		if err != nil {
			limit = 50
		}
		offset, err := strconv.Atoi(offsetParam)
		if err != nil {
			offset = 0
		}
		paginationDetails := PaginationDetails{
			Limit:  limit,
			Offset: offset,
		}
		ctx := context.WithValue(r.Context(), ctxKeyPagination, &paginationDetails)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func PaginationFromContext(ctx context.Context) *PaginationDetails {
	if paginationDetails, ok := ctx.Value(ctxKeyPagination).(*PaginationDetails); ok {
		return paginationDetails
	}
	return nil
}
