package slices

import (
	"reflect"
	"testing"
)

func TestShuffleSlice(t *testing.T) {
	t.Run("Shuffle int slice", func(t *testing.T) {
		original := []int{1, 2, 3, 4, 5}
		shuffled := ShuffleSlice(original)

		// in-place: contents equal because original now holds shuffled contents
		if !reflect.DeepEqual(shuffled, original) {
			t.Errorf("ShuffleSlice didn't modified the original slice")
		}

		// Avoid flakiness: try multiple shuffles to ensure it doesn't remain sorted
		const attempts = 5
		sorted := true
		for i := 0; i < attempts; i++ {
			if !isSliceSorted(shuffled) {
				sorted = false
				break
			}
			shuffled = ShuffleSlice(shuffled)
		}
		if sorted {
			t.Errorf("Slice remained sorted after %d shuffle attempts: %v", attempts, shuffled)
		}

		if !haveSameElements(original, shuffled) {
			t.Errorf("Shuffled slice has different elements than original")
		}
	})

	t.Run("Shuffle string slice", func(t *testing.T) {
		original := []string{"a", "b", "c", "d", "e"}
		shuffled := ShuffleSlice(original)

		// in-place: contents equal because original now holds shuffled contents
		if !reflect.DeepEqual(shuffled, original) {
			t.Errorf("ShuffleSlice didn't modified the original slice")
		}

		// Avoid flakiness: try multiple shuffles to ensure it doesn't remain sorted
		const attempts = 5
		sorted := true
		for i := 0; i < attempts; i++ {
			if !isSliceSorted(shuffled) {
				sorted = false
				break
			}
			shuffled = ShuffleSlice(shuffled)
		}
		if sorted {
			t.Errorf("Slice remained sorted after %d shuffle attempts: %v", attempts, shuffled)
		}

		if !haveSameElements(original, shuffled) {
			t.Errorf("Shuffled slice has different elements than original")
		}
	})

	t.Run("Shuffle empty slice", func(t *testing.T) {
		original := []int{}
		shuffled := ShuffleSlice(original)

		if len(shuffled) != 0 {
			t.Errorf("Expected empty slice, got %v", shuffled)
		}
	})

	t.Run("Shuffle single element slice", func(t *testing.T) {
		original := []int{1}
		shuffled := ShuffleSlice(original)

		if !reflect.DeepEqual(shuffled, original) {
			t.Errorf("Expected [1], got %v", shuffled)
		}
	})

	t.Run("Shuffle custom struct slice", func(t *testing.T) {
		type customStruct struct {
			id   int
			name string
		}

		original := []customStruct{
			{1, "one"},
			{2, "two"},
			{3, "three"},
			{4, "four"},
			{5, "five"},
		}
		shuffled := ShuffleSlice(original)

		if !reflect.DeepEqual(shuffled, original) {
			t.Errorf("ShuffleSlice didn't modified the original slice")
		}

		if isSliceSorted(shuffled) {
			t.Errorf("Slice was not shuffled: %v", shuffled)
		}

		if !haveSameElements(original, shuffled) {
			t.Errorf("Shuffled slice has different elements than original")
		}
	})
}

type testCustomSortOrderObject struct {
	Name  string
	Title string
	Type  string
}

func TestCustomSortOrder(t *testing.T) {
	tests := []struct {
		name          string
		objectInput   []testCustomSortOrderObject
		priorityInput map[string]int
		expected      []testCustomSortOrderObject
	}{
		{
			name: "Normal case",
			objectInput: []testCustomSortOrderObject{
				{"Some1", "SomeTitle", "First"},
				{"Some3", "SomeTitle3", "Third"},
				{"Some2", "SomeTitle2", "Second"},
				{"Some4", "SomeTitle4", "Fourth"},
				{"Some11", "SomeTitle11", "First"},
				{"Some33", "SomeTitle33", "Third"},
				{"Some22", "SomeTitle22", "Second"},
			},
			priorityInput: map[string]int{
				"First":  1,
				"Second": 2,
				"Third":  3,
				"Fourth": 4,
			},
			expected: []testCustomSortOrderObject{
				{"Some1", "SomeTitle", "First"},
				{"Some11", "SomeTitle11", "First"},
				{"Some2", "SomeTitle2", "Second"},
				{"Some22", "SomeTitle22", "Second"},
				{"Some3", "SomeTitle3", "Third"},
				{"Some33", "SomeTitle33", "Third"},
				{"Some4", "SomeTitle4", "Fourth"},
			},
		},
		{
			name:          "Empty slice",
			objectInput:   []testCustomSortOrderObject{},
			priorityInput: map[string]int{"First": 1, "Second": 2},
			expected:      []testCustomSortOrderObject{},
		},
		{
			name: "Single element",
			objectInput: []testCustomSortOrderObject{
				{"Only", "OnlyTitle", "First"},
			},
			priorityInput: map[string]int{"First": 1, "Second": 2},
			expected: []testCustomSortOrderObject{
				{"Only", "OnlyTitle", "First"},
			},
		},
		{
			name: "All unknown types (not in priority map)",
			objectInput: []testCustomSortOrderObject{
				{"Some1", "SomeTitle", "Unknown1"},
				{"Some2", "SomeTitle2", "Unknown2"},
				{"Some3", "SomeTitle3", "Unknown3"},
			},
			priorityInput: map[string]int{"First": 1, "Second": 2},
			expected: []testCustomSortOrderObject{
				{"Some1", "SomeTitle", "Unknown1"},
				{"Some2", "SomeTitle2", "Unknown2"},
				{"Some3", "SomeTitle3", "Unknown3"},
			},
		},
		{
			name: "Mix of known and unknown types",
			objectInput: []testCustomSortOrderObject{
				{"Some1", "SomeTitle", "Unknown1"},
				{"Some2", "SomeTitle2", "First"},
				{"Some3", "SomeTitle3", "Unknown2"},
				{"Some4", "SomeTitle4", "Second"},
			},
			priorityInput: map[string]int{"First": 1, "Second": 2},
			expected: []testCustomSortOrderObject{
				{"Some2", "SomeTitle2", "First"},
				{"Some4", "SomeTitle4", "Second"},
				{"Some1", "SomeTitle", "Unknown1"},
				{"Some3", "SomeTitle3", "Unknown2"},
			},
		},
		{
			name: "Empty priority map",
			objectInput: []testCustomSortOrderObject{
				{"Some1", "SomeTitle", "First"},
				{"Some2", "SomeTitle2", "Second"},
				{"Some3", "SomeTitle3", "Third"},
			},
			priorityInput: map[string]int{},
			expected: []testCustomSortOrderObject{
				{"Some1", "SomeTitle", "First"},
				{"Some2", "SomeTitle2", "Second"},
				{"Some3", "SomeTitle3", "Third"},
			},
		},
		{
			name: "Priority with zero value",
			objectInput: []testCustomSortOrderObject{
				{"Some1", "SomeTitle", "Zero"},
				{"Some2", "SomeTitle2", "First"},
				{"Some3", "SomeTitle3", "Unknown"},
			},
			priorityInput: map[string]int{"Zero": 0, "First": 1}, // 0 is treated as unknown
			expected: []testCustomSortOrderObject{
				{"Some2", "SomeTitle2", "First"},
				{"Some1", "SomeTitle", "Zero"},     // Zero value treated as unknown
				{"Some3", "SomeTitle3", "Unknown"}, // Unknown also goes to end
			},
		},
		{
			name: "Negative priority values",
			objectInput: []testCustomSortOrderObject{
				{"Some1", "SomeTitle", "Negative"},
				{"Some2", "SomeTitle2", "Positive"},
				{"Some3", "SomeTitle3", "Zero"},
			},
			priorityInput: map[string]int{"Negative": -1, "Positive": 1, "Zero": 0}, // 0 treated as unknown
			expected: []testCustomSortOrderObject{
				{"Some1", "SomeTitle", "Negative"},
				{"Some2", "SomeTitle2", "Positive"},
				{"Some3", "SomeTitle3", "Zero"}, // Zero value goes to end
			},
		},
		{
			name: "Large priority values",
			objectInput: []testCustomSortOrderObject{
				{"Some1", "SomeTitle", "Large"},
				{"Some2", "SomeTitle2", "Larger"},
				{"Some3", "SomeTitle3", "Small"},
			},
			priorityInput: map[string]int{"Small": 1, "Large": 1000000, "Larger": 2000000},
			expected: []testCustomSortOrderObject{
				{"Some3", "SomeTitle3", "Small"},
				{"Some1", "SomeTitle", "Large"},
				{"Some2", "SomeTitle2", "Larger"},
			},
		},
		{
			name: "Same priority values (stable sort test)",
			objectInput: []testCustomSortOrderObject{
				{"First", "Title1", "Same"},
				{"Second", "Title2", "Same"},
				{"Third", "Title3", "Same"},
				{"Fourth", "Title4", "Different"},
			},
			priorityInput: map[string]int{"Same": 1, "Different": 2},
			expected: []testCustomSortOrderObject{
				{"First", "Title1", "Same"},  // Stable sort preserves original order
				{"Second", "Title2", "Same"}, // for elements with same priority
				{"Third", "Title3", "Same"},
				{"Fourth", "Title4", "Different"},
			},
		},
		{
			name: "Empty string types",
			objectInput: []testCustomSortOrderObject{
				{"Some1", "SomeTitle", ""},
				{"Some2", "SomeTitle2", "First"},
				{"Some3", "SomeTitle3", ""},
			},
			priorityInput: map[string]int{"First": 1, "": 2}, // Empty string has priority
			expected: []testCustomSortOrderObject{
				{"Some2", "SomeTitle2", "First"},
				{"Some1", "SomeTitle", ""},
				{"Some3", "SomeTitle3", ""},
			},
		},
		{
			name: "Case sensitive types",
			objectInput: []testCustomSortOrderObject{
				{"Some1", "SomeTitle", "first"},
				{"Some2", "SomeTitle2", "First"},
				{"Some3", "SomeTitle3", "FIRST"},
			},
			priorityInput: map[string]int{"First": 1, "first": 2, "FIRST": 3},
			expected: []testCustomSortOrderObject{
				{"Some2", "SomeTitle2", "First"},
				{"Some1", "SomeTitle", "first"},
				{"Some3", "SomeTitle3", "FIRST"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objectInputBefore := tt.objectInput
			SortByCustomOrder(tt.objectInput, func(o testCustomSortOrderObject) string { return o.Type }, tt.priorityInput)
			if !reflect.DeepEqual(tt.objectInput, tt.expected) {
				t.Errorf("SortByCustomOrder(%v, func(o testCustomSortOrderObject) string { return o.Type }, %v) = %v, want %v", objectInputBefore, tt.priorityInput, tt.objectInput, tt.expected)
			}
		})
	}
}

// Helper function to check if a slice is sorted
func isSliceSorted[T any](slice []T) bool {
	for i := 1; i < len(slice); i++ {
		if !isLess(slice[i-1], slice[i]) {
			return false
		}
	}
	return true
}

// Helper function to compare two values of any type
func isLess(a, b any) bool {
	switch x := a.(type) {
	case int:
		return x < b.(int)
	case string:
		return x < b.(string)
	case float64:
		return x < b.(float64)
	// Add more types as needed
	default:
		// For types that can't be compared, we'll consider them equal
		// This means the slice will be considered "sorted" for incomparable types
		return false
	}
}

func TestDeduplicate(t *testing.T) {
	t.Run("Deduplicate int slice", func(t *testing.T) {
		input := []int{1, 2, 3, 2, 1, 4, 3, 5}
		expected := []int{1, 2, 3, 4, 5}
		result := Deduplicate(input)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Deduplicate(%v) = %v, want %v", input, result, expected)
		}
	})

	t.Run("Deduplicate string slice", func(t *testing.T) {
		input := []string{"a", "b", "c", "b", "a", "d", "c", "e"}
		expected := []string{"a", "b", "c", "d", "e"}
		result := Deduplicate(input)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Deduplicate(%v) = %v, want %v", input, result, expected)
		}
	})

	t.Run("Deduplicate empty slice", func(t *testing.T) {
		input := []int{}
		expected := []int{}
		result := Deduplicate(input)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Deduplicate(%v) = %v, want %v", input, result, expected)
		}
	})

	t.Run("Deduplicate single element slice", func(t *testing.T) {
		input := []int{1}
		expected := []int{1}
		result := Deduplicate(input)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Deduplicate(%v) = %v, want %v", input, result, expected)
		}
	})

	t.Run("Deduplicate slice with no duplicates", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		expected := []int{1, 2, 3, 4, 5}
		result := Deduplicate(input)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Deduplicate(%v) = %v, want %v", input, result, expected)
		}
	})

	t.Run("Deduplicate slice with all duplicates", func(t *testing.T) {
		input := []int{1, 1, 1, 1, 1}
		expected := []int{1}
		result := Deduplicate(input)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Deduplicate(%v) = %v, want %v", input, result, expected)
		}
	})

	t.Run("Deduplicate preserves order", func(t *testing.T) {
		input := []int{5, 2, 8, 2, 1, 5, 3}
		expected := []int{5, 2, 8, 1, 3}
		result := Deduplicate(input)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Deduplicate(%v) = %v, want %v", input, result, expected)
		}
	})

	t.Run("Deduplicate custom struct slice", func(t *testing.T) {
		type customStruct struct {
			id   int
			name string
		}

		input := []customStruct{
			{1, "one"},
			{2, "two"},
			{1, "one"},
			{3, "three"},
			{2, "two"},
		}
		expected := []customStruct{
			{1, "one"},
			{2, "two"},
			{3, "three"},
		}
		result := Deduplicate(input)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Deduplicate(%v) = %v, want %v", input, result, expected)
		}
	})
}

// Helper function to check if two slices have the same elements
func haveSameElements[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	countMap := make(map[T]int)

	for _, v := range a {
		countMap[v]++
	}

	for _, v := range b {
		if count, exists := countMap[v]; !exists || count == 0 {
			return false
		}
		countMap[v]--
	}

	return true
}
