package database

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func (db *DB) RunMigrations() error {
	if err := db.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrations, err := db.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	for _, migration := range migrations {
		if err := db.runMigration(migration); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration, err)
		}
	}

	return nil
}

func (db *DB) createMigrationsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT NOW()
	)`
	
	_, err := db.Exec(query)
	return err
}

func (db *DB) getMigrationFiles() ([]string, error) {
	var migrations []string
	
	err := filepath.WalkDir("migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !d.IsDir() && strings.HasSuffix(path, ".sql") {
			migrations = append(migrations, path)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	sort.Strings(migrations)
	return migrations, nil
}

func (db *DB) runMigration(migrationFile string) error {
	migrationName := filepath.Base(migrationFile)
	
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", migrationName).Scan(&exists)
	if err != nil {
		return err
	}
	
	if exists {
		fmt.Printf("Migration %s already applied, skipping\n", migrationName)
		return nil
	}
	
	content, err := fs.ReadFile(os.DirFS("."), migrationFile)
	if err != nil {
		return err
	}
	
	fmt.Printf("Running migration: %s\n", migrationName)
	
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	if _, err := tx.Exec(string(content)); err != nil {
		return err
	}
	
	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migrationName); err != nil {
		return err
	}
	
	return tx.Commit()
}
