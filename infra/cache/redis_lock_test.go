package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/pixie-sh/errors-go"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedisLocker(t *testing.T) (*RedisLocker, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := RedisLockConfiguration{
		DefaultExpiration: 30 * time.Second,
		DefaultRetryDelay: 100 * time.Millisecond,
		MaxRetries:        3,
	}

	locker, err := NewRedisLock(context.Background(), client, config)
	require.NoError(t, err)

	return locker, mr
}

func TestRedisLocker_Lock(t *testing.T) {
	locker, mr := setupTestRedisLocker(t)
	defer mr.Close()

	t.Run("successful lock", func(t *testing.T) {
		ctx := context.Background()
		lock, err := locker.Lock(ctx, "test_key")
		assert.NoError(t, err)
		assert.NotNil(t, lock)

		// Verify that the key exists in Redis
		assert.True(t, mr.Exists("test_key"))
	})

	t.Run("lock already held", func(t *testing.T) {
		ctx := context.Background()
		// First lock
		lock1, err := locker.Lock(ctx, "test_key2")
		assert.NoError(t, err)
		assert.NotNil(t, lock1)

		// Try to acquire the same lock
		lock2, err := locker.Lock(ctx, "test_key2")
		assert.Error(t, err)
		assert.Nil(t, lock2)

		e, ok := errors.Has(err, errors.FailedToAcquireLockErrorCode)
		assert.True(t, ok)
		assert.NotNil(t, e)
	})

	t.Run("lock with custom duration", func(t *testing.T) {
		ctx := context.Background()
		customDuration := 5 * time.Second
		lock, err := locker.Lock(ctx, "test_key3", customDuration)
		assert.NoError(t, err)
		assert.NotNil(t, lock)

		// Verify that the key exists in Redis with the correct TTL
		ttl := mr.TTL("test_key3")
		assert.True(t, ttl > 4*time.Second && ttl <= 5*time.Second)
	})
}

func TestRedisLock_Unlock(t *testing.T) {
	locker, mr := setupTestRedisLocker(t)
	defer mr.Close()

	t.Run("successful unlock", func(t *testing.T) {
		ctx := context.Background()
		lock, err := locker.Lock(ctx, "test_key")
		require.NoError(t, err)
		require.NotNil(t, lock)

		err = lock.Unlock()
		assert.NoError(t, err)

		// Verify that the key no longer exists in Redis
		assert.False(t, mr.Exists("test_key"))
	})

	t.Run("unlock already unlocked lock", func(t *testing.T) {
		ctx := context.Background()
		lock, err := locker.Lock(ctx, "test_key")
		require.NoError(t, err)
		require.NotNil(t, lock)

		err = lock.Unlock()
		assert.NoError(t, err)

		// Try to unlock again
		err = lock.Unlock()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lock not held by this instance")
	})
}

func TestRedisLock_Refresh(t *testing.T) {
	locker, mr := setupTestRedisLocker(t)
	defer mr.Close()

	t.Run("successful refresh", func(t *testing.T) {
		ctx := context.Background()
		lock, err := locker.Lock(ctx, "test_key", 1*time.Second)
		require.NoError(t, err)
		require.NotNil(t, lock)

		// Wait a bit
		time.Sleep(500 * time.Millisecond)

		// Refresh the lock
		refreshed, err := lock.Refresh()
		assert.NoError(t, err)
		assert.True(t, refreshed)

		// Verify that the key still exists in Redis with an extended TTL
		assert.True(t, mr.Exists("test_key"))
		ttl := mr.TTL("test_key")
		assert.True(t, ttl > 500*time.Millisecond)
	})

	t.Run("refresh unlocked lock", func(t *testing.T) {
		ctx := context.Background()
		lock, err := locker.Lock(ctx, "test_key2")
		require.NoError(t, err)
		require.NotNil(t, lock)

		err = lock.Unlock()
		assert.NoError(t, err)

		// Try to refresh the unlocked lock
		refreshed, err := lock.Refresh()
		assert.Error(t, err)
		assert.False(t, refreshed)
		assert.Contains(t, err.Error(), "lock not held by this instance")
	})
}

func TestRedisLocker_ConcurrentLocks(t *testing.T) {
	locker, mr := setupTestRedisLocker(t)
	defer mr.Close()

	t.Run("concurrent locks on different keys", func(t *testing.T) {
		ctx := context.Background()
		lock1, err := locker.Lock(ctx, "key1")
		assert.NoError(t, err)
		assert.NotNil(t, lock1)

		lock2, err := locker.Lock(ctx, "key2")
		assert.NoError(t, err)
		assert.NotNil(t, lock2)

		// Verify that both keys exist in Redis
		assert.True(t, mr.Exists("key1"))
		assert.True(t, mr.Exists("key2"))

		// Unlock both
		err = lock1.Unlock()
		assert.NoError(t, err)
		err = lock2.Unlock()
		assert.NoError(t, err)

		// Verify that both keys no longer exist in Redis
		assert.False(t, mr.Exists("key1"))
		assert.False(t, mr.Exists("key2"))
	})
}

func TestRedisLocker_Concurrency(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	require.NotNil(t, redisClient)

	config := RedisLockConfiguration{
		DefaultExpiration: 20 * time.Millisecond,
		DefaultRetryDelay: 50 * time.Millisecond,
		MaxRetries:        50,
	}

	locker, err := NewRedisLock(context.Background(), redisClient, config)
	require.NoError(t, err)
	require.NotNil(t, locker)

	const routines = 10
	const lockKey = "concurrency-test-key"
	var lockedCount []int

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx := context.Background()
			fmt.Printf("%d acquring dist lock...\n", i)
			now := time.Now()
			lock, err := locker.Lock(ctx, lockKey)
			if err != nil {
				fmt.Printf("%d unable to acquire lock; elapsed %d \n", i, time.Since(now).Milliseconds())
				return
			}

			mu.Lock()
			lockedCount = append(lockedCount, i)
			mu.Unlock()

			fmt.Printf("%d waiting...\n", i)
			time.Sleep(20 * time.Millisecond)

			fmt.Printf("%d unlocking...\n", i)
			err = lock.Unlock()
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	assert.Equal(t, routines, len(lockedCount))
}
