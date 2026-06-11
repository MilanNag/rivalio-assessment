package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/milann/taskflow/internal/auth"
)

type contextKey string

const claimsKey contextKey = "claims"

// requireAuth validates the JWT from the Authorization header (or, for SSE
// where custom headers are impossible, the access_token query parameter).
func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := ""
		if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
			tokenString = strings.TrimPrefix(h, "Bearer ")
		} else if t := r.URL.Query().Get("access_token"); t != "" {
			tokenString = t
		}

		if tokenString == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
			return
		}

		claims, err := auth.ParseToken(s.cfg.JWTSecret, tokenString)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid or expired token.")
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func claimsFrom(ctx context.Context) *auth.Claims {
	claims, _ := ctx.Value(claimsKey).(*auth.Claims)
	return claims
}
