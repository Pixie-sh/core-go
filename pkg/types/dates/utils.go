package dates

import (
	"time"

	"github.com/pixie-sh/core-go/pkg/models/database_models"
)

func FromNullableTime(t database_models.NullableTime) *time.Time {
	if t.Valid {
		return &t.Time
	}

	return nil
}
