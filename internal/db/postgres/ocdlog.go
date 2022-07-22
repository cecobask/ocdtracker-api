package postgres

import (
	"context"
	"fmt"
	"github.com/cecobask/ocd-tracker-api/internal/api/middleware"
	"github.com/cecobask/ocd-tracker-api/internal/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"time"
)

type Log struct {
	ID              uuid.UUID  `json:"id"`
	AccountID       string     `json:"account_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
	RuminateMinutes int        `json:"ruminate_minutes"`
	AnxietyLevel    int        `json:"anxiety_level"`
	Notes           *string    `json:"notes,omitempty"`
}

type LogList struct {
	Logs       []Log                        `json:"logs"`
	Pagination middleware.PaginationDetails `json:"pagination"`
}

const (
	getRowCountQuery   = `SELECT count(*) FROM ocdlog WHERE account_id = $1;`
	getAllLogsQuery    = `SELECT id, account_id, created_at, updated_at, ruminate_minutes, anxiety_level, notes FROM ocdlog WHERE account_id = $1 ORDER BY created_at LIMIT $2 OFFSET $3;`
	deleteAllLogsQuery = `DELETE FROM ocdlog WHERE account_id = $1;`
	getLogQuery        = `SELECT id, account_id, created_at, updated_at, ruminate_minutes, anxiety_level, notes FROM ocdlog WHERE account_id = $1 AND id = $2 LIMIT 1;`
	createLogQuery     = `INSERT INTO ocdlog (account_id, ruminate_minutes, anxiety_level, notes) VALUES ($1, $2, $3, $4);`
	updateLogQuery     = `UPDATE ocdlog SET ruminate_minutes = $3, anxiety_level = $4, notes = $5, updated_at = CURRENT_TIMESTAMP WHERE account_id = $1 AND id = $2;`
	deleteLogQuery     = `DELETE FROM ocdlog WHERE account_id = $1 AND id = $2;`
)

func (pg *Client) GetAllLogs(ctx context.Context) (*LogList, error) {
	logList := LogList{
		Logs: make([]Log, 0),
	}
	user, err := middleware.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var rowCount int
	err = pg.Connection.QueryRow(ctx, getRowCountQuery, user.UID).Scan(&rowCount)
	if err != nil {
		return nil, err
	}
	pagination := middleware.PaginationFromContext(ctx)
	paginationDetails := middleware.PaginationDetails{
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
		Total:  rowCount,
	}
	rows, err := pg.Connection.Query(ctx, getAllLogsQuery, user.UID, pagination.Limit, pagination.Offset)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var l Log
		err = rows.Scan(&l.ID, &l.AccountID, &l.CreatedAt, &l.UpdatedAt, &l.RuminateMinutes, &l.AnxietyLevel, &l.Notes)
		if err != nil {
			return nil, err
		}
		logList.Logs = append(logList.Logs, l)
	}
	paginationDetails.Count = len(logList.Logs)
	logList.Pagination = paginationDetails
	log.LoggerFromContext(ctx).Info(fmt.Sprintf("retrieved %d logs", len(logList.Logs)))
	return &logList, nil
}

func (pg *Client) DeleteAllLogs(ctx context.Context) error {
	err := pg.exec(ctx, deleteAllLogsQuery, "delete")
	if err != nil {
		return err
	}
	return nil
}

func (pg *Client) GetLog(ctx context.Context, id uuid.UUID) (*Log, error) {
	l := Log{}
	user, err := middleware.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	row := pg.Connection.QueryRow(ctx, getLogQuery, user.UID, id)
	err = row.Scan(&l.ID, &l.AccountID, &l.CreatedAt, &l.UpdatedAt, &l.RuminateMinutes, &l.AnxietyLevel, &l.Notes)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (pg Client) CreateLog(ctx context.Context, l Log) error {
	err := pg.exec(ctx, createLogQuery, "create", l.RuminateMinutes, l.AnxietyLevel, l.Notes)
	if err != nil {
		return err
	}
	return nil
}

func (pg Client) UpdateLog(ctx context.Context, id uuid.UUID, l Log) error {
	err := pg.exec(ctx, updateLogQuery, "update", id, l.RuminateMinutes, l.AnxietyLevel, l.Notes)
	if err != nil {
		return err
	}
	return nil
}

func (pg *Client) DeleteLog(ctx context.Context, id string) error {
	err := pg.exec(ctx, deleteLogQuery, "delete", id)
	if err != nil {
		return err
	}
	return nil
}

func (pg Client) exec(ctx context.Context, query, action string, args ...interface{}) error {
	user, err := middleware.UserFromContext(ctx)
	if err != nil {
		return err
	}
	args = append([]interface{}{user.UID}, args...)
	commandTag, err := pg.Connection.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	log.LoggerFromContext(ctx).Info(fmt.Sprintf("%sd %d log/s", action, commandTag.RowsAffected()))
	return nil
}

func (pg Client) LogExists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := pg.GetLog(ctx, id)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}
