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
	{
		ID:   2,
		Name: "add_proxy_url",
	},
	{
		ID:   3,
		Name: "change_id_to_string",
	},
	{
		ID:   4,
		Name: "add_s3_support",
	},
	{
		ID:   5,
		Name: "add_message_history",
	},
	{
		ID:   6,
		Name: "add_quoted_message_id",
	},
	{
		ID:   7,
		Name: "add_hmac_key",
	},
	{
		ID:   8,
		Name: "add_data_json",
	},
	{
		ID:   9,
		Name: "add_events_table",
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
		err = createTableIfNotExistsSQLite(tx, "users", `
			CREATE TABLE users (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				token TEXT NOT NULL,
				webhook TEXT NOT NULL DEFAULT '',
				jid TEXT NOT NULL DEFAULT '',
				qrcode TEXT NOT NULL DEFAULT '',
				connected INTEGER,
				expiration INTEGER,
				events TEXT NOT NULL DEFAULT '',
				proxy_url TEXT DEFAULT ''
			)`)
	case 2:
		err = addColumnIfNotExistsSQLite(tx, "users", "proxy_url", "TEXT DEFAULT ''")
	case 3:
		err = migrateSQLiteIDToString(tx)
	case 4:
		// Handle S3 columns
		err = addColumnIfNotExistsSQLite(tx, "users", "s3_enabled", "BOOLEAN DEFAULT 0")
		if err == nil {
			err = addColumnIfNotExistsSQLite(tx, "users", "s3_endpoint", "TEXT DEFAULT ''")
		}
		if err == nil {
			err = addColumnIfNotExistsSQLite(tx, "users", "s3_region", "TEXT DEFAULT ''")
		}
		if err == nil {
			err = addColumnIfNotExistsSQLite(tx, "users", "s3_bucket", "TEXT DEFAULT ''")
		}
		if err == nil {
			err = addColumnIfNotExistsSQLite(tx, "users", "s3_access_key", "TEXT DEFAULT ''")
		}
		if err == nil {
			err = addColumnIfNotExistsSQLite(tx, "users", "s3_secret_key", "TEXT DEFAULT ''")
		}
		if err == nil {
			err = addColumnIfNotExistsSQLite(tx, "users", "s3_path_style", "BOOLEAN DEFAULT 1")
		}
		if err == nil {
			err = addColumnIfNotExistsSQLite(tx, "users", "s3_public_url", "TEXT DEFAULT ''")
		}
		if err == nil {
			err = addColumnIfNotExistsSQLite(tx, "users", "media_delivery", "TEXT DEFAULT 'base64'")
		}
		if err == nil {
			err = addColumnIfNotExistsSQLite(tx, "users", "s3_retention_days", "INTEGER DEFAULT 30")
		}
	case 5:
		// Handle message_history table creation
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
				UNIQUE(user_id, message_id)
			)`)
		if err == nil {
			// Create index
			_, err = tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_message_history_user_chat_timestamp
				ON message_history (user_id, chat_jid, timestamp DESC)`)
		}
		if err == nil {
			// Add history column to users table
			err = addColumnIfNotExistsSQLite(tx, "users", "history", "INTEGER DEFAULT 0")
		}
	case 6:
		// Add quoted_message_id column to message_history table
		err = addColumnIfNotExistsSQLite(tx, "message_history", "quoted_message_id", "TEXT")
	case 7:
		// Add hmac_key column for webhook signing (feature not implemented)
		err = addColumnIfNotExistsSQLite(tx, "users", "hmac_key", "BLOB")
	case 8:
		// Add dataJson column to message_history table
		err = addColumnIfNotExistsSQLite(tx, "message_history", "datajson", "TEXT")
	case 9:
		// Create events table for storing all WhatsApp events
		err = createTableIfNotExistsSQLite(tx, "events", `
			CREATE TABLE events (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id TEXT NOT NULL,
				event_type TEXT NOT NULL,
				payload TEXT NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`)
		if err == nil {
			// Create index on user_id and created_at for efficient queries
			_, err = tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_events_user_created
				ON events (user_id, created_at DESC)`)
		}
		if err == nil {
			// Create index on event_type for filtering by event type
			_, err = tx.Exec(`
				CREATE INDEX IF NOT EXISTS idx_events_type
				ON events (event_type)`)
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

func migrateSQLiteIDToString(tx *sqlx.Tx) error {
	// 1. Check if we need to do the migration
	var currentType string
	err := tx.QueryRow(`
        SELECT type FROM pragma_table_info('users')
        WHERE name = 'id'`).Scan(&currentType)
	if err != nil {
		return fmt.Errorf("failed to check column type: %w", err)
	}

	if currentType != "INTEGER" {
		// No migration needed
		return nil
	}

	// 2. Create new table with string ID
	_, err = tx.Exec(`
        CREATE TABLE users_new (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            token TEXT NOT NULL,
            webhook TEXT NOT NULL DEFAULT '',
            jid TEXT NOT NULL DEFAULT '',
            qrcode TEXT NOT NULL DEFAULT '',
            connected INTEGER,
            expiration INTEGER,
            events TEXT NOT NULL DEFAULT '',
            proxy_url TEXT DEFAULT ''
        )`)
	if err != nil {
		return fmt.Errorf("failed to create new table: %w", err)
	}

	// 3. Copy data with new UUIDs
	_, err = tx.Exec(`
        INSERT INTO users_new
        SELECT
            hex(randomblob(16)),
            name, token, webhook, jid, qrcode,
            connected, expiration, events, proxy_url
        FROM users`)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// 4. Drop old table
	_, err = tx.Exec(`DROP TABLE users`)
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// 5. Rename new table
	_, err = tx.Exec(`ALTER TABLE users_new RENAME TO users`)
	if err != nil {
		return fmt.Errorf("failed to rename table: %w", err)
	}

	return nil
}

func addColumnIfNotExistsSQLite(tx *sqlx.Tx, tableName, columnName, columnDef string) error {
	var exists int
	err := tx.Get(&exists, `
        SELECT COUNT(*) FROM pragma_table_info(?)
        WHERE name = ?`, tableName, columnName)
	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if exists == 0 {
		_, err = tx.Exec(fmt.Sprintf(
			"ALTER TABLE %s ADD COLUMN %s %s",
			tableName, columnName, columnDef))
		if err != nil {
			return fmt.Errorf("failed to add column: %w", err)
		}
	}
	return nil
}
