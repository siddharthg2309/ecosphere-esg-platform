package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Addr        string
	DatabaseURL string
	JWTSecret   string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
	CORSOrigin  string
	SMTPAddr    string
}

func Load() (Config, error) {
	accessTTL, err := duration("ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}
	refreshTTL, err := duration("REFRESH_TTL", 7*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Addr: value("ADDR", ":8080"), DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"), AccessTTL: accessTTL, RefreshTTL: refreshTTL,
		CORSOrigin: value("CORS_ORIGIN", "http://localhost:5173"), SMTPAddr: value("SMTP_ADDR", "localhost:1025"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 24 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 24 characters")
	}
	return cfg, nil
}

func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func duration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return d, nil
}
