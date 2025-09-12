package workflows_models

import (
	"time"

	"github.com/pixie-sh/core-go/pkg/models"
	"github.com/pixie-sh/core-go/pkg/models/database_models"
	"github.com/pixie-sh/core-go/pkg/models/workflows_enums"
	"github.com/pixie-sh/core-go/pkg/uid"
)

type WorkflowWebhookConfiguration struct {
	PayloadTemplate string `json:"payload_template"`
	URL             string `json:"url"`
}

type WorkflowAPICallConfiguration struct {
	PayloadTemplate string `json:"payload_template"`
	EndpointName    string `json:"endpoint_name"`
}

type Workflow struct {
	ID                    uid.UID                          `json:"id"`
	EventName             string                           `json:"event_name"`
	DisplayName           string                           `json:"display_name"`
	WorkflowProvider      workflows_enums.WorkflowProvider `json:"workflow_provider"`
	WorkflowType          workflows_enums.WorkflowType     `json:"workflow_type"`
	WorkflowConfiguration interface{}                      `json:"workflow_configuration"`

	models.SoftDeletable
}

type WorkflowActivityLog struct {
	ID                uid.UID                               `json:"id"`
	EventName         string                                `json:"event_name"`
	WorkflowID        uid.UID                               `json:"workflow_id"`
	Workflow          database_models.JSONB                 `json:"workflow"`
	Payload           database_models.JSONB                 `json:"payload"`
	Result            database_models.JSONB                 `json:"result"`
	TriggeredAt       time.Time                             `json:"triggered_at"`
	TriggerSource     workflows_enums.WorkflowTriggerSource `json:"trigger_source"`
	TriggeredByUserID uint64                                `json:"triggered_by_user_id"`

	models.SoftDeletable
}
