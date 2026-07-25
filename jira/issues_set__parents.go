package jira

import (
	"errors"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/mogo/errors/errorsutil"
	"github.com/grokify/mogo/pointer"
	"github.com/grokify/mogo/type/slicesutil"
	"github.com/grokify/mogo/type/stringsutil"
)

/*
func (set *IssuesSet) RetrieveParentsIssuesSet(client *Client) (*IssuesSet, error) {
	parIssuesSet := NewIssuesSet(set.Config)
	parIDs := set.UnknownParents()
	if len(parIDs) == 0 {
		return parIssuesSet, nil
	}

	err := parIssuesSet.RetrieveIssues(client, parIDs)
	if err != nil {
		return nil, err
	}

	err = parIssuesSet.RetrieveParents(client)

	return parIssuesSet, err
}
*/

func (set *IssuesSet) RetrieveParents(client *Client) error {
	if client == nil {
		return errorsutil.Wrap(ErrClientCannotBeNil, "called in IssuesSet.RetrieveParents")
	}
	parIDs := set.KeysParentsUnpopulated()
	i := 0
	for len(parIDs) > 0 {
		err := set.RetrieveIssues(client, parIDs)
		if err != nil {
			return err
		}
		parIDs = set.KeysParentsUnpopulated()
		i++
		if i > 10 {
			return errors.New("exceeded max retrieve parent iterations")
		}
	}
	return nil
}

func (set *IssuesSet) RetrieveIssues(client *Client, ids []string) error {
	if client == nil {
		return errorsutil.Wrap(ErrClientCannotBeNil, "called in IssuesSet.RetrieveIssues")
	}
	ids = stringsutil.SliceCondenseSpace(ids, true, true)
	if len(ids) == 0 {
		return nil
	}

	idsSlicesMaxResults := slicesutil.SplitMaxLength(ids, JQLMaxResults)

	for _, idsSlice := range idsSlicesMaxResults {
		jqls := JQLStringsSimple(FieldKey, false, idsSlice, 0)

		for _, jql := range jqls {
			if iss, err := client.IssueAPI.SearchIssuesPages(jql, 0, 0, 0); err != nil {
				return err
			} else {
				return set.Add(iss...)
			}
		}
	}
	return nil
}

func (set *IssuesSet) IssueOrParent(key string) (*gojira.Issue, bool) {
	if iss, ok := set.Items[key]; ok {
		return &iss, true
	} else if set.Parents == nil {
		return nil, false
	} else if iss, ok := set.Parents.Items[key]; ok {
		return &iss, true
	} else {
		return nil, false
	}
}

func (set *IssuesSet) KeysParents() []string {
	var parKeys []string
	for _, iss := range set.Items {
		im := NewIssueMore(pointer.Pointer(iss))
		if parKey := im.ParentKey(); parKey != "" {
			parKeys = append(parKeys, parKey)
		}
	}
	return stringsutil.SliceCondenseSpace(parKeys, true, true)
}

// ParentsPopulated returns issue ids that are in the current set or current parent set.
func (set *IssuesSet) KeysParentsPopulated() []string {
	var parKeysPop []string
	parKeysAll := set.KeysParents()
	for _, parKey := range parKeysAll {
		parIss, ok := set.IssueOrParent(parKey)
		if ok && parIss != nil {
			parKeysPop = append(parKeysPop, parKey)
		}
	}

	return stringsutil.SliceCondenseSpace(parKeysPop, true, true)
}

// ParentsUnpopulated returns issue ids that are not in the current set or current parent set.
func (set *IssuesSet) KeysParentsUnpopulated() []string {
	var parKeysUnpop []string
	parKeysAll := set.KeysParents()
	for _, parKey := range parKeysAll {
		parIss, ok := set.IssueOrParent(parKey)
		if !ok || parIss == nil {
			parKeysUnpop = append(parKeysUnpop, parKey)
		}
	}

	return stringsutil.SliceCondenseSpace(parKeysUnpop, true, true)
}

/*
func (set *IssuesSet) GetLineage(key string) (Issues, error) {
	key = strings.TrimSpace(key)
	lineage := Issues{}
	if key == "" {
		return lineage, nil
	}
	iss, err := set.Get(key)
	if err != nil {
		return lineage, err
	}

	im := IssueMore{Issue: &iss}
	parKey := im.ParentKey()
	for {
		if parKey == "" {
			return lineage, nil
		}
		if parIss, err := set.Get(parKey); err != nil {
			return lineage, err
		} else {
			lineage = append(lineage, parIss)
			parIM := IssueMore{Issue: &parIss}
			parKey = parIM.ParentKey()
		}
	}
	return lineage, nil
}
*/
