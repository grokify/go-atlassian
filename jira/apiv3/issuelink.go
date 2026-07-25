package apiv3

// IssueLink represents an issue link
type IssueLink struct {
	ID           string         `json:"id"`
	InwardIssue  *Issue         `json:"inwardIssue"`
	OutwardIssue *Issue         `json:"outwardIssue"`
	Self         string         `json:"self"`
	Type         *IssueLinkType `json:"type"`
}

// IssueLinkType represents an issue link type
type IssueLinkType struct {
	ID      string `json:"id"`
	Inward  string `json:"inward"`
	Name    string `json:"name"`
	Outward string `json:"outward"`
	Self    string `json:"self"`
}
