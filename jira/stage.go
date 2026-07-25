package jira

import (
	"strings"

	"github.com/grokify/mogo/type/stringsutil"
)

type StageConfig struct {
	StageNamePlanning            string
	StageNameDesign              string
	StageNameDevelopment         string
	StageNameTesting             string
	StageNameDeployment          string
	StageNameReview              string
	StageNameDone                string
	MetaStageInPlanning          string // Override auto-build
	MetaStageReadyForPlanning    string
	MetaStageInDesign            string
	MetaStageReadyForDesign      string
	MetaStageInDevelopment       string
	MetaStageReadyForDevelopment string
	MetaStageReadyForTesting     string
	MetaStageInTesting           string
	MetaStageReadyForDeployment  string
	MetaStageInDeployment        string
	MetaStageReadyForReview      string
	MetaStageInReview            string
	MetaStageDone                string
	MetaStagePrefixIn            string
	MetaStagePrefixReadyFor      string
}

func NewStageConfigEmpty() *StageConfig {
	return &StageConfig{}
}

func DefaultStageSet() *StageConfig {
	return &StageConfig{
		StageNamePlanning:       StagePlanning,
		StageNameDesign:         StageDesign,
		StageNameDevelopment:    StageDevelopment,
		StageNameTesting:        StageTesting,
		StageNameDeployment:     StageDeployment,
		StageNameReview:         StageReview,
		StageNameDone:           StatusDone,
		MetaStagePrefixIn:       MetaStagePrefixIn,
		MetaStagePrefixReadyFor: MetaStagePrefixReadyFor,
	}
}

func (ss *StageConfig) InPlanningName() string {
	return buildMetaStageName(ss.MetaStageInPlanning, ss.MetaStagePrefixIn, ss.StageNamePlanning)
}

func (ss *StageConfig) ReadyForPlanningName() string {
	return buildMetaStageName(ss.MetaStageReadyForPlanning, ss.MetaStagePrefixReadyFor, ss.StageNamePlanning)
}

func (ss *StageConfig) InDesignName() string {
	return buildMetaStageName(ss.MetaStageInDesign, ss.MetaStagePrefixIn, ss.StageNameDesign)
}

func (ss *StageConfig) ReadyForDesignName() string {
	return buildMetaStageName(ss.MetaStageReadyForDesign, ss.MetaStagePrefixReadyFor, ss.StageNameDesign)
}

func (ss *StageConfig) InDevelopmentName() string {
	return buildMetaStageName(ss.MetaStageInDevelopment, ss.MetaStagePrefixIn, ss.StageNameDevelopment)
}

func (ss *StageConfig) ReadyForDevelopmentName() string {
	return buildMetaStageName(ss.MetaStageReadyForDevelopment, ss.MetaStagePrefixReadyFor, ss.StageNameDevelopment)
}

func (ss *StageConfig) InTestingName() string {
	return buildMetaStageName(ss.MetaStageInTesting, ss.MetaStagePrefixIn, ss.StageNameTesting)
}

func (ss *StageConfig) ReadyForTestingName() string {
	return buildMetaStageName(ss.MetaStageReadyForTesting, ss.MetaStagePrefixReadyFor, ss.StageNameTesting)
}

func (ss *StageConfig) InDeploymentName() string {
	return buildMetaStageName(ss.MetaStageInDeployment, ss.MetaStagePrefixIn, ss.StageNameDeployment)
}

func (ss *StageConfig) ReadyForDeploymentName() string {
	return buildMetaStageName(ss.MetaStageReadyForDeployment, ss.MetaStagePrefixReadyFor, ss.StageNameDeployment)
}

func (ss *StageConfig) InReviewName() string {
	return buildMetaStageName(ss.MetaStageInReview, ss.MetaStagePrefixIn, ss.StageNameTesting)
}

func (ss *StageConfig) ReadyForReviewName() string {
	return buildMetaStageName(ss.MetaStageReadyForReview, ss.MetaStagePrefixReadyFor, ss.StageNameTesting)
}

func (ss *StageConfig) DoneName() string {
	return buildMetaStageName(ss.MetaStageDone, "", ss.StageNameDone)
}

func (ss *StageConfig) Exists(stageName string) bool {
	m := ss.Map()
	_, ok := m[stageName]
	return ok
}

func (ss *StageConfig) TrimSpaceNames() {
	ss.StageNamePlanning = strings.TrimSpace(ss.StageNamePlanning)
	ss.StageNameDesign = strings.TrimSpace(ss.StageNameDesign)
	ss.StageNameDevelopment = strings.TrimSpace(ss.StageNameDevelopment)
	ss.StageNameTesting = strings.TrimSpace(ss.StageNameTesting)
	ss.StageNameDeployment = strings.TrimSpace(ss.StageNameDeployment)
	ss.StageNameReview = strings.TrimSpace(ss.StageNameReview)
	ss.StageNameDone = strings.TrimSpace(ss.StageNameDone)
}

func (ss *StageConfig) Order() []string {
	return stringsutil.SliceCondenseSpace([]string{
		ss.ReadyForPlanningName(),
		ss.InPlanningName(),
		ss.ReadyForDesignName(),
		ss.InDesignName(),
		ss.ReadyForDevelopmentName(),
		ss.InDevelopmentName(),
		ss.ReadyForTestingName(),
		ss.InTestingName(),
		ss.ReadyForDeploymentName(),
		ss.InDeploymentName(),
		ss.ReadyForReviewName(),
		ss.InReviewName(),
		ss.DoneName(),
	}, true, false)
}

func (ss *StageConfig) Map() map[string]int {
	names := ss.Order()
	out := map[string]int{}
	for i, name := range names {
		out[name] = i
	}
	return out
}

func buildMetaStageName(metaStageName, prefix, stageName string) string {
	metaStageName = strings.TrimSpace(metaStageName)
	stageName = strings.TrimSpace(stageName)
	if metaStageName != "" {
		return metaStageName
	} else if stageName == "" {
		return ""
	} else {
		return prefix + stageName
	}
}
