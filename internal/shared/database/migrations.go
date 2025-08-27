package database

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func (db *DB) RunMigrations() error {
	logger := slog.With("component", "migrations")
	logger.Info("Starting database migrations")

	if err := db.createMigrationsTable(); err != nil {
		logger.Error("Failed to create migrations table", "error", err)
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrations, err := db.getMigrationFiles()
	if err != nil {
		logger.Error("Failed to get migration files", "error", err)
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	logger.Info("Found migration files", "count", len(migrations))

	for _, migration := range migrations {
		if err := db.runMigration(migration); err != nil {
			logger.Error("Failed to run migration", "migration", migration, "error", err)
			return fmt.Errorf("failed to run migration %s: %w", migration, err)
		}
	}

	logger.Info("All migrations completed successfully")
	return nil
}

func (db *DB) createMigrationsTable() error {
	logger := slog.With("component", "migrations", "operation", "create_table")
	logger.Debug("Creating schema_migrations table if not exists")

	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT NOW()
	)`

	_, err := db.Exec(query)
	if err != nil {
		logger.Error("Failed to create schema_migrations table", "error", err)
	} else {
		logger.Debug("schema_migrations table ready")
	}
	return err
}

func (db *DB) getMigrationFiles() ([]string, error) {
	logger := slog.With("component", "migrations", "operation", "scan_files")
	logger.Debug("Scanning for migration files in migrations/ directory")

	var migrations []string

	err := filepath.WalkDir("migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logger.Warn("Error accessing migration file", "path", path, "error", err)
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".sql") {
			migrations = append(migrations, path)
			logger.Debug("Found migration file", "file", path)
		}

		return nil
	})

	if err != nil {
		logger.Error("Failed to scan migration directory", "error", err)
		return nil, err
	}

	sort.Strings(migrations)
	logger.Debug("Migration files collected", "count", len(migrations), "files", migrations)
	return migrations, nil
}

func (db *DB) runMigration(migrationFile string) error {
	migrationName := filepath.Base(migrationFile)
	logger := slog.With(
		"component", "migrations",
		"operation", "run_migration",
		"migration", migrationName,
	)

	// Check if migration already applied
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", migrationName).Scan(&exists)
	if err != nil {
		logger.Error("Failed to check migration status", "error", err)
		return err
	}

	if exists {
		logger.Debug("Migration already applied, skipping")
		return nil
	}

	// Read migration file
	content, err := fs.ReadFile(os.DirFS("."), migrationFile)
	if err != nil {
		logger.Error("Failed to read migration file", "error", err)
		return err
	}

	logger.Info("Running migration", "size_bytes", len(content))

	// Execute migration in transaction
	tx, err := db.Begin()
	if err != nil {
		logger.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err.Error() != "sql: transaction has already been committed or rolled back" {
			logger.Error("Failed to rollback transaction", "error", err)
		}
	}()

	// Execute migration SQL
	if _, err := tx.Exec(string(content)); err != nil {
		logger.Error("Failed to execute migration SQL", "error", err)
		return err
	}

	// Record migration as applied
	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migrationName); err != nil {
		logger.Error("Failed to record migration", "error", err)
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit migration transaction", "error", err)
		return err
	}

	logger.Info("Migration completed successfully")
	return nil
}
