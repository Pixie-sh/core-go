package notification_models

import "github.com/pixie-sh/core-go/pkg/uid"

type NotificationActivityLog struct {
	ID          uid.UID                        `json:"id"`
	PublicToken string                         `json:"public_token"`
	ActionID    uid.UID                        `json:"action_id"`
	TemplateID  uid.UID                        `json:"template_id"`
	Payload     NotificationActivityLogPayload `json:"payload"`
	Action      Action                         `json:"action,omitempty"`
	Template    Template                       `json:"template,omitempty"`
}

type NotificationActivityLogPayload struct {
	Notification         any            `json:"notification"`
	ActionPayloadRequest any            `json:"action_payload_request"`
	AdditionalData       map[string]any `json:"additional_data"`
}
