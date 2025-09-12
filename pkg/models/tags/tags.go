package tags

import (
	"database/sql/driver"
	"strings"

	"github.com/lib/pq"
)

// Tags represent a slice of Tag
type Tags []Tag

func (t Tags) String() string {
	var builder strings.Builder
	builder.WriteString("[")
	for i, tag := range t {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(`"`)
		builder.WriteString(string(tag))
		builder.WriteString(`"`)
	}
	builder.WriteString("]")
	return builder.String()
}

func (t Tags) Array() []string {
	var arr = make([]string, len(t))
	for i, tag := range t {
		arr[i] = tag.String()
	}

	return arr
}

func (a *Tags) Scan(src interface{}) error {
	var tags pq.StringArray
	err := tags.Scan(src)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		*a = append(*a, Tag(tag))
	}

	return nil
}

func (a Tags) Value() (driver.Value, error) {
	var tags pq.StringArray
	for _, tag := range a {
		tags = append(tags, tag.String())
	}

	return tags, nil
}
