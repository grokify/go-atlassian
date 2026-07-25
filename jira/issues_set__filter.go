package jira

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/grokify/mogo/pointer"
	"github.com/grokify/mogo/type/slicesutil"
)

func (set *IssuesSet) FilterByKeys(keys []string, errOnUnfound bool) (*IssuesSet, error) {
	filteredIssuesSet := NewIssuesSet(set.Config)
	var keysNotFound []string
	for _, key := range keys {
		if iss, ok := set.Items[key]; ok {
			filteredIssuesSet.Items[key] = iss
		} else if errOnUnfound {
			keysNotFound = append(keysNotFound, key)
		}
	}
	if len(keysNotFound) > 0 {
		keysNotFound = slicesutil.Dedupe(keysNotFound)
		sort.Strings(keysNotFound)
		return nil, fmt.Errorf("key not found (%s)", strings.Join(keysNotFound, ","))
	}
	return filteredIssuesSet, nil
}

func (set *IssuesSet) FilterByStatus(inclStatuses, exclStatuses []string) (*IssuesSet, error) {
	filteredIssuesSet := NewIssuesSet(set.Config)
	inclStatusesMap := map[string]int{}
	for _, s := range inclStatuses {
		inclStatusesMap[s]++
	}
	exclStatusesMap := map[string]int{}
	for _, s := range exclStatuses {
		exclStatusesMap[s]++
	}
	for _, iss := range set.Items {
		im := NewIssueMore(pointer.Pointer(iss))
		statusName := im.Status()
		_, inclStatusOk := inclStatusesMap[statusName]
		_, exclStatusOk := exclStatusesMap[statusName]
		if len(inclStatusesMap) > 0 && !inclStatusOk {
			continue
		} else if len(exclStatuses) > 0 && exclStatusOk {
			continue
		}
		err := filteredIssuesSet.Add(iss)
		if err != nil {
			return nil, err
		}
	}
	return filteredIssuesSet, nil
}

func (set *IssuesSet) FilterByStatusCategory(scatsRef StatusCategories, scatsIncl []string) (*IssuesSet, error) {
	var inclStatuses []string
	for _, scatIncl := range scatsIncl {
		if scatStatuses, ok := scatsRef.MapCategoryToStatuses[scatIncl]; ok {
			inclStatuses = append(inclStatuses, scatStatuses...)
		}
	}
	inclStatuses = slicesutil.Dedupe(inclStatuses)
	sort.Strings(inclStatuses)
	return set.FilterByStatus(inclStatuses, []string{})
}

/*
func (set *IssuesSet) FilterStatus(inclStatuses ...string) (*IssuesSet, error) {
	n := NewIssuesSet(set.Config)
	if len(inclStatuses) == 0 {
		return n, nil
	}
	for _, iss := range set.Items {
		im := NewIssueMore(pointer.Pointer(iss))
		if slices.Index(inclStatuses, im.Status()) >= 0 {
			err := n.Add(iss)
			if err != nil {
				return nil, err
			}
		}
	}
	return n, nil
}
*/

func (set *IssuesSet) FilterByType(inclTypes ...string) (*IssuesSet, error) {
	out := NewIssuesSet(set.Config)
	if len(inclTypes) == 0 {
		return out, nil
	}
	for _, iss := range set.Items {
		im := NewIssueMore(pointer.Pointer(iss))
		if slices.Index(inclTypes, im.Type()) >= 0 {
			err := out.Add(iss)
			if err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}
