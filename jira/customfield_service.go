package jira

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grokify/mogo/encoding/jsonutil"
	"github.com/grokify/mogo/net/urlutil"
)

type CustomFieldService struct {
	JRClient *Client
}

func NewCustomFieldService(client *Client) *CustomFieldService {
	return &CustomFieldService{JRClient: client}
}

func (svc *CustomFieldService) GetCustomFields() (CustomFields, error) {
	var cfs CustomFields
	if svc.JRClient == nil {
		return cfs, ErrJiraRESTClientCannotBeNil
	}
	apiURL := urlutil.JoinAbsolute(svc.JRClient.Config.ServerURL, APIV2URLListCustomFields)
	hclient := svc.JRClient.HTTPClient
	if hclient == nil {
		hclient = &http.Client{}
	}

	resp, err := hclient.Get(apiURL)
	if err != nil {
		return cfs, err
	}
	if resp.StatusCode >= 300 {
		return cfs, fmt.Errorf("error status code (%d)", resp.StatusCode)
	}
	_, err = jsonutil.UnmarshalReader(resp.Body, &cfs)
	return cfs, err
}

func (svc *CustomFieldService) GetCustomFieldEpicLink() (CustomField, error) {
	return svc.GetCustomField(CustomFieldNameEpicLink)
}

// GetCustomField returns a single custom field by name. If no field matches or
// multiple fields match, an error is returned. Use GetCustomFieldsByName when
// you need to handle duplicate names, or GetCustomFieldByID for exact ID matches.
func (svc *CustomFieldService) GetCustomField(customFieldName string) (CustomField, error) {
	cfs, err := svc.GetCustomFields()
	if err != nil {
		return CustomField{}, err
	}
	cfsName := cfs.FilterByNames(customFieldName)
	switch len(cfsName) {
	case 0:
		return CustomField{}, fmt.Errorf("custom field not found: %q", customFieldName)
	case 1:
		return cfsName[0], nil
	default:
		return CustomField{}, fmt.Errorf("multiple custom fields found with name %q (count: %d); use GetCustomFieldsByName or filter by ID", customFieldName, len(cfsName))
	}
}

// GetCustomFieldsByName returns all custom fields matching the given name.
// This handles the case where multiple fields share the same display name.
func (svc *CustomFieldService) GetCustomFieldsByName(name string) (CustomFields, error) {
	cfs, err := svc.GetCustomFields()
	if err != nil {
		return nil, err
	}
	return cfs.FilterByNames(name), nil
}

// GetCustomFieldByID returns the custom field with the exact ID.
func (svc *CustomFieldService) GetCustomFieldByID(id string) (CustomField, error) {
	cfs, err := svc.GetCustomFields()
	if err != nil {
		return CustomField{}, err
	}
	filtered := cfs.FilterByIDs(id)
	if len(filtered) == 0 {
		return CustomField{}, fmt.Errorf("custom field not found: %s", id)
	}
	return filtered[0], nil
}

func (svc *CustomFieldService) GetCustomFieldSet() (*CustomFieldSet, error) {
	if cfs, err := svc.GetCustomFields(); err != nil {
		return nil, err
	} else {
		set := NewCustomFieldSet()
		if err := set.Add(cfs...); err != nil {
			return nil, err
		} else {
			return set, nil
		}
	}
}

// GetCustomFieldsForProject returns custom fields available for a specific project,
// with full metadata from the global field list. It uses the createmeta API to
// identify which fields are available in the project, then enriches them with
// full metadata from the field list.
func (svc *CustomFieldService) GetCustomFieldsForProject(ctx context.Context, projectKey string) (CustomFields, error) {
	if svc.JRClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	// Get all custom fields with full metadata
	allFields, err := svc.GetCustomFields()
	if err != nil {
		return nil, fmt.Errorf("getting custom fields: %w", err)
	}

	// Get fields available in the project via createmeta API
	createMetaSvc := NewCreateMetaService(svc.JRClient)
	projectFields, err := createMetaSvc.GetAllFieldsForProject(ctx, projectKey)
	if err != nil {
		return nil, fmt.Errorf("getting project fields: %w", err)
	}

	// Get only custom fields from createmeta
	customFieldKeys := projectFields.CustomOnly().Keys()
	if len(customFieldKeys) == 0 {
		return CustomFields{}, nil
	}

	// Filter full metadata by the project's custom field keys
	return allFields.FilterByIDs(customFieldKeys...), nil
}
