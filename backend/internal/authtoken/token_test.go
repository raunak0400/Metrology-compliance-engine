package authtoken

import (
	"testing"
	"time"
)

func TestAccessToken_RoundTrip(t *testing.T) {
	tok := NewTokenizer("access-secret-value", "refresh-secret-value", time.Minute, time.Hour)

	signed, err := tok.SignAccessToken("user-1", "a@b.com", []string{"ADMIN"}, []string{"*"})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	claims, err := tok.VerifyAccessToken(signed)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.Subject != "user-1" || claims.Email != "a@b.com" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "ADMIN" {
		t.Fatalf("unexpected roles: %+v", claims.Roles)
	}
}

func TestAccessToken_RejectsWrongSecret(t *testing.T) {
	tok := NewTokenizer("access-secret-value", "refresh-secret-value", time.Minute, time.Hour)
	signed, _ := tok.SignAccessToken("user-1", "a@b.com", nil, nil)

	other := NewTokenizer("different-secret-value", "refresh-secret-value", time.Minute, time.Hour)
	if _, err := other.VerifyAccessToken(signed); err == nil {
		t.Fatal("expected verification to fail with a different secret")
	}
}

func TestRefreshToken_RoundTripAndExpiry(t *testing.T) {
	tok := NewTokenizer("access-secret-value", "refresh-secret-value", time.Minute, time.Hour)
	signed, expiresAt, err := tok.SignRefreshToken("user-1", "jti-123")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if !expiresAt.After(time.Now()) {
		t.Fatal("expected expiry in the future")
	}

	claims, err := tok.VerifyRefreshToken(signed)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.JTI != "jti-123" || claims.Subject != "user-1" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}
