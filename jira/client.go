package jira

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/goauth"
	"github.com/grokify/goauth/authutil"
	"github.com/grokify/mogo/errors/errorsutil"
	"github.com/grokify/mogo/log/slogutil"
	"github.com/grokify/mogo/net/http/httpsimple"
)

// Client provides access to the Jira REST API. It wraps both an HTTP client
// for custom requests and a go-jira client for standard operations.
// Use one of the NewClient* functions to create a properly initialized client.
type Client struct {
	Config         *Config
	HTTPClient     *http.Client
	JiraClient     *gojira.Client
	simpleClient   *httpsimple.Client
	Logger         *slog.Logger
	BacklogAPI     *BacklogService
	BoardAPI       *BoardService
	CreateMetaAPI  *CreateMetaService
	CustomFieldAPI *CustomFieldService
	IssueAPI       *IssueService
	CustomFieldSet *CustomFieldSet
}

// NewClientFromBasicAuth creates a new Client using basic authentication.
// If addCustomFieldSet is true, custom fields are loaded during initialization.
func NewClientFromBasicAuth(serverURL, username, password string, addCustomFieldSet bool) (*Client, error) {
	if hclient, err := authutil.NewClientBasicAuth(username, password, false); err != nil {
		return nil, err
	} else if jclient, err := JiraClientBasicAuth(serverURL, username, password); err != nil {
		return nil, err
	} else {
		return newClientFromClients(hclient, jclient, serverURL, addCustomFieldSet)
	}
}

// NewClientFromGoauthCLI creates a new Client by interactively selecting credentials
// from the goauth CLI. If inclAccountsOnError is true, available accounts are shown
// when an error occurs during selection.
func NewClientFromGoauthCLI(inclAccountsOnError, addCustomFieldSet bool) (*Client, error) {
	if creds, err := goauth.NewCredentialsFromCLI(inclAccountsOnError); err != nil {
		return nil, err
	} else {
		return NewClientFromGoauthCredentials(&creds, addCustomFieldSet)
	}
}

// NewClientFromGoauthCredentials creates a new Client from goauth credentials.
// Currently only basic authentication is supported.
func NewClientFromGoauthCredentials(c *goauth.Credentials, addCustomFieldSet bool) (*Client, error) {
	if c == nil {
		return nil, errors.New("goauth.Credentials cannot be nil")
	} else if c.Type == goauth.TypeBasic && c.Basic != nil {
		return NewClientFromBasicAuth(c.Basic.ServerURL, c.Basic.Username, c.Basic.Password, addCustomFieldSet)
	} else {
		return nil, errors.New("auth method not supported or populated")
	}
}

// NewClientGoauthCredentialsSetFile creates a new Client from a goauth credentials file.
// The accountkey specifies which account to use from the credentials set.
func NewClientGoauthCredentialsSetFile(filename, accountkey string, addCustomFieldSet, inclAccountsOnError bool) (*Client, error) {
	if creds, err := goauth.NewCredentialsFromSetFile(filename, accountkey, inclAccountsOnError); err != nil {
		return nil, err
	} else if client, err := NewClientFromGoauthCredentials(&creds, addCustomFieldSet); err != nil {
		return nil, err
	} else {
		return client, nil
	}
}

// NewClientFromHTTPClient creates a new Client from an existing http.Client.
// This is useful when the HTTP client already has authentication configured
// (e.g., via omnitoken or OAuth2).
func NewClientFromHTTPClient(serverURL string, httpClient *http.Client) (*Client, error) {
	jclient, err := gojira.NewClient(httpClient, serverURL)
	if err != nil {
		return nil, fmt.Errorf("create go-jira client: %w", err)
	}
	return newClientFromClients(httpClient, jclient, serverURL, false)
}

func newClientFromClients(hclient *http.Client, jclient *gojira.Client, serverURL string, addCustomFieldSet bool) (*Client, error) {
	c := &Client{
		HTTPClient: hclient,
		JiraClient: jclient}
	cfg := NewConfigDefault()
	cfg.ServerURL = serverURL
	c.Config = cfg
	sc := httpsimple.NewClient(hclient, serverURL)
	c.simpleClient = &sc
	if err := c.Inflate(addCustomFieldSet); err != nil {
		return nil, err
	}
	return c, nil
}

// NewClientGoauthBasicAuthFile creates a new Client from a goauth credentials file
// using basic authentication. This is the recommended way to create a client when
// credentials are stored in a file.
func NewClientGoauthBasicAuthFile(filename, credsKey string, addCustomFieldSet bool) (*Client, error) {
	if hclient, serverURL, err := NewClientHTTPBasicAuthFile(filename, credsKey); err != nil {
		return nil, errorsutil.Wrapf(err, `rest.ClientsBasicAuthFile() (%s)`, filename)
	} else if jclient, err := NewClientJiraBasicAuthFile(filename, credsKey); err != nil {
		return nil, errorsutil.Wrap(err, `rest.ClientsBasicAuthFile()..JiraClientBasicAuthFile()`)
	} else {
		return newClientFromClients(hclient, jclient, serverURL, addCustomFieldSet)
	}
}

// NewCredentialsGoauthFile reads credentials from a goauth credentials file.
func NewCredentialsGoauthFile(filename, credsKey string) (*goauth.Credentials, error) {
	if cs, err := goauth.ReadFileCredentialsSet(filename, true); err != nil {
		return nil, err
	} else if creds, err := cs.Get(credsKey); err != nil {
		return nil, err
	} else {
		return &creds, nil
	}
}

// NewCredentialsBasicAuthGoauthFile reads basic auth credentials from a goauth file.
func NewCredentialsBasicAuthGoauthFile(filename, credsKey string) (*goauth.CredentialsBasicAuth, error) {
	if creds, err := NewCredentialsGoauthFile(filename, credsKey); err != nil {
		return nil, err
	} else {
		return creds.Basic, nil
	}
}

// NewClientHTTPBasicAuthFile creates an HTTP client with basic auth from a goauth file.
func NewClientHTTPBasicAuthFile(filename, credsKey string) (hclient *http.Client, serverURL string, err error) {
	if creds, err := NewCredentialsBasicAuthGoauthFile(filename, credsKey); err != nil {
		return nil, "", err
	} else if hclient, err = creds.NewClient(); err != nil {
		return hclient, "", err
	} else {
		serverURL = creds.ServerURL
	}
	return
}

// NewClientJiraBasicAuthFile creates a go-jira client with basic auth from a goauth file.
func NewClientJiraBasicAuthFile(filename, credsKey string) (*gojira.Client, error) {
	if creds, err := NewCredentialsBasicAuthGoauthFile(filename, credsKey); err != nil {
		return nil, err
	} else {
		return JiraClientBasicAuthGoauth(creds)
	}
}

// JiraClientBasicAuth creates a go-jira client with basic authentication.
func JiraClientBasicAuth(serverURL, username, password string) (*gojira.Client, error) {
	tp := gojira.BasicAuthTransport{
		Username: username,
		Password: password}
	return gojira.NewClient(tp.Client(), serverURL)
}

// JiraClientBasicAuthGoauth creates a go-jira client from goauth basic credentials.
func JiraClientBasicAuthGoauth(creds *goauth.CredentialsBasicAuth) (*gojira.Client, error) {
	if creds == nil {
		return nil, errors.New("goauth.CredentialsBasicAuth cannot be nil")
	}
	return JiraClientBasicAuth(creds.ServerURL, creds.Username, creds.Password)
}

// Inflate initializes the client's service APIs (BacklogAPI, BoardAPI, CreateMetaAPI, CustomFieldAPI, IssueAPI).
// If addCustomFieldSet is true, custom fields are loaded from the Jira server.
func (c *Client) Inflate(addCustomFieldSet bool) error {
	c.BacklogAPI = NewBacklogService(c)
	c.BoardAPI = NewBoardService(c)
	c.CreateMetaAPI = NewCreateMetaService(c)
	c.CustomFieldAPI = NewCustomFieldService(c)
	c.IssueAPI = NewIssueService(c)
	if addCustomFieldSet {
		if err := c.LoadCustomFields(); err != nil {
			return err
		}
	}
	return nil
}

// LoadCustomFields fetches and caches custom field definitions from the Jira server.
func (c *Client) LoadCustomFields() error {
	if c.CustomFieldAPI == nil {
		c.CustomFieldAPI = NewCustomFieldService(c)
	}
	if fields, err := c.CustomFieldAPI.GetCustomFields(); err != nil {
		return err
	} else if len(fields) > 0 {
		c.CustomFieldSet = NewCustomFieldSet()
		if err := c.CustomFieldSet.Add(fields...); err != nil {
			return err
		}
	}
	return nil
}

// LogOrNotAny logs a message if the client's Logger is set. This is a no-op if Logger is nil.
func (c *Client) LogOrNotAny(ctx context.Context, level slog.Level, msg string, attrs ...any) {
	if c.Logger != nil {
		slogutil.LogOrNotAny(ctx, c.Logger, level, msg, attrs...)
	}
}
