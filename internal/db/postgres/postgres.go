package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/cecobask/ocd-tracker-api/internal/log"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

	"os"
	"time"
)

// Database ...
type Database struct {
	Connection *pgx.Conn
}

// Credentials ...
type Credentials struct {
	User     string
	Password string
}

// ConnectionConfig ...
type ConnectionConfig struct {
	DBName           string
	Port             string
	RetryMaxAttempts int
	RetryDelay       time.Duration
}

// ConnectWithConfig connects to a postgres database with the specified configuration
func ConnectWithConfig(ctx context.Context, credentials Credentials, connectionConfig ConnectionConfig) (*Database, error) {
	logger := log.LoggerFromContext(ctx)
	connString := fmt.Sprintf("host=postgres port=%s user=%s password=%s dbname=%s sslmode=disable",
		connectionConfig.Port, credentials.User, credentials.Password, connectionConfig.DBName,
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
		return &Database{
			Connection: postgresConn,
		}, nil
	}
}

// Connect connects to a postgres database
func Connect(ctx context.Context) (*Database, error) {
	credentials := Credentials{
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
	}
	connectionConfig := ConnectionConfig{
		DBName:           os.Getenv("POSTGRES_DB"),
		Port:             os.Getenv("POSTGRES_PORT"),
		RetryMaxAttempts: 10,
		RetryDelay:       time.Second * 5,
	}
	return ConnectWithConfig(ctx, credentials, connectionConfig)
}
