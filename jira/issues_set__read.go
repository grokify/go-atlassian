package jira

import (
	"encoding/json"
	"os"
)

// IssuesSetReadDirIssuesFiles reads a list of JSON issues files.
func IssuesSetReadDirIssuesFiles(filepaths []string) (*IssuesSet, error) {
	is := NewIssuesSet(nil)
	for _, fp := range filepaths {
		if ii, err := IssuesReadFileJSON(fp); err != nil {
			return nil, err
		} else if len(ii) == 0 {
			continue
		} else if err := is.Add(ii...); err != nil {
			return nil, err
		}
	}
	return is, nil
}

func IssuesSetReadFileJSON(filename string) (*IssuesSet, error) {
	if b, err := os.ReadFile(filename); err != nil {
		return nil, err
	} else {
		is := &IssuesSet{}
		return is, json.Unmarshal(b, is)
	}
}
