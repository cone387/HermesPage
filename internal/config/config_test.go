package config

import "testing"

func TestServerURLFromEnv(t *testing.T) {
	t.Setenv("HERMES_SERVER_URL", "https://reports.example.com")
	t.Setenv("HERMES_PORT", "5487")

	cfg := Load()
	if cfg.ServerURL != "https://reports.example.com" {
		t.Errorf("ServerURL = %q, want %q", cfg.ServerURL, "https://reports.example.com")
	}
}

func TestServerURLTrimsTrailingSlash(t *testing.T) {
	t.Setenv("HERMES_SERVER_URL", "https://reports.example.com/")
	t.Setenv("HERMES_PORT", "5487")

	cfg := Load()
	if cfg.ServerURL != "https://reports.example.com" {
		t.Errorf("ServerURL = %q, want %q (trailing slash trimmed)", cfg.ServerURL, "https://reports.example.com")
	}
}

func TestServerURLDefault(t *testing.T) {
	t.Setenv("HERMES_SERVER_URL", "")
	t.Setenv("HERMES_PORT", "9000")

	cfg := Load()
	if cfg.ServerURL != "http://localhost:9000" {
		t.Errorf("ServerURL = %q, want %q", cfg.ServerURL, "http://localhost:9000")
	}
}
