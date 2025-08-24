package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"planets-server/internal/utils"

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

	// Log connection details (without password)
	logger.Info("Connecting to database",
		"host", host,
		"port", port,
		"user", user,
		"database", dbname,
		"sslmode", sslmode,
	)

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Open database connection
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("Failed to open database connection", 
			"error", err,
			"host", host,
			"database", dbname,
		)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	logger.Debug("Testing database connection with ping")
	if err := sqlDB.Ping(); err != nil {
		logger.Error("Failed to ping database", 
			"error", err,
			"host", host,
			"database", dbname,
		)
		sqlDB.Close() // Clean up the connection
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established successfully",
		"host", host,
		"database", dbname,
	)

	return &DB{sqlDB}, nil
}
