package config

import (
	"os"
	"path/filepath"
	"testing"
)

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
