package omniroadmap

import (
	"context"

	jira "github.com/grokify/go-atlassian/jira"
	"github.com/grokify/omniroadmap-core/provider"
)

// ListCustomFieldDefinitions lists custom field definitions. If a project
// is configured (WithProjectKey), definitions are scoped to that project;
// otherwise falls back to the account-wide unscoped list. req.Kind is
// ignored — Jira's custom field APIs aren't filterable by issue type.
func (p *Provider) ListCustomFieldDefinitions(ctx context.Context, req *provider.ListCustomFieldDefinitionsRequest) (*provider.ListCustomFieldDefinitionsResponse, error) {
	defs, err := p.listCustomFields(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]provider.CustomFieldDefinition, len(defs))
	for i, d := range defs {
		out[i] = provider.CustomFieldDefinition{
			ID:   d.ID,
			Key:  d.Key,
			Name: d.Name,
			Type: d.Schema.Type,
			Metadata: map[string]any{
				"jpd.custom": d.Custom,
			},
		}
	}
	return &provider.ListCustomFieldDefinitionsResponse{Definitions: out}, nil
}

func (p *Provider) listCustomFields(ctx context.Context) (jira.CustomFields, error) {
	if p.projectKey != "" {
		defs, err := p.client.CustomFieldAPI.GetCustomFieldsForProject(ctx, p.projectKey)
		if err != nil {
			return nil, wrapErr("GetCustomFieldsForProject", err)
		}
		return defs, nil
	}
	defs, err := p.client.CustomFieldAPI.GetCustomFields()
	if err != nil {
		return nil, wrapErr("GetCustomFields", err)
	}
	return defs, nil
}
