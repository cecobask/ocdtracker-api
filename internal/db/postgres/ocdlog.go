package postgres

import (
	"context"
	"errors"
	"github.com/cecobask/ocd-tracker-api/internal/log"
	"time"
)

type Log struct {
	ID               string        `json:"id"`
	AccountID        string        `json:"account_id"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        *time.Time    `json:"updated_at,omitempty"`
	RuminateDuration time.Duration `json:"ruminate_duration"`
	AnxietyLevel     int           `json:"anxiety_level"`
	Notes            *string       `json:"notes,omitempty"`
}

type LogList struct {
	Logs []Log `json:"logs"`
}

func (pg *Client) GetAllLogs(ctx context.Context) (*LogList, error) {
	logger := log.LoggerFromContext(ctx)
	logger.Info("GetAllLogs() query invoked")
	return nil, errors.New("unimplemented GetAllLogs()")
}

func (pg *Client) DeleteAllLogs(ctx context.Context) error {
	logger := log.LoggerFromContext(ctx)
	logger.Info("DeleteAllLogs() query invoked")
	return errors.New("unimplemented DeleteAllLogs()")
}

func (pg *Client) GetLog(ctx context.Context) (*Log, error) {
	logger := log.LoggerFromContext(ctx)
	logger.Info("GetLog() query invoked")
	return nil, errors.New("unimplemented GetLog()")
}

func (pg *Client) CreateOrUpdateLog(ctx context.Context) error {
	logger := log.LoggerFromContext(ctx)
	logger.Info("CreateOrUpdateLog() query invoked")
	return errors.New("unimplemented CreateOrUpdateLog()")
}

func (pg *Client) DeleteLog(ctx context.Context) error {
	logger := log.LoggerFromContext(ctx)
	logger.Info("DeleteLog() query invoked")
	return errors.New("unimplemented DeleteLog()")
}
