package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cecobask/ocd-tracker-api/internal/entity"
	"github.com/jackc/pgx/v4"
	"strings"
)

const (
	createAccountQuery = `INSERT INTO account (id, email, display_name, wake_time, sleep_time, notification_interval) VALUES ($1, $2, $3, $4, $5, $6);`
	getAccountQuery    = `SELECT id, email, created_at, updated_at, display_name, wake_time, sleep_time, notification_interval FROM account WHERE id = $1 LIMIT 1;`
	deleteAccountQuery = `DELETE FROM account WHERE id = $1`
)

func (pg *Client) CreateAccount(ctx context.Context, account entity.Account) error {
	err := pg.logExec(ctx, createAccountQuery, "create", "account", account.ID, account.Email, account.DisplayName, account.WakeTime, account.SleepTime, account.NotificationInterval)
	if err != nil {
		return err
	}
	return nil
}

func (pg *Client) UpdateAccount(ctx context.Context, id string, account entity.Account) error {
	var (
		allowedFields = []string{"email", "display_name", "wake_time", "sleep_time", "notification_interval"}
		fields        []string
		fieldValues   []interface{}
		fieldUpdates  map[string]interface{}
	)
	jsonData, err := json.Marshal(account)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonData, &fieldUpdates)
	if err != nil {
		return err
	}
	// start at 2 because 1 is reserved for the id field
	index := 2
	fieldValues = append(fieldValues, id)
	for _, allowedField := range allowedFields {
		for fieldName, fieldValue := range fieldUpdates {
			if fieldName == allowedField {
				fields = append(fields, fmt.Sprintf("%s = $%d,", fieldName, index))
				fieldValues = append(fieldValues, fieldValue)
				index++
			}
		}
	}
	if len(fields) > 0 {
		fields = append(fields, "updated_at = CURRENT_TIMESTAMP")
		query := "UPDATE account SET " + strings.Join(fields, " ") + " WHERE id = $1;"
		err = pg.logExec(ctx, query, "update", "account", fieldValues...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pg *Client) GetAccount(ctx context.Context, id string) (*entity.Account, error) {
	row := pg.Connection.QueryRow(ctx, getAccountQuery, id)
	account := entity.Account{}
	err := row.Scan(
		&account.ID,
		&account.Email,
		&account.CreatedAt,
		&account.UpdatedAt,
		&account.DisplayName,
		&account.WakeTime,
		&account.SleepTime,
		&account.NotificationInterval,
	)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (pg *Client) DeleteAccount(ctx context.Context, id string) error {
	err := pg.logExec(ctx, deleteAccountQuery, "delete", "account", id)
	if err != nil {
		return err
	}

	return nil
}

func (pg *Client) AccountExists(ctx context.Context, id string) (bool, *entity.Account, error) {
	account, err := pg.GetAccount(ctx, id)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return false, nil, nil
		default:
			return false, nil, err
		}
	}
	return true, account, nil
}
