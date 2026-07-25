package jira

import (
	"strings"
	"time"

	"github.com/grokify/go-atlassian/web"
	"github.com/grokify/mogo/text/markdown"
)

type IssueMetas []IssueMeta

// HighestEpic returns the highest most Epic.
func (ims IssueMetas) HighestEpic() *IssueMeta {
	return ims.HighestType(TypeEpic)
}

func (ims IssueMetas) HighestType(issueType string) *IssueMeta {
	im := IssueMeta{}
	typeFound := true
	for _, imx := range ims {
		if imx.Type == issueType {
			im = imx
			typeFound = true
		}
	}
	if !typeFound {
		return nil
	}
	return &im
}

// HighestAboveEpic returns the highest item that follows an Epic.
func (ims IssueMetas) HighestAboveEpic() *IssueMeta {
	im := IssueMeta{}
	epicFound := false
	postEpicFound := false
	for _, imx := range ims {
		if imx.Type == TypeEpic {
			epicFound = true
		} else if epicFound && imx.Type != TypeEpic {
			im = imx
			postEpicFound = true
		}
	}
	if !postEpicFound {
		return nil
	}
	return &im
}

type IssueMeta struct {
	AdditionalFields map[string]*string
	AssigneeName     string
	CreateTime       *time.Time
	CreatorName      string
	Description      string
	EpicName         string
	Key              string
	KeyURL           string
	Labels           []string
	ParentKey        string
	Project          string
	ProjectKey       string
	Resolution       string
	Status           string
	Summary          string
	Type             string
	UpdateTime       *time.Time
}

func (im *IssueMeta) String() string {
	k := strings.TrimSpace(im.Key)
	s := strings.TrimSpace(im.Summary)
	if k == "" && s == "" {
		return ""
	}
	parts := []string{}
	if len(k) > 0 {
		parts = append(parts, k)
	}
	if len(s) > 0 {
		parts = append(parts, s)
	}
	return strings.Join(parts, ": ")
}

func (im *IssueMeta) BuildKeyURL(baseURL string) {
	if strings.TrimSpace(baseURL) != "" && strings.TrimSpace(im.Key) != "" {
		im.KeyURL = web.IssueLinkWebMarkdownOrEmptyFromIssueKey(
			strings.TrimSpace(baseURL),
			strings.TrimSpace(im.Key))
	}
}

// KeyLinkMarkdown returns a link of both `Key` and `KeyURL` are non-empty,`Key` if `Key` is non-empty or
// an empty string if both are empty.
func (im *IssueMeta) KeyLinkMarkdown() string {
	if strings.TrimSpace(im.Key) == "" {
		return ""
	} else if strings.TrimSpace(im.KeyURL) == "" {
		return im.Key
	} else {
		return markdown.Linkify(im.KeyURL, im.Key)
	}
}
