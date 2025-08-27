package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	Enabled           bool
}

type RateLimiter struct {
	config  RateLimitConfig
	clients map[string]*rate.Limiter
	mu      sync.RWMutex
	cleanup *time.Ticker
}

func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		config:  config,
		clients: make(map[string]*rate.Limiter),
		cleanup: time.NewTicker(time.Minute),
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
	for range rl.cleanup.C {
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

		ip := getClientIP(r)
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

func (rl *RateLimiter) Stop() {
	if rl.cleanup != nil {
		rl.cleanup.Stop()
	}
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header (for some proxies/load balancers)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to remote address
	return r.RemoteAddr
}