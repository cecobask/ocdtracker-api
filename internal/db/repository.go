package db

import (
	"context"
	"github.com/cecobask/ocd-tracker-api/pkg/entity"
	"github.com/google/uuid"
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, account entity.Account) error
	DeleteAccount(ctx context.Context, id string) error
	GetAccount(ctx context.Context, id string) (*entity.Account, error)
	UpdateAccount(ctx context.Context, id string, account entity.Account) error
}

type OCDLogRepository interface {
	CreateLog(ctx context.Context, accountID string, ocdLog entity.OCDLog) error
	DeleteAllLogs(ctx context.Context, accountID string) error
	DeleteLog(ctx context.Context, accountID string, id uuid.UUID) error
	GetAllLogs(ctx context.Context, accountID string, limit, offset int) (*entity.OCDLogList, error)
	GetLog(ctx context.Context, accountID string, id uuid.UUID) (*entity.OCDLog, error)
	UpdateLog(ctx context.Context, accountID string, id uuid.UUID, ocdLog entity.OCDLog) error
}
