package workflows_models

import "github.com/pixie-sh/core-go/pkg/models/workflows_enums"

type CreateWorkflowRequest struct {
	//Common fields
	EventName        string                           `json:"event_name" validate:"required"`
	DisplayName      string                           `json:"display_name" validate:"required"`
	WorkflowProvider workflows_enums.WorkflowProvider `json:"workflow_provider" validate:"required,isEnum:WorkflowProvider"`
	WorkflowType     workflows_enums.WorkflowType     `json:"workflow_type" validate:"required,isEnum:WorkflowType"`
	PayloadTemplate  string                           `json:"payload_template" validate:"required"`

	//Webhook specific
	URL string `json:"url" validate:"required_if=WorkflowType webhook,omitempty"`

	//API Call specific
	EndpointName string `json:"endpoint_name" validate:"required_if=WorkflowType api_call,omitempty"`
}

type UpdateWorkflowRequest struct {
	//Common fields
	EventName        string                           `json:"event_name" validate:"required"`
	DisplayName      string                           `json:"display_name" validate:"required"`
	WorkflowProvider workflows_enums.WorkflowProvider `json:"workflow_provider" validate:"required,isEnum:WorkflowProvider"`
	WorkflowType     workflows_enums.WorkflowType     `json:"workflow_type" validate:"required,isEnum:WorkflowType"`
	PayloadTemplate  string                           `json:"payload_template" validate:"required"`

	//Webhook specific
	URL string `json:"url" validate:"required_if=WorkflowType webhook,omitempty"`

	//API Call specific
	EndpointName string `json:"endpoint_name" validate:"required_if=WorkflowType api_call,omitempty"`
}

type TriggerEventWorkflowsRequest struct {
	Data any `json:"data" validate:"required"`
}
