package auth

import (
	"context"
	"net/http"
)

type contextKey string

const usernameKey contextKey = "username"

// RequireAuth validates the session cookie on each request.
// Redirects to /admin/login if the session is missing or invalid.
func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := GetSessionToken(r)
			if err != nil {
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				return
			}

			claims, err := ValidateToken(secret, token)
			if err != nil {
				ClearSessionCookie(w)
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), usernameKey, claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UsernameFromContext returns the authenticated username set by RequireAuth.
func UsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameKey).(string)
	return username, ok
}
