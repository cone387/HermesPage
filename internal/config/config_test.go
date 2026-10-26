package config

import "testing"

func TestServerURLFromEnv(t *testing.T) {
	t.Setenv("ACTA_SERVER_URL", "https://acta.example.com")
	t.Setenv("ACTA_PORT", "5487")

	cfg := Load()
	if cfg.ServerURL != "https://acta.example.com" {
		t.Errorf("ServerURL = %q, want %q", cfg.ServerURL, "https://acta.example.com")
	}
}

func TestServerURLTrimsTrailingSlash(t *testing.T) {
	t.Setenv("ACTA_SERVER_URL", "https://acta.example.com/")
	t.Setenv("ACTA_PORT", "5487")

	cfg := Load()
	if cfg.ServerURL != "https://acta.example.com" {
		t.Errorf("ServerURL = %q, want %q (trailing slash trimmed)", cfg.ServerURL, "https://acta.example.com")
	}
}

func TestServerURLDefault(t *testing.T) {
	t.Setenv("ACTA_SERVER_URL", "")
	t.Setenv("ACTA_PORT", "9000")

	cfg := Load()
	if cfg.ServerURL != "http://localhost:9000" {
		t.Errorf("ServerURL = %q, want %q", cfg.ServerURL, "http://localhost:9000")
	}
}

func TestAllEnvPrefixes(t *testing.T) {
	cfg := Load()
	if cfg.Port != "5487" {
		t.Errorf("Port = %q, want 5487", cfg.Port)
	}
}
