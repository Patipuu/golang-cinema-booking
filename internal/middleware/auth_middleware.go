package middleware

import (
	"context"
	"net/http"
	"strings"

	"booking_cinema_golang/internal/utils"
)

type contextKey string

const UserClaimsKey contextKey = "user_claims"

// AuthMiddleware validates JWT and sets claims in context. Use JWT_SECRET from config.
func AuthMiddleware(jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				utils.JSONUnauthorized(w, "missing authorization header")
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.JSONUnauthorized(w, "invalid authorization format")
				return
			}
			claims, err := utils.ParseToken(jwtSecret, parts[1])
			if err != nil {
				utils.JSONUnauthorized(w, "invalid or expired token")
				return
			}
			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminMiddleware checks if the authenticated user has admin role.
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r.Context())
		if claims == nil {
			utils.JSONUnauthorized(w, "authentication required")
			return
		}

		if claims.Role != "admin" {
			utils.JSONForbidden(w, "admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetClaims returns Claims from request context (nil if not set).
func GetClaims(ctx context.Context) *utils.Claims {
	c, _ := ctx.Value(UserClaimsKey).(*utils.Claims)
	return c
}
