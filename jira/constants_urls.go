package jira

const (
	APIV2URLListCustomFields = `/rest/api/2/field`
	APIV3URLIssue            = `/rest/api/3/issue` // /rest/api/3/issue/{issueIdOrKey}
	APIV3URLSearchJQL        = `/rest/api/3/search/jql`
	APIV3URLCreateMeta       = `/rest/api/3/issue/createmeta` // /rest/api/3/issue/createmeta/{projectKey}/issuetypes

	StatusCustomClosed = "Closed"

	MaxResults    = 1000
	MetaParamRank = "_rank"

	OperationAdd    = "add"
	OperationRemove = "remove"

	TimeTimeSpent                     = "Time Spent"
	TimeTimeEstimate                  = "Time Estimate"
	TimeTimeOriginalEstimate          = "Time Original Estimate"
	TimeAggregateTimeOriginalEstimate = "Aggregate Time Original Estimate"
	TimeAggregateTimeSpent            = "Aggregate Time Spent"
	TimeAggregateTimeEstimate         = "Aggregate Time Estimate"
	TimeTimeRemaining                 = "Time Remaining"
	TimeTimeRemainingOriginal         = "Time Remaining Original"

	FieldSlugType       = "type"
	FieldSlugProjectkey = "projectkey"
)

type Status struct {
	Name        string
	Description string
}

func IssueStatuses() []Status {
	return []Status{
		{
			Name:        "Open",
			Description: "The issue is open and ready for the assignee to start work on it.",
		},
	}
}
