// Rate limiting backed by Redis (fixed window per key) so limits are shared
// across API instances.
package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiterConfig struct {
	Window time.Duration
	Max    int
	Prefix string
}

// RateLimit implements a Redis-backed fixed-window limiter keyed by the
// authenticated user id, falling back to client IP.
func RateLimit(rdb *redis.Client, cfg RateLimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if u, ok := CurrentUser(c); ok {
			key = u.ID
		}
		redisKey := "rl:" + cfg.Prefix + ":" + key

		ctx := context.Background()
		count, err := rdb.Incr(ctx, redisKey).Result()
		if err != nil {
			// Fail open: a Redis outage should not take down the API.
			c.Next()
			return
		}
		if count == 1 {
			rdb.Expire(ctx, redisKey, cfg.Window)
		}

		ttl, _ := rdb.TTL(ctx, redisKey).Result()
		c.Header("RateLimit-Limit", strconv.Itoa(cfg.Max))
		remaining := cfg.Max - int(count)
		if remaining < 0 {
			remaining = 0
		}
		c.Header("RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("RateLimit-Reset", strconv.Itoa(int(ttl.Seconds())))

		if int(count) > cfg.Max {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMITED",
					"message": "Too many requests, please try again later.",
				},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
