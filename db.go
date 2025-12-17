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

func InitializeDatabase(exPath, dataDirFlag string) (*sqlx.DB, error) {
	config := getDatabaseConfig(exPath, dataDirFlag)
	return initializeSQLite(config)
}

func getDatabaseConfig(exPath, dataDirFlag string) DatabaseConfig {
	// Use datadir flag if provided, otherwise fall back to executable directory
	dataPath := exPath
	if dataDirFlag != "" {
		dataPath = dataDirFlag
	}

	return DatabaseConfig{
		Path: filepath.Join(dataPath, "dbdata"),
	}
}

func initializeSQLite(config DatabaseConfig) (*sqlx.DB, error) {
	if err := os.MkdirAll(config.Path, 0751); err != nil {
		return nil, fmt.Errorf("could not create dbdata directory: %w", err)
	}

	dbPath := filepath.Join(config.Path, "users.db")
	db, err := sqlx.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_busy_timeout=3000")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	return db, nil
}
