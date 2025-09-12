package enums

type EnumInfo[T any] struct {
	Label string `json:"label"`
	Value T      `json:"value"`
}
