package jira

import "sort"

func (set *IssuesSet) KeysByField(fieldLabel string) map[string]SliceMore {
	out := map[string]SliceMore{}
	for _, iss := range set.Items {
		im := NewIssueMore(&iss)
		if v, ok := im.Value(fieldLabel); ok {
			if _, ok := out[v]; !ok {
				out[v] = SliceMore{Values: []string{}}
			}
			if sm, ok := out[v]; !ok {
				panic("vals out found")
			} else {
				sm.Values = append(sm.Values, im.Key())
				out[v] = sm
			}
		}
	}
	for k, sm := range out {
		sm.Inflate()
		out[k] = sm
	}
	return out
}

type SliceMore struct {
	Count  int      `json:"count"`
	Values []string `json:"values"`
}

func (sm *SliceMore) Inflate() {
	sort.Strings(sm.Values)
	sm.Count = len(sm.Values)
}
