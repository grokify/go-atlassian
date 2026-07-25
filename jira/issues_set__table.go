package jira

import (
	"fmt"
	"strings"
	"time"

	"github.com/grokify/gocharts/v2/data/table"
	"github.com/grokify/mogo/errors/errorsutil"
	"github.com/grokify/mogo/pointer"
	"github.com/grokify/mogo/strconv/strconvutil"
)

type CustomTableCols struct {
	Cols []CustomCol
}

func (cols CustomTableCols) Names(defaultToSlug bool) []string {
	var names []string
	for _, c := range cols.Cols {
		if c.Name != "" {
			names = append(names, c.Name)
		} else if defaultToSlug {
			names = append(names, c.Slug)
		} else {
			names = append(names, "")
		}
	}
	return names
}

type CustomCol struct {
	Name       string
	Slug       string
	Type       string
	Func       func(im IssueMore) (string, error)
	RenderSkip bool
}

func (c CustomCol) NameOrSlug() string {
	if c.Name != "" {
		return c.Name
	} else {
		return c.Slug
	}
}

func CustomTableColsFromStrings(cols []string) CustomTableCols {
	ccols := CustomTableCols{Cols: []CustomCol{}}
	for _, c := range cols {
		ccols.Cols = append(ccols.Cols, CustomCol{
			Name: c,
			Slug: c,
			Type: table.FormatString,
		})
	}
	return ccols
}

func DefaultIssuesSetTableColumns(inclInitiative, inclEpic bool) table.ColumnDefinitionSet {
	var defs []table.ColumnDefinition
	if inclInitiative {
		initiativeCols := []table.ColumnDefinition{
			{Name: "Initiative Key", Format: table.FormatURL},
			{Name: "Initiative Name"}}
		defs = append(defs, initiativeCols...)
	}
	if inclEpic {
		epicCols := []table.ColumnDefinition{
			{Name: "Epic Key", Format: table.FormatURL},
			{Name: "Epic Name"}}
		defs = append(defs, epicCols...)
	}
	stdCols := []table.ColumnDefinition{
		{Name: "Issue Key", Format: table.FormatURL},
		{Name: "Issue Type"},
		{Name: "Project"},
		{Name: "Summary"},
		{Name: "Status"},
		{Name: "Resolution"},
		// {Name: "Aggregate Original Time Estimate Seconds", Format: table.FormatInt},
		// {Name: "Original Estimate Seconds", Format: table.FormatInt},
		{Name: "Original Estimate Days", Format: table.FormatFloat},
		{Name: "Estimate Days", Format: table.FormatFloat},
		{Name: "Time Spent Days", Format: table.FormatFloat},
		{Name: "Time Remaining Days", Format: table.FormatFloat},
		{Name: "Created", Format: table.FormatString},
	}
	defs = append(defs, stdCols...)

	// defs = append(defs, {Name: "Epic Key", Format: table.FormatURL},

	return table.ColumnDefinitionSet{
		Definitions: defs,
	}
}

// TableSet is designed to return a `table.TableSet` where the tables include a list of issues and optionally, epics, and/or initiatives.
func (set *IssuesSet) TableSet(customCols *CustomTableCols, inclEpic bool, initiativeType string, customFieldLabels []string) (*table.TableSet, error) {
	ts := table.NewTableSet("Jira Issues")
	tbl1Issues, err := set.TableDefault(customCols, inclEpic, initiativeType, customFieldLabels)
	if err != nil {
		return nil, err
	}
	tbl1Issues.Name = TypeIssue
	ts.TableMap[tbl1Issues.Name] = tbl1Issues
	ts.Order = append(ts.Order, tbl1Issues.Name)
	if inclEpic {
		isEpic, err := set.IssuesSetHighestType(TypeEpic)
		if err != nil {
			return nil, errorsutil.Wrapf(err, "error on `is.IssuesSetHighestType(%s)`", TypeEpic)
		}
		tbl2Epics, err := isEpic.TableDefault(customCols, false, initiativeType, customFieldLabels)
		if err != nil {
			return nil, err
		}
		tbl2Epics.Name = TypeEpic
		ts.TableMap[tbl2Epics.Name] = tbl2Epics
		ts.Order = append(ts.Order, tbl2Epics.Name)
	}

	if initiativeType != "" {
		isInit, err := set.IssuesSetHighestType(initiativeType)
		if err != nil {
			return nil, err
		}
		tbl3Initiatives, err := isInit.TableDefault(customCols, false, "", customFieldLabels)
		if err != nil {
			return nil, err
		}
		tbl3Initiatives.Name = initiativeType
		ts.TableMap[tbl3Initiatives.Name] = tbl3Initiatives
		ts.Order = append(ts.Order, tbl3Initiatives.Name)
	}
	return ts, nil
}

// TableDefault returns a `table.Table` where each record is a Jira issue starting with a linked issue key.
func (set *IssuesSet) TableDefault(customCols *CustomTableCols, inclEpic bool, initiativeType string, customFieldLabels []string) (*table.Table, error) {
	if set.Config == nil {
		set.Config = NewConfigDefault()
	}
	initiativeType = strings.TrimSpace(initiativeType)
	var inclInitiative bool
	if initiativeType != "" {
		inclInitiative = true
	} else {
		inclInitiative = false
	}
	baseURL := strings.TrimSpace(set.Config.ServerURL)

	tbl := table.NewTable("issues")

	tbl.LoadColumnDefinitionSet(DefaultIssuesSetTableColumns(inclInitiative, inclEpic))

	if customCols != nil {
		lenCols := len(tbl.Columns)
		for i, customCol := range customCols.Cols {
			if customCol.RenderSkip {
				continue
			}
			j := lenCols + i
			customCol.Type = strings.TrimSpace(customCol.Type)
			if customCol.Type != "" {
				tbl.FormatMap[j] = customCol.Type
			}
			if name := strings.TrimSpace(customCol.Name); name != "" {
				tbl.Columns = append(tbl.Columns, name)
			} else {
				tbl.Columns = append(tbl.Columns, fmt.Sprintf("Column %d", j+1))
			}
		}
	}

	for key, iss := range set.Items {
		issMore := NewIssueMore(pointer.Pointer(iss))
		issMeta := issMore.Meta(baseURL, customFieldLabels)

		lineage, err := set.Lineage(key, customFieldLabels)
		if err != nil {
			return nil, errorsutil.Wrapf(err, "is.Lineage(key) key=\"%s\"", key)
		}

		timeRemainingSecs := iss.Fields.TimeEstimate - iss.Fields.TimeSpent
		if timeRemainingSecs < 0 ||
			// issMore.Status() == StatusClosed ||
			issMore.Status() == StatusDone {
			timeRemainingSecs = 0
		}

		var row []string

		if inclInitiative {
			initKeyDispplay := ""
			initName := ""
			if initiative := lineage.HighestType(initiativeType); initiative != nil {
				initiative.BuildKeyURL(baseURL) // should not be needed.
				initKeyDispplay = initiative.KeyLinkMarkdown()
				initName = initiative.Summary
			}
			row = append(row, initKeyDispplay, initName)
		}

		if inclEpic {
			epicKeyDisplay := issMore.EpicKey()
			epicName := ""
			epic := lineage.HighestEpic()
			if epic != nil {
				epic.BuildKeyURL(baseURL) // should not be needed.
				epicKeyDisplay = epic.KeyLinkMarkdown()
				epicName = epic.Summary
			}
			row = append(row, epicKeyDisplay, epicName)
		}

		stdCells := []string{
			issMeta.KeyLinkMarkdown(),
			issMore.Type(),
			issMore.Project(),
			issMore.Summary(),
			issMore.Status(),
			issMore.Resolution(),
			// strconv.Itoa(iss.Fields.AggregateTimeOriginalEstimate),
			// strconv.Itoa(iss.Fields.TimeOriginalEstimate),
			strconvutil.Ftoa(set.Config.SecondsToWorkingDays(iss.Fields.TimeOriginalEstimate), -1),
			strconvutil.Ftoa(set.Config.SecondsToWorkingDays(iss.Fields.TimeEstimate), -1),
			strconvutil.Ftoa(set.Config.SecondsToWorkingDays(iss.Fields.TimeSpent), -1),
			strconvutil.Ftoa(set.Config.SecondsToWorkingDays(timeRemainingSecs), -1),
			issMore.CreateTime().Format(time.RFC3339),
			// time.Time(iss.Fields.Created).Format(time.RFC3339),
			// strconvutil.FormatFloat64Simple(float64(ix.TimeRemainingEstimate.Days(is.Config.WorkingHoursPerDay))),
		}
		row = append(row, stdCells...)

		if customCols != nil {
			for _, cc := range customCols.Cols {
				if cc.RenderSkip {
					continue
				} else if cc.Func == nil {
					row = append(row, "")
				} else if val, err := cc.Func(issMore); err != nil {
					return nil, err
				} else {
					row = append(row, val)
				}
			}
		}

		tbl.Rows = append(tbl.Rows, row)
	}
	return &tbl, nil
}

func (set *IssuesSet) TableSimple(cols []string) (*table.Table, error) {
	ccols := CustomTableColsFromStrings(cols)
	return set.Table(ccols)
}

func (set *IssuesSet) Table(cols CustomTableCols) (*table.Table, error) {
	tbl := table.NewTable("issues")
	tbl.Columns = cols.Names(true)
	for i, col := range cols.Cols {
		if col.Type != "" {
			tbl.FormatMap[i] = col.Type
		} else if col.Slug == CalcCreatedAgeDays {
			tbl.FormatMap[i] = table.FormatInt
		} else if col.Slug == FieldCreatedDate {
			tbl.FormatMap[i] = table.FormatDate
		} else if col.Slug == CalcCreatedMonth {
			tbl.FormatMap[i] = table.FormatDate
		}
	}
	for _, iss := range set.Items {
		issMore := NewIssueMore(pointer.Pointer(iss))
		var row []string
		for _, col := range cols.Cols {
			colSlug := strings.ToLower(strings.TrimSpace(col.Slug))
			if col.Func != nil {
				if val, err := col.Func(issMore); err != nil {
					return nil, err
				} else {
					row = append(row, val)
				}
			} else if val, ok := issMore.Value(colSlug); ok {
				row = append(row, val)
			} else if canonicalCustomKey, ok := IsCustomFieldKey(col.Slug); ok {
				row = append(row, issMore.CustomFieldStringOrDefault(canonicalCustomKey, ""))
			} else {
				row = append(row, "")
			}
		}
		tbl.Rows = append(tbl.Rows, row)
	}
	return &tbl, nil
}
