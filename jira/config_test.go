package jira

import (
	"testing"
	"time"
)

var configTests = []struct {
	hoursPerDay float32
	daysPerWeek float32
	days        float32
	people      float32
	capacity    time.Duration
}{
	{8.0, 5.0, 1.0, 1.0, time.Duration(int64(8.0*60*60)) * time.Second},
}

func TestConfig(t *testing.T) {
	for _, tt := range configTests {
		cfg := Config{
			WorkingHoursPerDay: tt.hoursPerDay,
			WorkingDaysPerWeek: tt.daysPerWeek}
		try := cfg.CapacityForDaysPeople(tt.days, tt.people)
		if try != tt.capacity {
			t.Errorf("atlassian.Config.CapacityForDaysPeople(%v,%v) mismatch: want (%v), got (%v)", tt.days, tt.people, tt.capacity, try)
		}
	}
}
