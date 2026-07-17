package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGCSConfig(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("GCS_BUCKET", "my-bucket")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.GCSBucket != "my-bucket" {
		t.Errorf("GCSBucket = %q, want %q", cfg.GCSBucket, "my-bucket")
	}
}

func TestLoadEnvFilePreservesExistingEnvironment(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://from-environment")

	filename := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(filename, []byte("DATABASE_URL=postgres://from-file\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	if err := loadEnvFile(filename); err != nil {
		t.Fatalf("loadEnvFile() error = %v", err)
	}

	if got := os.Getenv("DATABASE_URL"); got != "postgres://from-environment" {
		t.Fatalf("DATABASE_URL = %q, want existing environment value", got)
	}
}
