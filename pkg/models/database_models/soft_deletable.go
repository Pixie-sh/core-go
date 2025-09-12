package database_models

import (
	"time"

	"gorm.io/gorm"
)

type NullableTime = gorm.DeletedAt

// SoftDeletable to be used as composition by other entities
type SoftDeletable struct {
	CreatedAt *time.Time   `gorm:"type:timestamp"`
	UpdatedAt *time.Time   `gorm:"type:timestamp"`
	DeletedAt NullableTime `gorm:"index"`
}
