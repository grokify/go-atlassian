package apiv3

import "strings"

type Description struct {
	Type    string               `json:"type"`
	Version int                  `json:"version"`
	Content []DescriptionContent `json:"content"`
}

func (desc Description) String() string {
	var parts []string
	for _, c := range desc.Content {
		for _, c2 := range c.Content {
			parts = append(parts, strings.TrimSpace(c2.Text))
		}
	}
	return strings.Join(parts, " ")
}

type DescriptionContent struct {
	Type    string                `json:"type"`
	Content []DescriptionContent2 `json:"content"`
}

type DescriptionContent2 struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

/*
	"type": "doc",
	"version": 1,
	"content": [
		{
			"type": "paragraph",
			"content": [
				{
					"type": "text",
					"text": "my Descriptiont."
				}
			]
		}
	]
*/
