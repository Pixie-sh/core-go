package tags

import (
	"context"
	"sync"

	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/infra/events"
	pixieerrors "github.com/pixie-sh/core-go/pkg/errors"
	"github.com/pixie-sh/core-go/pkg/models/tags"
	"github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/types/sets"
)

type entityIDtype interface{}

type TagResult[T entityIDtype] struct {
	Scope  tags.TagScope
	ID     T
	Tags   []tags.Tag
	Events []events.UntypedEventWrapper
}

// TagChecker defines interface for checking tag-related information
type TagChecker[T entityIDtype] interface {
	Tags(ctx context.Context) (TagResult[T], error)
	Contains(ctx context.Context, tags ...tags.Tag) (bool, error)
}

// TagMutator extends TagChecker with methods for modifying tags
type TagMutator[T entityIDtype] interface {
	TagChecker[T]

	Add(ctx context.Context, tags ...tags.Tag) error
	Remove(ctx context.Context, tags ...tags.Tag) error
	Save(ctx context.Context) (TagResult[T], error)
}

// EntityStorage defines the interface for persisting entity tags
// most likely implemented by a repository, transactional or not is the implementer responsibility
type EntityStorage[T entityIDtype] interface {
	GetTags(ctx context.Context, id T) ([]tags.Tag, error)
	UpsertTags(ctx context.Context, id T, tags ...tags.Tag) error
}

// EntityActivityLogStorage defines the interface for logging tag changes
type EntityActivityLogStorage[T entityIDtype] interface {
	InsertTagsActivityLog(ctx context.Context, executorID uint64, id T, previousTags, newTags []tags.Tag) error
}

// tagService implements TagMutator interface and provides thread-safe operations for managing entity tags.
// It supports both mutable and immutable tags, with optional activity logging and event emission capabilities.
type tagService[T entityIDtype] struct {
	mu          sync.RWMutex
	mutableTags sets.Set[tags.Tag] // Current set of tags that can be modified

	managedScope       tags.TagScope
	immutableTags      []tags.Tag // Original tags that cannot be modified
	executorUserID     uint64
	entityID           T                           // Unique identifier for the entity
	entityStorage      EntityStorage[T]            // Storage for persisting entity tags
	activityLogStorage EntityActivityLogStorage[T] // Optional storage for logging tag changes
	generateEvents     func(
		ctx context.Context,
		entityID T,
		tags []tags.Tag,
	) []events.UntypedEventWrapper // Optional event emitter for when Save occurs; called within locked context
}

// Save persists the current tags and logs the changes if activity logging is enabled
// calls emitEvent after both storages execute without error
func (t *tagService[T]) Save(ctx context.Context) (TagResult[T], error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	slice := sets.Slice(t.mutableTags)
	err := t.entityStorage.UpsertTags(ctx, t.entityID, slice...)
	if err != nil {
		return TagResult[T]{}, err
	}

	err = t.activityLogStorage.InsertTagsActivityLog(ctx, t.executorUserID, t.entityID, t.immutableTags, sets.Slice(t.mutableTags))
	if err != nil {
		return TagResult[T]{}, err
	}

	return TagResult[T]{
		Scope:  t.managedScope,
		ID:     t.entityID,
		Tags:   slice,
		Events: t.generateEvents(ctx, t.entityID, slice),
	}, nil
}

// Add adds new tags to the mutable tag set
func (t *tagService[T]) Add(_ context.Context, tagz ...tags.Tag) error {
	for _, tag := range tagz {
		_, _, err := tags.HasScope(tag, t.managedScope)
		if err != nil {
			return err
		}
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, tag := range tagz {
		t.mutableTags[tag] = struct{}{}
	}
	return nil
}

// Remove removes tags from the mutable tag set
func (t *tagService[T]) Remove(_ context.Context, tagz ...tags.Tag) error {
	for _, tag := range tagz {
		_, _, err := tags.HasScope(tag, t.managedScope)
		if err != nil {
			return err
		}
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, tag := range tagz {
		delete(t.mutableTags, tag)
	}
	return nil
}

// Tags returns the current set of mutable tags
func (t *tagService[T]) Tags(_ context.Context) (TagResult[T], error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return TagResult[T]{
		ID:     t.entityID,
		Tags:   sets.Slice(t.mutableTags),
		Events: nil,
	}, nil
}

// Contains checks if all provided tags exist in the mutable tag set
func (t *tagService[T]) Contains(_ context.Context, tags ...tags.Tag) (bool, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, tag := range tags {
		if _, exists := t.mutableTags[tag]; !exists {
			return false, nil
		}
	}
	return true, nil
}

type TagServiceConfiguration[T entityIDtype] struct {
	ManagedScope       tags.TagScope
	ExecutorUserID     uint64
	EntityID           T                                                                                   //mandatory
	EntityStorage      EntityStorage[T]                                                                    //mandatory
	GenerateEvents     func(ctx context.Context, entityID T, tags []tags.Tag) []events.UntypedEventWrapper //optional
	ActivityLogStorage EntityActivityLogStorage[T]                                                         //optional
}

// NewTagService creates a new TagMutator instance
func NewTagService[T entityIDtype](ctx context.Context, configuration TagServiceConfiguration[T]) (TagMutator[T], error) {
	if configuration.ManagedScope == "" {
		return nil, errors.New("managed scope is empty", pixieerrors.TagsDependencyErrorCode)
	}

	if types.Nil(configuration.EntityStorage) {
		return nil, errors.New("entity storage is nil", pixieerrors.TagsDependencyErrorCode)
	}

	if types.Nil(configuration.EntityID) {
		return nil, errors.New("entityID is nil", pixieerrors.TagsDependencyErrorCode)
	}

	if types.Nil(configuration.GenerateEvents) {
		configuration.GenerateEvents = func(context.Context, T, []tags.Tag) []events.UntypedEventWrapper { return nil }
	}

	if types.Nil(configuration.ActivityLogStorage) {
		configuration.ActivityLogStorage = &loggerActivityLogStorage[T]{}
	}

	immutableTags, err := configuration.EntityStorage.GetTags(ctx, configuration.EntityID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get entity %s tags", pixieerrors.TagsNotFoundErrorCode)
	}

	return &tagService[T]{
		mu:                 sync.RWMutex{},
		mutableTags:        sets.From(immutableTags),
		immutableTags:      immutableTags,
		executorUserID:     configuration.ExecutorUserID,
		managedScope:       configuration.ManagedScope,
		entityID:           configuration.EntityID,
		entityStorage:      configuration.EntityStorage,
		activityLogStorage: configuration.ActivityLogStorage,
		generateEvents:     configuration.GenerateEvents,
	}, nil
}
