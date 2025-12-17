package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Migration struct {
	ID      int
	Name    string
	UpSQL   string
	DownSQL string
}

var migrations = []Migration{
	{
		ID:   1,
		Name: "initial_schema",
	},
}

// GenerateRandomID creates a random string ID
func GenerateRandomID() (string, error) {
	bytes := make([]byte, 16) // 128 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// Initialize the database with migrations
func initializeSchema(db *sqlx.DB) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get already applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply missing migrations
	for _, migration := range migrations {
		if _, ok := applied[migration.ID]; !ok {
			if err := applyMigration(db, migration); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.ID, err)
			}
		}
	}

	return nil
}

func createMigrationsTable(db *sqlx.DB) error {
	var tableExists bool

	err := db.Get(&tableExists, `
		SELECT EXISTS (
			SELECT 1 FROM sqlite_master
			WHERE type='table' AND name='migrations'
		)`)

	if err != nil {
		return fmt.Errorf("failed to check migrations table existence: %w", err)
	}

	if tableExists {
		return nil
	}

	_, err = db.Exec(`
		CREATE TABLE migrations (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

func getAppliedMigrations(db *sqlx.DB) (map[int]struct{}, error) {
	applied := make(map[int]struct{})
	var rows []struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	err := db.Select(&rows, "SELECT id, name FROM migrations ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}

	for _, row := range rows {
		applied[row.ID] = struct{}{}
	}

	return applied, nil
}

func applyMigration(db *sqlx.DB, migration Migration) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	switch migration.ID {
	case 1:
		// Create users table
		// TODO: remove days_to_sync_history, always sync whole history
		err = createTableIfNotExistsSQLite(tx, "users", `
			CREATE TABLE users (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				jid TEXT NOT NULL DEFAULT '',
				qrcode TEXT NOT NULL DEFAULT '',
				connected INTEGER,
				history INTEGER DEFAULT 0,
				days_to_sync_history INTEGER DEFAULT 0
			)`)
		if err != nil {
			return fmt.Errorf("failed to create users table: %w", err)
		}

		// Create message_history table
		err = createTableIfNotExistsSQLite(tx, "message_history", `
			CREATE TABLE message_history (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id TEXT NOT NULL,
				chat_jid TEXT NOT NULL,
				sender_jid TEXT NOT NULL,
				message_id TEXT NOT NULL,
				timestamp DATETIME NOT NULL,
				message_type TEXT NOT NULL,
				text_content TEXT,
				media_link TEXT,
				quoted_message_id TEXT,
				datajson TEXT,
				UNIQUE(user_id, message_id)
			)`)
		if err != nil {
			return fmt.Errorf("failed to create message_history table: %w", err)
		}

		// Create message_history index
		_, err = tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_message_history_user_chat_timestamp
			ON message_history (user_id, chat_jid, timestamp DESC)`)
		if err != nil {
			return fmt.Errorf("failed to create message_history index: %w", err)
		}

		// Create events table
		err = createTableIfNotExistsSQLite(tx, "events", `
			CREATE TABLE events (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id TEXT NOT NULL,
				event_type TEXT NOT NULL,
				payload TEXT NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`)
		if err != nil {
			return fmt.Errorf("failed to create events table: %w", err)
		}

		// Create events indexes
		_, err = tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_events_user_created
			ON events (user_id, created_at DESC)`)
		if err != nil {
			return fmt.Errorf("failed to create events user index: %w", err)
		}

		_, err = tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_events_type
			ON events (event_type)`)
		if err != nil {
			return fmt.Errorf("failed to create events type index: %w", err)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record the migration
	if _, err = tx.Exec(`INSERT INTO migrations (id, name) VALUES (?, ?)`, migration.ID, migration.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

func createTableIfNotExistsSQLite(tx *sqlx.Tx, tableName, createSQL string) error {
	var exists int
	err := tx.Get(&exists, `
        SELECT COUNT(*) FROM sqlite_master
        WHERE type='table' AND name=?`, tableName)
	if err != nil {
		return err
	}

	if exists == 0 {
		_, err = tx.Exec(createSQL)
		return err
	}
	return nil
}
