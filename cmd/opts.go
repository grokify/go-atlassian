package cmd

import (
	"github.com/grokify/go-atlassian/jira"
	"github.com/grokify/goauth"
	flags "github.com/jessevdk/go-flags"
)

type Options struct {
	Options goauth.Options
	// Authfile        string `short:"a" long:"Goauth authfile" description:"Goauth auth File"`
	// Authkey         string `short:"k" long:"Goauth key" description:"Goauth credentials Key"`
	BoardID           uint   `short:"b" long:"boardid" description:"Jira Board ID"`
	IssueKey          string `short:"i" long:"key" description:"Jira Issue Key"`
	JQL               string `short:"j" long:"jql" description:"Jira Query Language"`
	AddCustomFieldSet []bool `long:"loadcustomfieldset" description:"Add custom field set"`
	Customfield       string `short:"c" long:"customfield" description:"Custom field"`
	CustomfieldName   string `short:"n" long:"customfield name" description:"Custom field name"` // 'Epic Link'
}

func (opts Options) Client() (*jira.Client, error) {
	return jira.NewClientGoauthBasicAuthFile(
		opts.Options.CredsPath,
		opts.Options.Account,
		len(opts.AddCustomFieldSet) > 0)
}

func NewOptions() (Options, error) {
	opts := Options{}
	_, err := flags.Parse(&opts)
	return opts, err
}
