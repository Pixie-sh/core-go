package common

import (
	"time"

	"github.com/pixie-sh/core-go/pkg/models/enums"

	"github.com/pixie-sh/core-go/pkg/models"
)

type ContentType struct {
	models.SoftDeletable

	ID   uint64 `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
} //@name ContentType

type ContentTypeView struct {
	ID   uint64 `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

type ContentTypeBOView struct {
	ID        uint64                  `json:"id,omitempty"`
	Name      string                  `json:"name,omitempty"`
	Slug      string                  `json:"slug,omitempty"`
	Status    enums.ContentTypeStatus `json:"status,omitempty"`
	CreatedAt *time.Time              `json:"createdAt,omitempty"`
	UpdatedAt *time.Time              `json:"updatedAt,omitempty"`
}

type CreateContentTypeRequest struct {
	Name   string                  `json:"name"`
	Slug   string                  `json:"slug"`
	Status enums.ContentTypeStatus `json:"status"`
}

type UpdateContentTypeRequest struct {
	Name   string                  `json:"name"`
	Slug   string                  `json:"slug"`
	Status enums.ContentTypeStatus `json:"status"`
}
