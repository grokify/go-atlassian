package jira

import (
	"strings"
	"time"
)

// TimeRemaining returns calculated timeRemainingOriginal and timeRemaiing and against the timeOriginalEstimate and timeEstimate respectively.
func TimeRemaining(status string, timeOriginalEstimate, timeEstimate, timeSpent int) (timeRemainingOriginal, timeRemaining int) {
	if timeOriginalEstimate < 0 {
		timeOriginalEstimate = 0
	}
	if timeEstimate < 0 {
		timeEstimate = 0
	}
	if timeSpent < 0 {
		timeEstimate = 0
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "closed" || status == "done" {
		return 0, 0
	}
	timeRemaining = timeEstimate - timeSpent
	if timeRemaining < 0 {
		timeRemaining = 0
	}
	timeRemainingOriginal = timeOriginalEstimate - timeSpent
	if timeRemainingOriginal < 0 {
		timeRemainingOriginal = 0
	}
	return
}

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
