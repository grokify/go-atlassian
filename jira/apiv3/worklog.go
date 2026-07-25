package apiv3

// WorklogContainer represents the worklog container
type WorklogContainer struct {
	MaxResults int       `json:"maxResults"`
	StartAt    int       `json:"startAt"`
	Total      int       `json:"total"`
	Worklogs   []Worklog `json:"worklogs"`
}

// Worklog represents a worklog entry
type Worklog struct {
	Author           *User  `json:"author"`
	Comment          any    `json:"comment"` // Can be string or ADF object
	Created          string `json:"created"`
	ID               string `json:"id"`
	IssueID          string `json:"issueId"`
	Self             string `json:"self"`
	Started          string `json:"started"`
	TimeSpent        string `json:"timeSpent"`
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
	UpdateAuthor     *User  `json:"updateAuthor"`
	Updated          string `json:"updated"`
}
