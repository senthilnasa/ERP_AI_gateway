package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/senthilnasa/ERP_AI_gateway/internal/config"
	"github.com/senthilnasa/ERP_AI_gateway/internal/logger"
	"github.com/senthilnasa/ERP_AI_gateway/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware injects a X-Request-ID header if not provided.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Set("RequestID", reqID)
		c.Header("X-Request-ID", reqID)
		c.Next()
	}
}

func AuthMiddleware(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Only enforce Bearer API Key authentication on /api/* endpoints
		if !strings.HasPrefix(path, "/api/") {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			reqID, _ := c.Get("RequestID")
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ApiResponse{
				Success: false,
				Error: models.ApiErrorDetail{
					Code:      "UNAUTHORIZED",
					Message:   "Missing Authorization header",
					RequestID: fmt.Sprintf("%v", reqID),
				},
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] != expectedKey {
			reqID, _ := c.Get("RequestID")
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ApiResponse{
				Success: false,
				Error: models.ApiErrorDetail{
					Code:      "UNAUTHORIZED",
					Message:   "Invalid Bearer API Key",
					RequestID: fmt.Sprintf("%v", reqID),
				},
			})
			return
		}

		c.Next()
	}
}

// RateLimiter tracks requests per client IP.
type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]int
	lastReset time.Time
	limit    int
}

func NewRateLimiter(limitPerMin int) *RateLimiter {
	return &RateLimiter{
		clients:   make(map[string]int),
		lastReset: time.Now(),
		limit:     limitPerMin,
	}
}

func RateLimitMiddleware(cfg config.RateLimitConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(cfg.RequestsPerMinute)

	return func(c *gin.Context) {
		limiter.mu.Lock()
		if time.Since(limiter.lastReset) > time.Minute {
			limiter.clients = make(map[string]int)
			limiter.lastReset = time.Now()
		}

		ip := c.ClientIP()
		count := limiter.clients[ip]
		if count >= limiter.limit {
			limiter.mu.Unlock()
			reqID, _ := c.Get("RequestID")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, models.ApiResponse{
				Success: false,
				Error: models.ApiErrorDetail{
					Code:      "RATE_LIMIT_EXCEEDED",
					Message:   "Too many requests. Please try again later.",
					RequestID: fmt.Sprintf("%v", reqID),
				},
			})
			return
		}

		limiter.clients[ip] = count + 1
		limiter.mu.Unlock()
		c.Next()
	}
}

// RecoveryMiddleware gracefully catches panics.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Get().Error("Panic recovered: %v", err)
				reqID, _ := c.Get("RequestID")
				c.AbortWithStatusJSON(http.StatusInternalServerError, models.ApiResponse{
					Success: false,
					Error: models.ApiErrorDetail{
						Code:      "INTERNAL_SERVER_ERROR",
						Message:   "An internal server error occurred",
						RequestID: fmt.Sprintf("%v", reqID),
					},
				})
			}
		}()
		c.Next()
	}
}

// LoggingMiddleware logs request metadata without logging prompts or sensitive text.
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		reqID, _ := c.Get("RequestID")
		logger.Get().Info("[%s] %s %s | Status: %d | Duration: %v",
			reqID, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency)
	}
}

// MaxBytesMiddleware enforces maximum payload body size limit.
func MaxBytesMiddleware(maxMB int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(maxMB)<<20)
		}
		c.Next()
	}
}
