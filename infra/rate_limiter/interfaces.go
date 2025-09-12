package rate_limiter

import (
	"context"
	"time"
)

type ICache interface {
	SetEX(ctx context.Context, key string, value []byte, expiration ...time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	TTL(ctx context.Context, key string) (time.Duration, error)
}

type IRequestRateLimiter interface {
	AllowRequest(ctx context.Context, uuid string, requestType string, customTTL ...time.Duration) bool
}

type IRateLimiter interface {
	Try(ctx context.Context, key string) (bool, time.Duration)
	Increment(ctx context.Context, key string, customTTL ...time.Duration) (int, error)
}

type ILimitRateLimiter interface {
	TryLimit(ctx context.Context, key string, limit int, customTTL ...time.Duration) (bool, time.Duration)
	Increment(ctx context.Context, key string, customTTL ...time.Duration) (int, error)
	GetLimit() int
	GetWindowSize() time.Duration
}
