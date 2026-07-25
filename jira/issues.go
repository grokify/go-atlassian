package jira

import (
	"encoding/json"
	"os"
	"strconv"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/gocharts/v2/data/histogram"
	"github.com/grokify/mogo/encoding/jsonutil"
	"github.com/grokify/mogo/type/slicesutil"
	"github.com/grokify/mogo/type/stringsutil"
)

type Issues []gojira.Issue

func (ii Issues) CountsByType() map[string]int {
	counts := map[string]int{}
	for _, iss := range ii {
		name := iss.Fields.Type.Name
		counts[name]++
		counts["_total"]++
	}
	return counts
}

// CountsByProjectTypeStatus returns a `*histogram.Histogram` with issue counts
// by project, type, and status. This can be used to export CSV and XLSX sheets
// for analysis.
func (ii Issues) CountsByProjectTypeStatus() *histogram.HistogramSets {
	hsets := histogram.NewHistogramSets("")
	for _, iss := range ii {
		hsets.Add(
			iss.Fields.Project.Key,
			iss.Fields.Type.Name,
			iss.Fields.Status.Name,
			1,
			true)
	}
	return hsets
}

func (ii Issues) AddRank() Issues {
	nii := Issues{}
	for i, iss := range ii {
		if iss.Fields == nil {
			iss.Fields = &gojira.IssueFields{}
		}
		iss.Fields.Unknowns[MetaParamRank] = strconv.Itoa(i)
		nii = append(nii, iss)
	}
	return nii
}

// IssuesSet converts the `Issues{}` to an `*IssuesSet{}`.
func (ii Issues) IssuesSet(cfg *Config) (*IssuesSet, error) {
	is := NewIssuesSet(cfg)
	err := is.Add(ii...)
	return is, err
}

// Keys returns a slice of sorted issue keys.
func (ii Issues) Keys() []string {
	var keys []string
	for _, iss := range ii {
		iss := iss
		im := NewIssueMore(&iss)
		keys = append(keys, im.Key())
	}
	return stringsutil.SliceCondenseSpace(keys, true, true)
}

func (ii Issues) KeysSplitVenn(keys []string) slicesutil.Venn2ResultOrdered[string] {
	keys = stringsutil.SliceCondenseSpace(keys, true, true)
	return slicesutil.Venn2Sort(ii.Keys(), keys)
}

func (ii Issues) Metas(baseURL string, customFieldLabels []string) IssueMetas {
	metas := IssueMetas{}
	for _, iss := range ii {
		iss := iss
		im := NewIssueMore(&iss)
		metas = append(metas, im.Meta(baseURL, customFieldLabels))
	}
	return metas
}

// WriteFileJSON outputs a JSON file that can be read in again.
func (ii Issues) WriteFileJSON(filename, prefix, indent string) error {
	if b, err := jsonutil.MarshalSimple(ii, prefix, indent); err != nil {
		return err
	} else {
		return os.WriteFile(filename, b, 0600)
	}
}

func IssuesReadFileJSON(filename string) (Issues, error) {
	var ii Issues
	if b, err := os.ReadFile(filename); err != nil {
		return ii, err
	} else {
		return ii, json.Unmarshal(b, &ii)
	}
}
