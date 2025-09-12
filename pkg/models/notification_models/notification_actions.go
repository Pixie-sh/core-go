package notification_models

import (
	"github.com/pixie-sh/core-go/pkg/models/language_models"
	"github.com/pixie-sh/core-go/pkg/uid"
	"github.com/pixie-sh/core-go/pkg/utils"
)

type ActionDeliveryMethodEnum string

func (e ActionDeliveryMethodEnum) String() string {
	return string(e)
}

type ActionStatusEnum string

func (e ActionStatusEnum) String() string {
	return string(e)
}

const (
	EmailActionDeliveryMethod      ActionDeliveryMethodEnum = "email"
	SMSActionDeliveryMethod        ActionDeliveryMethodEnum = "sms"
	PUSHMobileActionDeliveryMethod ActionDeliveryMethodEnum = "push_mobile"
	PUSHWebActionDeliveryMethod    ActionDeliveryMethodEnum = "push_web"

	ActiveActionStatusEnum   ActionStatusEnum = "active"
	InactiveActionStatusEnum ActionStatusEnum = "inactive"
	PreviewActionStatusEnum  ActionStatusEnum = "preview"
)

var ActionStatusEnumList = []ActionStatusEnum{
	ActiveActionStatusEnum,
	InactiveActionStatusEnum,
	PreviewActionStatusEnum,
}

var ActionDeliveryMethodList = []ActionDeliveryMethodEnum{
	EmailActionDeliveryMethod,
	SMSActionDeliveryMethod,
	PUSHMobileActionDeliveryMethod,
	PUSHWebActionDeliveryMethod,
}

type Action struct {
	ID                uid.UID                      `json:"id,omitempty"`
	Name              string                       `json:"name,omitempty"`
	Language          language_models.LanguageEnum `json:"language,omitempty"`
	EventType         string                       `json:"event_type,omitempty"`
	DeliveryMethod    ActionDeliveryMethodEnum     `json:"delivery_method,omitempty"`
	ActionStatus      ActionStatusEnum             `json:"action_status,omitempty"`
	ActionPayload     any                          `json:"action_payload,omitempty"`
	TemplateID        uid.UID                      `json:"template_id,omitempty"`
	ConditionTemplate string                       `json:"condition_template,omitempty"`
}

// EmailDeliveryMethodPayload to be used along with EmailActionDeliveryMethod
type EmailDeliveryMethodPayload struct {
	To      string `json:"to" validate:"required" field_description:"Receiver email address"`
	Subject string `json:"subject" validate:"required" field_description:"Email subject"`
}

// SMSDeliveryMethodPayload to be used along with SMSActionDeliveryMethod
type SMSDeliveryMethodPayload struct {
	PhoneNumber string `json:"phone_number" validate:"required" field_definition:"SMS receiver phone number"`
}

// PUSHMobileDeliveryMethodPayload to be used along with PUSHMobileActionDeliveryMethod
type PUSHMobileDeliveryMethodPayload struct {
	UserToken      string            `json:"user_token" validate:"required" field_description:"User's registered mobile token (do not use user id here); use helper: {{ GetUserPushToken user_id }}'"`
	Title          string            `json:"title" validate:"required" field_description:"Push notification title."`
	AdditionalData map[string]string `json:"additional_data" validate:"required" field_description:"Push additional data, used on the mobile app to trigger correct pages."`
}

// PUSHWebDeliveryMethodPayload to be used along with PUSHWebActionDeliveryMethod
type PUSHWebDeliveryMethodPayload struct {
	PartyID string `json:"user_id" validate:"required" field_description:"User's registered mobile token. Use helper: {{ GetUserPartyID user_id }}'"`
}

var ActionDeliveryMethodEnumActionDeliveryMethodPayloadSchemasList = map[ActionDeliveryMethodEnum]struct {
	Schema       any `json:"schema"`
	Descriptions any `json:"descriptions"`
}{
	EmailActionDeliveryMethod:      {Schema: utils.SchemaJSON(EmailDeliveryMethodPayload{}), Descriptions: utils.SchemaDescriptions(EmailDeliveryMethodPayload{})},
	SMSActionDeliveryMethod:        {Schema: utils.SchemaJSON(SMSDeliveryMethodPayload{}), Descriptions: utils.SchemaDescriptions(SMSDeliveryMethodPayload{})},
	PUSHMobileActionDeliveryMethod: {Schema: utils.SchemaJSON(PUSHMobileDeliveryMethodPayload{}), Descriptions: utils.SchemaDescriptions(PUSHMobileDeliveryMethodPayload{})},
	PUSHWebActionDeliveryMethod:    {Schema: utils.SchemaJSON(PUSHWebDeliveryMethodPayload{}), Descriptions: utils.SchemaDescriptions(PUSHWebDeliveryMethodPayload{})},
}
