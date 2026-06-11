package config

import (
	"testing"
	"time"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
}

func TestLoadDefaults(t *testing.T) {
	setRequiredEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Port != "8090" {
		t.Errorf("port = %s", cfg.Port)
	}
	if cfg.JWTExpiry != 24*time.Hour {
		t.Errorf("expiry = %v", cfg.JWTExpiry)
	}
	if cfg.MaxUploadBytes != 10<<20 {
		t.Errorf("max upload = %d", cfg.MaxUploadBytes)
	}
}

func TestLoadOverrides(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("PORT", "9999")
	t.Setenv("JWT_EXPIRY_HOURS", "2")
	t.Setenv("ALLOWED_ORIGINS", "https://a.com,https://b.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Port != "9999" {
		t.Errorf("port = %s", cfg.Port)
	}
	if cfg.JWTExpiry != 2*time.Hour {
		t.Errorf("expiry = %v", cfg.JWTExpiry)
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("origins = %v", cfg.AllowedOrigins)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	if _, err := Load(); err == nil {
		t.Error("expected error for missing DATABASE_URL")
	}
}

func TestLoadRequiresStrongSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "short")
	if _, err := Load(); err == nil {
		t.Error("expected error for short JWT_SECRET")
	}
}
