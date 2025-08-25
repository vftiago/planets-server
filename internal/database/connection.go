package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"planets-server/internal/utils"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func Connect() (*DB, error) {
	logger := slog.With("component", "database", "operation", "connect")
	logger.Debug("Initializing database connection")

	// Get database configuration
	host := utils.GetEnv("DB_HOST", "localhost")
	port := utils.GetEnv("DB_PORT", "5432")
	user := utils.GetEnv("DB_USER", "postgres")
	password := utils.GetEnv("DB_PASSWORD", "postgres")
	dbname := utils.GetEnv("DB_NAME", "planets")
	sslmode := utils.GetEnv("DB_SSLMODE", "disable")

	// Connection pool settings
	maxOpenConns, _ := strconv.Atoi(utils.GetEnv("DB_MAX_OPEN_CONNS", "25"))
	maxIdleConns, _ := strconv.Atoi(utils.GetEnv("DB_MAX_IDLE_CONNS", "5"))
	connMaxLifetime, _ := strconv.Atoi(utils.GetEnv("DB_CONN_MAX_LIFETIME_MINUTES", "5"))

	logger.Info("Connecting to database",
		"host", host,
		"port", port,
		"user", user,
		"database", dbname,
		"sslmode", sslmode,
		"max_open_conns", maxOpenConns,
		"max_idle_conns", maxIdleConns,
	)

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("Failed to open database connection", 
			"error", err, "host", host, "database", dbname)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Minute)

	// Test the connection
	logger.Debug("Testing database connection with ping")
	if err := sqlDB.Ping(); err != nil {
		logger.Error("Failed to ping database", 
			"error", err, "host", host, "database", dbname)
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established successfully",
		"host", host, "database", dbname)

	return &DB{sqlDB}, nil
}
