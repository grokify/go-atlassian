package jira

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/grokify/mogo/type/slicesutil"
	"github.com/grokify/mogo/type/stringsutil"
	"github.com/grokify/mogo/type/stringsutil/join"
)

type JQLMeta struct {
	Name            string
	Key             string
	Description     string
	FilterID        int
	QueryTime       time.Time
	QueryTotalCount int
	Directions      []string
}

// JQL is a JQL builder. It will create a JQL string using `JQL.String()` from the supplied infomration.
type JQL struct {
	Meta            JQLMeta // Not part of JQL
	CreatedGT       *time.Time
	CreatedGTE      *time.Time
	CreatedLT       *time.Time
	CreatedLTE      *time.Time
	DueGT           *time.Time
	DueGTE          *time.Time
	DueLT           *time.Time
	DueLTE          *time.Time
	UpdatedGT       *time.Time
	UpdatedGTE      *time.Time
	UpdatedLT       *time.Time
	UpdatedLTE      *time.Time
	FiltersIncl     [][]string // outer level is `AND`, inner level is `IN`.
	FiltersExcl     [][]string
	IssuesIncl      [][]string
	IssuesExcl      [][]string
	KeysIncl        [][]string
	KeysExcl        [][]string
	LabelsIncl      [][]string
	LabelsExcl      [][]string
	ParentsIncl     [][]string
	ParentsExcl     [][]string
	ProjectsIncl    [][]string
	ProjectsExcl    [][]string
	ResolutionIncl  [][]string
	ResolutionExcl  [][]string
	StatusesIncl    [][]string
	StatusesExcl    [][]string
	TypesIncl       [][]string
	TypesExcl       [][]string
	SummaryLike     []string
	SummaryNotLike  []string
	TextLike        []string
	TextNotLike     []string
	CustomFieldIncl map[string][]string // slice is `IN`
	CustomFieldExcl map[string][]string
	AnyIncl         JQLAndOrStringer
	AnyExcl         JQLAndOrStringer
	Raw             []string
}

type JQLAndOrStringer [][]fmt.Stringer

func (j JQLAndOrStringer) Fields(not bool) []string {
	if len(j) == 0 {
		return []string{}
	}
	var andConditions []string
	for _, orCondsRaw := range j {
		if len(orCondsRaw) == 0 {
			continue
		}
		var orCondsStrs []string
		for _, orCondRaw := range orCondsRaw {
			if orCondStr := strings.TrimSpace(orCondRaw.String()); orCondStr != "" {
				orCondsStrs = append(orCondsStrs, addParen(orCondStr))
			}
		}
		if len(orCondsStrs) > 0 {
			andCondition := addParen(strings.Join(orCondsStrs, operatorORSpaces))
			if not {
				andConditions = append(andConditions, addParen("NOT "+andCondition))
			} else {
				andConditions = append(andConditions, andCondition)
			}
		}
	}
	return andConditions
}

// JQLOrAndStringer combines `fmt.Stringer` slice of slice to create a JQL.
// Outer slide is "AND", Inner slice is "OR".
func (j JQLAndOrStringer) String(not bool) string {
	if len(j) == 0 {
		return ""
	} else if andConditions := j.Fields(not); len(andConditions) == 0 {
		return ""
	} else {
		return strings.TrimSpace(strings.Join(andConditions, operatorANDSpaces))
	}
}

func addParen(s string) string {
	return addPrefixSuffix("(", ")", s)
}

func addPrefixSuffix(prefix, suffix, s string) string {
	return prefix + s + suffix
}

func (j JQL) String() string {
	conditions := slicesutil.AppendBulk(
		[]string{},
		[][]string{
			j.conditionsStringFields(),
			j.conditionsDateFields(),
			j.conditionsLikeNotLike(),
			j.conditionsCustomFields(),
			j.AnyIncl.Fields(false),
			j.AnyExcl.Fields(true),
			j.Raw,
		},
	)
	conditions = stringsutil.SliceCondenseSpace(conditions, true, false)
	return strings.TrimSpace(strings.Join(conditions, operatorANDSpaces))
}

func (j JQL) conditionsStringFields() []string {
	var conditions []string

	type inclExclProc struct {
		Field   string
		Values  [][]string
		Exclude bool
	}
	procs := []inclExclProc{
		{Field: FieldFilter, Values: j.FiltersIncl, Exclude: false},
		{Field: FieldFilter, Values: j.FiltersExcl, Exclude: true},
		{Field: FieldIssue, Values: j.IssuesIncl, Exclude: false},
		{Field: FieldIssue, Values: j.IssuesExcl, Exclude: true},
		{Field: FieldKey, Values: j.KeysIncl, Exclude: false},
		{Field: FieldKey, Values: j.KeysExcl, Exclude: true},
		{Field: FieldLabels, Values: j.LabelsIncl, Exclude: false},
		{Field: FieldLabels, Values: j.LabelsExcl, Exclude: true},
		{Field: FieldParent, Values: j.ParentsIncl, Exclude: false},
		{Field: FieldParent, Values: j.ParentsExcl, Exclude: true},
		{Field: FieldProject, Values: j.ProjectsIncl, Exclude: false},
		{Field: FieldProject, Values: j.ProjectsExcl, Exclude: true},
		{Field: FieldResolution, Values: j.ResolutionIncl, Exclude: false},
		{Field: FieldResolution, Values: j.ResolutionExcl, Exclude: true},
		{Field: FieldStatus, Values: j.StatusesIncl, Exclude: false},
		{Field: FieldStatus, Values: j.StatusesExcl, Exclude: true},
		{Field: FieldType, Values: j.TypesIncl, Exclude: false},
		{Field: FieldType, Values: j.TypesExcl, Exclude: true},
	}
	for _, proc := range procs {
		if field := strings.TrimSpace(proc.Field); field == "" {
			panic("field is empty")
		} else if len(proc.Values) > 0 {
			for _, inConditionVals := range proc.Values {
				if cond := inCondition(proc.Field, inConditionVals, proc.Exclude); cond != "" {
					conditions = append(conditions, cond)
				}
			}
		}
	}
	return conditions
}

func (j JQL) conditionsDateFields() []string {
	var conditions []string

	type dateProc struct {
		Field    string
		Operator string
		Time     *time.Time
	}
	procs := []dateProc{
		{Field: FieldCreatedDate, Operator: OperatorGT, Time: j.CreatedGT},
		{Field: FieldCreatedDate, Operator: OperatorGTE, Time: j.CreatedGTE},
		{Field: FieldCreatedDate, Operator: OperatorLT, Time: j.CreatedLT},
		{Field: FieldCreatedDate, Operator: OperatorLTE, Time: j.CreatedLTE},
		{Field: FieldDueDate, Operator: OperatorGT, Time: j.DueGT},
		{Field: FieldDueDate, Operator: OperatorGTE, Time: j.DueGTE},
		{Field: FieldDueDate, Operator: OperatorLT, Time: j.DueLT},
		{Field: FieldDueDate, Operator: OperatorLTE, Time: j.DueLTE},
		{Field: FieldUpdated, Operator: OperatorGT, Time: j.UpdatedGT},
		{Field: FieldUpdated, Operator: OperatorGTE, Time: j.UpdatedGTE},
		{Field: FieldUpdated, Operator: OperatorLT, Time: j.UpdatedLT},
		{Field: FieldUpdated, Operator: OperatorLTE, Time: j.UpdatedLTE},
	}
	for _, proc := range procs {
		if field := strings.TrimSpace(proc.Field); field == "" {
			panic("field is empty")
		} else if op := strings.TrimSpace(proc.Operator); op == "" {
			panic("operator is empty")
		} else if proc.Time != nil && !proc.Time.IsZero() {
			conditions = append(conditions,
				fmt.Sprintf("%s %s %s", field, op, proc.Time.Format(time.DateOnly)),
			)
		}
	}
	return conditions
}

func (j JQL) conditionsLikeNotLike() []string {
	var conditions []string
	for _, v := range j.SummaryLike {
		conditions = append(conditions, fmt.Sprintf("%s %s \"%s\"", FieldSummary, OperatorLike, v))
	}
	for _, v := range j.SummaryNotLike {
		conditions = append(conditions, fmt.Sprintf("%s %s \"%s\"", FieldSummary, OperatorNotLike, v))
	}
	for _, v := range j.TextLike {
		conditions = append(conditions, fmt.Sprintf("%s %s \"%s\"", FieldText, OperatorLike, v))
	}
	for _, v := range j.TextNotLike {
		conditions = append(conditions, fmt.Sprintf("%s %s \"%s\"", FieldText, OperatorNotLike, v))
	}

	return conditions
}

func (j JQL) conditionsCustomFields() []string {
	var conditions []string
	for cfk, cfv := range j.CustomFieldIncl {
		cfv = stringsutil.SliceCondenseSpace(cfv, true, false)
		if len(cfv) == 0 {
			continue
		}
		if cfkCanonicalID, err := CustomFieldLabelToID(cfk); err == nil {
			cfk = cfkCanonicalID.StringBrackets()
		}
		if cond := inCondition(cfk, cfv, false); cond != "" {
			conditions = append(conditions, cond)
		}
	}
	for cfk, cfv := range j.CustomFieldExcl {
		cfv = stringsutil.SliceCondenseSpace(cfv, true, false)
		if len(cfv) == 0 {
			continue
		}
		if cfkCanonicalID, err := CustomFieldLabelToID(cfk); err == nil {
			cfk = cfkCanonicalID.StringBrackets()
		}
		if cond := inCondition(cfk, cfv, true); cond != "" {
			conditions = append(conditions, cond)
		}
	}
	return conditions
}

func (j JQL) QueryString() string {
	return "jql=" + url.QueryEscape(j.String())
}

// quoteFieldIfNeeded wraps field names containing spaces in double quotes for JQL.
func quoteFieldIfNeeded(field string) string {
	if strings.Contains(field, " ") {
		return `"` + field + `"`
	}
	return field
}

func inCondition(field string, values []string, exclude bool) string {
	field = strings.TrimSpace(field)
	values = stringsutil.SliceCondenseSpace(values, true, true)
	if field == "" || len(values) == 0 {
		return ""
	}
	field = quoteFieldIfNeeded(field)
	if len(values) == 1 {
		operator := "="
		if exclude {
			operator = "!="
		}
		qtr := stringsutil.Quoter{
			Beg:         "'",
			End:         "'",
			SkipNesting: true,
		}
		return fmt.Sprintf("%s %s %s", field, operator, qtr.Quote(values[0]))
	} else if len(values) > 1 {
		operator := "IN"
		if exclude {
			operator = "NOT IN"
		}
		return fmt.Sprintf("%s %s (%s)", field, operator, join.JoinQuote(values, "'", "'", JQLInSep))
	} else {
		return ""
	}
}

// JQLStringsSimple provides a set of JQLs for a single field and values. The purpose of this function
// is to split very long lists of values so that each JQL is under a certain length limit.
func JQLStringsSimple(field string, exclude bool, vals []string, jqlMaxLength int) []string {
	if jqlMaxLength < 0 {
		jqlMaxLength *= -1
	}
	field = strings.TrimSpace(field)
	if field == "" {
		return []string{}
	}
	vals = stringsutil.SliceCondenseSpace(vals, true, true)
	if len(vals) == 0 {
		return []string{}
	}
	var jqls []string
	operator := "IN"
	if exclude {
		operator = "NOT IN"
	}
	baseString := fmt.Sprintf("%s %s ()", field, operator)
	baseStringLen := len(baseString)
	quoter := stringsutil.Quoter{
		Beg:         "'",
		End:         "'",
		SkipNesting: true,
	}
	if jqlMaxLength == 0 {
		jqlMaxLength = JQLMaxLength
	}
	valsMaxLength := jqlMaxLength - baseStringLen
	valsString := ""
	for i, val := range vals {
		valQuoted := quoter.Quote(val)
		if len(valsString)+len(valQuoted) > valsMaxLength {
			jqls = append(jqls, fmt.Sprintf("%s %s (%s)", field, operator, valsString))
			valsString = ""
		}
		valsString += valQuoted
		if i < len(vals)-1 {
			valsString += JQLInSep
		}
	}
	if len(valsString) > 0 {
		jqls = append(jqls, fmt.Sprintf("%s %s (%s)", field, operator, valsString))
	}
	return jqls
}
