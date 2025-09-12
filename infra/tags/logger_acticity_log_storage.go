package tags

import (
	"context"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/models/tags"
)

type loggerActivityLogStorage[T entityIDtype] struct{}

func (n loggerActivityLogStorage[T]) InsertCollectionTagsActivityLog(ctx context.Context, executorID uint64, withTags tags.Tags, ids ...T) error {
	pixiecontext.GetCtxLogger(ctx).
		With("entity_ids", ids).
		With("executorID", executorID).
		With("affected_rows", len(ids)).
		With("new_tags", withTags).
		Log("logging add tags collection activity")

	return nil
}

func (n loggerActivityLogStorage[T]) RemoveCollectionTagsActivityLog(ctx context.Context, executorID uint64, withTags tags.Tags, ids ...T) error {
	pixiecontext.GetCtxLogger(ctx).
		With("entity_ids", ids).
		With("executorID", executorID).
		With("affected_rows", len(ids)).
		With("new_tags", withTags).
		Log("logging remove tags collection activity")

	return nil
}

func (n loggerActivityLogStorage[T]) InsertTagsActivityLog(ctx context.Context, executorID uint64, id T, previousTags, newTags []tags.Tag) error {
	pixiecontext.GetCtxLogger(ctx).
		With("entity_id", id).
		With("executorID", executorID).
		With("previous_tags", previousTags).
		With("new_tags", newTags).
		Log("logging add tags activity")

	return nil
}

func (n loggerActivityLogStorage[T]) RemoveTagsActivityLog(ctx context.Context, executorID uint64, id T, previousTags, newTags []tags.Tag) error {
	pixiecontext.GetCtxLogger(ctx).
		With("entity_id", id).
		With("executorID", executorID).
		With("previous_tags", previousTags).
		With("new_tags", newTags).
		Log("logging remove tags activity")

	return nil
}
