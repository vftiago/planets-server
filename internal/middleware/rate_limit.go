package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	Enabled           bool
	TrustProxy        bool
}

type RateLimiter struct {
	config  RateLimitConfig
	clients map[string]*rate.Limiter
	mu      sync.RWMutex
}

func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		config:  config,
		clients: make(map[string]*rate.Limiter),
	}

	if config.Enabled {
		go rl.cleanupClients()
	}

	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.clients[ip]
	rl.mu.RUnlock()

	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.config.RequestsPerSecond), rl.config.BurstSize)

		rl.mu.Lock()
		rl.clients[ip] = limiter
		rl.mu.Unlock()
	}

	return limiter
}

func (rl *RateLimiter) cleanupClients() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		// Remove clients that haven't been used recently
		for ip, limiter := range rl.clients {
			if limiter.TokensAt(time.Now()) == float64(rl.config.BurstSize) {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		ip := getClientIP(r, rl.config.TrustProxy)
		limiter := rl.getLimiter(ip)

		logger := slog.With(
			"middleware", "rate_limit",
			"client_ip", ip,
			"method", r.Method,
			"path", r.URL.Path,
		)

		if !limiter.Allow() {
			logger.Warn("Rate limit exceeded",
				"requests_per_second", rl.config.RequestsPerSecond,
				"burst_size", rl.config.BurstSize,
			)

			w.Header().Set("Retry-After", "1")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		logger.Debug("Request allowed through rate limiter")
		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			// X-Forwarded-For can be comma-separated; first entry is the client
			if i := strings.IndexByte(xff, ','); i != -1 {
				return strings.TrimSpace(xff[:i])
			}
			return xff
		}

		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return xri
		}
	}

	// Strip port from RemoteAddr (e.g. "192.168.1.1:12345" -> "192.168.1.1")
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
