package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cecobask/ocd-tracker-api/internal/entity"
	"github.com/cecobask/ocd-tracker-api/internal/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"strings"
)

const (
	createLogQuery     = `INSERT INTO ocdlog (account_id, ruminate_minutes, anxiety_level, notes) VALUES ($1, $2, $3, $4);`
	deleteAllLogsQuery = `DELETE FROM ocdlog WHERE account_id = $1;`
	deleteLogQuery     = `DELETE FROM ocdlog WHERE account_id = $1 AND id = $2;`
	getAllLogsQuery    = `SELECT id, account_id, created_at, updated_at, ruminate_minutes, anxiety_level, notes FROM ocdlog WHERE account_id = $1 ORDER BY created_at LIMIT $2 OFFSET $3;`
	getLogQuery        = `SELECT id, account_id, created_at, updated_at, ruminate_minutes, anxiety_level, notes FROM ocdlog WHERE account_id = $1 AND id = $2 LIMIT 1;`
	getRowCountQuery   = `SELECT count(*) FROM ocdlog WHERE account_id = $1;`
)

func (pg *Client) GetAllLogs(ctx context.Context, accountID string, limit, offset int) (*entity.OCDLogList, error) {
	ocdLogList := entity.OCDLogList{
		Logs: make([]entity.OCDLog, 0),
	}
	var rowCount int
	err := pg.Connection.QueryRow(ctx, getRowCountQuery, accountID).Scan(&rowCount)
	if err != nil {
		return nil, err
	}
	paginationDetails := entity.PaginationDetails{
		Limit:  limit,
		Offset: offset,
		Total:  rowCount,
	}
	rows, err := pg.Connection.Query(ctx, getAllLogsQuery, accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var ocdLog entity.OCDLog
		err = rows.Scan(&ocdLog.ID, &ocdLog.AccountID, &ocdLog.CreatedAt, &ocdLog.UpdatedAt, &ocdLog.RuminateMinutes, &ocdLog.AnxietyLevel, &ocdLog.Notes)
		if err != nil {
			return nil, err
		}
		ocdLogList.Logs = append(ocdLogList.Logs, ocdLog)
	}
	paginationDetails.Count = len(ocdLogList.Logs)
	ocdLogList.Pagination = paginationDetails
	log.LoggerFromContext(ctx).Info(fmt.Sprintf("retrieved %d logs", len(ocdLogList.Logs)))
	return &ocdLogList, nil
}

func (pg *Client) DeleteAllLogs(ctx context.Context, accountID string) error {
	err := pg.logExec(ctx, deleteAllLogsQuery, "delete", "log", accountID)
	if err != nil {
		return err
	}
	return nil
}

func (pg *Client) GetLog(ctx context.Context, accountID string, id uuid.UUID) (*entity.OCDLog, error) {
	ocdLog := entity.OCDLog{}
	row := pg.Connection.QueryRow(ctx, getLogQuery, accountID, id)
	err := row.Scan(&ocdLog.ID, &ocdLog.AccountID, &ocdLog.CreatedAt, &ocdLog.UpdatedAt, &ocdLog.RuminateMinutes, &ocdLog.AnxietyLevel, &ocdLog.Notes)
	if err != nil {
		return nil, err
	}
	return &ocdLog, nil
}

func (pg *Client) CreateLog(ctx context.Context, ocdLog entity.OCDLog) error {
	err := pg.logExec(ctx, createLogQuery, "create", "log", ocdLog.AccountID, ocdLog.RuminateMinutes, ocdLog.AnxietyLevel, ocdLog.Notes)
	if err != nil {
		return err
	}
	return nil
}

func (pg *Client) UpdateLog(ctx context.Context, accountID string, id uuid.UUID, ocdLog entity.OCDLog) error {
	var (
		allowedFields = []string{"ruminate_minutes", "anxiety_level", "notes"}
		fields        []string
		fieldValues   []interface{}
		fieldUpdates  map[string]interface{}
	)
	jsonData, err := json.Marshal(ocdLog)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonData, &fieldUpdates)
	if err != nil {
		return err
	}
	// start at 3 because 1 & 2 are reserved for account_id & id fields
	index := 3
	fieldValues = append(fieldValues, accountID, id)
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
		query := "UPDATE ocdlog SET " + strings.Join(fields, " ") + " WHERE account_id = $1 AND id = $2;"
		err = pg.logExec(ctx, query, "update", "log", fieldValues...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pg *Client) DeleteLog(ctx context.Context, accountID string, id uuid.UUID) error {
	err := pg.logExec(ctx, deleteLogQuery, "delete", "log", accountID, id)
	if err != nil {
		return err
	}
	return nil
}

func (pg *Client) LogExists(ctx context.Context, accountID string, id uuid.UUID) (bool, *entity.OCDLog, error) {
	ocdLog, err := pg.GetLog(ctx, accountID, id)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return false, nil, nil
		default:
			return false, nil, err
		}
	}
	return true, ocdLog, nil
}

func (pg *Client) logExec(ctx context.Context, query, action, entity string, args ...interface{}) error {
	commandTag, err := pg.Connection.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	log.LoggerFromContext(ctx).Info(fmt.Sprintf("%sd %d %s/s", action, commandTag.RowsAffected(), entity))
	return nil
}
