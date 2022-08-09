package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cecobask/ocd-tracker-api/internal/db"
	"github.com/cecobask/ocd-tracker-api/pkg/entity"
	"github.com/cecobask/ocd-tracker-api/pkg/log"
	"github.com/georgysavva/scany/sqlscan"
	"github.com/google/uuid"
)

type OCDLogRepository struct {
	Connection *sql.Conn
}

var _ db.OCDLogRepository = (*OCDLogRepository)(nil)

const (
	deleteAllLogsQuery = `DELETE FROM ocdlog WHERE account_id = $1;`
	deleteLogQuery     = `DELETE FROM ocdlog WHERE account_id = $1 AND id = $2;`
	getAllLogsQuery    = `SELECT id, account_id, created_at, updated_at, ruminate_minutes, anxiety_level, notes FROM ocdlog WHERE account_id = $1 ORDER BY created_at LIMIT $2 OFFSET $3;`
	getLogQuery        = `SELECT id, account_id, created_at, updated_at, ruminate_minutes, anxiety_level, notes FROM ocdlog WHERE account_id = $1 AND id = $2 LIMIT 1;`
	getRowCountQuery   = `SELECT count(*) FROM ocdlog WHERE account_id = $1;`
)

func NewOCDLogRepository(conn *sql.Conn) *OCDLogRepository {
	return &OCDLogRepository{
		Connection: conn,
	}
}

func (repo *OCDLogRepository) GetAllLogs(ctx context.Context, accountID string, limit, offset int) (*entity.OCDLogList, error) {
	ocdLogList := entity.OCDLogList{
		Logs: make([]entity.OCDLog, 0),
	}
	var rowCount int
	err := sqlscan.Get(ctx, repo.Connection, &rowCount, getRowCountQuery, accountID)
	if err != nil {
		return nil, err
	}
	paginationDetails := entity.PaginationDetails{
		Limit:  limit,
		Offset: offset,
		Total:  rowCount,
	}
	var ocdLogs []entity.OCDLog
	err = sqlscan.Select(ctx, repo.Connection, &ocdLogs, getAllLogsQuery, accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	ocdLogList.Logs = ocdLogs
	paginationDetails.Count = len(ocdLogList.Logs)
	ocdLogList.Pagination = paginationDetails
	log.LoggerFromContext(ctx).Info(fmt.Sprintf("retrieved %d logs", len(ocdLogList.Logs)))
	return &ocdLogList, nil
}

func (repo *OCDLogRepository) DeleteAllLogs(ctx context.Context, accountID string) error {
	err := logExec(ctx, repo.Connection, deleteAllLogsQuery, "delete", accountID)
	if err != nil {
		return err
	}
	return nil
}

func (repo *OCDLogRepository) GetLog(ctx context.Context, accountID string, id uuid.UUID) (*entity.OCDLog, error) {
	ocdLog := entity.OCDLog{}
	err := sqlscan.Get(ctx, repo.Connection, &ocdLog, getLogQuery, accountID, id)
	if err != nil {
		return nil, err
	}
	return &ocdLog, nil
}

func (repo *OCDLogRepository) CreateLog(ctx context.Context, accountID string, ocdLog *entity.OCDLog) error {
	query, fieldValues, err := buildCreateQuery(ocdLog, accountID)
	if err != nil {
		return err
	}
	err = logExec(ctx, repo.Connection, *query, "create", fieldValues...)
	if err != nil {
		return err
	}
	return nil
}

func (repo *OCDLogRepository) UpdateLog(ctx context.Context, accountID string, id uuid.UUID, ocdLog *entity.OCDLog) error {
	query, fieldValues, err := buildUpdateQuery(ocdLog, accountID, &id)
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

func (repo *OCDLogRepository) DeleteLog(ctx context.Context, accountID string, id uuid.UUID) error {
	err := logExec(ctx, repo.Connection, deleteLogQuery, "delete", accountID, id)
	if err != nil {
		return err
	}
	return nil
}
