package apiv3

// CommentContainer represents the comment container
type CommentContainer struct {
	Comments   []Comment `json:"comments"`
	MaxResults int       `json:"maxResults"`
	Self       string    `json:"self"`
	StartAt    int       `json:"startAt"`
	Total      int       `json:"total"`
}

// Comment represents a comment
type Comment struct {
	Author       *User  `json:"author"`
	Body         any    `json:"body"` // Can be string or ADF object
	Created      string `json:"created"`
	ID           string `json:"id"`
	Self         string `json:"self"`
	UpdateAuthor *User  `json:"updateAuthor"`
	Updated      string `json:"updated"`
}
