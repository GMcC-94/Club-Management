package middleware

import (
	"club-management/internal/logger"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/csrf"
	"golang.org/x/time/rate"
)

// CSRFTokenKey is the context key for the CSRF token
type contextKey string

const CSRFTokenKey contextKey = "csrf_token"

// RequestLogger logs every incoming request
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"duration", time.Since(start).String(),
		)
	})
}

// InjectCSRFToken pulls the CSRF token from gorilla/csrf and stores it in context
// so templ components can access it without needing it passed as a parameter
func InjectCSRFToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := csrf.Token(r)
		ctx := context.WithValue(r.Context(), CSRFTokenKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// --- Rate Limiting ---

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	loginLimiters = make(map[string]*ipLimiter)
	mu            sync.Mutex
)

func getLoginLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if l, exists := loginLimiters[ip]; exists {
		l.lastSeen = time.Now()
		return l.limiter
	}

	// 5 attempts per minute per IP
	l := rate.NewLimiter(rate.Every(time.Minute/5), 5)
	loginLimiters[ip] = &ipLimiter{limiter: l, lastSeen: time.Now()}
	return l
}

// Cleanup stale IP limiters every 10 minutes
func init() {
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			mu.Lock()
			for ip, l := range loginLimiters {
				if time.Since(l.lastSeen) > 15*time.Minute {
					delete(loginLimiters, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

// LoginRateLimit limits POST /login to 5 attempts per minute per IP
func LoginRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ip := r.RemoteAddr
			if !getLoginLimiter(ip).Allow() {
				logger.Warn("rate limit exceeded", "ip", ip, "path", r.URL.Path)
				http.Error(w, "Too many login attempts. Please try again in a minute.", http.StatusTooManyRequests)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
