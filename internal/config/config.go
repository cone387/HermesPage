package config

import "os"

type Config struct {
	Port    string
	DataDir string
	APIKey  string
	WebDir  string
}

func Load() *Config {
	return &Config{
		Port:    getEnv("HERMES_PORT", "8080"),
		DataDir: getEnv("HERMES_DATA_DIR", "./reports"),
		APIKey:  getEnv("HERMES_API_KEY", ""),
		WebDir:  getEnv("HERMES_WEB_DIR", "./web"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
