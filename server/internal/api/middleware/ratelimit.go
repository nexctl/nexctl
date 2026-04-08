package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/nexctl/nexctl/server/pkg/errcode"
	"github.com/nexctl/nexctl/server/pkg/response"
	"golang.org/x/time/rate"
)

// clientIP returns the client address for rate limiting. Prefer running after chi RealIP so RemoteAddr is normalized.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type ipRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	lim      rate.Limit
	burst    int
	maxKeys  int
}

func newIPRateLimiter(interval time.Duration, burst, maxKeys int) *ipRateLimiter {
	return &ipRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		lim:      rate.Every(interval),
		burst:    burst,
		maxKeys:  maxKeys,
	}
}

func (l *ipRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if ip == "" {
		ip = "unknown"
	}
	if len(l.limiters) >= l.maxKeys && l.limiters[ip] == nil {
		for k := range l.limiters {
			delete(l.limiters, k)
			break
		}
	}
	lim, ok := l.limiters[ip]
	if !ok {
		lim = rate.NewLimiter(l.lim, l.burst)
		l.limiters[ip] = lim
	}
	return lim.Allow()
}

// RateLimitFunc wraps a handler with per-IP rate limiting (uses chi RealIP; ensure RealIP runs before this).
func RateLimitFunc(interval time.Duration, burst, maxKeys int) func(http.HandlerFunc) http.HandlerFunc {
	lim := newIPRateLimiter(interval, burst, maxKeys)
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !lim.allow(clientIP(r)) {
				response.WriteError(w, http.StatusTooManyRequests, errcode.RateLimited, "too many requests")
				return
			}
			next(w, r)
		}
	}
}
