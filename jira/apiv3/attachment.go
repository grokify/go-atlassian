package apiv3

// Attachment represents an attachment
type Attachment struct {
	Author    *User  `json:"author"`
	Content   string `json:"content"`
	Created   string `json:"created"`
	Filename  string `json:"filename"`
	ID        string `json:"id"`
	MimeType  string `json:"mimeType"`
	Self      string `json:"self"`
	Size      int    `json:"size"`
	Thumbnail string `json:"thumbnail"`
}

// Component represents a component
type Component struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Self        string `json:"self"`
}

// Version represents a version
type Version struct {
	Archived        bool    `json:"archived"`
	Description     string  `json:"description"`
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	ProjectID       int     `json:"projectId"`
	ReleaseDate     *string `json:"releaseDate"`
	Released        bool    `json:"released"`
	Self            string  `json:"self"`
	StartDate       *string `json:"startDate"`
	UserReleaseDate *string `json:"userReleaseDate"`
	UserStartDate   *string `json:"userStartDate"`
}
