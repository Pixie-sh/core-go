package sets

type Set[T comparable] map[T]struct{}

func Contains[T comparable](s Set[T], item T) (T, bool) {
	_, exists := s[item]
	return item, exists
}

func From[T comparable](slice []T) Set[T] {
	result := make(Set[T])
	for _, item := range slice {
		result[item] = struct{}{}
	}

	return result
}

func Slice[T comparable](s Set[T]) []T {
	result := make([]T, 0, len(s))
	for item := range s {
		result = append(result, item)
	}

	return result
}

func SliceAdapter[T comparable, C any](s Set[T], adapter func(T) C) []C {
	if adapter == nil {
		panic("sets.SliceAdapter: adapter cannot be nil")
	}

	result := make([]C, 0, len(s))
	for item := range s {
		result = append(result, adapter(item))
	}

	return result
}
