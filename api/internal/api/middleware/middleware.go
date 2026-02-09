package middleware

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func CacheControl(maxAge time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(maxAge.Seconds())))
			next.ServeHTTP(w, r)
		})
	}
}

var apiKeyWarningLogged bool

func APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := os.Getenv("DEVARCH_API_KEY")
		if apiKey == "" {
			if !apiKeyWarningLogged {
				log.Println("WARNING: DEVARCH_API_KEY is not set — API authentication is disabled. Set DEVARCH_API_KEY to enable auth.")
				apiKeyWarningLogged = true
			}
			next.ServeHTTP(w, r)
			return
		}

		provided := r.Header.Get("X-API-Key")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
}

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

const maxVisitors = 10000

func RateLimit(requestsPerSecond float64, burst int) func(http.Handler) http.Handler {
	rl := &rateLimiter{visitors: make(map[string]*visitor)}

	go func() {
		for {
			time.Sleep(time.Minute)
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			rl.mu.Lock()
			v, exists := rl.visitors[ip]
			if !exists {
				if len(rl.visitors) >= maxVisitors {
					rl.mu.Unlock()
					http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
					return
				}
				v = &visitor{tokens: float64(burst)}
				rl.visitors[ip] = v
			}

			elapsed := time.Since(v.lastSeen).Seconds()
			v.tokens += elapsed * requestsPerSecond
			if v.tokens > float64(burst) {
				v.tokens = float64(burst)
			}
			v.lastSeen = time.Now()

			if v.tokens < 1 {
				rl.mu.Unlock()
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			v.tokens--
			rl.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

func MaxBodySize(bytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, bytes)
			next.ServeHTTP(w, r)
		})
	}
}
