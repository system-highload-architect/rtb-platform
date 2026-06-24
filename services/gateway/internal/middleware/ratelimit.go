package middleware

import (
	"net/http"

	"rtb-platform/pkg/ratelimit"
)

// RateLimitMiddleware ограничивает частоту запросов от одного IP.
type RateLimitMiddleware struct {
	limiter *ratelimit.Limiter
}

func NewRateLimitMiddleware(limiter *ratelimit.Limiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: limiter}
}

func (m *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !m.limiter.Allow(ip) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
