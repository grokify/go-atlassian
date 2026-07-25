package jira

import (
	"sort"

	"github.com/grokify/mogo/type/slicesutil"
)

type IssueMores []IssueMore

func (ii IssueMores) Keys(sortAsc, dedupe bool) []string {
	var out []string
	for _, iss := range ii {
		out = append(out, iss.Key())
	}
	if sortAsc {
		sort.Strings(out)
	}
	if dedupe {
		slicesutil.Dedupe(out)
	}
	return out
}

func (ii IssueMores) ProjectKeys(sortAsc, dedupe bool) []string {
	var out []string
	for _, iss := range ii {
		out = append(out, iss.Key())
	}
	if sortAsc {
		sort.Strings(out)
	}
	return out
}
