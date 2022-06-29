package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"log"
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
	connString := fmt.Sprintf("host=postgres port=%s user=%s password=%s dbname=%s sslmode=disable",
		connectionConfig.Port, credentials.User, credentials.Password, connectionConfig.DBName,
	)
	attempts := 0
	for {
		attempts++
		if attempts == connectionConfig.RetryMaxAttempts {
			return nil, errors.New("postgres connection max retries attempts reached")
		}
		postgresConn, err := pgx.Connect(ctx, connString)
		if err != nil {
			fmt.Printf("could not connect to postgres: attempts=%d, error=%v\n", attempts, err)
			time.Sleep(connectionConfig.RetryDelay)
			continue
		}
		log.Println("postgres connection established")
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
