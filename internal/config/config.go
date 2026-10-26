package config

import (
	"os"
	"strings"
)

type Config struct {
	Port       string
	DataDir    string
	WebDir     string
	JWTSecret  string
	AdminUser  string
	AdminPass  string
	ServerURL  string
}

func Load() *Config {
	port := getEnv("HERMES_PORT", "5487")
	serverURL := strings.TrimRight(os.Getenv("HERMES_SERVER_URL"), "/")
	if serverURL == "" {
		serverURL = "http://localhost:" + port
	}
	return &Config{
		Port:       port,
		DataDir:    getEnv("HERMES_DATA_DIR", "./reports"),
		WebDir:     getEnv("HERMES_WEB_DIR", "./web"),
		JWTSecret:  os.Getenv("HERMES_JWT_SECRET"),
		AdminUser:  os.Getenv("HERMES_ADMIN_USER"),
		AdminPass:  os.Getenv("HERMES_ADMIN_PASS"),
		ServerURL:  serverURL,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
