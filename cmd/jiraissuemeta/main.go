package main

import (
	"context"
	"fmt"

	"github.com/grokify/go-atlassian/cmd"
	"github.com/grokify/go-atlassian/jira"
	"github.com/grokify/mogo/errors/errorsutil"
	"github.com/grokify/mogo/fmt/fmtutil"
	"github.com/grokify/mogo/log/logutil"
	flags "github.com/jessevdk/go-flags"
)

func main() {
	opts := cmd.Options{}
	_, err := flags.Parse(&opts)
	logutil.FatalErr(err)

	fmtutil.MustPrintJSON(opts)

	jrClient, err := opts.Client()
	logutil.FatalErr(errorsutil.Wrap(err, "Client"))

	iss, err := jrClient.IssueAPI.Issue(context.Background(), opts.IssueKey, nil)
	logutil.FatalErr(err)

	im := jira.NewIssueMore(iss)

	fmtutil.MustPrintJSON(im.Meta("", []string{}))
	fmtutil.MustPrintJSON(im.Meta(jrClient.Config.ServerURL, []string{}))

	fmt.Println("DONE")
}
