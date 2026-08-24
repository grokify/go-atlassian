package omniroadmap

import (
	"net/http"
	"net/http/httptest"
	"testing"

	jira "github.com/grokify/go-atlassian/jira"
	"github.com/grokify/omniroadmap-core/provider"
	"github.com/grokify/omniroadmap-core/provider/providertest"
)

func newTestProvider(t *testing.T, handler http.HandlerFunc, opts ...Option) *Provider {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := jira.NewClientFromHTTPClient(server.URL, http.DefaultClient)
	if err != nil {
		t.Fatalf("jira.NewClientFromHTTPClient: %v", err)
	}
	return NewProvider(client, opts...)
}

func TestConformance(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[],"total":0}`))
	}, WithProjectKey("PROJ"))

	providertest.RunAll(t, providertest.Config{
		Provider:        p,
		SkipIntegration: true,
	})
}

func TestListItems_RequiresProjectKey(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no HTTP call expected without a configured projectKey")
	})

	_, err := p.ListItems(t.Context(), &provider.ListItemsRequest{})
	if err == nil {
		t.Fatal("expected ErrUnsupportedOperation without a projectKey, got nil")
	}
}

func TestListReleases_Unsupported(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no HTTP call expected — ListReleases is unsupported in this adapter's scope")
	})

	_, err := p.ListReleases(t.Context(), &provider.ListReleasesRequest{})
	if err == nil {
		t.Fatal("expected ErrUnsupportedOperation, got nil")
	}
}

func TestGetItem_UnsupportedKind(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no HTTP call expected for an unsupported kind")
	})

	_, err := p.GetItem(t.Context(), &provider.GetItemRequest{
		Kind: provider.ItemKindInitiative,
		ID:   "PROJ-1",
	})
	if err == nil {
		t.Fatal("expected ErrUnsupportedOperation for a non-Feature kind, got nil")
	}
}
