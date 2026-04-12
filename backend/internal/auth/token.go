package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrTokenInvalid = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type TokenClaims struct {
	Username  string    `json:"u"`
	ExpiresAt time.Time `json:"e"`
}

// CreateToken generates an HMAC-SHA256 signed token with format:
// base64url(payload).base64url(signature)
func CreateToken(secret, username string, duration time.Duration) (string, error) {
	claims := TokenClaims{
		Username:  username,
		ExpiresAt: time.Now().Add(duration),
	}
	data, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(data)
	return encoded + "." + sign(secret, encoded), nil
}

// ValidateToken verifies the token signature and expiry.
func ValidateToken(secret, token string) (*TokenClaims, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return nil, ErrTokenInvalid
	}

	expected := sign(secret, parts[0])
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return nil, ErrTokenInvalid
	}

	data, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrTokenInvalid
	}

	var claims TokenClaims
	if err := json.Unmarshal(data, &claims); err != nil {
		return nil, ErrTokenInvalid
	}

	if time.Now().After(claims.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}

func sign(secret, data string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
