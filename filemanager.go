package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// FileManager manages local file storage operations
type FileManager struct {
	mu       sync.RWMutex
	mediaDir string
	mediaURL string
}

// Global file manager instance
var fileManager *FileManager
var fileManagerOnce sync.Once

// GetFileManager returns the global FileManager instance (singleton)
func GetFileManager() *FileManager {
	fileManagerOnce.Do(func() {
		mediaDir := os.Getenv("WUZAPI_MEDIA_DIR")
		if mediaDir == "" {
			mediaDir = "./media"
		}

		mediaURL := os.Getenv("WUZAPI_MEDIA_URL")
		if mediaURL == "" {
			// Default to relative path, will be served by static handler
			mediaURL = "/media"
		}

		fileManager = &FileManager{
			mediaDir: mediaDir,
			mediaURL: strings.TrimRight(mediaURL, "/"),
		}

		// Ensure media directory exists
		if err := os.MkdirAll(mediaDir, 0755); err != nil {
			log.Error().Err(err).Str("path", mediaDir).Msg("Failed to create media directory")
		} else {
			log.Info().Str("path", mediaDir).Str("url", mediaURL).Msg("File manager initialized")
		}
	})
	return fileManager
}

// GetMediaDir returns the configured media directory
func (m *FileManager) GetMediaDir() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mediaDir
}

// GenerateFilePath generates file path based on message metadata
func (m *FileManager) GenerateFilePath(userID, contactJID, messageID string, mimeType string, isIncoming bool) string {
	// Determine direction
	direction := "outbox"
	if isIncoming {
		direction = "inbox"
	}

	// Clean contact JID for filesystem safety
	contactJID = strings.ReplaceAll(contactJID, "@", "_")
	contactJID = strings.ReplaceAll(contactJID, ":", "_")

	// Get current time
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	// Determine media type folder
	mediaType := "documents"
	if strings.HasPrefix(mimeType, "image/") {
		mediaType = "images"
	} else if strings.HasPrefix(mimeType, "video/") {
		mediaType = "videos"
	} else if strings.HasPrefix(mimeType, "audio/") {
		mediaType = "audio"
	}

	// Get file extension
	ext := ".bin"
	switch {
	case strings.Contains(mimeType, "jpeg"), strings.Contains(mimeType, "jpg"):
		ext = ".jpg"
	case strings.Contains(mimeType, "png"):
		ext = ".png"
	case strings.Contains(mimeType, "gif"):
		ext = ".gif"
	case strings.Contains(mimeType, "webp"):
		ext = ".webp"
	case strings.Contains(mimeType, "mp4"):
		ext = ".mp4"
	case strings.Contains(mimeType, "webm"):
		ext = ".webm"
	case strings.Contains(mimeType, "ogg"):
		ext = ".ogg"
	case strings.Contains(mimeType, "opus"):
		ext = ".opus"
	case strings.Contains(mimeType, "pdf"):
		ext = ".pdf"
	case strings.Contains(mimeType, "doc"):
		if strings.Contains(mimeType, "docx") {
			ext = ".docx"
		} else {
			ext = ".doc"
		}
	}

	// Build relative path (same structure as S3 keys)
	relativePath := fmt.Sprintf("users/%s/%s/%s/%s/%s/%s/%s/%s%s",
		userID,
		direction,
		contactJID,
		year,
		month,
		day,
		mediaType,
		messageID,
		ext,
	)

	return relativePath
}

// SaveFile saves file to disk and returns metadata
func (m *FileManager) SaveFile(userID, contactJID, messageID string,
	data []byte, mimeType string, fileName string, isIncoming bool) (map[string]interface{}, error) {

	// Generate relative path
	relativePath := m.GenerateFilePath(userID, contactJID, messageID, mimeType, isIncoming)

	// Build full filesystem path
	fullPath := filepath.Join(m.mediaDir, relativePath)

	// Create directory structure
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	// Generate public URL
	publicURL := fmt.Sprintf("%s/%s", m.mediaURL, relativePath)

	// Return file metadata (compatible with previous S3 response format)
	fileData := map[string]interface{}{
		"url":      publicURL,
		"path":     relativePath,
		"size":     len(data),
		"mimeType": mimeType,
		"fileName": fileName,
	}

	log.Debug().
		Str("userID", userID).
		Str("path", relativePath).
		Int("size", len(data)).
		Msg("File saved to disk")

	return fileData, nil
}

// DeleteUserFiles deletes all files for a user
func (m *FileManager) DeleteUserFiles(userID string) error {
	userDir := filepath.Join(m.mediaDir, "users", userID)

	// Check if directory exists
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		log.Debug().Str("userID", userID).Msg("No files to delete for user")
		return nil
	}

	// Remove entire user directory
	if err := os.RemoveAll(userDir); err != nil {
		return fmt.Errorf("failed to delete user files: %w", err)
	}

	log.Info().Str("userID", userID).Msg("All user files deleted from disk")
	return nil
}
