package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	db "github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/db/generated"
)

var ErrNoAPIKey = errors.New("no api key configured")

// GenerateAPIKey creates a cryptographically random API key.
// Returns the raw key (to show to the admin once) and its SHA-256 hex hash
// (to store in the database).
func GenerateAPIKey() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	hash = hashKey(raw)
	return raw, hash, nil
}

// HashAPIKey returns the SHA-256 hex hash of a raw key string.
func HashAPIKey(raw string) string {
	return hashKey(raw)
}

func hashKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// RequireAPIKey is an HTTP middleware that enforces API key authentication.
// Option A: always required — returns 401 if no key is configured or the
// provided key does not match.
//
// The client must send:
//
//	Authorization: Bearer <raw_key>
func RequireAPIKey(queries db.Querier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := bearerToken(r)
			if !ok {
				http.Error(w, `{"error":"missing or malformed Authorization header"}`, http.StatusUnauthorized)
				return
			}

			storedKey, err := queries.GetAPIKey(r.Context())
			if err != nil {
				// No key in DB — always reject (option A)
				http.Error(w, `{"error":"api key not configured"}`, http.StatusUnauthorized)
				return
			}

			if hashKey(raw) != storedKey.KeyHash {
				http.Error(w, `{"error":"invalid api key"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// bearerToken extracts the token from "Authorization: Bearer <token>".
func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}
	return token, true
}

// APIKeyFromContext is a placeholder — the API key middleware does not inject
// data into context, but we keep this for symmetry with the session middleware.
func APIKeyFromContext(ctx context.Context) bool {
	return ctx != nil
}
