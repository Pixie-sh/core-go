package notification_models

import (
	"github.com/pixie-sh/core-go/pkg/models/language_models"
)

type CreateTemplateRequest struct {
	Name     string                       `json:"name" validate:"required"`
	Language language_models.LanguageEnum `json:"language" validate:"required"`
	Template []byte                       `json:"template" validate:"required"`
}

type CreateActionRequest struct {
	Name              string                       `json:"name,omitempty" validate:"required"`
	EventType         string                       `json:"event_type,omitempty" validate:"required"`
	Language          language_models.LanguageEnum `json:"language,omitempty" validate:"required,isEnum:LanguageEnum"`
	DeliveryMethod    ActionDeliveryMethodEnum     `json:"delivery_method,omitempty" validate:"required,isEnum:ActionDeliveryMethodEnum"`
	ActionStatus      ActionStatusEnum             `json:"action_status,omitempty" validate:"required,isEnum:ActionStatusEnum"`
	ActionPayload     any                          `json:"action_payload,omitempty" validate:"required"`
	TemplateID        uint64                       `json:"template_id,omitempty" validate:"required"`
	ConditionTemplate string                       `json:"condition_template,omitempty" validate:"omitempty,required"`
}

type CreatePushNotificationRequest struct {
	Title   string `json:"title" validate:"required" field_description:"Push notification title"`
	Message string `json:"message" validate:"required" field_description:"Push notification message body"`
	Action  string `json:"action" field_description:"Action for push notification"`
	UserID  uint64 `json:"user_id" validate:"required" field_description:"Target user ID"`
}
