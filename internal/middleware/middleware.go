package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Logger adds structured logging to each request
func Logger() gin.HandlerFunc {
	logger, _ := zap.NewProduction()
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("request completed",
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("method", c.Request.Method),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("request_id", c.GetString("request_id")),
		)
	}
}

// Recovery handles panics and returns a 500 error
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger, _ := zap.NewProduction()
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("request_id", c.GetString("request_id")),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
		}()
		c.Next()
	}
}

// CORS adds Cross-Origin Resource Sharing headers
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimit creates a rate limiting middleware
func RateLimit(requestsPerMinute int, cleanupInterval time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate.Limit(requestsPerMinute)/60, requestsPerMinute)

	// Start cleanup goroutine
	go func() {
		for {
			time.Sleep(cleanupInterval)
			limiter.cleanup(cleanupInterval)
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

// Client represents a rate-limited client
type Client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter manages rate limiting for multiple clients
type RateLimiter struct {
	clients map[string]*Client
	mu      sync.Mutex
	rate    rate.Limit
	burst   int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*Client),
		rate:    r,
		burst:   burst,
	}
}

// allow checks if a client is allowed to make a request
func (rl *RateLimiter) allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	client, exists := rl.clients[clientIP]
	if !exists {
		client = &Client{
			limiter: rate.NewLimiter(rl.rate, rl.burst),
		}
		rl.clients[clientIP] = client
	}

	client.lastSeen = time.Now()
	return client.limiter.Allow()
}

// cleanup removes old clients that haven't been seen recently
func (rl *RateLimiter) cleanup(cleanupInterval time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, client := range rl.clients {
		if time.Since(client.lastSeen) > cleanupInterval {
			delete(rl.clients, ip)
		}
	}
}
