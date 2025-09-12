package slices

import (
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/utils"
)

func Contains[T comparable](s []T, e T) bool {
	return slices.Contains(s, e)
}

func Map[S ~[]E, E any, R any](model S, f func(item E) R) []R {
	var result []R
	for _, item := range model {
		res := f(item)
		if !types.Nil(res) {
			result = append(result, res)
		}
	}

	return result
}

func Find[S ~[]E, E any](slice S, f func(item E) bool) E {
	var result E
	for _, item := range slice {
		if f(item) {
			return item
		}
	}

	return result
}

func MapOrError[S ~[]E, E any, R any](model S, f func(item E) (R, error)) ([]R, error) {
	var result []R
	for _, item := range model {
		res, err := f(item)
		if err != nil {
			return nil, err
		}

		if !types.Nil(res) {
			result = append(result, res)
		}
	}

	return result, nil
}

// ShuffleSlice shuffles the elements of a slice in place and returns the shuffled slice.
// This function is type-agnostic and works with slices of any type.
// !! original slice is modified; return the same slice just for readability
func ShuffleSlice[T any](slice []T) []T {
	r := utils.Random()
	r.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
	return slice
}

func JoinUint64Slice(slice []uint64, sep string) string {
	strSlice := make([]string, len(slice))
	for i, num := range slice {
		strSlice[i] = strconv.FormatUint(num, 10)
	}

	return strings.Join(strSlice, sep)
}

func Copy[S ~[]E, E any](states S) S {
	cp := make(S, len(states))
	copy(cp, states)
	return cp
}

func SortByCustomOrder[T any](slice []T, fieldExtractor func(T) string, priorityMap map[string]int) {
	sort.SliceStable(slice, func(i, j int) bool {
		valueI := fieldExtractor(slice[i])
		valueJ := fieldExtractor(slice[j])

		priorityI := priorityMap[valueI]
		priorityJ := priorityMap[valueJ]

		if priorityI == 0 {
			priorityI = 999999
		}
		if priorityJ == 0 {
			priorityJ = 999999
		}

		return priorityI < priorityJ
	})
}

func Deduplicate[T comparable](slice []T) []T {
	if len(slice) == 0 {
		return slice
	}

	seen := make(map[T]bool)
	result := make([]T, 0)

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
