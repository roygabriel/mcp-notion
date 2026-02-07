package config

import "testing"

func TestLoad_MissingToken(t *testing.T) {
	t.Setenv("NOTION_API_TOKEN", "")
	t.Setenv("NOTION_API_VERSION", "")
	t.Setenv("NOTION_TIMEOUT", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when NOTION_API_TOKEN is missing")
	}
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("NOTION_API_TOKEN", "secret_test_token")
	t.Setenv("NOTION_API_VERSION", "")
	t.Setenv("NOTION_TIMEOUT", "")

	// Prevent .env file from interfering
	t.Setenv("NOTION_API_VERSION", "")
	t.Setenv("NOTION_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.NotionAPIToken != "secret_test_token" {
		t.Errorf("expected token 'secret_test_token', got %q", cfg.NotionAPIToken)
	}
	if cfg.NotionAPIVersion != "2022-06-28" {
		t.Errorf("expected default version '2022-06-28', got %q", cfg.NotionAPIVersion)
	}
	if cfg.NotionTimeout != 30 {
		t.Errorf("expected default timeout 30, got %d", cfg.NotionTimeout)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("NOTION_API_TOKEN", "secret_abc")
	t.Setenv("NOTION_API_VERSION", "2023-08-01")
	t.Setenv("NOTION_TIMEOUT", "60")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.NotionAPIVersion != "2023-08-01" {
		t.Errorf("expected version '2023-08-01', got %q", cfg.NotionAPIVersion)
	}
	if cfg.NotionTimeout != 60 {
		t.Errorf("expected timeout 60, got %d", cfg.NotionTimeout)
	}
}

func TestLoad_InvalidTimeout(t *testing.T) {
	t.Setenv("NOTION_API_TOKEN", "secret_abc")
	t.Setenv("NOTION_TIMEOUT", "notanumber")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Invalid timeout should fall back to default
	if cfg.NotionTimeout != 30 {
		t.Errorf("expected default timeout 30 for invalid input, got %d", cfg.NotionTimeout)
	}
}

func TestLoad_NegativeTimeout(t *testing.T) {
	t.Setenv("NOTION_API_TOKEN", "secret_abc")
	t.Setenv("NOTION_TIMEOUT", "-5")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.NotionTimeout != 30 {
		t.Errorf("expected default timeout 30 for negative input, got %d", cfg.NotionTimeout)
	}
}
