// Package omniroadmap implements omniroadmap-core's provider.Provider for
// Jira Product Discovery (JPD), wrapping go-atlassian's existing jira.Client.
//
// Scope is deliberately narrow, per feasibility research: JPD "Ideas" are
// ordinary Jira issues (a JPD-specific issue type in a team-managed
// project), reachable via Jira's standard public REST API with zero new
// client code. JPD's distinctive features — Views (prioritization
// matrices), Insights (weighted scoring), and roadmap/timeline
// visualization — have no officially supported public API as of this
// writing (Atlassian staff have confirmed no public API for Views; Insights
// is only reachable via an undocumented internal GraphQL endpoint not safe
// to build on). This adapter does not attempt to expose any of that.
//
// Like aha-go and productboard-go, go-atlassian is a from-scratch SDK, so
// this adapter lives as an embedded omniroadmap/ subpackage rather than a
// separate repo.
package omniroadmap

import (
	jira "github.com/grokify/go-atlassian/jira"
	omniroadmap "github.com/grokify/omniroadmap-core"
	"github.com/grokify/omniroadmap-core/provider"
)

// providerName is the registry key and Provider.Name() value for this
// adapter. Named "jpd", not "jira" — this adapter only covers the narrow
// JPD-ideas-as-issues slice, not general Jira issue tracking.
const providerName = "jpd"

// defaultIdeaIssueType is the conventional issue type name JPD uses for
// Ideas. It's configurable via WithIdeaIssueType since it can be renamed
// per-project.
const defaultIdeaIssueType = "Idea"

// Provider implements provider.Provider using go-atlassian's *jira.Client,
// scoped to a single JPD project.
type Provider struct {
	client        *jira.Client
	projectKey    string
	ideaIssueType string
}

var _ provider.Provider = (*Provider)(nil)

// Option configures a Provider.
type Option func(*Provider)

// WithProjectKey scopes ListItems/GetItem to a single Jira project (JPD
// ideas live in a team-managed project). Required for ListItems — without
// it, ListItems returns omniroadmap.ErrUnsupportedOperation, since
// searching without any project scope would return every issue the
// credentials can see, not just Ideas.
func WithProjectKey(projectKey string) Option {
	return func(p *Provider) { p.projectKey = projectKey }
}

// WithIdeaIssueType overrides the issue type name used to identify Ideas
// (defaults to "Idea"). Set this if the project renamed its JPD issue type.
func WithIdeaIssueType(issueType string) Option {
	return func(p *Provider) { p.ideaIssueType = issueType }
}

// NewProvider wraps client as a provider.Provider.
func NewProvider(client *jira.Client, opts ...Option) *Provider {
	p := &Provider{client: client, ideaIssueType: defaultIdeaIssueType}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Provider) Name() string { return providerName }

// Close is a no-op: jira.Client owns no closable resources beyond its HTTP
// client, which the standard library manages for us.
func (p *Provider) Close() error { return nil }

func (p *Provider) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		Kinds:                []provider.ItemKind{provider.ItemKindFeature},
		SupportsReleases:     false,
		SupportsObjectives:   false,
		SupportsCustomFields: true,
		SupportsWrite:        false,
	}
}

func init() {
	_ = omniroadmap.RegisterProvider(providerName, func(config any) (provider.Provider, error) {
		client, ok := config.(*jira.Client)
		if !ok {
			return nil, omniroadmap.NewAPIError(providerName, 0, "invalid_config",
				"omniroadmap/jpd: expected *jira.Client config")
		}
		return NewProvider(client), nil
	})
}
