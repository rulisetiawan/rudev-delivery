package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rdb:    rdb,
		limit:  limit,
		window: window,
	}
}

func (rl *RateLimiter) LimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		identifier := r.Header.Get("X-User-ID")
		if identifier == "" {
			identifier = getClientIP(r)
		}

		key := fmt.Sprintf("rate_limit:%s", identifier)

		count, err := rl.rdb.Incr(ctx, key).Result()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if count == 1 {
			rl.rdb.Expire(ctx, key, rl.window)
		}

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(rl.limit)-count)))

		if count > int64(rl.limit) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Too Many Requests", "message": "Kuota akses terlampaui. Coba lagi nanti."}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
