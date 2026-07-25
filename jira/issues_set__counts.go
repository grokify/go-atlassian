package jira

import (
	"errors"
	"fmt"

	"github.com/grokify/gocharts/v2/data/histogram"
	"github.com/grokify/mogo/pointer"
	"github.com/grokify/mogo/time/timeutil"
	"github.com/grokify/mogo/type/stringsutil"
)

// Counts returns a `map[string]map[string]uint{}` where the first key
// is the category and the second is the category value.
func (set *IssuesSet) Counts() map[string]map[string]uint {
	mm := map[string]map[string]uint{
		"byProject":    set.CountsByProject(),
		"byProjectKey": set.CountsByProjectKey(),
		"byStatus":     set.CountsByStatus(),
		"byType":       set.CountsByType(true, false),
		"byTime":       set.CountsByTime(),
	}
	return mm
}

/*
func (set *IssuesSet) TimeSeriesCreated() (timeseries.TimeSeries, error) {
	ts := timeseries.NewTimeSeries("")
	for _, iss := range set.Items {
		iss := iss
		im := IssueMore{Issue: &iss}
		ts.AddInt64(im.CreateTime().UTC(), 1)
	}
	return ts, nil
}
*/

// CountsByCustomFieldValues returns a list of custom field value counts where `customField` is in
// the format `customfield_12345`.
func (set *IssuesSet) CountsByCustomFieldValues(customField string) (map[string]uint, error) {
	out := map[string]uint{}
	for _, iss := range set.Items {
		iss := iss
		im := NewIssueMore(&iss)
		cfInfo, err := im.CustomField(customField)
		if err != nil {
			return out, err
		}
		out[cfInfo.Value]++
	}
	return out, nil
}

// CountsByCustomFieldName resolves a custom field by display name using the CustomFieldSet,
// handling the common case where multiple field IDs share the same name. For each issue,
// it checks all matching field IDs and uses the first populated value it finds.
// If onlyPopulated is true, issues with no populated matching field are skipped.
func (set *IssuesSet) CountsByCustomFieldName(fieldName string, cfSet *CustomFieldSet, onlyPopulated bool) (map[string]uint, error) {
	if cfSet == nil {
		return nil, errors.New("CustomFieldSet is required for name-based lookup")
	}
	ids := cfSet.NameToIDs(fieldName)
	if len(ids) == 0 {
		return nil, fmt.Errorf("no custom field found with name %q", fieldName)
	}

	out := map[string]uint{}
	for _, iss := range set.Items {
		iss := iss
		im := NewIssueMore(&iss)
		// Try each matching field ID, use the first populated value
		value := ""
		for _, id := range ids {
			if v := im.CustomFieldStringOrDefault(id, ""); v != "" {
				value = v
				break
			}
		}
		if value == "" && onlyPopulated {
			continue
		}
		if value == "" {
			value = "(empty)"
		}
		out[value]++
	}
	return out, nil
}

func (set *IssuesSet) CountsByProjectAndCustomFieldValues(customField string) (*histogram.HistogramSet, error) {
	hset := histogram.NewHistogramSet("")
	for _, iss := range set.Items {
		iss := iss
		im := NewIssueMore(&iss)
		project := im.Project()
		if cfInfo, err := im.CustomField(customField); err != nil {
			return nil, err
		} else {
			hset.Add(project, cfInfo.Value, 1)
		}
	}
	return hset, nil
}

// CountsByProject returns `map[string]uint` representing issue counts by project.
func (set *IssuesSet) CountsByProject() map[string]uint {
	m := map[string]uint{}
	for _, iss := range set.Items {
		im := NewIssueMore(pointer.Pointer(iss))
		m[im.Project()]++
	}
	return m
}

// CountsByProjectKey returns `map[string]uint` representing issue counts by project key.
func (set *IssuesSet) CountsByProjectKey() map[string]uint {
	m := map[string]uint{}
	for _, iss := range set.Items {
		im := NewIssueMore(pointer.Pointer(iss))
		m[im.ProjectKey()]++
	}
	return m
}

// CountsByStatus returns `map[string]uint` representing issue counts by status.
func (set *IssuesSet) CountsByStatus() map[string]uint {
	m := map[string]uint{}
	for _, iss := range set.Items {
		im := NewIssueMore(pointer.Pointer(iss))
		m[im.Status()]++
	}
	return m
}

func (set *IssuesSet) CountsByMetaStage(inclTypeFilter []string) map[string]uint {
	inclTypeFilter = stringsutil.SliceCondenseSpace(inclTypeFilter, true, true)
	inclTypeFilterMap := map[string]int{}
	for _, filter := range inclTypeFilter {
		inclTypeFilterMap[filter]++
	}
	out := map[string]uint{}
	count := uint(0)
	unknownStatus := map[string]uint{}
	for _, iss := range set.Items {
		im := NewIssueMore(pointer.Pointer(iss))
		if len(inclTypeFilterMap) > 0 {
			if _, ok := inclTypeFilterMap[im.Type()]; !ok {
				continue
			}
		}
		metaStage := ""
		if set.Config != nil && set.Config.StatusConfig != nil {
			metaStage = set.Config.StatusConfig.MetaStage(im.Status())
		}
		if metaStage == "" {
			unknownStatus[im.Status()]++
		}
		out[metaStage]++
		count++
	}
	if msuCount(out) != count {
		panic("count mismatch")
	}
	return out
}

func (set *IssuesSet) CountsByProjectAndMetaStage(inclTypeFilter []string) *histogram.HistogramSet {
	out := histogram.NewHistogramSet("")
	inclTypeFilter = stringsutil.SliceCondenseSpace(inclTypeFilter, true, true)
	inclTypeFilterMap := map[string]int{}
	for _, filter := range inclTypeFilter {
		inclTypeFilterMap[filter]++
	}
	count := int(0)
	for _, iss := range set.Items {
		im := NewIssueMore(pointer.Pointer(iss))
		if len(inclTypeFilterMap) > 0 {
			if _, ok := inclTypeFilterMap[im.Type()]; !ok {
				continue
			}
		}
		projectName := im.Project()
		metaStage := ""
		if set.Config != nil && set.Config.StatusConfig != nil {
			metaStage = set.Config.StatusConfig.MetaStage(im.Status())
		}
		out.Add(projectName, metaStage, 1)
		count++
	}
	if count != out.Sum() {
		panic("count mismatch")
	}
	return out
}

func msuCount(m map[string]uint) uint {
	c := uint(0)
	for _, v := range m {
		c += v
	}
	return c
}

func (set *IssuesSet) CountWithTypeFilter(inclTypeFilter []string) uint {
	inclTypeFilter = stringsutil.SliceCondenseSpace(inclTypeFilter, true, true)
	inclTypeFilterMap := map[string]int{}
	for _, filter := range inclTypeFilter {
		inclTypeFilterMap[filter]++
	}
	count := uint(0)
	for _, iss := range set.Items {
		im := NewIssueMore(pointer.Pointer(iss))
		if len(inclTypeFilterMap) > 0 {
			if _, ok := inclTypeFilterMap[im.Type()]; !ok {
				continue
			}
		}
		count++
	}
	return count
}

func (set *IssuesSet) CountsByType(inclLeafs, inclParents bool) map[string]uint {
	m := map[string]uint{}
	if inclLeafs {
		for _, iss := range set.Items {
			iss := iss
			im := NewIssueMore(&iss)
			m[im.Type()]++
		}
	}
	if inclParents && set.Parents != nil {
		for _, iss := range set.Parents.Items {
			iss := iss
			im := NewIssueMore(&iss)
			m[im.Type()]++
		}
	}
	return m
}

func (set *IssuesSet) CountsByTime() map[string]uint {
	out := map[string]uint{}
	for _, iss := range set.Items {
		if iss.Fields == nil {
			continue
		}
		if iss.Fields.TimeEstimate <= 0 {
			out["TimeEstimateLTEZ"]++
		} else {
			out["TimeEstimateGTZ"]++
		}
		if iss.Fields.TimeOriginalEstimate <= 0 {
			out["TimeOriginalEstimateLTEZ"]++
		} else {
			out["TimeOriginalEstimateGTZ"]++
		}
		/*
			TimeTimeSpent                     = "Time Spent"
			TimeTimeEstimate                  = "Time Estimate"
			TimeTimeOriginalEstimate          = "Time Original Estimate"
			TimeAggregateTimeOriginalEstimate = "Aggregate Time Original Estimate"
			TimeAggregateTimeSpent            = "Aggregate Time Spent"
			TimeAggregateTimeEstimate         = "Aggregate Time Estimate"
			TimeTimeRemaining                 = "Time Remaining"
			TimeTimeRemainingOriginal         = "Time Remaining Original"
		*/
	}
	return out
}

func (set *IssuesSet) CountsByWorkstream(wsFuncMake WorkstreamFuncMake, inclTypeFilter []string) (map[string]uint, error) {
	inclTypeFilter = stringsutil.SliceCondenseSpace(inclTypeFilter, true, true)
	inclTypeFilterMap := map[string]int{}
	for _, filter := range inclTypeFilter {
		inclTypeFilterMap[filter]++
	}
	out := map[string]uint{}
	for _, iss := range set.Items {
		im := NewIssueMore(pointer.Pointer(iss))
		if len(inclTypeFilterMap) > 0 {
			if _, ok := inclTypeFilterMap[im.Type()]; !ok {
				continue
			}
		}
		if ws, err := wsFuncMake(im.Key()); err != nil {
			return nil, err
		} else {
			out[ws]++
		}
	}
	return out, nil
}

func (set *IssuesSet) TimeStats() TimeStats {
	if set.Config == nil {
		set.Config = NewConfigDefault()
	}
	ts := TimeStats{
		TimeUnit:           timeutil.SecondString,
		ItemCount:          len(set.Items),
		WorkingDaysPerWeek: set.Config.WorkingDaysPerWeek,
		WorkingHoursPerDay: set.Config.WorkingHoursPerDay,
	}
	for _, iss := range set.Items {
		if iss.Fields == nil {
			continue
		}
		timeRemainingOriginal, timeRemaining := TimeRemaining(iss.Fields.Status.Name, iss.Fields.TimeOriginalEstimate, iss.Fields.TimeEstimate, iss.Fields.TimeSpent)
		ts.TimeSpent += float32(iss.Fields.TimeSpent)
		ts.TimeEstimate += float32(iss.Fields.TimeEstimate)
		ts.TimeOriginalEstimate += float32(iss.Fields.TimeOriginalEstimate)
		ts.AggregateTimeOriginalEstimate += float32(iss.Fields.AggregateTimeOriginalEstimate)
		ts.AggregateTimeSpent += float32(iss.Fields.AggregateTimeSpent)
		ts.AggregateTimeEstimate += float32(iss.Fields.AggregateTimeEstimate)
		ts.TimeRemaining += float32(timeRemaining)
		ts.TimeRemainingOriginal += float32(timeRemainingOriginal)
	}
	return ts
}
