package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cecobask/ocd-tracker-api/pkg/entity"
	"github.com/cecobask/ocd-tracker-api/pkg/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"reflect"
	"strings"

	"os"
	"time"
)

type Credentials struct {
	User     string
	Password string
}

type ConnectionConfig struct {
	Host             string
	DBName           string
	Port             string
	RetryMaxAttempts int
	RetryDelay       time.Duration
}

type entityType string

const (
	entityTypeAccount entityType = "account"
	entityTypeOCDLog  entityType = "ocdlog"
)

// ConnectWithConfig connects to a postgres database with custom config
func ConnectWithConfig(ctx context.Context, credentials Credentials, connectionConfig ConnectionConfig) (*pgx.Conn, error) {
	logger := log.LoggerFromContext(ctx)
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		connectionConfig.Host, connectionConfig.Port, credentials.User, credentials.Password, connectionConfig.DBName,
	)
	attempts := 0
	for {
		attempts++
		if attempts == connectionConfig.RetryMaxAttempts {
			return nil, errors.New("reached max postgres connection retry attempts")
		}
		postgresConn, err := pgx.Connect(ctx, connString)
		if err != nil {
			logger.Debug("failed attempt to establish postgres connection", zap.Int("attempts", attempts), zap.Error(err))
			time.Sleep(connectionConfig.RetryDelay)
			continue
		}
		logger.Info("established postgres connection")
		return postgresConn, nil
	}
}

// Connect connects to a postgres database with default config
func Connect(ctx context.Context) (*pgx.Conn, error) {
	credentials := Credentials{
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
	}
	connectionConfig := ConnectionConfig{
		Host:             os.Getenv("POSTGRES_HOST"),
		DBName:           os.Getenv("POSTGRES_DB"),
		Port:             os.Getenv("POSTGRES_PORT"),
		RetryMaxAttempts: 10,
		RetryDelay:       time.Second * 5,
	}
	return ConnectWithConfig(ctx, credentials, connectionConfig)
}

func logExec(ctx context.Context, conn *pgx.Conn, query, action string, args ...interface{}) error {
	commandTag, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	log.LoggerFromContext(ctx).Info(fmt.Sprintf("%sd %d record/s", action, commandTag.RowsAffected()))
	return nil
}

func buildCreateQuery(object interface{}, accountID string) (query *string, fieldValues []interface{}, err error) {
	var (
		fieldsAllowed []string
		fieldNames    []string
		fieldIndexes  []string
		jsonData      []byte
	)
	entityType, err := getEntityType(object)
	if err != nil {
		return nil, nil, err
	}
	switch entityType {
	case entityTypeAccount:
		fieldsAllowed = append(fieldsAllowed, "email", "display_name", "wake_time", "sleep_time", "notification_interval", "photo_url")
		fieldNames = append(fieldNames, "id")
		fieldIndexes = append(fieldIndexes, "$1")
		fieldValues = append(fieldValues, accountID)
		jsonData, err = json.Marshal(object.(*entity.Account))
	case entityTypeOCDLog:
		fieldsAllowed = append(fieldsAllowed, "ruminate_minutes", "anxiety_level", "notes")
		fieldNames = append(fieldNames, "account_id")
		fieldIndexes = append(fieldIndexes, "$1")
		fieldValues = append(fieldValues, accountID)
		jsonData, err = json.Marshal(object.(*entity.OCDLog))
	}
	fieldCreates := make(map[string]interface{})
	err = json.Unmarshal(jsonData, &fieldCreates)
	if err != nil {
		return nil, nil, err
	}
	index := len(fieldNames) + 1 // start at x because of reserved field/s
	for _, allowedField := range fieldsAllowed {
		for fieldName, fieldValue := range fieldCreates {
			if fieldName == allowedField {
				fieldNames = append(fieldNames, fieldName)
				fieldIndexes = append(fieldIndexes, fmt.Sprintf("$%d", index))
				fieldValues = append(fieldValues, fieldValue)
				index++
			}
		}
	}
	fieldNamesStr := strings.TrimSuffix(strings.Join(fieldNames, ", "), ",")
	fieldIndexesStr := strings.TrimSuffix(strings.Join(fieldIndexes, ", "), ",")
	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", entityType, fieldNamesStr, fieldIndexesStr)
	return &q, fieldValues, nil
}

func buildUpdateQuery(object interface{}, accountID string, logID *uuid.UUID) (query *string, fieldValues []interface{}, err error) {
	var (
		fieldsAllowed []string
		fields        []string
		whereClause   string
		jsonData      []byte
	)
	entityType, err := getEntityType(object)
	if err != nil {
		return nil, nil, err
	}
	switch entityType {
	case entityTypeAccount:
		fieldsAllowed = append(fieldsAllowed, "email", "display_name", "wake_time", "sleep_time", "notification_interval", "photo_url")
		fieldValues = append(fieldValues, accountID)
		whereClause = "id = $1"
		jsonData, err = json.Marshal(object.(*entity.Account))
	case entityTypeOCDLog:
		fieldsAllowed = append(fieldsAllowed, "ruminate_minutes", "anxiety_level", "notes")
		fieldValues = append(fieldValues, accountID, logID)
		whereClause = "account_id = $1 AND id = $2"
		jsonData, err = json.Marshal(object.(*entity.OCDLog))
	}
	fieldUpdates := make(map[string]interface{})
	err = json.Unmarshal(jsonData, &fieldUpdates)
	if err != nil {
		return nil, nil, err
	}
	index := len(fieldValues) + 1 // start at x because of reserved field/s
	for _, allowedField := range fieldsAllowed {
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
		fieldsStr := strings.Join(fields, " ")
		q := fmt.Sprintf("UPDATE %s SET %s WHERE %s;", entityType, fieldsStr, whereClause)
		return &q, fieldValues, nil
	}
	return nil, nil, nil
}

func getEntityType(object interface{}) (entityType, error) {
	entityTypeStr := reflect.TypeOf(object).String()
	packageName := strings.Split(entityTypeStr, ".")
	entityType := strings.ToLower(packageName[len(packageName)-1])
	switch entityType {
	case "account":
		return entityTypeAccount, nil
	case "ocdlog":
		return entityTypeOCDLog, nil
	default:
		return "", fmt.Errorf("unknown entity type %s", entityTypeStr)
	}
}
