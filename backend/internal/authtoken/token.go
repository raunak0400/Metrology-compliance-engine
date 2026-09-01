// Package authtoken issues and verifies JWTs. Split out from internal/auth so
// both internal/app (shared Deps) and internal/middleware can depend on the
// Tokenizer type without an import cycle through the auth business-logic
// package (which itself depends on internal/app-registered routes).
package authtoken

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenClaims struct {
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	JTI string `json:"jti"`
	jwt.RegisteredClaims
}

type Tokenizer struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

func NewTokenizer(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *Tokenizer {
	return &Tokenizer{AccessSecret: accessSecret, RefreshSecret: refreshSecret, AccessTTL: accessTTL, RefreshTTL: refreshTTL}
}

func (t *Tokenizer) SignAccessToken(userID, email string, roles, permissions []string) (string, error) {
	claims := AccessTokenClaims{
		Email:       email,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(t.AccessSecret))
}

// SignRefreshToken returns the signed token and its expiry time (persisted
// alongside the token's SHA-256 hash in refresh_tokens).
func (t *Tokenizer) SignRefreshToken(userID, jti string) (token string, expiresAt time.Time, err error) {
	expiresAt = time.Now().Add(t.RefreshTTL)
	claims := RefreshTokenClaims{
		JTI: jti,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(t.RefreshSecret))
	return token, expiresAt, err
}

func (t *Tokenizer) VerifyAccessToken(token string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(tok *jwt.Token) (interface{}, error) {
		return []byte(t.AccessSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func (t *Tokenizer) VerifyRefreshToken(token string) (*RefreshTokenClaims, error) {
	claims := &RefreshTokenClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(tok *jwt.Token) (interface{}, error) {
		return []byte(t.RefreshSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}
