package jira

import (
	"slices"

	"github.com/grokify/mogo/errors/errorsutil"
)

// RetrieveChildrenOfType retrieves all children of supplied parent types. If the child matches a base type,
// it is inserted into the current `IssuesSet`. If it is not a baseType, it is inserted into `Parents`. Of
// note, this will only load children of parent types that are already in the `IssuesSet`.
func (set *IssuesSet) RetrieveChildrenOfType(client *Client, parentTypes, baseTypes []string) error {
	if len(parentTypes) == 0 {
		return nil
	} else if client == nil {
		return errorsutil.Wrap(ErrClientCannotBeNil, "called in IssuesSet.RetrieveChildrenOfType")
	}
	parentKeys := set.KeysForTypes(parentTypes, true, true)

	if len(parentKeys) == 0 {
		return nil
	}

	children, err := client.IssueAPI.SearchChildrenIssues(parentKeys)
	if err != nil {
		return err
	} else if set.Parents == nil {
		set.Parents = NewIssuesSet(set.Config)
	}

	for len(children) > 0 {
		var unknownChildrenKeys []string
		for _, c := range children {
			c := c
			im := NewIssueMore(&c)
			childKey := im.Key()
			if childKey == "" || set.KeyExists(childKey, true) {
				continue
			} else if slices.Index(baseTypes, im.Type()) > -1 {
				set.Items[childKey] = c
				continue
			} else { // not base type.
				set.Parents.Items[childKey] = c
				unknownChildrenKeys = append(unknownChildrenKeys, childKey)
			}
		}

		if len(unknownChildrenKeys) == 0 {
			break
		} else {
			children, err = client.IssueAPI.SearchChildrenIssues(unknownChildrenKeys)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
