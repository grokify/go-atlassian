package apiv3

// IssueType represents an issue type in the V3 API
type IssueType struct {
	AvatarID       int    `json:"avatarId"`
	Description    string `json:"description"`
	HierarchyLevel int    `json:"hierarchyLevel"`
	IconURL        string `json:"iconUrl"`
	ID             string `json:"id"`
	Name           string `json:"name"`
	Self           string `json:"self"`
	Subtask        bool   `json:"subtask"`
}

// Status represents a status in the V3 API
type Status struct {
	Description    string          `json:"description"`
	IconURL        string          `json:"iconUrl"`
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Self           string          `json:"self"`
	StatusCategory *StatusCategory `json:"statusCategory"`
}

// StatusCategory represents a status category
type StatusCategory struct {
	ColorName string `json:"colorName"`
	ID        int    `json:"id"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	Self      string `json:"self"`
}

// Priority represents a priority in the V3 API
type Priority struct {
	IconURL string `json:"iconUrl"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Self    string `json:"self"`
}

// Progress represents progress information
type Progress struct {
	Progress int `json:"progress"`
	Total    int `json:"total"`
}

// Resolution represents a resolution
type Resolution struct {
	Description string `json:"description"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Self        string `json:"self"`
}

// SecurityLevel represents a security level
type SecurityLevel struct {
	Description string `json:"description"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Self        string `json:"self"`
}

// TimeTracking represents time tracking information
type TimeTracking struct {
	OriginalEstimate         string `json:"originalEstimate"`
	OriginalEstimateSeconds  int    `json:"originalEstimateSeconds"`
	RemainingEstimate        string `json:"remainingEstimate"`
	RemainingEstimateSeconds int    `json:"remainingEstimateSeconds"`
	TimeSpent                string `json:"timeSpent"`
	TimeSpentSeconds         int    `json:"timeSpentSeconds"`
}

// Votes represents vote information
type Votes struct {
	HasVoted bool   `json:"hasVoted"`
	Self     string `json:"self"`
	Votes    int    `json:"votes"`
}

// Watches represents watch information
type Watches struct {
	IsWatching bool   `json:"isWatching"`
	Self       string `json:"self"`
	WatchCount int    `json:"watchCount"`
}

// CustomFieldOption represents a custom field option value
type CustomFieldOption struct {
	ID    string `json:"id"`
	Self  string `json:"self"`
	Value string `json:"value"`
}

// IssueRestriction represents issue restrictions
type IssueRestriction struct {
	IssueRestrictions map[string]any `json:"issuerestrictions"`
	ShouldDisplay     bool           `json:"shouldDisplay"`
}
