package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	NotionAPIToken   string
	NotionAPIVersion string
	NotionTimeout    int
	LogLevel         string
	LogFormat        string
	MetricsAddr      string
}

// Load reads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	token := os.Getenv("NOTION_API_TOKEN")
	version := os.Getenv("NOTION_API_VERSION")
	timeoutStr := os.Getenv("NOTION_TIMEOUT")
	logLevel := os.Getenv("MCP_SERVER_LOG_LEVEL")
	logFormat := os.Getenv("MCP_SERVER_LOG_FORMAT")
	metricsAddr := os.Getenv("MCP_METRICS_ADDR")

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

	if logLevel == "" {
		logLevel = "INFO"
	}
	if logFormat == "" {
		logFormat = "json"
	}
	if metricsAddr == "" {
		metricsAddr = ":9090"
	}

	return &Config{
		NotionAPIToken:   token,
		NotionAPIVersion: version,
		NotionTimeout:    timeout,
		LogLevel:         strings.ToUpper(strings.TrimSpace(logLevel)),
		LogFormat:        strings.ToLower(strings.TrimSpace(logFormat)),
		MetricsAddr:      strings.TrimSpace(metricsAddr),
	}, nil
}
