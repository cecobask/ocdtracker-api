package postgres

import (
	"context"
	"github.com/cecobask/ocd-tracker-api/internal/db"
	"github.com/cecobask/ocd-tracker-api/pkg/entity"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type AccountRepository struct {
	Connection *pgx.Conn
}

var _ db.AccountRepository = (*AccountRepository)(nil)

const (
	getAccountQuery    = `SELECT id, email, created_at, updated_at, display_name, wake_time, sleep_time, notification_interval FROM account WHERE id = $1 LIMIT 1;`
	deleteAccountQuery = `DELETE FROM account WHERE id = $1`
)

func NewAccountRepository(conn *pgx.Conn) *AccountRepository {
	return &AccountRepository{
		Connection: conn,
	}
}

func (repo *AccountRepository) CreateAccount(ctx context.Context, account entity.Account) error {
	query, fieldValues, err := buildCreateQuery(account, account.ID)
	if err != nil {
		return err
	}
	err = logExec(ctx, repo.Connection, *query, "create", fieldValues...)
	if err != nil {
		return err
	}
	return nil
}

func (repo *AccountRepository) UpdateAccount(ctx context.Context, id string, account entity.Account) error {
	query, fieldValues, err := buildUpdateQuery(account, id, nil)
	if err != nil {
		return err
	}
	if query != nil && fieldValues != nil {
		err = logExec(ctx, repo.Connection, *query, "update", fieldValues...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repo *AccountRepository) GetAccount(ctx context.Context, id string) (*entity.Account, error) {
	account := entity.Account{}
	err := pgxscan.Get(ctx, repo.Connection, &account, getAccountQuery, id)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (repo *AccountRepository) DeleteAccount(ctx context.Context, id string) error {
	err := logExec(ctx, repo.Connection, deleteAccountQuery, "delete", id)
	if err != nil {
		return err
	}
	return nil
}
