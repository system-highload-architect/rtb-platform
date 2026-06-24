package middleware

import (
	"context"
	"net/http"
	"strings"

	"rtb-platform/services/gateway/internal/ports"
)

type AuthMiddleware struct {
	authClient ports.AuthPort // может быть nil, если auth не настроен
}

func NewAuthMiddleware(authClient ports.AuthPort) *AuthMiddleware {
	return &AuthMiddleware{authClient: authClient}
}

func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.authClient == nil {
			// Auth отключён — пропускаем
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		resp, err := m.authClient.Validate(r.Context(), parts[1])
		if err != nil || !resp.Valid {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Можно положить userID и role в контекст
		ctx := context.WithValue(r.Context(), "user_id", resp.UserId)
		ctx = context.WithValue(ctx, "role", resp.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
