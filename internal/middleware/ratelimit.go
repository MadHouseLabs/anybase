package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// limiterEntry stores a rate limiter with last access time for cleanup
type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// PerIPRateLimiter creates a rate limiter per IP address with automatic cleanup
type PerIPRateLimiter struct {
	limiters map[string]*limiterEntry
	mu       sync.RWMutex
	rps      int
	burst    int
	ttl      time.Duration
	stop     chan struct{}
}

func NewPerIPRateLimiter(rps, burst int) *PerIPRateLimiter {
	rl := &PerIPRateLimiter{
		limiters: make(map[string]*limiterEntry),
		rps:      rps,
		burst:    burst,
		ttl:      15 * time.Minute, // Clean up entries after 15 minutes of inactivity
		stop:     make(chan struct{}),
	}
	
	// Start cleanup goroutine
	go rl.cleanupLoop()
	
	return rl
}

// cleanupLoop removes expired entries periodically
func (rl *PerIPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute) // Run cleanup every 5 minutes
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stop:
			return
		}
	}
}

// cleanup removes entries that haven't been accessed recently
func (rl *PerIPRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	for ip, entry := range rl.limiters {
		if now.Sub(entry.lastAccess) > rl.ttl {
			delete(rl.limiters, ip)
		}
	}
}

// Stop gracefully stops the cleanup goroutine
func (rl *PerIPRateLimiter) Stop() {
	close(rl.stop)
}

func (rl *PerIPRateLimiter) getLimiter(ip string) *rate.Limiter {
	now := time.Now()
	
	rl.mu.RLock()
	entry, exists := rl.limiters[ip]
	rl.mu.RUnlock()
	
	if exists {
		// Update last access time
		rl.mu.Lock()
		entry.lastAccess = now
		rl.mu.Unlock()
		return entry.limiter
	}
	
	// Create new limiter
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	// Double-check in case another goroutine created it
	entry, exists = rl.limiters[ip]
	if !exists {
		entry = &limiterEntry{
			limiter:    rate.NewLimiter(rate.Limit(rl.rps), rl.burst),
			lastAccess: now,
		}
		rl.limiters[ip] = entry
	}
	return entry.limiter
}

func (rl *PerIPRateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getLimiter(ip)
		
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}