package rate_limiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pixie-sh/core-go/pkg/types"

	"github.com/pixie-sh/core-go/infra/cache"

	"github.com/pixie-sh/logger-go/logger"
)

type RateLimiterConfiguration struct {
	Limit              int    `json:"limit"`
	WindowSizeDuration string `json:"window_size_duration"`
}

type RateLimiter struct {
	cache      ICache
	limit      int
	windowSize time.Duration
}

func NewRateLimiter(cache ICache, config RateLimiterConfiguration) (*RateLimiter, error) {
	dur, err := time.ParseDuration(config.WindowSizeDuration)
	if err != nil {
		return nil, err
	}

	return &RateLimiter{
		cache:      cache,
		limit:      config.Limit,
		windowSize: dur,
	}, nil
}

// AllowRequest check and increments occurrences if allowed
func (rl *RateLimiter) AllowRequest(ctx context.Context, uuid string, requestType string, customTTL ...time.Duration) bool {
	key := fmt.Sprintf("%s:%s", uuid, requestType)
	count, err := rl.getCount(ctx, key)
	if err != nil {
		logger.Logger.Warn("error get count from cache; %s", err)
		return false
	}

	if count <= rl.limit {
		if err := rl.incrementCount(ctx, key, customTTL...); err != nil {
			logger.Logger.Warn("error increment from cache; %s", err)
			return false
		}

		return true
	}

	return false
}

// Increment Occurrences of pair: uuid+requestType
func (rl *RateLimiter) Increment(ctx context.Context, key string, customTTL ...time.Duration) (int, error) {
	err := rl.incrementCount(ctx, key, customTTL...)
	if err != nil {
		logger.Logger.Warn("error get count from cache; %s", err)
		return -1, err
	}

	count, err := rl.getCount(ctx, key)
	if err != nil {
		logger.Logger.Warn("error get count from cache; %s", err)
		return -1, err
	}

	return count, nil
}

// Try check if the key occurrence is below or equal the threshold
func (rl *RateLimiter) Try(ctx context.Context, key string) (bool, time.Duration) {
	count, err := rl.getCount(ctx, key)
	if err != nil {
		logger.Logger.Warn("error get count from cache, returning default %s; %s", rl.windowSize, err)
		return false, rl.windowSize
	}

	allowed := count <= rl.limit
	if allowed {
		return true, rl.windowSize
	}

	ttl, err := rl.cache.TTL(ctx, key)
	if err != nil {
		logger.Logger.Warn("error get TTL from cache, returning default %s; %s", rl.windowSize, err)
		return false, rl.windowSize
	}

	return allowed, ttl
}

func (rl *RateLimiter) TryLimit(ctx context.Context, key string, limit int, customTTL ...time.Duration) (bool, time.Duration) {
	windowSize := rl.windowSize
	if len(customTTL) > 0 {
		windowSize = customTTL[0]
	}

	count, err := rl.getCount(ctx, key)
	if err != nil {
		logger.Logger.Warn("error get count from cache, returning default %s; %s", windowSize, err)
		return false, windowSize
	}

	allowed := count <= limit
	if allowed {
		return true, windowSize
	}

	ttl, err := rl.cache.TTL(ctx, key)
	if err != nil {
		logger.Logger.Warn("error get TTL from cache, returning default %s; %s", windowSize, err)
		return false, windowSize
	}

	return allowed, ttl
}

func (rl *RateLimiter) GetLimit() int {
	return rl.limit
}

func (rl *RateLimiter) GetWindowSize() time.Duration {
	return rl.windowSize
}

func (rl *RateLimiter) getCount(ctx context.Context, key string) (int, error) {
	value, err := rl.cache.Get(ctx, key)
	if err != nil {
		if cache.IsEmptyError(err) {
			return 0, nil
		}

		return 0, err
	}

	return strconv.Atoi(string(value))
}

func (rl *RateLimiter) incrementCount(ctx context.Context, key string, customTTL ...time.Duration) error {
	ttl := rl.windowSize
	if len(customTTL) > 0 {
		ttl = customTTL[0]
	}

	currentCount, _ := rl.getCount(ctx, key)
	return rl.cache.SetEX(ctx, key, types.UnsafeBytes(strconv.Itoa(currentCount+1)), ttl)
}
