package workflows

import (
	"fmt"

	"github.com/pixie-sh/core-go/pkg/models/database_models"
	"github.com/pixie-sh/core-go/pkg/models/serializer"
	"github.com/pixie-sh/core-go/pkg/models/workflows_enums"
	"github.com/pixie-sh/core-go/pkg/models/workflows_models"
)

type WorkflowConfigurationParser struct {
	Type          workflows_enums.WorkflowType
	Configuration database_models.JSONB
}

func (p WorkflowConfigurationParser) Parse() (interface{}, error) {
	switch p.Type {
	case workflows_enums.WebhookWorflowType:
		var cfg workflows_models.WorkflowWebhookConfiguration
		err := serializer.FromJSONB(p.Configuration, &cfg, false)
		if err != nil {
			return nil, err
		}
		return cfg, nil

	case workflows_enums.ApiCallWorflowType:
		var cfg workflows_models.WorkflowAPICallConfiguration
		err := serializer.FromJSONB(p.Configuration, &cfg, false)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	return interface{}(nil), fmt.Errorf("unknown workflow type: %s", p.Type)
}
