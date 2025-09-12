package cache

import (
	"context"
	goErrors "errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pixie-sh/di-go"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/types"
)

type BatchSetExModel struct {
	Key      string
	Value    []byte
	Duration *time.Duration //if nil, duration is not set
}

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Peek(ctx context.Context, key string) error
	Increment(ctx context.Context, key string) (int64, error)
	Decrement(ctx context.Context, key string, by ...int64) (int64, error)
	SetEX(ctx context.Context, key string, value []byte, expiration ...time.Duration) error
	Delete(ctx context.Context, key string) error
	Scan(ctx context.Context, pattern string, rows int64, cursor ...uint64) ([]string, uint64, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
}

type BatchCache interface {
	Cache

	BatchSetEX(ctx context.Context, keyPairs []BatchSetExModel) error
	BatchDelete(ctx context.Context, key []string) error
	BatchOperations(ctx context.Context, operations []BatchOperation) error
}

type SetCache interface {
	BatchCache

	SetAdd(ctx context.Context, key string, members ...interface{}) (int64, error)
	SetMembers(ctx context.Context, key string) ([]string, error)
	SetRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	SetCard(ctx context.Context, key string) (int64, error)
	SetIsMember(ctx context.Context, key string, member interface{}) (bool, error)
}

type BatchOperationType string

const SETEX = BatchOperationType("SETEX")
const SADD = BatchOperationType("SADD")

type BatchOperation struct {
	Type     BatchOperationType
	Key      string
	Value    []byte         // For SETEX operations
	Duration *time.Duration // Optional For SETEX expiration
	Members  []interface{}  // For SADD operations
}

type RedisCacheConfiguration struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

func (r RedisCacheConfiguration) LookupNode(lookupPath string) (any, error) {
	return di.ConfigurationNodeLookup(r, lookupPath)
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(ctx context.Context, configuration RedisCacheConfiguration) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     configuration.Address,
		Password: configuration.Password,
		DB:       configuration.DB,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &RedisCache{
		client: rdb,
	}, nil
}

func (r *RedisCache) SetEX(ctx context.Context, key string, value []byte, duration ...time.Duration) error {
	return r.setEX(ctx, r.client, key, value, duration...).Err()
}

func (r *RedisCache) Keys(ctx context.Context) ([]string, error) {
	return GetCollection(ctx, r, "*", 1000, true)
}

// BatchOperations executes multiple SETEX and SADD operations in a single pipeline
func (r *RedisCache) BatchOperations(ctx context.Context, operations []BatchOperation) error {
	if len(operations) == 0 {
		return nil
	}

	pipeline := r.client.TxPipeline()

	for _, op := range operations {
		switch op.Type {
		case SETEX:
			if op.Duration != nil {
				pipeline.SetEX(ctx, op.Key, op.Value, *op.Duration)
			} else {
				pipeline.Set(ctx, op.Key, op.Value, 0)
			}
		case SADD:
			pipeline.SAdd(ctx, op.Key, op.Members...)

			// If duration is specified, set expiration on the set
			if op.Duration != nil {
				pipeline.Expire(ctx, op.Key, *op.Duration)
			}
		default:
			pixiecontext.GetCtxLogger(ctx).Error("unsupported batch operation type: %s", op.Type)
			continue
		}
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		pixiecontext.GetCtxLogger(ctx).Error("failed to execute batch operations: %v", err)
	}

	return err
}

func (r *RedisCache) BatchSetEX(ctx context.Context, keyPairs []BatchSetExModel) error {
	var pipeline = r.client.TxPipeline()
	for _, i := range keyPairs {
		if i.Duration != nil {
			r.setEX(ctx, pipeline, i.Key, i.Value, *i.Duration)
		} else {
			r.setEX(ctx, pipeline, i.Key, i.Value)
		}
	}

	_, err := pipeline.Exec(ctx)
	return err
}

func (r *RedisCache) setEX(ctx context.Context, txClient redis.Cmdable, key string, value []byte, duration ...time.Duration) *redis.StatusCmd {
	if len(duration) == 0 {
		return txClient.Set(ctx, key, value, 0)
	}

	return txClient.Set(ctx, key, value, duration[0])
}

func (r *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	res := r.client.Incr(ctx, key)
	if res.Err() != nil {
		return -2, res.Err()
	}

	return res.Val(), nil
}

func (r *RedisCache) Decrement(ctx context.Context, key string, by ...int64) (int64, error) {
	if len(by) > 0 {
		res := r.client.DecrBy(ctx, key, by[0])
		if res.Err() != nil {
			return -2, res.Err()
		}

		return res.Val(), nil
	}

	res := r.client.Decr(ctx, key)
	if res.Err() != nil {
		return -2, res.Err()
	}

	return res.Val(), nil
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	return types.UnsafeBytes(val), nil
}

func (r *RedisCache) Peek(ctx context.Context, key string) error {
	res, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return err
	}

	switch res {
	case 1:
		return nil
	default:
		return errors.New("redis: key not found")
	}
}

func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

func (r *RedisCache) Scan(ctx context.Context, pattern string, rows int64, fromCursor ...uint64) ([]string, uint64, error) {
	var (
		cursor uint64
		keys   []string
		err    error
	)

	if len(fromCursor) > 0 {
		cursor = fromCursor[0]
	}

	keys, cursor, err = r.client.Scan(ctx, cursor, pattern, rows).Result()
	if err != nil {
		logger.Logger.Error("failed to scan keys for pattern %s", pattern)
		return nil, 0, err
	}

	return keys, cursor, nil
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) BatchDelete(ctx context.Context, key []string) error {
	pipeline := r.client.TxPipeline()
	for _, i2 := range key {
		pipeline.Del(ctx, i2)
	}

	_, err := pipeline.Exec(ctx)
	return err
}

// SAdd adds one or more members to a set
func (r *RedisCache) SetAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	result, err := r.client.SAdd(ctx, key, members...).Result()
	if err != nil {
		return 0, err
	}

	return result, nil
}

// SMembers returns all members of a set
func (r *RedisCache) SetMembers(ctx context.Context, key string) ([]string, error) {
	members, err := r.client.SMembers(ctx, key).Result()
	if err != nil && !IsEmptyError(err) {
		return nil, err
	}

	return members, nil
}

// SRem removes one or more members from a set
func (r *RedisCache) SetRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	result, err := r.client.SRem(ctx, key, members...).Result()
	if err != nil {
		return 0, err
	}

	return result, nil
}

// SCard returns the number of members in a set
func (r *RedisCache) SetCard(ctx context.Context, key string) (int64, error) {
	count, err := r.client.SCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	return count, nil
}

// SIsMember checks if a member exists in a set
func (r *RedisCache) SetIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	exists, err := r.client.SIsMember(ctx, key, member).Result()
	if err != nil && !IsEmptyError(err) {
		return false, err
	}

	return exists, nil
}

func IsEmptyError(err error) bool {
	return goErrors.Is(err, redis.Nil)
}
