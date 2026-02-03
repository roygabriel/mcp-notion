package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	NotionAPIToken   string
	NotionAPIVersion string
	NotionTimeout    int
}

// Load reads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	token := os.Getenv("NOTION_API_TOKEN")
	version := os.Getenv("NOTION_API_VERSION")
	timeoutStr := os.Getenv("NOTION_TIMEOUT")

	// Validate required fields
	if token == "" {
		return nil, fmt.Errorf("NOTION_API_TOKEN environment variable is required (create an integration at https://www.notion.so/my-integrations)")
	}

	// Set defaults
	if version == "" {
		version = "2022-06-28"
	}

	timeout := 30 // default timeout in seconds
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
			timeout = t
		}
	}

	return &Config{
		NotionAPIToken:   token,
		NotionAPIVersion: version,
		NotionTimeout:    timeout,
	}, nil
}
