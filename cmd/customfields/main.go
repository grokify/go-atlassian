package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/grokify/go-atlassian/cmd"
	"github.com/grokify/go-atlassian/jira"
	flags "github.com/jessevdk/go-flags"
)

func main() {
	opts := cmd.Options{}
	_, err := flags.Parse(&opts)
	if err != nil {
		slog.Error("failed to parse flags", "error", err)
		os.Exit(1)
	}

	jrClient, err := opts.Client()
	if err != nil {
		slog.Error("failed to create client", "error", err)
		os.Exit(2)
	}

	cfs, err := jrClient.CustomFieldAPI.GetCustomFields()
	if err != nil {
		slog.Error("failed to get custom fields", "error", err)
		os.Exit(3)
	}

	err = cfs.WriteTable(os.Stdout)
	if err != nil {
		slog.Error("failed to write table", "error", err)
		os.Exit(4)
	}

	if opts.Customfield != "" {
		ids := strings.Split(opts.Customfield, ",")
		filtered := cfs.FilterByIDs(ids...)
		if err := filtered.WriteTable(os.Stdout); err != nil {
			slog.Error("failed to write filtered table", "error", err)
			os.Exit(5)
		}
	}

	if opts.CustomfieldName != "" {
		names := strings.Split(opts.CustomfieldName, ",")
		filtered := cfs.FilterByNames(names...)
		if err := filtered.WriteTable(os.Stdout); err != nil {
			slog.Error("failed to write filtered table", "error", err)
			os.Exit(6)
		}
	}

	// Get Epic Link Custom Field
	cfName, err := jrClient.CustomFieldAPI.GetCustomFieldEpicLink()
	if err != nil {
		slog.Error("failed to get epic link custom field", "error", err)
		os.Exit(7)
	}

	cfsName := jira.CustomFields{cfName}
	if err := cfsName.WriteTable(os.Stdout); err != nil {
		slog.Error("failed to write table", "error", err)
		os.Exit(8)
	}

	slog.Info("DONE")
	os.Exit(0)
}
