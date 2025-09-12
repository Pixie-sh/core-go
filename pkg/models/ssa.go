package models

import "time"

type ServerSideAcknowledge struct {
	PayloadType string    `json:"payload_type" validate:"required"`
	ID          string    `json:"id" validate:"required,uuid4"`
	When        time.Time `json:"when" validate:"required"`
}
