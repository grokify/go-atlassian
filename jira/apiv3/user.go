package apiv3

// User represents a user in the V3 API
type User struct {
	AccountID    string            `json:"accountId"`
	AccountType  string            `json:"accountType"`
	Active       bool              `json:"active"`
	AvatarUrls   map[string]string `json:"avatarUrls"`
	DisplayName  string            `json:"displayName"`
	EmailAddress string            `json:"emailAddress"`
	Self         string            `json:"self"`
	TimeZone     string            `json:"timeZone"`
}
