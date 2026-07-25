package jira

import (
	"net/url"
)

// JQLsReportMarkdownLines provides Markdownlines for a set of JQLs, including querying the number
// of results for each JQL via the Jira API.
func (svc *IssueService) JQLsReportMarkdownLines(jqls JQLs, opts *JQLsReportMarkdownOpts) ([]string, error) {
	if jqls, err := svc.JQLsAddMetadata(jqls); err != nil {
		return []string{}, err
	} else {
		if opts == nil {
			opts = &JQLsReportMarkdownOpts{}
		}
		if opts.IssuesWebURL == "" && svc.Client != nil && svc.Client.Config != nil {
			opts.IssuesWebURL = svc.Client.Config.WebURLIssues(url.Values{})
		}
		return jqls.ReportMarkdownLines(opts)
	}
}
