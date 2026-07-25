package jira

import (
	"fmt"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/mogo/time/timeutil"
)

type TimeStatsSets struct {
	Map map[string]TimeStatsSet
}

func (tss *TimeStatsSets) AddIssue(iss gojira.Issue) {
	if iss.Fields == nil {
		return
	}
}

type TimeStatsSet struct {
	Map map[string]TimeStats
}

type TimeStats struct {
	TimeUnit                      string
	WorkingHoursPerDay            float32
	WorkingDaysPerWeek            float32
	ItemCount                     int
	TimeSpent                     float32
	TimeEstimate                  float32
	TimeOriginalEstimate          float32
	AggregateTimeOriginalEstimate float32
	AggregateTimeSpent            float32
	AggregateTimeEstimate         float32
	TimeRemaining                 float32
	TimeRemainingOriginal         float32
}

func (ts TimeStats) SecondsToDays() (TimeStats, error) {
	if ts.TimeUnit != timeutil.SecondString {
		return ts, fmt.Errorf("time unit is not seconds, is (%s)", ts.TimeUnit)
	}
	if ts.WorkingHoursPerDay <= 0 {
		return ts, fmt.Errorf("workingHoursPerDay must be greater than 0, is (%v)", ts.WorkingHoursPerDay)
	}
	whpd := ts.WorkingHoursPerDay
	tsDays := TimeStats{
		TimeUnit:                      timeutil.DayString,
		WorkingHoursPerDay:            ts.WorkingHoursPerDay,
		WorkingDaysPerWeek:            ts.WorkingDaysPerWeek,
		ItemCount:                     ts.ItemCount,
		TimeEstimate:                  ts.TimeEstimate / 60 / 60 / whpd,
		TimeOriginalEstimate:          ts.TimeOriginalEstimate / 60 / 60 / whpd,
		AggregateTimeOriginalEstimate: ts.AggregateTimeOriginalEstimate / 60 / 60 / whpd,
		AggregateTimeSpent:            ts.AggregateTimeSpent / 60 / 60 / whpd,
		AggregateTimeEstimate:         ts.AggregateTimeEstimate / 60 / 60 / whpd,
		TimeRemaining:                 ts.TimeRemaining / 60 / 60 / whpd,
		TimeRemainingOriginal:         ts.TimeRemainingOriginal / 60 / 60 / whpd,
	}
	return tsDays, nil
}
