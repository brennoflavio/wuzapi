package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func GetConfigPath(appName string) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("appName cannot be empty")
	}

	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		xdgConfig = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(xdgConfig, appName), nil
}

func GetCachePath(appName string) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("appName cannot be empty")
	}

	xdgCache := os.Getenv("XDG_CACHE_HOME")
	if xdgCache == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		xdgCache = filepath.Join(homeDir, ".cache")
	}

	return filepath.Join(xdgCache, appName), nil
}

func GetAppDataPath() (string, error) {
	path := os.Getenv("APP_DIR")
	if path == "" {
		return "", fmt.Errorf("could not find path")
	}
	return path, nil
}

func OutputJSON(data interface{}) {
	jsonOutput, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonOutput))
}
