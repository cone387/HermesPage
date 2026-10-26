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
	t.Setenv("ACTA_DATA_DIR", "/tmp/data")
	t.Setenv("ACTA_WEB_DIR", "/tmp/web")
	t.Setenv("ACTA_JWT_SECRET", "s3cret")
	t.Setenv("ACTA_ADMIN_USER", "admin")
	t.Setenv("ACTA_ADMIN_PASS", "pass")
	t.Setenv("ACTA_PORT", "")

	cfg := Load()
	if cfg.DataDir != "/tmp/data" {
		t.Errorf("DataDir = %q, want /tmp/data", cfg.DataDir)
	}
	if cfg.WebDir != "/tmp/web" {
		t.Errorf("WebDir = %q, want /tmp/web", cfg.WebDir)
	}
	if cfg.JWTSecret != "s3cret" {
		t.Errorf("JWTSecret = %q, want s3cret", cfg.JWTSecret)
	}
	if cfg.AdminUser != "admin" {
		t.Errorf("AdminUser = %q, want admin", cfg.AdminUser)
	}
	if cfg.AdminPass != "pass" {
		t.Errorf("AdminPass = %q, want pass", cfg.AdminPass)
	}
	if cfg.Port != "5487" {
		t.Errorf("Port = %q, want default 5487", cfg.Port)
	}
}
