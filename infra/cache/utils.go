package cache

import (
	"context"
	goErrors "errors"
	"strconv"
	"time"

	"github.com/pixie-sh/errors-go"
	"github.com/redis/go-redis/v9"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"

	"github.com/pixie-sh/logger-go/logger"
)

func GetInt64(value interface{}, err error) (int64, error) {
	if err != nil {
		if goErrors.Is(err, redis.Nil) {
			return 0, nil
		}

		return 0, err
	}

	countStr, ok := value.(string)
	if !ok {
		return 0, errors.New("invalid int64 value")
	}

	return strconv.ParseInt(countStr, 10, 64)
}

func GetCollection(ctx context.Context, cache Cache, pattern string, rows int64, untilEnd ...bool) ([]string, error) {
	var allKeys []string
	var cursor uint64 = 0
	var err error

	// Always get at least the first batch
	firstKeys, firstCursor, err := cache.Scan(ctx, pattern, rows)
	if err != nil {
		if goErrors.Is(err, redis.Nil) {
			return []string{}, nil
		}
		return nil, err
	}

	allKeys = append(allKeys, firstKeys...)
	cursor = firstCursor

	// Continue scanning if we have a non-zero cursor and untilEnd is true
	if cursor > 0 && (len(untilEnd) > 0 && untilEnd[0]) {
		timeout := 45 * time.Second
		startTime := time.Now()

		for cursor > 0 {
			if time.Since(startTime) > timeout {
				logger.Logger.
					Warn("scan operation timed out after %v for pattern %s; returning gathered keys.", time.Since(startTime), pattern)
				break
			}

			moreKeys, nextCursor, err := cache.Scan(ctx, pattern, rows, cursor)
			if err != nil {
				logger.Logger.
					With("error", err).
					Error("failed to scan keys for pattern %s; cursor %d", pattern, cursor)
				return nil, err
			}

			pixiecontext.GetCtxLogger(ctx).With("keys", moreKeys).Debug("found keys for cursor %d", cursor)
			allKeys = append(allKeys, moreKeys...)
			cursor = nextCursor
		}
	}

	return allKeys, nil
}
