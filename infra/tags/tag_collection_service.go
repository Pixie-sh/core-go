package tags

import (
	"context"

	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/infra/events"
	pixieerror "github.com/pixie-sh/core-go/pkg/errors"
	"github.com/pixie-sh/core-go/pkg/models/tags"
	"github.com/pixie-sh/core-go/pkg/types"
)

type TagCollectionResult[T entityIDtype] struct {
	Tags []tags.EntityTag[T]
}

type TagCollectionExecResult[T entityIDtype] struct {
	RowsAffected uint64
	Events       []events.UntypedEventWrapper
	AffectedIDs  []T
}

type TagCollectionChecker[T entityIDtype] interface {
	Get(ctx context.Context, idGetters ...T) (TagCollectionResult[T], error)
}

type TagCollectionMutator[T entityIDtype] interface {
	TagCollectionChecker[T]

	Add(ctx context.Context, tags tags.Tags, idGetters ...T) (TagCollectionExecResult[T], error)
	Remove(ctx context.Context, tags tags.Tags, idGetters ...T) (TagCollectionExecResult[T], error)
}

type tagCollectionService[T entityIDtype] struct {
	*TagCollectionServiceConfiguration[T]
}

func (t *tagCollectionService[T]) Get(ctx context.Context, ids ...T) (TagCollectionResult[T], error) {
	res, err := t.Repository.Get(ctx, ids...)
	if err != nil {
		return TagCollectionResult[T]{}, err
	}

	return TagCollectionResult[T]{
		Tags: res,
	}, nil
}

func (t *tagCollectionService[T]) Add(ctx context.Context, tagz tags.Tags, ids ...T) (TagCollectionExecResult[T], error) {
	for _, tag := range tagz {
		_, _, err := tags.HasScope(tag, t.ManagedScope)
		if err != nil {
			return TagCollectionExecResult[T]{}, err
		}
	}

	res, err := t.Repository.Add(ctx, tagz, ids...)
	if err != nil {
		return TagCollectionExecResult[T]{}, err
	}

	err = t.ActivityLog.InsertCollectionTagsActivityLog(ctx, t.ExecutorUserID, tagz, ids...)
	if err != nil {
		return TagCollectionExecResult[T]{}, err
	}

	return TagCollectionExecResult[T]{
		RowsAffected: uint64(len(res)),
		AffectedIDs:  res,
		Events:       t.GenerateEvents(ctx, res, tagz),
	}, nil
}

func (t *tagCollectionService[T]) Remove(ctx context.Context, tagz tags.Tags, ids ...T) (TagCollectionExecResult[T], error) {
	for _, tag := range tagz {
		_, _, err := tags.HasScope(tag, t.ManagedScope)
		if err != nil {
			return TagCollectionExecResult[T]{}, err
		}
	}

	res, err := t.Repository.Remove(ctx, tagz, ids...)
	if err != nil {
		return TagCollectionExecResult[T]{}, err
	}

	err = t.ActivityLog.RemoveCollectionTagsActivityLog(ctx, t.ExecutorUserID, tagz, res...)
	if err != nil {
		return TagCollectionExecResult[T]{}, err
	}

	return TagCollectionExecResult[T]{
		RowsAffected: uint64(len(res)),
		AffectedIDs:  res,
		Events:       t.GenerateEvents(ctx, res, tagz),
	}, nil
}

type EntityCollectionActivityLogStorage[T entityIDtype] interface {
	InsertCollectionTagsActivityLog(ctx context.Context, executorID uint64, withTags tags.Tags, ids ...T) error
	RemoveCollectionTagsActivityLog(ctx context.Context, executorID uint64, withTags tags.Tags, ids ...T) error
}

type EntityCollectionStorage[T entityIDtype] interface {
	Get(ctx context.Context, ids ...T) ([]tags.EntityTag[T], error)
	Add(ctx context.Context, withTags tags.Tags, ids ...T) ([]T, error)
	Remove(ctx context.Context, withTags tags.Tags, ids ...T) ([]T, error)
}

type TagCollectionServiceConfiguration[T entityIDtype] struct {
	ExecutorUserID uint64 //TODO move to uid.UID when
	ManagedScope   tags.TagScope
	Repository     EntityCollectionStorage[T]
	ActivityLog    EntityCollectionActivityLogStorage[T]
	GenerateEvents func(ctx context.Context, ids []T, tags tags.Tags) []events.UntypedEventWrapper
}

func NewTagCollectionService[T entityIDtype](_ context.Context, config TagCollectionServiceConfiguration[T]) (TagCollectionMutator[T], error) {
	if config.ManagedScope == "" {
		return nil, errors.New("managed scope is empty", pixieerror.TagsDependencyErrorCode)
	}

	if config.ExecutorUserID == 0 {
		return nil, errors.New("executor user id is empty", pixieerror.TagsDependencyErrorCode)
	}

	if types.Nil(config.Repository) {
		return nil, errors.New("repository is nil", pixieerror.TagsDependencyErrorCode)
	}

	if types.Nil(config.ActivityLog) {
		config.ActivityLog = &loggerActivityLogStorage[T]{}
	}

	if types.Nil(config.GenerateEvents) {
		config.GenerateEvents = func(context.Context, []T, tags.Tags) []events.UntypedEventWrapper { return nil }
	}

	return &tagCollectionService[T]{
		TagCollectionServiceConfiguration: &config,
	}, nil
}
