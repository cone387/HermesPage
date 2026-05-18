package config

import "os"

type Config struct {
	Port       string
	DataDir    string
	WebDir     string
	JWTSecret  string
	AdminUser  string
	AdminPass  string
}

func Load() *Config {
	return &Config{
		Port:       getEnv("HERMES_PORT", "8080"),
		DataDir:    getEnv("HERMES_DATA_DIR", "./reports"),
		WebDir:     getEnv("HERMES_WEB_DIR", "./web"),
		JWTSecret:  os.Getenv("HERMES_JWT_SECRET"),
		AdminUser:  os.Getenv("HERMES_ADMIN_USER"),
		AdminPass:  os.Getenv("HERMES_ADMIN_PASS"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
