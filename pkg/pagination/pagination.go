package pagination

import (
	"context"
	"fmt"

	"github.com/pixie-sh/logger-go/logger"
)

func PaginatedFunctionProcessor[T any](
	ctx context.Context,
	fetchPageFunc func(ctx context.Context, page uint64, batchSize uint64) ([]T, error),
	processBatchFunc func(ctx context.Context, items []T) error,
	batchSize uint64,
	logger logger.Interface,
) error {
	var page uint64 = 0

	for {
		// Fetch a page of data
		items, err := fetchPageFunc(ctx, page, batchSize)
		if err != nil {
			if logger != nil {
				logger.With("error", err).Error("error fetching data")
			}
			return fmt.Errorf("error fetching data at page %d: %w", page, err)
		}

		if logger != nil {
			logger.Log("processing %v items for page %d", len(items), page+1)
		}

		// Process the batch
		if len(items) > 0 {
			err = processBatchFunc(ctx, items)
			if err != nil {
				if logger != nil {
					logger.With("error", err).Error("failed to process batch at page %d", page)
				}
				return fmt.Errorf("error processing batch at page %d: %w", page, err)
			}
		}

		// Increment page and check if we need to continue
		page++
		if len(items) < int(batchSize) {
			if logger != nil {
				logger.Log("no more items to process, ending...")
			}
			break
		}
	}

	return nil
}
