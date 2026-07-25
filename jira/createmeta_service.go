package jira

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grokify/mogo/encoding/jsonutil"
	"github.com/grokify/mogo/net/urlutil"
)

// CreateMetaService provides access to Jira's createmeta API for discovering
// which fields are available for issue creation in specific projects.
type CreateMetaService struct {
	JRClient *Client
}

// NewCreateMetaService creates a new CreateMetaService.
func NewCreateMetaService(client *Client) *CreateMetaService {
	return &CreateMetaService{JRClient: client}
}

// GetIssueTypes returns available issue types for a project.
// Uses GET /rest/api/3/issue/createmeta/{projectKey}/issuetypes
func (svc *CreateMetaService) GetIssueTypes(ctx context.Context, projectKey string) ([]CreateMetaIssueType, error) {
	if svc.JRClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	apiURL := urlutil.JoinAbsolute(
		svc.JRClient.Config.ServerURL,
		APIV3URLCreateMeta,
		projectKey,
		"issuetypes",
	)

	hclient := svc.JRClient.HTTPClient
	if hclient == nil {
		hclient = &http.Client{}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := hclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("error status code (%d) for project %q", resp.StatusCode, projectKey)
	}

	var result CreateMetaIssueTypesResponse
	if _, err := jsonutil.UnmarshalReader(resp.Body, &result); err != nil {
		return nil, err
	}

	return result.IssueTypes, nil
}

// GetFields returns available fields for a project/issue-type combination.
// Uses GET /rest/api/3/issue/createmeta/{projectKey}/issuetypes/{issueTypeId}
func (svc *CreateMetaService) GetFields(ctx context.Context, projectKey, issueTypeID string) (CreateMetaFields, error) {
	if svc.JRClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	apiURL := urlutil.JoinAbsolute(
		svc.JRClient.Config.ServerURL,
		APIV3URLCreateMeta,
		projectKey,
		"issuetypes",
		issueTypeID,
	)

	hclient := svc.JRClient.HTTPClient
	if hclient == nil {
		hclient = &http.Client{}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := hclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("error status code (%d) for project %q, issue type %q", resp.StatusCode, projectKey, issueTypeID)
	}

	var result CreateMetaFieldsResponse
	if _, err := jsonutil.UnmarshalReader(resp.Body, &result); err != nil {
		return nil, err
	}

	return result.Values, nil
}

// GetAllFieldsForProject returns all custom fields available across all issue types
// in a project. It queries each issue type and aggregates the unique fields.
func (svc *CreateMetaService) GetAllFieldsForProject(ctx context.Context, projectKey string) (CreateMetaFields, error) {
	issueTypes, err := svc.GetIssueTypes(ctx, projectKey)
	if err != nil {
		return nil, fmt.Errorf("getting issue types for project %q: %w", projectKey, err)
	}

	seen := make(map[string]CreateMetaField)
	for _, it := range issueTypes {
		fields, err := svc.GetFields(ctx, projectKey, it.ID)
		if err != nil {
			return nil, fmt.Errorf("getting fields for project %q, issue type %q: %w", projectKey, it.ID, err)
		}
		for _, f := range fields {
			if _, exists := seen[f.Key]; !exists {
				seen[f.Key] = f
			}
		}
	}

	result := make(CreateMetaFields, 0, len(seen))
	for _, f := range seen {
		result = append(result, f)
	}
	return result, nil
}

// GetRequiredFields returns only the required fields for issue creation
// in a specific project and issue type combination.
func (svc *CreateMetaService) GetRequiredFields(ctx context.Context, projectKey, issueTypeID string) (CreateMetaFields, error) {
	fields, err := svc.GetFields(ctx, projectKey, issueTypeID)
	if err != nil {
		return nil, err
	}
	return fields.RequiredOnly(), nil
}

// GetRequiredFieldsForProject returns all required fields across all issue types
// in a project. A field is included if it is required for at least one issue type.
// The result includes which issue types require each field via RequiredFieldInfo.
func (svc *CreateMetaService) GetRequiredFieldsForProject(ctx context.Context, projectKey string) ([]RequiredFieldInfo, error) {
	issueTypes, err := svc.GetIssueTypes(ctx, projectKey)
	if err != nil {
		return nil, fmt.Errorf("getting issue types for project %q: %w", projectKey, err)
	}

	// Map field key to info about that field
	fieldInfo := make(map[string]*RequiredFieldInfo)

	for _, it := range issueTypes {
		fields, err := svc.GetFields(ctx, projectKey, it.ID)
		if err != nil {
			return nil, fmt.Errorf("getting fields for project %q, issue type %q: %w", projectKey, it.ID, err)
		}
		for _, f := range fields {
			if !f.Required {
				continue
			}
			if info, exists := fieldInfo[f.Key]; exists {
				info.IssueTypes = append(info.IssueTypes, it.Name)
			} else {
				fieldInfo[f.Key] = &RequiredFieldInfo{
					Field:      f,
					IssueTypes: []string{it.Name},
				}
			}
		}
	}

	result := make([]RequiredFieldInfo, 0, len(fieldInfo))
	for _, info := range fieldInfo {
		result = append(result, *info)
	}
	return result, nil
}
