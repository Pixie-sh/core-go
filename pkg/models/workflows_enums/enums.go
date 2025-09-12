package workflows_enums

type WorkflowProvider string

const (
	HubspotWorflowProvider WorkflowProvider = "hubspot"
) //@Field WorkflowProvider

var WorkflowProviderList = []WorkflowProvider{
	HubspotWorflowProvider,
}

type WorkflowType string

const (
	WebhookWorflowType WorkflowType = "webhook"
	ApiCallWorflowType WorkflowType = "api_call"
) //@Field WorkflowType

var WorkflowTypeList = []WorkflowType{
	WebhookWorflowType,
	ApiCallWorflowType,
}

type WorkflowTriggerSource string

const (
	ManualWorkflowTriggerSource WorkflowTriggerSource = "manual"
	EventWorkflowTriggerSource  WorkflowTriggerSource = "event"
) //@Field WorkflowTriggerSource
