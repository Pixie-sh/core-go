package models

// UntypedCursorPaginatedResult struct received when apis with cursor pagination are called
type UntypedCursorPaginatedResult struct {
	Data             interface{} `json:"data"`
	PerPage          int         `json:"per_page"`
	NextStartAfter   *string     `json:"next_start_after,omitempty"`
	HasMore          bool        `json:"has_more"`
	AvailablePerPage []int       `json:"available_per_page"`
}

// CursorPaginatedResult struct received when apis with cursor pagination are called
type CursorPaginatedResult[T any] struct {
	UntypedCursorPaginatedResult

	Data T `json:"data"`
}
