package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// Lua script to implement atomic sliding-window rate limiting in Redis
const rateLimitLua = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

local clearBefore = now - window
redis.call('zremrangebyscore', key, 0, clearBefore)
local current_requests = redis.call('zcard', key)

if current_requests < limit then
    redis.call('zadd', key, now, now)
    redis.call('expire', key, math.ceil(window / 1000))
    return 1
else
    return 0
end
`

func RateLimiter(rdb *redis.Client, maxRequests int, windowDuration time.Duration) func(http.Handler) http.Handler {
	script := redis.NewScript(rateLimitLua)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			// Key based on IP
			key := "rate_limit:" + ip
			nowMs := time.Now().UnixNano() / int64(time.Millisecond)
			windowMs := windowDuration.Milliseconds()

			ctx := context.Background()
			res, err := script.Run(ctx, rdb, []string{key}, nowMs, windowMs, maxRequests).Int()
			if err != nil {
				slog.Error("Redis rate limit check failed", "error", err, "ip", ip)
				// Proceed on redis failure to avoid blocking traffic, but log the error
				next.ServeHTTP(w, r)
				return
			}

			if res == 0 {
				slog.Warn("Rate limit exceeded", "ip", ip, "path", r.URL.Path)
				http.Error(w, `{"error": "Too Many Requests", "message": "Rate limit exceeded. Please try again later."}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
