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
	port := getEnv("ACTA_PORT", "5487")
	serverURL := strings.TrimRight(os.Getenv("ACTA_SERVER_URL"), "/")
	if serverURL == "" {
		serverURL = "http://localhost:" + port
	}
	return &Config{
		Port:       port,
		DataDir:    getEnv("ACTA_DATA_DIR", "./reports"),
		WebDir:     getEnv("ACTA_WEB_DIR", "./web"),
		JWTSecret:  os.Getenv("ACTA_JWT_SECRET"),
		AdminUser:  os.Getenv("ACTA_ADMIN_USER"),
		AdminPass:  os.Getenv("ACTA_ADMIN_PASS"),
		ServerURL:  serverURL,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
