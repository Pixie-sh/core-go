package tags

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pixie-sh/core-go/pkg/models/tags"
	"github.com/pixie-sh/core-go/pkg/uid"
)

// MockEntityStorage is a mock implementation of EntityStorage
type MockEntityStorage struct {
	mock.Mock
}

func (m *MockEntityStorage) GetTags(ctx context.Context, id uid.UID) ([]tags.Tag, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]tags.Tag), args.Error(1)
}

func (m *MockEntityStorage) UpsertTags(ctx context.Context, id uid.UID, tags ...tags.Tag) error {
	args := m.Called(ctx, id, tags)
	return args.Error(0)
}

// MockEntityActivityLogStorage is a mock implementation of EntityActivityLogStorage
type MockEntityActivityLogStorage struct {
	mock.Mock
}

func (m *MockEntityActivityLogStorage) InsertTagsActivityLog(ctx context.Context, executorID uint64, id uid.UID, previousTags, newTags []tags.Tag) error {
	args := m.Called(ctx, executorID, id, previousTags, newTags)
	return args.Error(0)
}

func TestNewTaggingService(t *testing.T) {
	ctx := context.Background()
	entityID := uid.New()
	mockStorage := &MockEntityStorage{}
	mockActivityLog := &MockEntityActivityLogStorage{}

	t.Run("successful creation with existing tags", func(t *testing.T) {
		existingTags := []tags.Tag{"deal.tag1", "deal.tag2", "deal.tag3"}
		mockStorage.On("GetTags", ctx, entityID).Return(existingTags, nil).Once()

		config := TagServiceConfiguration[uid.UID]{
			ManagedScope:       tags.TagDealScope,
			EntityID:           entityID,
			EntityStorage:      mockStorage,
			ActivityLogStorage: mockActivityLog,
		}
		service, err := NewTagService(ctx, config)

		require.NoError(t, err)
		require.NotNil(t, service)

		tags, err := service.Tags(ctx)
		require.NoError(t, err)
		assert.ElementsMatch(t, existingTags, tags.Tags)

		mockStorage.AssertExpectations(t)
	})

	t.Run("successful creation with no existing tags", func(t *testing.T) {
		mockStorage.On("GetTags", ctx, entityID).Return([]tags.Tag{}, nil).Once()

		config := TagServiceConfiguration[uid.UID]{
			ManagedScope:  tags.TagDealScope,
			EntityID:      entityID,
			EntityStorage: mockStorage,
		}
		service, err := NewTagService(ctx, config)

		require.NoError(t, err)
		require.NotNil(t, service)

		tags, err := service.Tags(ctx)
		require.NoError(t, err)
		assert.Empty(t, tags.Tags)

		mockStorage.AssertExpectations(t)
	})

	t.Run("creation fails when GetTags returns error", func(t *testing.T) {
		expectedErr := assert.AnError
		mockStorage.On("GetTags", ctx, entityID).Return([]tags.Tag{}, expectedErr).Once()

		config := TagServiceConfiguration[uid.UID]{
			ManagedScope:  tags.TagDealScope,
			EntityID:      entityID,
			EntityStorage: mockStorage,
		}
		service, err := NewTagService(ctx, config)

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "unable to get entity")

		mockStorage.AssertExpectations(t)
	})
}

func TestTagging_Add(t *testing.T) {
	ctx := context.Background()
	entityID := uid.New()
	mockStorage := &MockEntityStorage{}

	existingTags := []tags.Tag{"deal.existing1", "deal.existing2"}
	mockStorage.On("GetTags", ctx, entityID).Return(existingTags, nil)

	config := TagServiceConfiguration[uid.UID]{
		ManagedScope:  tags.TagDealScope,
		EntityID:      entityID,
		EntityStorage: mockStorage,
	}
	service, err := NewTagService(ctx, config)
	require.NoError(t, err)

	t.Run("add single tag", func(t *testing.T) {
		newTag := tags.Tag("deal.new-tag")
		err := service.Add(ctx, newTag)
		require.NoError(t, err)

		tags, err := service.Tags(ctx)
		require.NoError(t, err)
		assert.Contains(t, tags.Tags, newTag)
		assert.Len(t, tags.Tags, 3) // 2 existing + 1 new
	})

	t.Run("add multiple tags", func(t *testing.T) {
		newTags := []tags.Tag{"deal.tag1", "deal.tag2", "deal.tag3"}
		err := service.Add(ctx, newTags...)
		require.NoError(t, err)

		tags, err := service.Tags(ctx)
		require.NoError(t, err)
		for _, tag := range newTags {
			assert.Contains(t, tags.Tags, tag)
		}
	})

	t.Run("add duplicate tag does not create duplicates", func(t *testing.T) {
		duplicateTag := tags.Tag("deal.existing1")
		err := service.Add(ctx, duplicateTag)
		require.NoError(t, err)

		tags, err := service.Tags(ctx)
		require.NoError(t, err)

		count := 0
		for _, tag := range tags.Tags {
			if tag == duplicateTag {
				count++
			}
		}
		assert.Equal(t, 1, count, "duplicate tag should not be added twice")
	})
}

func TestTagging_Remove(t *testing.T) {
	ctx := context.Background()
	entityID := uid.New()
	mockStorage := &MockEntityStorage{}

	existingTags := []tags.Tag{"deal.tag1", "deal.tag2", "deal.tag3", "deal.tag4"}
	mockStorage.On("GetTags", ctx, entityID).Return(existingTags, nil)

	config := TagServiceConfiguration[uid.UID]{
		ManagedScope:  tags.TagDealScope,
		EntityID:      entityID,
		EntityStorage: mockStorage,
	}
	service, err := NewTagService(ctx, config)
	require.NoError(t, err)

	t.Run("remove single existing tag", func(t *testing.T) {
		tagToRemove := tags.Tag("deal.tag2")
		err := service.Remove(ctx, tagToRemove)
		require.NoError(t, err)

		tags, err := service.Tags(ctx)
		require.NoError(t, err)
		assert.NotContains(t, tags.Tags, tagToRemove)
		assert.Len(t, tags.Tags, 3) // 4 original - 1 removed
	})

	t.Run("remove multiple tags", func(t *testing.T) {
		tagsToRemove := []tags.Tag{"deal.tag1", "deal.tag3"}
		err := service.Remove(ctx, tagsToRemove...)
		require.NoError(t, err)

		tags, err := service.Tags(ctx)
		require.NoError(t, err)
		for _, tag := range tagsToRemove {
			assert.NotContains(t, tags.Tags, tag)
		}
		assert.Len(t, tags.Tags, 1) // Only deal.tag4 should remain
	})

	t.Run("remove non-existing tag does not cause error", func(t *testing.T) {
		nonExistingTag := tags.Tag("deal.non-existing")
		err := service.Remove(ctx, nonExistingTag)
		require.NoError(t, err)

		tags, err := service.Tags(ctx)
		require.NoError(t, err)
		assert.NotContains(t, tags.Tags, nonExistingTag)
	})
}

func TestTagging_Contains(t *testing.T) {
	ctx := context.Background()
	entityID := uid.New()
	mockStorage := &MockEntityStorage{}

	existingTags := []tags.Tag{"deal.tag1", "deal.tag2", "deal.tag3"}
	mockStorage.On("GetTags", ctx, entityID).Return(existingTags, nil)

	config := TagServiceConfiguration[uid.UID]{
		ManagedScope:  tags.TagDealScope,
		EntityID:      entityID,
		EntityStorage: mockStorage,
	}
	service, err := NewTagService(ctx, config)
	require.NoError(t, err)

	t.Run("contains existing single tag", func(t *testing.T) {
		contains, err := service.Contains(ctx, tags.Tag("deal.tag1"))
		require.NoError(t, err)
		assert.True(t, contains)
	})

	t.Run("contains multiple existing tags", func(t *testing.T) {
		contains, err := service.Contains(ctx, tags.Tag("deal.tag1"), tags.Tag("deal.tag2"))
		require.NoError(t, err)
		assert.True(t, contains)
	})

	t.Run("does not contain non-existing tag", func(t *testing.T) {
		contains, err := service.Contains(ctx, tags.Tag("deal.non-existing"))
		require.NoError(t, err)
		assert.False(t, contains)
	})

	t.Run("does not contain when one of multiple tags is missing", func(t *testing.T) {
		contains, err := service.Contains(ctx, tags.Tag("deal.tag1"), tags.Tag("deal.non-existing"))
		require.NoError(t, err)
		assert.False(t, contains)
	})

	t.Run("contains after adding new tag", func(t *testing.T) {
		newTag := tags.Tag("deal.new-tag")
		err := service.Add(ctx, newTag)
		require.NoError(t, err)

		contains, err := service.Contains(ctx, newTag)
		require.NoError(t, err)
		assert.True(t, contains)
	})
}

func TestTagging_Save(t *testing.T) {
	ctx := context.Background()
	entityID := uid.New()
	mockStorage := &MockEntityStorage{}
	mockActivityLog := &MockEntityActivityLogStorage{}

	t.Run("save without activity log", func(t *testing.T) {
		existingTags := []tags.Tag{"deal.tag1", "deal.tag2"}
		mockStorage.On("GetTags", ctx, entityID).Return(existingTags, nil).Once()
		mockStorage.On("UpsertTags", ctx, entityID, mock.AnythingOfType("[]tags.Tag")).Return(nil).Once()

		config := TagServiceConfiguration[uid.UID]{
			ManagedScope:  tags.TagDealScope,
			EntityID:      entityID,
			EntityStorage: mockStorage,
		}
		service, err := NewTagService(ctx, config)
		require.NoError(t, err)

		// Add a new tag
		err = service.Add(ctx, tags.Tag("deal.new-tag"))
		require.NoError(t, err)

		// Save
		_, err = service.Save(ctx)
		require.NoError(t, err)

		mockStorage.AssertExpectations(t)
	})

	t.Run("save with activity log", func(t *testing.T) {
		existingTags := []tags.Tag{"deal.tag1", "deal.tag2"}
		mockStorage.On("GetTags", ctx, entityID).Return(existingTags, nil).Once()
		mockStorage.On("UpsertTags", ctx, entityID, mock.AnythingOfType("[]tags.Tag")).Return(nil).Once()
		mockActivityLog.On("InsertTagsActivityLog", ctx, uint64(0), entityID, existingTags, mock.AnythingOfType("[]tags.Tag")).Return(nil).Once()

		config := TagServiceConfiguration[uid.UID]{
			ManagedScope:       tags.TagDealScope,
			EntityID:           entityID,
			EntityStorage:      mockStorage,
			ActivityLogStorage: mockActivityLog,
		}
		service, err := NewTagService(ctx, config)
		require.NoError(t, err)

		// Add a new tag
		err = service.Add(ctx, tags.Tag("deal.new-tag"))
		require.NoError(t, err)

		// Save
		_, err = service.Save(ctx)
		require.NoError(t, err)

		mockStorage.AssertExpectations(t)
		mockActivityLog.AssertExpectations(t)
	})

	t.Run("save fails when UpsertTags returns error", func(t *testing.T) {
		existingTags := []tags.Tag{"deal.tag1"}
		expectedErr := assert.AnError
		mockStorage.On("GetTags", ctx, entityID).Return(existingTags, nil).Once()
		mockStorage.On("UpsertTags", ctx, entityID, mock.AnythingOfType("[]tags.Tag")).Return(expectedErr).Once()

		config := TagServiceConfiguration[uid.UID]{
			ManagedScope:  tags.TagDealScope,
			EntityID:      entityID,
			EntityStorage: mockStorage,
		}
		service, err := NewTagService(ctx, config)
		require.NoError(t, err)

		_, err = service.Save(ctx)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)

		mockStorage.AssertExpectations(t)
	})

	t.Run("save fails when activity log returns error", func(t *testing.T) {
		existingTags := []tags.Tag{"deal.tag1"}
		expectedErr := assert.AnError
		mockStorage.On("GetTags", ctx, entityID).Return(existingTags, nil).Once()
		mockStorage.On("UpsertTags", ctx, entityID, mock.AnythingOfType("[]tags.Tag")).Return(nil).Once()
		mockActivityLog.On("InsertTagsActivityLog", ctx, uint64(0), entityID, existingTags, mock.AnythingOfType("[]tags.Tag")).Return(expectedErr).Once()

		config := TagServiceConfiguration[uid.UID]{
			ManagedScope:       tags.TagDealScope,
			EntityID:           entityID,
			EntityStorage:      mockStorage,
			ActivityLogStorage: mockActivityLog,
		}
		service, err := NewTagService(ctx, config)
		require.NoError(t, err)

		_, err = service.Save(ctx)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)

		mockStorage.AssertExpectations(t)
		mockActivityLog.AssertExpectations(t)
	})
}

func TestTagging_IntegrationWorkflow(t *testing.T) {
	ctx := context.Background()
	entityID := uid.New()
	mockStorage := &MockEntityStorage{}
	mockActivityLog := &MockEntityActivityLogStorage{}

	existingTags := []tags.Tag{"deal.initial1", "deal.initial2"}
	mockStorage.On("GetTags", ctx, entityID).Return(existingTags, nil)

	config := TagServiceConfiguration[uid.UID]{
		ManagedScope:       tags.TagDealScope,
		EntityID:           entityID,
		EntityStorage:      mockStorage,
		ActivityLogStorage: mockActivityLog,
	}
	service, err := NewTagService(ctx, config)
	require.NoError(t, err)

	// Add some tags
	err = service.Add(ctx, tags.Tag("deal.added1"), tags.Tag("deal.added2"))
	require.NoError(t, err)

	// Remove one existing tag
	err = service.Remove(ctx, tags.Tag("deal.initial1"))
	require.NoError(t, err)

	// Check final state
	tagz, err := service.Tags(ctx)
	require.NoError(t, err)
	expectedTags := []tags.Tag{"deal.initial2", "deal.added1", "deal.added2"}
	assert.ElementsMatch(t, expectedTags, tagz.Tags)

	// Verify contains works correctly
	contains, err := service.Contains(ctx, tags.Tag("deal.initial2"), tags.Tag("deal.added1"))
	require.NoError(t, err)
	assert.True(t, contains)

	contains, err = service.Contains(ctx, tags.Tag("deal.initial1")) // removed tag
	require.NoError(t, err)
	assert.False(t, contains)

	// Mock save operation
	mockStorage.On("UpsertTags", ctx, entityID, mock.MatchedBy(func(tags []tags.Tag) bool {
		return len(tags) == 3 // Should have 3 tags after operations
	})).Return(nil).Once()

	// Fix: Change uint64(1) to uint64(0) to match the actual call
	mockActivityLog.On("InsertTagsActivityLog", ctx, uint64(0), entityID, existingTags, mock.AnythingOfType("[]tags.Tag")).Return(nil).Once()

	_, err = service.Save(ctx)
	require.NoError(t, err)

	mockStorage.AssertExpectations(t)
	mockActivityLog.AssertExpectations(t)
}

func TestTag_String(t *testing.T) {
	tag := tags.Tag("deal.test-tag")
	assert.Equal(t, "deal.test-tag", tag.String())
}
func TestTagService_Concurrency(t *testing.T) {
	ctx := context.Background()
	entityID := uid.New()
	mockStorage := &MockEntityStorage{}
	mockActivityLog := &MockEntityActivityLogStorage{}

	// Initial tags for the service
	initialTags := []tags.Tag{"deal.initial1", "deal.initial2", "deal.initial3"}
	mockStorage.On("GetTags", ctx, entityID).Return(initialTags, nil)

	// Mock storage operations to be thread-safe
	mockStorage.On("UpsertTags", ctx, entityID, mock.AnythingOfType("[]tags.Tag")).Return(nil).Maybe()
	mockActivityLog.On("InsertTagsActivityLog", ctx, uint64(0), entityID, mock.AnythingOfType("[]tags.Tag"), mock.AnythingOfType("[]tags.Tag")).Return(nil).Maybe()

	config := TagServiceConfiguration[uid.UID]{
		ManagedScope:       tags.TagDealScope,
		EntityID:           entityID,
		EntityStorage:      mockStorage,
		ActivityLogStorage: mockActivityLog,
	}

	service, err := NewTagService(ctx, config)
	require.NoError(t, err)

	t.Run("concurrent add operations", func(t *testing.T) {
		const numGoroutines = 10
		const tagsPerGoroutine = 5

		var wg sync.WaitGroup
		errorChan := make(chan error, numGoroutines)

		// Launch multiple goroutines adding tags concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()

				for j := 0; j < tagsPerGoroutine; j++ {
					tagName := fmt.Sprintf("deal.concurrent-tag-%d-%d", routineID, j)
					if err := service.Add(ctx, tags.Tag(tagName)); err != nil {
						errorChan <- err
						return
					}
				}
			}(i)
		}

		wg.Wait()
		close(errorChan)

		// Check for any errors
		for err := range errorChan {
			t.Errorf("Concurrent add operation failed: %v", err)
		}

		// Verify all tags were added
		finalTags, err := service.Tags(ctx)
		require.NoError(t, err)

		expectedCount := len(initialTags) + (numGoroutines * tagsPerGoroutine)
		assert.Equal(t, expectedCount, len(finalTags.Tags), "Expected %d tags after concurrent additions", expectedCount)
	})

	t.Run("concurrent mixed operations", func(t *testing.T) {
		const numGoroutines = 15

		var wg sync.WaitGroup
		errorChan := make(chan error, numGoroutines*3) // Buffer for potential errors

		// Add operations
		for i := 0; i < numGoroutines/3; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()
				tagName := fmt.Sprintf("deal.add-tag-%d", routineID)
				if err := service.Add(ctx, tags.Tag(tagName)); err != nil {
					errorChan <- err
				}
			}(i)
		}

		// Remove operations (trying to remove initial tags)
		for i := 0; i < numGoroutines/3; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()
				// Try to remove initial tags or some that might have been added
				tagToRemove := initialTags[routineID%len(initialTags)]
				if err := service.Remove(ctx, tagToRemove); err != nil {
					errorChan <- err
				}
			}(i)
		}

		// Contains operations
		for i := 0; i < numGoroutines/3; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()
				tagToCheck := initialTags[routineID%len(initialTags)]
				if _, err := service.Contains(ctx, tagToCheck); err != nil {
					errorChan <- err
				}
			}(i)
		}

		wg.Wait()
		close(errorChan)

		// Check for any errors
		for err := range errorChan {
			t.Errorf("Concurrent mixed operation failed: %v", err)
		}

		// Verify service is still functional
		finalTags, err := service.Tags(ctx)
		require.NoError(t, err)
		assert.NotNil(t, finalTags)
	})

	t.Run("concurrent read operations", func(t *testing.T) {
		const numGoroutines = 20

		var wg sync.WaitGroup
		errorChan := make(chan error, numGoroutines*2)

		// Multiple goroutines reading Tags() and Contains() simultaneously
		for i := 0; i < numGoroutines; i++ {
			wg.Add(2)

			// Tags() operation
			go func() {
				defer wg.Done()
				if _, err := service.Tags(ctx); err != nil {
					errorChan <- err
				}
			}()

			// Contains() operation
			go func(routineID int) {
				defer wg.Done()
				tagToCheck := initialTags[routineID%len(initialTags)]
				if _, err := service.Contains(ctx, tagToCheck); err != nil {
					errorChan <- err
				}
			}(i)
		}

		wg.Wait()
		close(errorChan)

		// Check for any errors
		for err := range errorChan {
			t.Errorf("Concurrent read operation failed: %v", err)
		}
	})

	t.Run("concurrent save operations", func(t *testing.T) {
		const numGoroutines = 5

		var wg sync.WaitGroup
		errorChan := make(chan error, numGoroutines)

		// Multiple goroutines trying to save simultaneously
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, err := service.Save(ctx); err != nil {
					errorChan <- err
				}
			}()
		}

		wg.Wait()
		close(errorChan)

		// Check for any errors
		for err := range errorChan {
			t.Errorf("Concurrent save operation failed: %v", err)
		}
	})

	t.Run("stress test - high volume concurrent operations", func(t *testing.T) {
		const numGoroutines = 50
		const operationsPerGoroutine = 10

		var wg sync.WaitGroup
		errorChan := make(chan error, numGoroutines*operationsPerGoroutine)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerGoroutine; j++ {
					operation := j % 4 // Cycle through different operations

					switch operation {
					case 0: // Add
						tagName := fmt.Sprintf("deal.stress-add-%d-%d", routineID, j)
						if err := service.Add(ctx, tags.Tag(tagName)); err != nil {
							errorChan <- err
							return
						}
					case 1: // Remove
						tagName := fmt.Sprintf("deal.stress-add-%d-%d", routineID, j-1)
						if err := service.Remove(ctx, tags.Tag(tagName)); err != nil {
							errorChan <- err
							return
						}
					case 2: // Contains
						tagToCheck := initialTags[routineID%len(initialTags)]
						if _, err := service.Contains(ctx, tagToCheck); err != nil {
							errorChan <- err
							return
						}
					case 3: // Tags
						if _, err := service.Tags(ctx); err != nil {
							errorChan <- err
							return
						}
					}
				}
			}(i)
		}

		wg.Wait()
		close(errorChan)

		// Check for any errors
		errorCount := 0
		for err := range errorChan {
			errorCount++
			t.Errorf("Stress test operation failed: %v", err)
		}

		if errorCount == 0 {
			t.Logf("Stress test completed successfully with %d goroutines and %d operations each",
				numGoroutines, operationsPerGoroutine)
		}

		// Verify service is still functional after stress test
		finalTags, err := service.Tags(ctx)
		require.NoError(t, err)
		assert.NotNil(t, finalTags)
	})

	mockStorage.AssertExpectations(t)
}
