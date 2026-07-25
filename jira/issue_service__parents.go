package jira

import "errors"

func (svc *IssueService) IssuesSetAddParents(set *IssuesSet) error {
	if set == nil {
		return errors.New("issues set is nil")
	} else if parents, err := svc.SearchIssuesSetParents(set); err != nil {
		return err
	} else {
		set.Parents = parents
		return nil
	}
}
