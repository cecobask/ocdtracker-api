package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cecobask/ocdtracker-api/internal/db"
	"github.com/cecobask/ocdtracker-api/pkg/entity"
	"github.com/cecobask/ocdtracker-api/pkg/log"
	"github.com/georgysavva/scany/sqlscan"
	"github.com/google/uuid"
)

type OCDLogRepository struct {
	DB *sql.DB
}

var _ db.OCDLogRepository = (*OCDLogRepository)(nil)

const (
	deleteAllLogsQuery = `DELETE FROM ocdlog WHERE account_id = $1;`
	deleteLogQuery     = `DELETE FROM ocdlog WHERE account_id = $1 AND id = $2;`
	getAllLogsQuery    = `SELECT id, account_id, created_at, updated_at, ruminate_minutes, anxiety_level, notes FROM ocdlog WHERE account_id = $1 ORDER BY created_at LIMIT $2 OFFSET $3;`
	getLogQuery        = `SELECT id, account_id, created_at, updated_at, ruminate_minutes, anxiety_level, notes FROM ocdlog WHERE account_id = $1 AND id = $2 LIMIT 1;`
	getRowCountQuery   = `SELECT count(*) FROM ocdlog WHERE account_id = $1;`
)

func NewOCDLogRepository(db *sql.DB) *OCDLogRepository {
	return &OCDLogRepository{
		DB: db,
	}
}

func (repo *OCDLogRepository) GetAllLogs(ctx context.Context, accountID string, limit, offset int) (*entity.OCDLogList, error) {
	ocdLogList := entity.OCDLogList{
		Logs: make([]entity.OCDLog, 0),
	}
	var rowCount int
	err := sqlscan.Get(ctx, repo.DB, &rowCount, getRowCountQuery, accountID)
	if err != nil {
		return nil, err
	}
	paginationDetails := entity.PaginationDetails{
		Limit:  limit,
		Offset: offset,
		Total:  rowCount,
	}
	var ocdLogs []entity.OCDLog
	err = sqlscan.Select(ctx, repo.DB, &ocdLogs, getAllLogsQuery, accountID, limit, offset)
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
	err := logExec(ctx, repo.DB, deleteAllLogsQuery, "delete", accountID)
	if err != nil {
		return err
	}
	return nil
}

func (repo *OCDLogRepository) GetLog(ctx context.Context, accountID string, id uuid.UUID) (*entity.OCDLog, error) {
	ocdLog := entity.OCDLog{}
	err := sqlscan.Get(ctx, repo.DB, &ocdLog, getLogQuery, accountID, id)
	if err != nil {
		return nil, err
	}
	return &ocdLog, nil
}

func (repo *OCDLogRepository) CreateLog(ctx context.Context, accountID string, ocdLog *entity.OCDLog) error {
	pgElems, err := buildCreateQuery(ocdLog, accountID)
	if err != nil {
		return err
	}
	err = logExec(ctx, repo.DB, pgElems.query, "create", pgElems.fieldValues...)
	if err != nil {
		return err
	}
	return nil
}

func (repo *OCDLogRepository) UpdateLog(ctx context.Context, accountID string, id uuid.UUID, ocdLog *entity.OCDLog) error {
	pgElems, err := buildUpdateQuery(ocdLog, accountID, &id)
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

func (repo *OCDLogRepository) DeleteLog(ctx context.Context, accountID string, id uuid.UUID) error {
	err := logExec(ctx, repo.DB, deleteLogQuery, "delete", accountID, id)
	if err != nil {
		return err
	}
	return nil
}
