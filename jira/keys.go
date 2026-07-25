package jira

import (
	"regexp"
	"strings"

	"github.com/grokify/mogo/type/stringsutil"
)

var rxJiraTicket = regexp.MustCompile(`([A-Z]+\-[0-9]+)`)

func KeyIsValid(id string) bool {
	return rxJiraTicket.MatchString(id)
}

func ParseKeys(s string, unique, asc bool) []string {
	var keys []string
	m := rxJiraTicket.FindAllStringSubmatch(s, -1)
	for _, n := range m {
		if len(n) > 1 {
			keys = append(keys, n[1])
		}
	}
	return stringsutil.SliceCondenseSpace(keys, unique, asc)
}

func KeysContainProject(keys []string, projectKey string) bool {
	pk := strings.ToUpper(strings.TrimSpace(projectKey))
	for _, k := range keys {
		k := strings.ToUpper(strings.TrimSpace(k))
		if strings.Index(k, pk+"-") == 0 {
			return true
		}
	}
	return false
}
