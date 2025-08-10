package database

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"smat/iot/simulation/iot-inventory-management/internal/config"
)

func NewPostgresDB(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// RunMigrations runs all pending migrations
func RunMigrations(db *sql.DB, migrationsDir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	log.Printf("Running migrations from directory: %s", absPath)

	if err := goose.Up(db, absPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	newVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get version after migration: %w", err)
	}

	log.Printf("Migrations completed successfully. Current version: %d", newVersion)
	return nil
}
