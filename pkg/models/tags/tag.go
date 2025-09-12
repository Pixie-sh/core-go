package tags

import (
	"database/sql/driver"
	"fmt"
)

type EntityTag[T any] struct {
	Scope TagScope `json:"scope"`
	Tags  Tags     `json:"tags"`
	ID    T        `json:"id"`
}

type TagScope string

func (t TagScope) String() string {
	return string(t)
}

// Tag represents a string-based identifier used for categorization
type Tag string

func (t Tag) String() string {
	return string(t)
}

func (t *Tag) Scan(src interface{}) error {
	switch v := src.(type) {
	case string:
		*t = Tag(v)
	case []byte:
		*t = Tag(v)
	case nil:
		*t = ""
	default:
		return fmt.Errorf("cannot scan %T into Tag", src)
	}
	return nil
}

func (t Tag) Value() (driver.Value, error) {
	return t.String(), nil
}
