package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type DatabaseConfig struct {
	Path string
}

func InitializeDatabase(appName string) (*sqlx.DB, error) {
	config, err := getDatabaseConfig(appName)
	if err != nil {
		return nil, err
	}
	return initializeSQLite(config)
}

func getDatabaseConfig(appName string) (DatabaseConfig, error) {
	configPath, err := GetConfigPath(appName)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to get config path: %w", err)
	}

	return DatabaseConfig{
		Path: configPath,
	}, nil
}

func initializeSQLite(config DatabaseConfig) (*sqlx.DB, error) {
	if err := os.MkdirAll(config.Path, 0751); err != nil {
		return nil, fmt.Errorf("could not create data directory: %w", err)
	}

	dbPath := filepath.Join(config.Path, "store.db")
	db, err := sqlx.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_busy_timeout=3000")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	return db, nil
}
