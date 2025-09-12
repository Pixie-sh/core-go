package notification_models

import (
	"github.com/pixie-sh/core-go/pkg/models"
	"github.com/pixie-sh/core-go/pkg/models/language_models"
	"github.com/pixie-sh/core-go/pkg/uid"
)

type Template struct {
	models.SoftDeletable

	ID       uid.UID                      `json:"id"`
	Name     string                       `json:"name"`
	Language language_models.LanguageEnum `json:"language"`
	Template []byte                       `json:"template,omitempty"`
}
