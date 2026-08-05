package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientInfo
	limit    int
	window   time.Duration
}

type clientInfo struct {
	count    int
	lastSeen time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientInfo),
		limit:   limit,
		window:  window,
	}

	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, client := range rl.clients {
			if time.Since(client.lastSeen) > rl.window {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	client, exists := rl.clients[ip]
	if !exists {
		rl.clients[ip] = &clientInfo{
			count:    1,
			lastSeen: time.Now(),
		}
		return true
	}

	if time.Since(client.lastSeen) > rl.window {
		client.count = 1
		client.lastSeen = time.Now()
		return true
	}

	if client.count >= rl.limit {
		return false
	}

	client.count++
	client.lastSeen = time.Now()
	return true
}

func RateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
