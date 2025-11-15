package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"user-services/internal/config"
	"user-services/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	Requests           int           // Number of requests allowed
	Window             time.Duration // Time window for rate limiting
	AccountRequests    int           // Number of requests allowed per account
	AccountWindow      time.Duration // Time window for account rate limiting
	ProgressiveBackoff bool          // Enable progressive backoff
}

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed   bool
	Remaining int
	ResetTime time.Time
	Reason    string
}

// RateLimiter interface for rate limiting implementations
type RateLimiter interface {
	CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (*RateLimitResult, error)
	CheckAccountRateLimit(ctx context.Context, email string, failedAttempts int) (*RateLimitResult, error)
	RecordFailedAttempt(ctx context.Context, email string) error
	ResetFailedAttempts(ctx context.Context, email string) error
}

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	client *redis.Client
	config *config.Config
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client, cfg *config.Config) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
		config: cfg,
	}
}

// CheckRateLimit checks if a request is allowed based on IP rate limiting
func (r *RedisRateLimiter) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (*RateLimitResult, error) {
	now := time.Now()
	windowStart := now.Add(-window)

	// Use sliding window algorithm with Redis sorted set
	pipe := r.client.Pipeline()

	// Remove old entries outside the window
	removeCmd := pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.UnixNano(), 10))

	// Count current requests in window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: now.UnixNano(),
	})

	// Set expiration on the key
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute rate limit pipeline: %w", err)
	}

	// Check results
	currentCount := countCmd.Val()
	remaining := limit - int(currentCount)

	// If we're at or over the limit, remove the current request we just added
	if currentCount > int64(limit) {
		r.client.ZRem(ctx, key, now.UnixNano())
		remaining = 0
	}

	result := &RateLimitResult{
		Allowed:   currentCount <= int64(limit),
		Remaining: max(0, remaining-1), // Subtract 1 for current request
		ResetTime: now.Add(window),
		Reason:    "ip_rate_limit",
	}

	// Clean up expired entries periodically
	if removeCmd.Val() > 0 {
		r.client.Expire(ctx, key, window)
	}

	return result, nil
}

// CheckAccountRateLimit checks account-specific rate limiting with progressive backoff
func (r *RedisRateLimiter) CheckAccountRateLimit(ctx context.Context, email string, failedAttempts int) (*RateLimitResult, error) {
	failedKey := fmt.Sprintf("failed_attempts:%s", email)
	blockKey := fmt.Sprintf("account_block:%s", email)

	// Check if account is currently blocked
	blocked, err := r.client.Exists(ctx, blockKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check account block: %w", err)
	}

	if blocked > 0 {
		ttl, _ := r.client.TTL(ctx, blockKey).Result()
		return &RateLimitResult{
			Allowed:   false,
			Remaining: 0,
			ResetTime: time.Now().Add(ttl),
			Reason:    "account_blocked",
		}, nil
	}

	// Get current failed attempts
	attempts, err := r.client.Get(ctx, failedKey).Int()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get failed attempts: %w", err)
	}

	if err == redis.Nil {
		attempts = 0
	}

	// Calculate progressive rate limits
	maxRequests, window := r.getProgressiveRateLimit(attempts)

	// Create rate limit key for this account
	rateLimitKey := fmt.Sprintf("account_rate_limit:%s", email)

	// Check rate limit for this account
	result, err := r.CheckRateLimit(ctx, rateLimitKey, maxRequests, window)
	if err != nil {
		return nil, err
	}

	result.Reason = "account_rate_limit"

	// If failed attempts exceed threshold, block the account
	if attempts >= 5 {
		blockDuration := r.getBlockDuration(attempts)
		r.client.SetEx(ctx, blockKey, "1", blockDuration)

		return &RateLimitResult{
			Allowed:   false,
			Remaining: 0,
			ResetTime: time.Now().Add(blockDuration),
			Reason:    "account_blocked_too_many_failures",
		}, nil
	}

	return result, nil
}

// RecordFailedAttempt records a failed authentication attempt
func (r *RedisRateLimiter) RecordFailedAttempt(ctx context.Context, email string) error {
	key := fmt.Sprintf("failed_attempts:%s", email)

	pipe := r.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 24*time.Hour) // Keep failed attempts for 24 hours

	_, err := pipe.Exec(ctx)
	return err
}

// ResetFailedAttempts resets failed attempts after successful login
func (r *RedisRateLimiter) ResetFailedAttempts(ctx context.Context, email string) error {
	keys := []string{
		fmt.Sprintf("failed_attempts:%s", email),
		fmt.Sprintf("account_block:%s", email),
		fmt.Sprintf("account_rate_limit:%s", email),
	}

	return r.client.Del(ctx, keys...).Err()
}

// getProgressiveRateLimit returns rate limits based on failed attempts
func (r *RedisRateLimiter) getProgressiveRateLimit(failedAttempts int) (int, time.Duration) {
	switch {
	case failedAttempts < 3:
		return 10, time.Minute // 10 requests per minute
	case failedAttempts < 5:
		return 5, time.Minute  // 5 requests per minute
	case failedAttempts < 10:
		return 2, time.Minute  // 2 requests per minute
	default:
		return 1, 5 * time.Minute // 1 request per 5 minutes
	}
}

// getBlockDuration returns block duration based on failed attempts
func (r *RedisRateLimiter) getBlockDuration(failedAttempts int) time.Duration {
	switch {
	case failedAttempts < 10:
		return 15 * time.Minute
	case failedAttempts < 20:
		return 1 * time.Hour
	default:
		return 24 * time.Hour
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter RateLimiter, config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s:%s", clientIP, c.Request.URL.Path)

		result, err := limiter.CheckRateLimit(c.Request.Context(), key, config.Requests, config.Window)
		if err != nil {
			// Log error but don't block requests on Redis failures
			if !limiter.(*RedisRateLimiter).config.IsProduction() {
				response.InternalServerError(c, "Rate limiting service unavailable")
				c.Abort()
				return
			}
			// In production, allow requests to proceed if rate limiting fails
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

		if !result.Allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(result.ResetTime.Sub(time.Now()).Seconds()), 10))

			response.TooManyRequests(c, "Too many requests. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// AccountRateLimitMiddleware creates account-specific rate limiting middleware
func AccountRateLimitMiddleware(limiter RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to extract email from request body for POST requests
		var email string

		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			// For login and register endpoints, extract email from request
			if c.Request.URL.Path == "/api/v1/users/login" || c.Request.URL.Path == "/api/v1/users/register" {
				type authRequest struct {
					Email string `json:"email"`
				}
				var req authRequest
				if err := c.ShouldBindJSON(&req); err == nil {
					email = req.Email
				}
			}
		}

		if email == "" {
			// If we can't extract email, skip account rate limiting
			c.Next()
			return
		}

		// Get failed attempts count
		failedKey := fmt.Sprintf("failed_attempts:%s", email)
		attempts, _ := limiter.(*RedisRateLimiter).client.Get(c.Request.Context(), failedKey).Int()

		result, err := limiter.CheckAccountRateLimit(c.Request.Context(), email, attempts)
		if err != nil {
			// Log error but don't block requests on Redis failures
			if !limiter.(*RedisRateLimiter).config.IsProduction() {
				response.InternalServerError(c, "Rate limiting service unavailable")
				c.Abort()
				return
			}
			c.Next()
			return
		}

		if !result.Allowed {
			var message string
			switch result.Reason {
			case "account_blocked":
				message = "Account temporarily locked due to too many failed attempts. Please try again later."
			case "account_blocked_too_many_failures":
				message = "Account locked due to suspicious activity. Please contact support or try again later."
			default:
				message = "Too many attempts for this account. Please try again later."
			}

			c.Header("Retry-After", strconv.FormatInt(int64(result.ResetTime.Sub(time.Now()).Seconds()), 10))
			response.TooManyRequests(c, message)
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthRateLimitMiddleware combines IP and account rate limiting for auth endpoints
func AuthRateLimitMiddleware(limiter RateLimiter, ipConfig RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Apply IP-based rate limiting
		clientIP := c.ClientIP()
		ipKey := fmt.Sprintf("auth_rate_limit:%s:%s", clientIP, c.Request.URL.Path)

		ipResult, err := limiter.CheckRateLimit(c.Request.Context(), ipKey, ipConfig.Requests, ipConfig.Window)
		if err != nil {
			if !limiter.(*RedisRateLimiter).config.IsProduction() {
				response.InternalServerError(c, "Rate limiting service unavailable")
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// Set rate limit headers for IP-based limiting
		c.Header("X-RateLimit-Limit", strconv.Itoa(ipConfig.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(ipResult.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(ipResult.ResetTime.Unix(), 10))

		if !ipResult.Allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(ipResult.ResetTime.Sub(time.Now()).Seconds()), 10))
			response.TooManyRequests(c, "Too many authentication attempts. Please try again later.")
			c.Abort()
			return
		}

		// Apply account-based rate limiting if we can extract email
		var email string
		if c.Request.Method == "POST" {
			if c.Request.URL.Path == "/api/v1/users/login" || c.Request.URL.Path == "/api/v1/users/register" {
				type authRequest struct {
					Email string `json:"email"`
				}
				var req authRequest
				if err := c.ShouldBindJSON(&req); err == nil {
					email = req.Email
				}
			} else if c.Request.URL.Path == "/api/v1/password/reset/request" {
				type resetRequest struct {
					Email string `json:"email"`
				}
				var req resetRequest
				if err := c.ShouldBindJSON(&req); err == nil {
					email = req.Email
				}
			}
		}

		if email != "" {
			failedKey := fmt.Sprintf("failed_attempts:%s", email)
			attempts, _ := limiter.(*RedisRateLimiter).client.Get(c.Request.Context(), failedKey).Int()

			accountResult, err := limiter.CheckAccountRateLimit(c.Request.Context(), email, attempts)
			if err != nil {
				if !limiter.(*RedisRateLimiter).config.IsProduction() {
					response.InternalServerError(c, "Rate limiting service unavailable")
					c.Abort()
					return
				}
				c.Next()
				return
			}

			if !accountResult.Allowed {
				var message string
				switch accountResult.Reason {
				case "account_blocked":
					message = "Account temporarily locked due to too many failed attempts. Please try again later."
				case "account_blocked_too_many_failures":
					message = "Account locked due to suspicious activity. Please contact support or try again later."
				default:
					message = "Too many attempts for this account. Please try again later."
				}

				c.Header("Retry-After", strconv.FormatInt(int64(accountResult.ResetTime.Sub(time.Now()).Seconds()), 10))
				response.TooManyRequests(c, message)
				c.Abort()
				return
			}
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