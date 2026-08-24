package omniroadmap

import (
	"context"

	omniroadmap "github.com/grokify/omniroadmap-core"
	"github.com/grokify/omniroadmap-core/provider"
)

// ListReleases is unsupported: this adapter's scope is deliberately
// narrowed to JPD ideas-as-issues (see package doc). Jira's "fixVersions"
// could loosely map to a release concept, but that's out of scope for this
// pass — Capabilities().SupportsReleases is false, matching this.
func (p *Provider) ListReleases(ctx context.Context, req *provider.ListReleasesRequest) (*provider.ListReleasesResponse, error) {
	return nil, omniroadmap.ErrUnsupportedOperation
}
