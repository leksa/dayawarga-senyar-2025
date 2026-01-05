package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leksa/datamapper-senyar/internal/dto"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

type visitor struct {
	tokens    int
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per window
// window: time window duration
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	// Cleanup old entries every minute
	go rl.cleanup()

	return rl
}

// cleanup removes old visitor entries
func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastReset) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request is allowed for the given IP
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			tokens:    rl.rate - 1,
			lastReset: time.Now(),
		}
		return true
	}

	// Reset tokens if window has passed
	if time.Since(v.lastReset) > rl.window {
		v.tokens = rl.rate - 1
		v.lastReset = time.Now()
		return true
	}

	// Check if tokens available
	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

// RemainingTokens returns the remaining tokens for an IP
func (rl *RateLimiter) RemainingTokens(ip string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if v, exists := rl.visitors[ip]; exists {
		if time.Since(v.lastReset) > rl.window {
			return rl.rate
		}
		return v.tokens
	}
	return rl.rate
}

// Middleware returns a Gin middleware for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !rl.Allow(ip) {
			remaining := rl.RemainingTokens(ip)
			c.Header("X-RateLimit-Limit", string(rune(rl.rate)))
			c.Header("X-RateLimit-Remaining", string(rune(remaining)))
			c.Header("Retry-After", rl.window.String())

			c.JSON(http.StatusTooManyRequests, dto.APIResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "RATE_LIMITED",
					Message: "Too many requests, please try again later",
					Details: map[string]interface{}{
						"retry_after": rl.window.Seconds(),
					},
				},
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := rl.RemainingTokens(ip)
		c.Header("X-RateLimit-Limit", string(rune(rl.rate)))
		c.Header("X-RateLimit-Remaining", string(rune(remaining)))

		c.Next()
	}
}

// DefaultRateLimiter returns a rate limiter with default settings
// 100 requests per minute
func DefaultRateLimiter() *RateLimiter {
	return NewRateLimiter(100, time.Minute)
}

// StrictRateLimiter returns a stricter rate limiter
// 30 requests per minute (for write operations)
func StrictRateLimiter() *RateLimiter {
	return NewRateLimiter(30, time.Minute)
}
