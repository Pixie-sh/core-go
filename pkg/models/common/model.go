package common

type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type StatusGroup[T any] struct {
	Name     string
	Statuses []T
}

type CountResponse struct {
	Count int `json:"count"`
}
