package xml

import (
	"strings"

	"github.com/grokify/go-atlassian/jira"
	"github.com/grokify/gocharts/v2/data/histogram"
	"github.com/grokify/gocharts/v2/data/table"
	"github.com/grokify/mogo/type/stringsutil"
)

type Issues []Issue

func (ii Issues) FilterByStatus(statuses ...string) Issues {
	new := Issues{}
	mStatuses := map[string]int{}
	for _, s := range statuses {
		mStatuses[s]++
	}
	for _, ix := range ii {
		if _, ok := mStatuses[ix.Status.DisplayName]; ok {
			new = append(new, ix)
		}
	}
	return new
}

/*
func (ii Issues) ValidateKeys() error {
	for i, ix := range ii {

	}
}
*/

func (ii Issues) Keys() []string {
	keys := []string{}
	for _, ix := range ii {
		keys = append(keys, ix.Key.DisplayName)
	}
	return stringsutil.SliceCondenseSpace(keys, true, true)
}

func (ii Issues) Stats(workingHoursPerDay, workingDaysPerWeek float32) jira.IssuesStats {
	if workingHoursPerDay == 0 {
		workingHoursPerDay = jira.WorkingHoursPerDayDefault
	}
	if workingDaysPerWeek == 0 {
		workingDaysPerWeek = jira.WorkingDaysPerWeekDefault
	}
	workingHoursPerDay64 := float64(workingHoursPerDay)
	stats := jira.IssuesStats{
		WorkingHoursPerDay:     workingHoursPerDay,
		WorkingDaysPerWeek:     workingDaysPerWeek,
		ItemCount:              len(ii),
		ItemCountByStatus:      map[string]int{},
		ItemCountByType:        map[string]int{},
		EstimateStatsByType:    map[string]jira.EstimateStats{},
		ClosedEstimateVsActual: jira.EstimateVsActual{},
	}
	for _, it := range ii {
		stats.TimeOriginalEstimate += it.TimeOriginalEstimate.Duration()
		stats.AggregateTimeSpent += it.AggregateTimeSpent.Duration()
		stats.ItemCountByStatus[it.Status.DisplayName]++
		stats.ItemCountByType[it.Type.DisplayName]++
		esStats, ok := stats.EstimateStatsByType[it.Type.DisplayName]
		if !ok {
			esStats = jira.EstimateStats{}
		}
		if it.TimeOriginalEstimate.Seconds > 0 {
			esStats.WithEstimate++
		} else {
			esStats.WithoutEstimate++
		}
		stats.EstimateStatsByType[it.Type.DisplayName] = esStats
		if it.Status.DisplayName == jira.StatusDone {
			stats.ClosedEstimateVsActual.ClosedCount++
			if it.TimeOriginalEstimate.Seconds > 0 {
				stats.ClosedEstimateVsActual.ClosedCountWithEstimate++
				stats.ClosedEstimateVsActual.EstimateDays += it.TimeOriginalEstimate.Duration().Hours() / workingHoursPerDay64
				stats.ClosedEstimateVsActual.ActualDays += it.AggregateTimeSpent.Duration().Hours() / workingHoursPerDay64
			}
		}
	}
	stats.TimeOriginalEstimateDays = stats.TimeOriginalEstimate.Hours() / workingHoursPerDay64
	stats.AggregateTimeSpentDays = stats.AggregateTimeSpent.Hours() / workingHoursPerDay64
	stats.ClosedEstimateVsActual.Inflate()
	return stats
}

/*
type IssuesStats struct {
	WorkingHoursPerDay       float32
	WorkingDaysPerWeek       float32
	ItemCount                int
	ItemCountByStatus        map[string]int
	ItemCountByType          map[string]int
	EstimateStatsByType      map[string]EstimateStats
	TimeOriginalEstimate     time.Duration
	TimeOriginalEstimateDays float64
	AggregateTimeSpent       time.Duration
	AggregateTimeSpentDays   float64
	ClosedEstimateVsActual   EstimateVsActual
}

type EstimateStats struct {
	WithEstimate    int
	WithoutEstimate int
}

type EstimateVsActual struct {
	ClosedCount             int
	ClosedCountWithEstimate int
	EstimateDays            float64
	ActualDays              float64
	EstimateRatio           float64
}

func (eva *EstimateVsActual) Inflate() {
	if eva.ActualDays > 0 {
		eva.EstimateRatio = eva.ActualDays / eva.EstimateDays
	}
}
*/

// TSRHistogramSets returns a `*histogram.HistogramSets` for Type, Status and Resolution.
func (ii Issues) TSRHistogramSets(name string) *histogram.HistogramSets {
	if strings.TrimSpace(name) == "" {
		name = "TSR"
	}
	hset := histogram.NewHistogramSets(name)
	for _, iss := range ii {
		hset.Add(
			iss.Type.DisplayName,
			iss.Status.DisplayName,
			iss.Resolution.DisplayName,
			1, true)
	}
	return hset
}

// TSRTable returns a `table.Table` for Type, Status and Resolution.
func (ii Issues) TSRTable(name string) table.Table {
	hset := ii.TSRHistogramSets(name)
	return hset.Table("Jira Issues", "Type", "Status", "Resolution", "Count")
}

// TSRWriteCSV writes a CSV file for Type, Status and Resolution.
func (ii Issues) TSRWriteCSV(filename string) error {
	tbl := ii.TSRTable("")
	return tbl.WriteCSV(filename)
}
