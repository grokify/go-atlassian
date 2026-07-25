package xml

import (
	"fmt"
	"strconv"
	"strings"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/go-atlassian/jira"
	"github.com/grokify/gocharts/v2/data/histogram"
	"github.com/grokify/gocharts/v2/data/table"
	"github.com/grokify/mogo/net/urlutil"
	"github.com/grokify/mogo/strconv/strconvutil"
	"github.com/grokify/mogo/text/markdown"
	"github.com/grokify/mogo/type/maputil"
)

type IssuesSet struct {
	Config    *jira.Config
	IssuesMap map[string]Issue
}

func NewIssuesSet(cfg *jira.Config) IssuesSet {
	if cfg == nil {
		cfg = jira.NewConfigDefault()
	}
	return IssuesSet{
		Config:    cfg,
		IssuesMap: map[string]Issue{}}
}

func (set *IssuesSet) AddFromAPI(issues ...gojira.Issue) error {
	for _, iss := range issues {
		err := set.Add(IssueFromAPI(iss))
		if err != nil {
			return err
		}
	}
	return nil
}

func (set *IssuesSet) Add(issues ...Issue) error {
	if set.IssuesMap == nil {
		set.IssuesMap = map[string]Issue{}
	}
	missingKeyIndexes := []string{}
	for i, ix := range issues {
		k := ix.GetKey()
		if k == "" {
			missingKeyIndexes = append(missingKeyIndexes, strconv.Itoa(i))
			continue
		}
		ix.TrimSpace()
		set.IssuesMap[k] = ix
	}
	if len(missingKeyIndexes) > 0 {
		return fmt.Errorf("missingkeyindexes (%s)", strings.Join(missingKeyIndexes, ","))
	}
	return nil
}

func (set *IssuesSet) ReadFile(filename string) error {
	x, err := ReadFile(filename)
	if err != nil {
		return err
	}
	return set.Add(x.Channel.Issues...)
}

func (set *IssuesSet) Keys() []string {
	return maputil.Keys(set.IssuesMap)
}

func (set *IssuesSet) Table(baseURL string) table.Table {
	baseURL = strings.TrimSpace(baseURL)

	if set.Config == nil {
		set.Config = jira.NewConfigDefault()
	}
	tbl := table.NewTable("issues")
	tbl.Columns = []string{"Type", "Key", "Summary", "Status", "Resolution", "Aggregate Original Time Estimate Seconds", "Original Estimate Seconds", "Original Estimate Days", "Estimate Days", "Time Spent", "Time Remaining"}
	tbl.FormatMap = map[int]string{
		1: table.FormatURL,
		5: table.FormatInt,
		6: table.FormatInt,
		7: table.FormatFloat,
		8: table.FormatFloat,
		9: table.FormatFloat,
	}
	for k, ix := range set.IssuesMap {
		keyDisplay := k
		if len(baseURL) > 0 {
			keyURL := urlutil.JoinAbsolute(baseURL, "/browse/", k)
			keyDisplay = markdown.Linkify(keyURL, k)
		}
		tbl.Rows = append(tbl.Rows, []string{
			ix.Type.DisplayName,
			keyDisplay,
			// ix.Title,
			ix.Summary,
			ix.Status.DisplayName,
			ix.Resolution.DisplayName,
			strconv.Itoa(int(ix.AggregateTimeOriginalEstimate.Seconds)),
			strconv.Itoa(int(ix.TimeOriginalEstimate.Seconds)),
			strconvutil.Ftoa(float64(ix.TimeOriginalEstimate.Days(set.Config.WorkingHoursPerDay)), -1),
			strconvutil.Ftoa(float64(ix.TimeEstimate.Days(set.Config.WorkingHoursPerDay)), -1),
			strconvutil.Ftoa(float64(ix.TimeSpent.Days(set.Config.WorkingHoursPerDay)), -1),
			// strconvutil.FormatFloat64Simple(float64(ix.TimeRemainingEstimate.Days(is.Config.WorkingHoursPerDay))),
		})
	}
	return tbl
}

func (set *IssuesSet) AggregateTimeEstimate() int64 {
	var agg int64
	for _, ix := range set.IssuesMap {
		agg += ix.TimeEstimate.Seconds
	}
	return agg
}

func (set *IssuesSet) AggregateTimeOriginalEstimate() int64 {
	var agg int64
	for _, ix := range set.IssuesMap {
		agg += ix.TimeOriginalEstimate.Seconds
	}
	return agg
}

func (set *IssuesSet) AggregateTimeRemainingEstimate() int64 {
	var agg int64
	for _, ix := range set.IssuesMap {
		agg += ix.TimeSpent.Seconds
	}
	return agg
}

func (set *IssuesSet) AggregateTimeSpent() int64 {
	var agg int64
	for _, ix := range set.IssuesMap {
		agg += ix.TimeSpent.Seconds
	}
	return agg
}

// TSRHistogramSets returns a `*histogram.HistogramSets` for Type, Status and Resolution.
func (set *IssuesSet) TSRHistogramSets(name string) *histogram.HistogramSets {
	if strings.TrimSpace(name) == "" {
		name = "TSR"
	}
	hset := histogram.NewHistogramSets(name)
	for _, iss := range set.IssuesMap {
		hset.Add(
			iss.Type.DisplayName,
			iss.Status.DisplayName,
			iss.Resolution.DisplayName,
			1, true)
	}
	return hset
}

// TSRTable returns a `table.Table` for Type, Status and Resolution.
func (set *IssuesSet) TSRTable(name string) table.Table {
	hset := set.TSRHistogramSets(name)
	return hset.Table("Jira Issues", "Type", "Status", "Resolution", "Count")
}

// TSRWriteCSV writes a CSV file for Type, Status and Resolution.
func (set *IssuesSet) TSRWriteCSV(filename string) error {
	tbl := set.TSRTable("")
	return tbl.WriteCSV(filename)
}

// TSRWriteCSV writes a CSV file for Type, Status and Resolution.
func (set *IssuesSet) TSRWriteXLSX(filename, sheetname string) error {
	tbl := set.TSRTable("")
	return tbl.WriteXLSX(filename, sheetname)
}
