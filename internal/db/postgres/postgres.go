package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cecobask/ocdtracker-api/internal/aws"
	"github.com/cecobask/ocdtracker-api/pkg/entity"
	"github.com/cecobask/ocdtracker-api/pkg/log"
	"github.com/golang-migrate/migrate/v4"
	pgMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"reflect"
	"strconv"
	"strings"

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

type postgresElements struct {
	query       string
	fieldValues []interface{}
}

type entityType string

const (
	entityTypeAccount entityType = "account"
	entityTypeOCDLog  entityType = "ocdlog"
)

// ConnectWithConfig connects to a database with custom config
func ConnectWithConfig(ctx context.Context, credentials Credentials, connectionConfig ConnectionConfig) (*sql.DB, error) {
	logger := log.LoggerFromContext(ctx)
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		connectionConfig.Host, connectionConfig.Port, credentials.User, credentials.Password, connectionConfig.DBName,
	)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare database: %w", err)
	}
	attempts := 0
	for {
		attempts++
		if attempts == connectionConfig.RetryMaxAttempts {
			return nil, fmt.Errorf("reached max database connection retry attempts")
		}
		err := db.Ping()
		if err != nil {
			logger.Warn("failed attempt to establish database connection", zap.Int("attempts", attempts), zap.Error(err))
			time.Sleep(connectionConfig.RetryDelay)
			continue
		}
		logger.Info("established database connection")
		return db, nil
	}
}

// Connect connects to a database with default config
func Connect(ctx context.Context, postgresCredsSecret *aws.PostgresCredsSecret) (*sql.DB, error) {
	credentials := Credentials{
		User:     postgresCredsSecret.Username,
		Password: postgresCredsSecret.Password,
	}
	connectionConfig := ConnectionConfig{
		Host:             postgresCredsSecret.Host,
		DBName:           postgresCredsSecret.DBName,
		Port:             strconv.Itoa(postgresCredsSecret.Port),
		RetryMaxAttempts: 10,
		RetryDelay:       time.Second * 5,
	}
	return ConnectWithConfig(ctx, credentials, connectionConfig)
}

func Migrate(db *sql.DB) error {
	driver, err := pgMigrate.WithInstance(db, &pgMigrate.Config{})
	if err != nil {
		return fmt.Errorf("failed to link database and migrator: %w", err)
	}
	instance, err := migrate.NewWithDatabaseInstance("file:///migration", "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to initialise database migrator: %w", err)
	}
	err = instance.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	return fmt.Errorf("failed to migrate database to latest version: %w", err)

}

func logExec(ctx context.Context, db *sql.DB, query, action string, args ...interface{}) error {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	log.LoggerFromContext(ctx).Info(fmt.Sprintf("%sd %d record/s", action, rowsAffected))
	return nil
}

func buildCreateQuery(object interface{}, accountID string) (*postgresElements, error) {
	var (
		fieldsAllowed []string
		fieldNames    []string
		jsonData      []byte
	)
	entityType, err := getEntityType(object)
	if err != nil {
		return nil, err
	}
	switch entityType {
	case entityTypeAccount:
		fieldsAllowed = append(fieldsAllowed, "email", "display_name", "wake_time", "sleep_time", "notification_interval", "photo_url")
		fieldNames = append(fieldNames, "id")
		jsonData, err = json.Marshal(object.(*entity.Account))
	case entityTypeOCDLog:
		fieldsAllowed = append(fieldsAllowed, "ruminate_minutes", "anxiety_level", "notes")
		fieldNames = append(fieldNames, "account_id")
		jsonData, err = json.Marshal(object.(*entity.OCDLog))
	}
	fieldValues := []interface{}{accountID}
	fieldIndexes := []string{"$1"}
	fieldCreates := make(map[string]interface{})
	err = json.Unmarshal(jsonData, &fieldCreates)
	if err != nil {
		return nil, err
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
	return &postgresElements{query: q, fieldValues: fieldValues}, nil
}

func buildUpdateQuery(object interface{}, accountID string, logID *uuid.UUID) (*postgresElements, error) {
	var (
		fieldsAllowed []string
		fields        []string
		fieldValues   []interface{}
		whereClause   string
		jsonData      []byte
	)
	entityType, err := getEntityType(object)
	if err != nil {
		return nil, err
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
		return nil, err
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
		return &postgresElements{query: q, fieldValues: fieldValues}, nil
	}
	return nil, nil // no action
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
