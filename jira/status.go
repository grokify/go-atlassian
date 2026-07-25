package jira

import (
	"errors"
	"net/url"
	"sort"
	"strings"

	"github.com/grokify/mogo/type/slicesutil"
	"github.com/grokify/mogo/type/stringsutil"
)

type StatusCategoryConfig struct {
	Map         map[string]string
	StageConfig StageConfig
	// MetaStageOrder []string
}

func NewStatusConfig(stageConfig StageConfig) StatusCategoryConfig {
	return StatusCategoryConfig{
		Map:         map[string]string{}, // status to metastatus
		StageConfig: stageConfig,
	}
}

// AddMapSlice should be a map where the keys are meta statuses and the values are slices of Jira statuses.
func (ss *StatusCategoryConfig) AddMapSlice(m map[string][]string) error {
	for metaStatus, vals := range m {
		for _, status := range vals {
			if err := ss.Add(status, metaStatus); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ss *StatusCategoryConfig) Add(status, metaStage string) error {
	if ss.Map == nil {
		ss.Map = map[string]string{}
	}
	if !ss.StageConfig.Exists(metaStage) {
		return errors.New("metaStage is not configured")
	}
	ss.Map[status] = metaStage
	return nil
}

func (ss *StatusCategoryConfig) MapMetaStageToStatuses() map[string][]string {
	out := url.Values{}
	for status, metaStage := range ss.Map {
		out.Add(metaStage, status)
	}
	return out
}

// MetaStage returns the metastatus for a status. If there is no metastatus, an empty string is returned.
func (ss *StatusCategoryConfig) MetaStage(status string) string {
	if cat, ok := ss.Map[status]; ok {
		return cat
	} else {
		return ""
	}
}

func (ss *StatusCategoryConfig) StatusesReadyForPlanning() []string {
	if metaStageName := ss.StageConfig.ReadyForPlanningName(); metaStageName == "" {
		return []string{}
	} else {
		return ss.StatusesForMetaStage(metaStageName)
	}
}

func (ss *StatusCategoryConfig) StatusesInDevelopment() []string {
	if metaStageName := ss.StageConfig.InDevelopmentName(); metaStageName == "" {
		return []string{}
	} else {
		return ss.StatusesForMetaStage(metaStageName)
	}
}

func (ss *StatusCategoryConfig) StatusesDone() []string { // not backlog
	if metaStageName := ss.StageConfig.DoneName(); metaStageName == "" {
		return []string{}
	} else {
		return ss.StatusesForMetaStage(metaStageName)
	}
}

func (ss *StatusCategoryConfig) StatusesForMetaStage(metaStatus string) []string {
	var statuses []string
	for k, v := range ss.Map {
		if v == metaStatus {
			statuses = append(statuses, k)
		}
	}
	return stringsutil.SliceCondenseSpace(statuses, true, true)
}

func (ss *StatusCategoryConfig) StatusesInDevelopmentAndDone() []string { // not backlog
	var statuses []string
	statuses = append(statuses, ss.StatusesInDevelopment()...)
	statuses = append(statuses, ss.StatusesDone()...)
	return stringsutil.SliceCondenseSpace(statuses, true, true)
}

type StatusCategories struct {
	CategoryOrder         []string
	UnknownCategory       string
	MapCategoryToStatuses map[string][]string
	MapStatusToCategory   map[string]string
}

func NewStatusCategories() StatusCategories {
	return StatusCategories{
		CategoryOrder:         []string{},
		MapCategoryToStatuses: map[string][]string{},
		MapStatusToCategory:   map[string]string{},
	}
}

func (sc *StatusCategories) AddMapCategoryToStatuses(m map[string][]string) {
	if sc.MapCategoryToStatuses == nil {
		sc.MapCategoryToStatuses = map[string][]string{}
	}
	for k2, vs2 := range m {
		vs1, ok := sc.MapCategoryToStatuses[k2]
		if !ok {
			sc.MapCategoryToStatuses[k2] = []string{}
		}
		vs1 = append(vs1, vs2...)
		sort.Strings(vs1)
		vs1 = slicesutil.Dedupe(vs1)
		sc.MapCategoryToStatuses[k2] = vs1
	}
	sc.buildMapStatusToCategory()
}

func (sc *StatusCategories) buildMapStatusToCategory() {
	if len(sc.MapCategoryToStatuses) == 0 {
		return
	}
	out := map[string]string{}
	for cat, stats := range sc.MapCategoryToStatuses {
		for _, stat := range stats {
			out[stat] = cat
		}
	}
	sc.MapStatusToCategory = out
}

func (sc *StatusCategories) StatusesForCategories(cats []string, matchLCTrimSpace bool) []string {
	var out []string
	for _, wantCat := range cats {
		if matchLCTrimSpace {
			wantCat = strings.ToLower(strings.TrimSpace(wantCat))
		}
		for tryCat, statuses := range sc.MapCategoryToStatuses {
			if matchLCTrimSpace {
				tryCat = strings.ToLower(strings.TrimSpace(tryCat))
			}
			if tryCat == wantCat {
				out = append(out, statuses...)
			}
		}
	}
	out = slicesutil.Dedupe(out)
	sort.Strings(out)
	return out
}
