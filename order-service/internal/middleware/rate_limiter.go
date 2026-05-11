package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client        *redis.Client
	maxRequests   int
	windowSeconds int
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	maxReq := 10
	windowSecs := 60

	if v := os.Getenv("RATE_LIMIT_REQUESTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxReq = n
		}
	}
	if v := os.Getenv("RATE_LIMIT_WINDOW_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			windowSecs = n
		}
	}

	return &RateLimiter{
		client:        redisClient,
		maxRequests:   maxReq,
		windowSeconds: windowSecs,
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)
		ctx := context.Background()
		window := time.Duration(rl.windowSeconds) * time.Second

		count, err := rl.client.Incr(ctx, key).Result()
		if err != nil {
			log.Printf("[RateLimit] Redis error: %v — allowing request", err)
			c.Next()
			return
		}

		if count == 1 {
			rl.client.Expire(ctx, key, window)
		}

		if count > int64(rl.maxRequests) {
			log.Printf("[RateLimit] IP %s exceeded limit (%d/%d)", clientIP, count, rl.maxRequests)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"limit":       rl.maxRequests,
				"window_secs": rl.windowSeconds,
			})
			c.Abort()
			return
		}

		log.Printf("[RateLimit] IP %s: request %d/%d", clientIP, count, rl.maxRequests)
		c.Next()
	}
}
