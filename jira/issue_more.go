package jira

import (
	"errors"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/go-atlassian/web"
	"github.com/grokify/mogo/encoding/jsonutil"
	"github.com/grokify/mogo/pointer"
	"github.com/grokify/mogo/time/duration"
	"github.com/grokify/mogo/time/timeutil"
	"github.com/grokify/mogo/type/slicesutil"
)

type IssueMore struct {
	Issue *gojira.Issue
}

func NewIssueMore(iss *gojira.Issue) IssueMore {
	return IssueMore{Issue: iss}
}

func (im *IssueMore) AdditionalFields(additionalFieldNames []string) map[string]*string {
	out := map[string]*string{}
	for _, name := range additionalFieldNames {
		cfStr, err := im.CustomFieldString(name)
		if err != nil {
			out[name] = nil
		} else {
			out[name] = pointer.Pointer(cfStr)
		}
	}
	return out
}

func (im *IssueMore) AssigneeName() string {
	if im.Issue == nil || im.Issue.Fields == nil || im.Issue.Fields.Assignee == nil {
		return ""
	}
	return im.Issue.Fields.Assignee.DisplayName
}

func (im *IssueMore) CreateTime() time.Time {
	if im.Issue == nil || im.Issue.Fields == nil {
		return time.Time{}
	}
	return time.Time(im.Issue.Fields.Created)
}

func (im *IssueMore) CreatorName() string {
	if im.Issue == nil || im.Issue.Fields == nil || im.Issue.Fields.Creator == nil {
		return ""
	}
	return im.Issue.Fields.Creator.DisplayName
}

// CustomField takes a custom value key such as `customfield_12345`.`
func (im *IssueMore) CustomField(customFieldLabel string) (IssueCustomField, error) {
	cf := IssueCustomField{}
	if im.Issue == nil {
		return cf, errors.New("issue not set")
	}
	err := GetUnmarshalCustomValue(*im.Issue, customFieldLabel, &cf)
	return cf, err
}

// CustomFieldString takes a custom value key such as `customfield_12345`.`
func (im *IssueMore) CustomFieldString(customFieldLabel string) (string, error) {
	cf, err := im.CustomField(customFieldLabel)
	return cf.Value, err
}

// CustomFieldStringOrEmpty takes a custom value key such as `customfield_12345`.`
func (im *IssueMore) CustomFieldStringOrDefault(customFieldLabel, def string) string {
	if cf, err := im.CustomField(customFieldLabel); err != nil {
		return def
	} else {
		return cf.Value
	}
}

func (im *IssueMore) Description() string {
	if im.Issue == nil || im.Issue.Fields == nil {
		return ""
	} else {
		return im.Issue.Fields.Description
	}
}

func (im *IssueMore) EpicKey() string {
	if im.Issue == nil || im.Issue.Fields == nil || im.Issue.Fields.Epic == nil {
		return ""
	} else {
		return im.Issue.Fields.Epic.Key
	}
}

func (im *IssueMore) EpicName() string {
	if im.Issue == nil || im.Issue.Fields == nil || im.Issue.Fields.Epic == nil {
		return ""
	} else if strings.TrimSpace(im.Issue.Fields.Epic.Name) != "" {
		return im.Issue.Fields.Epic.Name
	} else {
		return ""
	}
}

func (im *IssueMore) EpicNameOrSummary() string {
	if im.Issue == nil || im.Issue.Fields == nil || im.Issue.Fields.Epic == nil {
		return ""
	} else if strings.TrimSpace(im.Issue.Fields.Epic.Name) != "" {
		return im.Issue.Fields.Epic.Name
	} else if strings.TrimSpace(im.Issue.Fields.Epic.Summary) != "" {
		return im.Issue.Fields.Epic.Summary
	} else {
		return ""
	}
}

func (im *IssueMore) Key() string {
	if im.Issue == nil {
		return ""
	}
	return strings.TrimSpace(im.Issue.Key)
}

// Keys returns a slice of all keys for this issue over time including the current
// key and all previous keys. The return slice is deduped sorted.
// This relies on pulling the changelog from the Jira API.
// NOTE: keys are sorted alphabetically, not by change date.
func (im *IssueMore) Keys() (keys []string, hasChangelog bool) {
	if im.Issue == nil {
		return keys, hasChangelog
	} else if key := im.Key(); key != "" {
		keys = []string{key}
	}

	if im.Issue.Changelog == nil {
		return keys, hasChangelog
	} else {
		hasChangelog = true
	}

	for _, history := range im.Issue.Changelog.Histories {
		for _, item := range history.Items {
			if strings.ToLower(strings.TrimSpace(item.Field)) == "key" {
				// "fromString" is the old key, "toString" is the new key
				if from := strings.TrimSpace(item.FromString); from != "" {
					keys = append(keys, from)
				}
				if to := strings.TrimSpace(item.ToString); to != "" {
					keys = append(keys, to)
				}
			}
		}
	}

	keys = slicesutil.Dedupe(keys)
	sort.Strings(keys)
	return keys, hasChangelog
}

func (im *IssueMore) KeyLinkWebMarkdown(baseURL string) string {
	return web.IssueURLWebOrEmptyFromIssueKey(
		baseURL, im.Key())
}

func (im *IssueMore) KeyURLWeb(baseURL string) string {
	key := im.Key()
	baseURL = strings.TrimSpace(baseURL)
	if key == "" || baseURL == "" {
		return ""
	}
	return web.IssueURLWebOrEmptyFromIssueKey(baseURL, key)
}

func (im *IssueMore) Labels(sortAsc bool) []string {
	if im.Issue == nil || im.Issue.Fields == nil || len(im.Issue.Fields.Labels) == 0 {
		return []string{}
	} else if !sortAsc || len(im.Issue.Fields.Labels) == 1 {
		return slices.Clone(im.Issue.Fields.Labels)
	} else {
		labels := slices.Clone(im.Issue.Fields.Labels)
		sort.Strings(labels)
		return labels
	}
}

func (im *IssueMore) LabelExists(label string, matchToLowerTrimSpace bool) bool {
	labels := im.Labels(true)
	if !matchToLowerTrimSpace {
		return slices.Contains(labels, label)
	}
	label = strings.ToLower(strings.TrimSpace(label))
	for _, lblTry := range labels {
		lblTry = strings.ToLower(strings.TrimSpace(lblTry))
		if lblTry == label {
			return true
		}
	}
	return false
}

func (im *IssueMore) ParentKey() string {
	if im.Issue == nil || im.Issue.Fields == nil || im.Issue.Fields.Parent == nil {
		return ""
	}
	return strings.TrimSpace(im.Issue.Fields.Parent.Key)
}

func (im *IssueMore) Project() string {
	if im.Issue == nil || im.Issue.Fields == nil {
		return ""
	}
	return im.Issue.Fields.Project.Name
}

func (im *IssueMore) ProjectKey() string {
	if im.Issue == nil || im.Issue.Fields == nil {
		return ""
	}
	return im.Issue.Fields.Project.Key
}

func (im *IssueMore) Resolution() string {
	if im.Issue == nil || im.Issue.Fields == nil || im.Issue.Fields.Resolution == nil {
		return ""
	}
	return im.Issue.Fields.Resolution.Name
}

func (im *IssueMore) ResolutionTime() time.Time {
	if im.Issue == nil || im.Issue.Fields == nil {
		return time.Time{}
	} else {
		return time.Time(im.Issue.Fields.Resolutiondate)
	}
}

func (im *IssueMore) Status() string {
	if im.Issue == nil || im.Issue.Fields == nil || im.Issue.Fields.Status == nil {
		return ""
	}
	return im.Issue.Fields.Status.Name
}

func (im *IssueMore) Summary() string {
	if im.Issue == nil {
		return ""
	}
	return im.Issue.Fields.Summary
}

func (im *IssueMore) Type() string {
	if im.Issue == nil {
		return ""
	}
	return im.Issue.Fields.Type.Name
}

func (im *IssueMore) UpdateTime() time.Time {
	if im.Issue == nil || im.Issue.Fields == nil {
		return time.Time{}
	}
	return time.Time(im.Issue.Fields.Updated)
}

func (im *IssueMore) Value(fieldSlug string) (string, bool) {
	fieldSlug = strings.ToLower(strings.TrimSpace(fieldSlug))
	switch fieldSlug {
	case FieldKey:
		return im.Key(), true
	case AliasIssueKey:
		return im.Key(), true
	case FieldProjectKey:
		return im.ProjectKey(), true
	case CalcCreatedAgeDays:
		t := im.CreateTime()
		tm := timeutil.NewTimeMore(t, 0)
		if tm.IsZeroAny() {
			return "0", true
		} else {
			days := duration.DurationDays(time.Since(t))
			return strconv.Itoa(int(days)), true
		}
	case FieldCreatedDate:
		t := im.CreateTime()
		return t.Format(time.RFC3339), true
	case CalcCreatedMonth:
		tm := timeutil.NewTimeMore(im.CreateTime().UTC(), 0)
		return tm.MonthStart().Format(time.RFC3339), true
	case FieldResolution:
		return im.Resolution(), true
	case FieldStatus:
		return im.Status(), true
	case FieldSummary:
		return im.Summary(), true
	case FieldType:
		return im.Type(), true
	default:
		if canonicalCustomKey, ok := IsCustomFieldKey(fieldSlug); ok {
			return im.CustomFieldStringOrDefault(canonicalCustomKey, ""), true
		}
	}
	return "", false
}

func (im *IssueMore) ValueOrDefault(fieldSlug, def string) string {
	if v, ok := im.Value(fieldSlug); ok {
		return v
	} else {
		return def
	}
}

func (im *IssueMore) Meta(serverURL string, additionalFieldNames []string) IssueMeta {
	created := im.CreateTime().UTC()
	var createdPtr *time.Time
	if !created.IsZero() {
		createdPtr = &created
	}
	updated := im.UpdateTime().UTC()
	var updatedPtr *time.Time
	if !updated.IsZero() {
		updatedPtr = &updated
	}

	return IssueMeta{
		AdditionalFields: im.AdditionalFields(additionalFieldNames),
		AssigneeName:     im.AssigneeName(),
		CreateTime:       createdPtr,
		CreatorName:      im.CreatorName(),
		Description:      im.Description(),
		EpicName:         im.EpicName(),
		Key:              im.Key(),
		KeyURL:           im.KeyURLWeb(serverURL),
		Labels:           im.Labels(true),
		ParentKey:        im.ParentKey(),
		Project:          im.Project(),
		ProjectKey:       im.ProjectKey(),
		Resolution:       im.Resolution(),
		Status:           im.Status(),
		Summary:          im.Summary(),
		Type:             im.Type(),
		UpdateTime:       updatedPtr,
	}
}

func (im *IssueMore) WriteFileJSON(filename string, perm os.FileMode, prefix, indent string) error {
	return jsonutil.MarshalFile(filename, im.Issue, prefix, indent, perm)
}
