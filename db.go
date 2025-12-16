package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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

type HistoryMessage struct {
	ID              int       `json:"id" db:"id"`
	UserID          string    `json:"user_id" db:"user_id"`
	ChatJID         string    `json:"chat_jid" db:"chat_jid"`
	SenderJID       string    `json:"sender_jid" db:"sender_jid"`
	MessageID       string    `json:"message_id" db:"message_id"`
	Timestamp       time.Time `json:"timestamp" db:"timestamp"`
	MessageType     string    `json:"message_type" db:"message_type"`
	TextContent     string    `json:"text_content" db:"text_content"`
	MediaLink       string    `json:"media_link" db:"media_link"`
	QuotedMessageID string    `json:"quoted_message_id,omitempty" db:"quoted_message_id"`
	DataJson        string    `json:"data_json" db:"datajson"`
}

func (s *server) saveMessageToHistory(userID, chatJID, senderJID, messageID, messageType, textContent, mediaLink, quotedMessageID, dataJson string) error {
	query := `INSERT INTO message_history (user_id, chat_jid, sender_jid, message_id, timestamp, message_type, text_content, media_link, quoted_message_id, datajson)
                 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, userID, chatJID, senderJID, messageID, time.Now(), messageType, textContent, mediaLink, quotedMessageID, dataJson)
	if err != nil {
		return fmt.Errorf("failed to save message to history: %w", err)
	}
	return nil
}

func (s *server) trimMessageHistory(userID, chatJID string, limit int) error {
	queryHistory := `
		DELETE FROM message_history
		WHERE id IN (
			SELECT id FROM message_history
			WHERE user_id = ? AND chat_jid = ?
			ORDER BY timestamp DESC
			LIMIT -1 OFFSET ?
		)`

	querySecrets := `
		DELETE FROM whatsmeow_message_secrets
		WHERE message_id IN (
			SELECT id FROM message_history
			WHERE user_id = ? AND chat_jid = ?
			ORDER BY timestamp DESC
			LIMIT -1 OFFSET ?
		)`

	if _, err := s.db.Exec(querySecrets, userID, chatJID, limit); err != nil {
		return fmt.Errorf("failed to trim message secrets: %w", err)
	}

	if _, err := s.db.Exec(queryHistory, userID, chatJID, limit); err != nil {
		return fmt.Errorf("failed to trim message history: %w", err)
	}

	return nil
}
