package middleware

import (
	"net/http"

	"rtb-platform/pkg/idempotent"
)

// IdempotentMiddleware проверяет заголовок Idempotency-Key и предотвращает дублирование запросов.
type IdempotentMiddleware struct {
	store *idempotent.Store
}

func NewIdempotentMiddleware(store *idempotent.Store) *IdempotentMiddleware {
	return &IdempotentMiddleware{store: store}
}

func (m *IdempotentMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Применяем только для методов, которые могут быть идемпотентными (POST / PATCH)
		if r.Method == http.MethodPost || r.Method == http.MethodPatch {
			key := r.Header.Get("Idempotency-Key")
			if key != "" {
				if !m.store.Check(key) {
					http.Error(w, "duplicate request", http.StatusConflict)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
