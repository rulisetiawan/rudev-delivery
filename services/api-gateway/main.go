package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Port            string
	JWTSecret       []byte
	AuthServiceURL  string
	OrderServiceURL string
	RedisHost       string
}

func main() {
	cfg := Config{
		Port:            getEnv("PORT", "8080"),
		JWTSecret:       []byte(getEnv("JWT_SECRET", "supersecretkey_desa123")),
		AuthServiceURL:  getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
		OrderServiceURL: getEnv("ORDER_SERVICE_URL", "http://order-service:8082"),
		RedisHost:       getEnv("REDIS_HOST", "redis:6379"),
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[WARNING] Redis connect error: %v", err)
	} else {
		log.Println("[API Gateway] Connected to Redis successfully!")
	}

	rateLimiter := NewRateLimiter(rdb, 60, 1*time.Minute)

	authURL, _ := url.Parse(cfg.AuthServiceURL)
	orderURL, _ := url.Parse(cfg.OrderServiceURL)

	authProxy := httputil.NewSingleHostReverseProxy(authURL)
	orderProxy := httputil.NewSingleHostReverseProxy(orderURL)

	mux := http.NewServeMux()

	// Public Routes (Auth)
	mux.Handle("/api/v1/auth/", corsMiddleware(
		correlationIDMiddleware(
			rateLimiter.LimitMiddleware(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					authProxy.ServeHTTP(w, r)
				}),
			),
		),
	))

	// Public Catalog & Share Links (No JWT Required)
	mux.Handle("/api/v1/public/", corsMiddleware(
		correlationIDMiddleware(
			rateLimiter.LimitMiddleware(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					orderProxy.ServeHTTP(w, r)
				}),
			),
		),
	))

	// Protected Routes (Orders & Admin Catalog)
	mux.Handle("/api/v1/orders/", corsMiddleware(
		correlationIDMiddleware(
			jwtAuthMiddleware(cfg.JWTSecret,
				rateLimiter.LimitMiddleware(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						orderProxy.ServeHTTP(w, r)
					}),
				),
			),
		),
	))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"api-gateway"}`))
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("[Custom Go API Gateway] Running on port %s...", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server Listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[API Gateway] Shutting down gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	_ = server.Shutdown(shutdownCtx)
	log.Println("[API Gateway] Server stopped.")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func correlationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", reqID)
		r.Header.Set("X-Request-ID", reqID)

		next.ServeHTTP(w, r)
	})
}

func jwtAuthMiddleware(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"Unauthorized: Missing token"}`, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error":"Unauthorized: Invalid token"}`, http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, exists := claims["user_id"]; exists {
				r.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))
			}
			if role, exists := claims["role"]; exists {
				r.Header.Set("X-User-Role", fmt.Sprintf("%v", role))
			}
		}

		next.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
