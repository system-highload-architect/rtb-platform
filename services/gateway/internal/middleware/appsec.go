package middleware

import (
	"net/http"
)

// AppsecMiddleware выполняет базовые проверки безопасности:
// - валидация заголовка Host
// - санитизация входных данных (если потребуется)
type AppsecMiddleware struct {
	allowedHosts []string
}

func NewAppsecMiddleware(allowedHosts []string) *AppsecMiddleware {
	return &AppsecMiddleware{allowedHosts: allowedHosts}
}

func (m *AppsecMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что запрос пришёл на разрешённый хост (защита от DNS rebinding)
		if len(m.allowedHosts) > 0 {
			host := r.Host
			allowed := false
			for _, h := range m.allowedHosts {
				if host == h {
					allowed = true
					break
				}
			}
			if !allowed {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
		// Санитизация может быть добавлена позже, если потребуется
		next.ServeHTTP(w, r)
	})
}
