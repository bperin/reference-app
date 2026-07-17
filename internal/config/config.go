package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	AppEnv           string
	HTTPAddress      string
	DatabaseURL      string
	JWTSecret        string
	JWTIssuer        string
	JWTAudience      string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	LogLevel         string
	GCSBucket        string
	GCSCredentials   string
	EventarcAudience string
	EventarcIssuer   string
}

// Load loads configuration from environment variables.
// It also reads a .env file if present in the current working directory.
func Load() (*Config, error) {
	// Attempt to load from .env file if it exists, reporting other file-system errors
	if err := loadEnvFile(".env"); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read .env file: %w", err)
	}

	accessTokenTTL, err := getEnvDuration("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_TTL: %w", err)
	}

	refreshTokenTTL, err := getEnvDuration("REFRESH_TOKEN_TTL", 720*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_TTL: %w", err)
	}

	cfg := &Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		HTTPAddress:      getEnv("HTTP_ADDRESS", ":8080"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTIssuer:        getEnv("JWTIssuer", "reference-api"),
		JWTAudience:      getEnv("JWTAudience", "reference-api-clients"),
		AccessTokenTTL:   accessTokenTTL,
		RefreshTokenTTL:  refreshTokenTTL,
		LogLevel:         getEnv("LOG_LEVEL", "INFO"),
		GCSBucket:        getEnv("GCS_BUCKET", ""),
		GCSCredentials:   getEnv("GCS_CREDENTIALS", ""),
		EventarcAudience: getEnv("EVENTARC_AUDIENCE", ""),
		EventarcIssuer:   getEnv("EVENTARC_ISSUER", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.GCSBucket == "" {
		return nil, fmt.Errorf("GCS_BUCKET is required")
	}

	return cfg, nil
}

// Print logs the configuration with sensitive values redacted.
func (c *Config) Print() {
	redactedDB := redactDSN(c.DatabaseURL)
	redactedSecret := "[REDACTED]"
	if len(c.JWTSecret) > 0 {
		redactedSecret = fmt.Sprintf("[REDACTED (%d chars)]", len(c.JWTSecret))
	}

	fmt.Printf("Configuration loaded:\n")
	fmt.Printf("  APP_ENV:           %s\n", c.AppEnv)
	fmt.Printf("  HTTP_ADDRESS:      %s\n", c.HTTPAddress)
	fmt.Printf("  DATABASE_URL:      %s\n", redactedDB)
	fmt.Printf("  JWT_ISSUER:        %s\n", c.JWTIssuer)
	fmt.Printf("  JWT_AUDIENCE:      %s\n", c.JWTAudience)
	fmt.Printf("  ACCESS_TOKEN_TTL:  %s\n", c.AccessTokenTTL)
	fmt.Printf("  REFRESH_TOKEN_TTL: %s\n", c.RefreshTokenTTL)
	fmt.Printf("  LOG_LEVEL:         %s\n", c.LogLevel)
	fmt.Printf("  JWT_SECRET:        %s\n", redactedSecret)
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) (time.Duration, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal, nil
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, err
	}
	return d, nil
}

func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			val = val[1 : len(val)-1]
		}
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, val); err != nil {
				return fmt.Errorf("set %s: %w", key, err)
			}
		}
	}
	return scanner.Err()
}

func redactDSN(dsn string) string {
	if !strings.HasPrefix(dsn, "postgres://") && !strings.HasPrefix(dsn, "postgresql://") {
		return "[REDACTED]"
	}
	parts := strings.SplitN(dsn, "@", 2)
	if len(parts) != 2 {
		return "[REDACTED]"
	}
	prefixParts := strings.SplitN(parts[0], "://", 2)
	if len(prefixParts) != 2 {
		return "[REDACTED]"
	}
	return fmt.Sprintf("%s://[REDACTED_USER_PASS]@%s", prefixParts[0], parts[1])
}
