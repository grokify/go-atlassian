package apiv3

// Project represents a project in the V3 API
type Project struct {
	AvatarUrls      map[string]string `json:"avatarUrls"`
	ID              string            `json:"id"`
	Key             string            `json:"key"`
	Name            string            `json:"name"`
	ProjectCategory *ProjectCategory  `json:"projectCategory"`
	ProjectTypeKey  string            `json:"projectTypeKey"`
	Self            string            `json:"self"`
	Simplified      bool              `json:"simplified"`
}

// ProjectCategory represents a project category
type ProjectCategory struct {
	Description string `json:"description"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Self        string `json:"self"`
}
