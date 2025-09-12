package cache

import (
	"context"
	goErrors "errors"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/pixie-sh/errors-go"
	"github.com/redis/go-redis/v9"
)

// SharedLocker defines the interface for a distributed lock
type SharedLocker interface {
	Lock(ctx context.Context, key string, lockDuration ...time.Duration) (SharedLock, error)
}

// SharedLock defines the interface for a distributed lock
type SharedLock interface {
	Unlock() error
	Refresh() (bool, error)
}

// RedisLockConfiguration holds the configuration for RedisLocker
type RedisLockConfiguration struct {
	DefaultExpiration time.Duration `json:"default_expiration"`
	DefaultRetryDelay time.Duration `json:"default_retry_delay"`
	MaxRetries        int           `json:"max_retries"`
}

// RedisLocker implements the SharedLocker interface using redsync.Redsync
type RedisLocker struct {
	rs     *redsync.Redsync
	config RedisLockConfiguration
}

// RedisLock implements the SharedLock interface using redsync.Mutex
type RedisLock struct {
	mutex *redsync.Mutex
	ctx   context.Context
}

// NewRedisLock creates a new RedisLocker instance
func NewRedisLock(_ context.Context, client *redis.Client, config RedisLockConfiguration) (*RedisLocker, error) {
	if client == nil {
		return nil, errors.New("redis client is nil")
	}

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	return &RedisLocker{
		rs:     rs,
		config: config,
	}, nil
}

// Lock implements the SharedLocker.Lock method
func (rl *RedisLocker) Lock(_ context.Context, key string, lockDuration ...time.Duration) (SharedLock, error) {
	duration := rl.config.DefaultExpiration
	if len(lockDuration) > 0 {
		duration = lockDuration[0]
	}

	mutex := rl.rs.NewMutex(key,
		redsync.WithExpiry(duration),
		redsync.WithTries(rl.config.MaxRetries),
		redsync.WithRetryDelay(rl.config.DefaultRetryDelay),
	)

	lockCtx := context.Background()
	err := mutex.LockContext(lockCtx)
	if err != nil {
		if goErrors.Is(err, redsync.ErrFailed) {
			return nil, errors.New("lock failed to acquire").WithErrorCode(errors.FailedToAcquireLockErrorCode)
		}

		return nil, errors.Wrap(err, "%s", err.Error()).WithErrorCode(errors.FailedToAcquireLockErrorCode)
	}

	return &RedisLock{
		mutex: mutex,
		ctx:   lockCtx,
	}, nil
}

// Unlock implements the SharedLocker.Unlock method
func (rl *RedisLock) Unlock() error {
	if rl.mutex == nil {
		return errors.New("lock not held by this instance")
	}

	ok, err := rl.mutex.UnlockContext(rl.ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("failed to release the lock")
	}

	rl.mutex = nil
	return nil
}

// Refresh implements the SharedLocker.Refresh method
func (rl *RedisLock) Refresh() (bool, error) {
	if rl.mutex == nil {
		return false, errors.New("lock not held by this instance")
	}

	ok, err := rl.mutex.ExtendContext(rl.ctx)
	if err != nil {
		return false, err
	}

	return ok, nil
}
