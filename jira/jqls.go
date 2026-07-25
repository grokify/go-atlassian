package jira

import (
	"fmt"
	"strings"
	"time"

	"github.com/grokify/mogo/text/markdown"
)

type JQLs []JQL

func (jqls JQLs) JoinString(keyword string) string {
	var parts []string
	for _, jql := range jqls {
		parts = append(parts, "("+jql.String()+")")
	}
	return strings.Join(parts, " "+keyword+" ")
}

type JQLsReportMarkdownOpts struct {
	IssuesWebURL   string // should end with `issues/?`, query string `jql=...` will be appended.
	HeaderPrefix   string
	TimeLocation   *time.Location
	TimeZone       string
	AddCount       bool
	AddExplicitURL bool
}

func (opts JQLsReportMarkdownOpts) GetTimeLocation() (*time.Location, error) {
	if opts.TimeLocation != nil {
		return opts.TimeLocation, nil
	} else if tz := strings.TrimSpace(opts.TimeZone); tz != "" {
		return time.LoadLocation(tz)
	} else {
		return nil, nil
	}
}

// ReportMarkdownLines provides Markdownlines for a set of JQLs.
// The `JQLsReportMarkdownOpts.AddCount` option adds a static count to the report. This is useful when the
// report isn't auto-updating, such as on a code repo.
// The `JQLsReportMarkdownOpts.AddExplicitURL` option adds a URL which can be pasted into Confluence. This
// will be interpreted to load a dynamic table in the Confluence page. This is not needed for git repo pages.
func (jqls JQLs) ReportMarkdownLines(opts *JQLsReportMarkdownOpts) ([]string, error) {
	if opts == nil {
		opts = &JQLsReportMarkdownOpts{}
	}
	var lines []string
	issuesWebURL := strings.TrimSpace(opts.IssuesWebURL) // should end with `issues/?`
	timeLoc, err := opts.GetTimeLocation()
	if err != nil {
		return lines, err
	}
	for i, j := range jqls {
		dtStr := ""
		count := -1
		if dt := j.Meta.QueryTime; !dt.IsZero() {
			if timeLoc != nil {
				dt = dt.In(timeLoc)
			}
			dtStr = fmt.Sprintf(" at %s", dt.Format(time.RFC1123))
			count = j.Meta.QueryTotalCount
		}

		name := strings.TrimSpace(j.Meta.Name)
		if name == "" {
			name = fmt.Sprintf("JQL #%d", i+1)
		}
		lines = append(lines, "", fmt.Sprintf("%s%s", opts.HeaderPrefix, name))

		if key := strings.TrimSpace(j.Meta.Key); key != "" {
			lines = append(lines, fmt.Sprintf("* Key: `%s`", key))
		}
		if desc := strings.TrimSpace(j.Meta.Description); desc != "" {
			lines = append(lines, fmt.Sprintf("* Description: %s", desc))
		}
		if opts.AddCount && count > -1 {
			lines = append(lines, fmt.Sprintf("* Issue Count: **%d**%s", count, dtStr))
		}
		if len(j.Meta.Directions) > 0 {
			lines = append(lines, "* Directions")
			lines = append(lines, j.Meta.Directions...)
		}

		js := strings.TrimSpace(j.String())
		jq := strings.TrimSpace(j.QueryString())
		if jq != "" {
			jq = issuesWebURL + jq
		}
		mk := markdown.Linkify(jq, "`"+js+"`")
		if mk != "" {
			lines = append(lines, fmt.Sprintf("* JQL: %s", mk))
		}
		if opts.AddExplicitURL && jq != "" {
			lines = append(lines, "", markdown.Linkify(jq, jq))
		}
	}
	return lines, nil
}
