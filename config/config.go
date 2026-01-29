package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Port           string
	LogLevel       slog.Level
	AllowedOrigins []string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logLevelStr := os.Getenv("LOG_LEVEL")
	logLevel := parseLogLevel(logLevelStr)

	originsStr := os.Getenv("ALLOWED_ORIGINS")
	if originsStr == "" {
		originsStr = "*"
	}
	origins := strings.Split(originsStr, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return &Config{
		Port:           port,
		LogLevel:       logLevel,
		AllowedOrigins: origins,
	}
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
