package auth

import (
	"strings"
	"testing"
	"time"
)

const testSecret = "test-secret-test-secret-test-secret!"

func TestPasswordHashing(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "password123" {
		t.Fatal("password stored in plain text")
	}
	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("expected bcrypt hash, got %q", hash[:4])
	}
	if !CheckPassword(hash, "password123") {
		t.Error("correct password rejected")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Error("wrong password accepted")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	token, err := GenerateToken(testSecret, "user-1", "a@b.com", "admin", time.Hour)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	claims, err := ParseToken(testSecret, token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != "user-1" || claims.Email != "a@b.com" || claims.Role != "admin" {
		t.Errorf("unexpected claims %+v", claims)
	}
}

func TestParseTokenRejectsBadInput(t *testing.T) {
	t.Run("wrong secret", func(t *testing.T) {
		token, _ := GenerateToken(testSecret, "user-1", "a@b.com", "user", time.Hour)
		if _, err := ParseToken("another-secret-another-secret-12", token); err == nil {
			t.Error("expected error for wrong secret")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		token, _ := GenerateToken(testSecret, "user-1", "a@b.com", "user", -time.Minute)
		if _, err := ParseToken(testSecret, token); err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("garbage token", func(t *testing.T) {
		if _, err := ParseToken(testSecret, "not.a.token"); err == nil {
			t.Error("expected error for garbage token")
		}
	})
}
