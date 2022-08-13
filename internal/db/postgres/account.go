package postgres

import (
	"context"
	"database/sql"
	"github.com/cecobask/ocdtracker-api/internal/db"
	"github.com/cecobask/ocdtracker-api/pkg/entity"
	"github.com/georgysavva/scany/sqlscan"
)

type AccountRepository struct {
	DB *sql.DB
}

var _ db.AccountRepository = (*AccountRepository)(nil)

const (
	getAccountQuery    = `SELECT id, email, created_at, updated_at, display_name, wake_time, sleep_time, notification_interval, photo_url FROM account WHERE id = $1 LIMIT 1;`
	deleteAccountQuery = `DELETE FROM account WHERE id = $1`
)

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{
		DB: db,
	}
}

func (repo *AccountRepository) CreateAccount(ctx context.Context, account *entity.Account) error {
	pgElems, err := buildCreateQuery(account, account.ID)
	if err != nil {
		return err
	}
	err = logExec(ctx, repo.DB, pgElems.query, "create", pgElems.fieldValues...)
	if err != nil {
		return err
	}
	return nil
}

func (repo *AccountRepository) UpdateAccount(ctx context.Context, id string, account *entity.Account) error {
	pgElems, err := buildUpdateQuery(account, id, nil)
	if err != nil {
		return err
	}
	if pgElems != nil {
		err = logExec(ctx, repo.DB, pgElems.query, "update", pgElems.fieldValues...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repo *AccountRepository) GetAccount(ctx context.Context, id string) (*entity.Account, error) {
	account := entity.Account{}
	err := sqlscan.Get(ctx, repo.DB, &account, getAccountQuery, id)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (repo *AccountRepository) DeleteAccount(ctx context.Context, id string) error {
	err := logExec(ctx, repo.DB, deleteAccountQuery, "delete", id)
	if err != nil {
		return err
	}
	return nil
}
