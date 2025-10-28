package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"planets-server/internal/shared/config"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type Tx struct {
	*sql.Tx
}

type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func (db *DB) BeginTx() (*Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &Tx{tx}, nil
}

func (db *DB) BeginTxContext(ctx context.Context) (*Tx, error) {
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &Tx{tx}, nil
}

func Connect() (*DB, error) {
	cfg := config.GlobalConfig
	logger := slog.With("component", "database", "operation", "connect")
	logger.Debug("Initializing database connection")

	logger.Info("Connecting to database",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"user", cfg.Database.User,
		"database", cfg.Database.Name,
		"sslmode", cfg.Database.SSLMode,
		"max_open_conns", cfg.Database.MaxOpenConns,
		"max_idle_conns", cfg.Database.MaxIdleConns,
	)

	sqlDB, err := sql.Open("postgres", cfg.ConnectionString())
	if err != nil {
		logger.Error("Failed to open database connection",
			"error", err, "host", cfg.Database.Host, "database", cfg.Database.Name)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	logger.Debug("Testing database connection with ping")
	if err := sqlDB.Ping(); err != nil {
		logger.Error("Failed to ping database",
			"error", err, "host", cfg.Database.Host, "database", cfg.Database.Name)
		if closeErr := sqlDB.Close(); closeErr != nil {
			logger.Error("Failed to close database after ping failure", "close_error", closeErr, "ping_error", err)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established successfully",
		"host", cfg.Database.Host, "database", cfg.Database.Name)

	return &DB{sqlDB}, nil
}
