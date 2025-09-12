package maps

import (
	"reflect"
	"testing"
)

type sample struct {
	Value int
}

func TestMapStructValue(t *testing.T) {
	// Test 1: Basic mapping without empty structs
	strukts1 := []sample{{1}, {2}, {3}}
	expected1 := []int{2, 4, 6}
	result1 := MapStructValue(strukts1, func(s sample) int {
		return s.Value * 2
	})
	if !reflect.DeepEqual(result1, expected1) {
		t.Errorf("Test 1 failed. Expected %v, got %v", expected1, result1)
	}

	// Test 2: Mapping with empty structs, without withEmpty flag
	strukts2 := []sample{{1}, {}, {3}}
	expected2 := []int{2, 0, 6}
	result2 := MapStructValue(strukts2, func(s sample) int {
		return s.Value * 2
	})
	if !reflect.DeepEqual(result2, expected2) {
		t.Errorf("Test 2 failed. Expected %v, got %v", expected2, result2)
	}

	// Test 3: Mapping with empty structs, with withEmpty flag
	strukts3 := []sample{{1}, {}, {3}}
	expected3 := []int{2, 0, 6}
	result3 := MapStructValue(strukts3, func(s sample) int {
		return s.Value * 2
	}, true)
	if !reflect.DeepEqual(result3, expected3) {
		t.Errorf("Test 3 failed. Expected %v, got %v", expected3, result3)
	}

	// Test 4: Mapping with a different struct and function
	type AnotherSample struct {
		Text string
	}

	strukts4 := []AnotherSample{{"Hello"}, {"World"}, {""}}
	expected4 := []string{"Hello!", "World!", "!"}
	result4 := MapStructValue(strukts4, func(s AnotherSample) string {
		return s.Text + "!"
	}, true)
	if !reflect.DeepEqual(result4, expected4) {
		t.Errorf("Test 4 failed. Expected %v, got %v", expected4, result4)
	}

	// Test 5: Mapping with empty structs and a different struct and function, with withEmpty flag
	expected5 := []string{"Hello!", "World!", "!"}
	result5 := MapStructValue(strukts4, func(s AnotherSample) string {
		return s.Text + "!"
	}, true)
	if !reflect.DeepEqual(result5, expected5) {
		t.Errorf("Test 5 failed. Expected %v, got %v", expected5, result5)
	}
}

type TestValue struct {
	ToDelete bool
	Name     string
}

func TestMapKeysFilteringOnValueField(t *testing.T) {
	// Test 1: Basic test with boolean filtering
	map1 := map[string]TestValue{
		"test1": TestValue{
			ToDelete: false,
			Name:     "Object1",
		},
		"test2": TestValue{
			ToDelete: true,
			Name:     "Object2",
		},
	}

	expected1 := []string{"test1"}
	result1 := MapKeysFilteringOnValueField(map1, "ToDelete", false)
	if !reflect.DeepEqual(result1, expected1) {
		t.Errorf("Test 1 failed. Expected %v, got %v", expected1, result1)
	}

	// Test 2: Basic test with string filtering
	map2 := map[string]TestValue{
		"test1": TestValue{
			ToDelete: false,
			Name:     "Object1",
		},
		"test2": TestValue{
			ToDelete: true,
			Name:     "Object2",
		},
		"test3": TestValue{
			ToDelete: true,
			Name:     "Object2",
		},
	}

	expected2 := []string{"test2", "test3"}
	result2 := MapKeysFilteringOnValueField(map2, "Name", "Object2")
	if len(result2) != len(expected2) {
		t.Errorf("Test 2 failed. Expected length %v, got length %v", len(expected2), len(result2))
	} else {
		resultSet := make(map[string]bool)
		for _, v := range result2 {
			resultSet[v] = true
		}
		for _, v := range expected2 {
			if !resultSet[v] {
				t.Errorf("Test 2 failed. Expected %v, got %v", expected2, result2)
				break
			}
		}
	}
}
