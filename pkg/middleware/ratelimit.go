package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter provides Redis-backed rate limiting.
type RateLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

// NewRateLimiter creates a RateLimiter.
func NewRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{rdb: rdb, limit: limit, window: window}
}

// Limit returns a gin middleware that rate-limits by IP.
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", ip)

		ctx := context.Background()
		count, err := rl.rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			rl.rdb.Expire(ctx, key, rl.window)
		}

		ttl, _ := rl.rdb.TTL(ctx, key).Result()
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(max(0, rl.limit-int(count))))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))

		if int(count) > rl.limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

// USSDRateLimiter limits USSD sessions to 10 per hour per phone.
func (rl *RateLimiter) USSDLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		phone := c.PostForm("phoneNumber")
		if phone == "" {
			c.Next()
			return
		}

		key := fmt.Sprintf("ussd_limit:%s", phone)
		ctx := context.Background()

		count, err := rl.rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			rl.rdb.Expire(ctx, key, time.Hour)
		}

		if count > 10 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "USSD session limit exceeded (10/hour)",
			})
			return
		}
		c.Next()
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
