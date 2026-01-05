package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CacheEntry represents a cached response
type CacheEntry struct {
	Status      int
	Body        []byte
	ContentType string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

// Cache implements a simple in-memory cache
type Cache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
	maxSize int
}

// NewCache creates a new cache with specified TTL and max entries
func NewCache(ttl time.Duration, maxSize int) *Cache {
	c := &Cache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}

	// Cleanup expired entries every minute
	go c.cleanup()

	return c
}

// cleanup removes expired entries
func (c *Cache) cleanup() {
	for {
		time.Sleep(time.Minute)
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.After(entry.ExpiresAt) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// Get retrieves an entry from cache
func (c *Cache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry, true
}

// Set stores an entry in cache
func (c *Cache) Set(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entries if at max capacity
	if len(c.entries) >= c.maxSize {
		var oldestKey string
		var oldestTime time.Time
		for k, e := range c.entries {
			if oldestKey == "" || e.CreatedAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = e.CreatedAt
			}
		}
		if oldestKey != "" {
			delete(c.entries, oldestKey)
		}
	}

	c.entries[key] = entry
}

// Delete removes an entry from cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Clear removes all entries from cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// Size returns the number of entries in cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Invalidate removes entries matching a prefix
func (c *Cache) Invalidate(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.entries {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.entries, key)
		}
	}
}

// responseWriter wraps gin.ResponseWriter to capture response
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// generateCacheKey generates a cache key from request
func generateCacheKey(c *gin.Context) string {
	// Include method, path, and query params in cache key
	data := c.Request.Method + c.Request.URL.Path + c.Request.URL.RawQuery
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Middleware returns a Gin middleware for caching GET requests
func (cache *Cache) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Skip caching for certain paths
		path := c.Request.URL.Path
		skipPaths := []string{"/health", "/ready", "/api/v1/events", "/api/v1/sync", "/api/v1/scheduler"}
		for _, skip := range skipPaths {
			if len(path) >= len(skip) && path[:len(skip)] == skip {
				c.Next()
				return
			}
		}

		key := generateCacheKey(c)

		// Check cache
		if entry, found := cache.Get(key); found {
			c.Header("X-Cache", "HIT")
			c.Header("X-Cache-Age", time.Since(entry.CreatedAt).String())
			c.Data(entry.Status, entry.ContentType, entry.Body)
			c.Abort()
			return
		}

		// Cache miss - capture response
		c.Header("X-Cache", "MISS")

		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = writer

		c.Next()

		// Only cache successful responses
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			entry := &CacheEntry{
				Status:      c.Writer.Status(),
				Body:        writer.body.Bytes(),
				ContentType: c.Writer.Header().Get("Content-Type"),
				CreatedAt:   time.Now(),
				ExpiresAt:   time.Now().Add(cache.ttl),
			}
			cache.Set(key, entry)
		}
	}
}

// DefaultCache returns a cache with default settings
// 30 second TTL, max 1000 entries
func DefaultCache() *Cache {
	return NewCache(30*time.Second, 1000)
}
