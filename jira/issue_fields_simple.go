package jira

/*
import (
	gojira "github.com/andygrunwald/go-jira"
)

type IssueFieldsSimple struct {
	Fields *gojira.IssueFields
}

func (ifs IssueFieldsSimple) EpicKey() string {
	if ifs.Fields == nil || ifs.Fields.Epic == nil {
		return ""
	} else {
		return ifs.Fields.Epic.Key
	}
}

func (ifs IssueFieldsSimple) EpicName() string {
	if ifs.Fields == nil || ifs.Fields.Epic == nil {
		return ""
	} else {
		if ifs.Fields.Epic.Name != "" {
			return ifs.Fields.Epic.Name
		}
		if ifs.Fields.Epic.Summary != "" {
			return ifs.Fields.Epic.Summary
		}
		return " "
	}
}

func (ifs IssueFieldsSimple) ResolutionName() string {
	if ifs.Fields == nil || ifs.Fields.Resolution == nil {
		return ""
	} else {
		return ifs.Fields.Resolution.Name
	}
}

func (ifs IssueFieldsSimple) StatusName() string {
	if ifs.Fields == nil || ifs.Fields.Status == nil {
		return ""
	} else {
		return ifs.Fields.Status.Name
	}
}
*/
