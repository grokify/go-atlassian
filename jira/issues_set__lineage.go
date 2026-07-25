package jira

import (
	"errors"
	"fmt"
	"strings"

	"github.com/grokify/mogo/errors/errorsutil"
	"github.com/grokify/mogo/type/stringsutil"
)

var ErrLineageNotFound = errors.New("lineage not found")

// Lineage returns a slice of `IssueMeta` where the supplied key is in index 0 and the most senior
// parent is the last element of the slice. If a parent is not found in the set, an error is returned.
func (set *IssuesSet) Lineage(key string, customFieldLabels []string) (IssueMetas, error) {
	if key == "Epic" {
		panic("Lineage Epic")
	}
	ims := IssueMetas{}
	iss, err := set.Issue(key)
	if err != nil {
		return ims, errorsutil.Wrapf(err, "key not found (%s)", key)
	}
	im := NewIssueMore(&iss)
	imeta := im.Meta(set.Config.ServerURL, customFieldLabels)
	ims = append(ims, imeta)
	parKey := im.ParentKey()

	if parKey != "" && set.Parents == nil {
		return ims, errors.New("parents not set")
	}

	for parKey != "" {
		parIss, err := set.Issue(parKey)
		if err != nil {
			return ims, errorsutil.Wrap(err, "parent not found")
		}
		parIM := NewIssueMore(&parIss)
		parM := parIM.Meta(set.Config.ServerURL, customFieldLabels)
		ims = append(ims, parM)
		parKey = parIM.ParentKey()
	}

	return ims, nil
}

func (set *IssuesSet) LineageValidateSet() (popLineage []string, unpopLineage []string, allValid bool) {
	issKeys := set.Keys()
	for _, issKey := range issKeys {
		_, err := set.LineageValidateKey(issKey)
		if err != nil {
			unpopLineage = append(unpopLineage, issKey)
		} else {
			popLineage = append(popLineage, issKey)
		}
	}
	popLineage = stringsutil.SliceCondenseSpace(popLineage, true, true)
	unpopLineage = stringsutil.SliceCondenseSpace(unpopLineage, true, true)
	if len(popLineage) == len(set.Items) && len(unpopLineage) == 0 {
		allValid = true
		return popLineage, unpopLineage, allValid
	}
	return popLineage, unpopLineage, allValid
}

func (set *IssuesSet) LineageTopKeysPopulated() ([]string, error) {
	var linPop []string
	issKeys := set.Keys()
	for _, issKey := range issKeys {
		issKey = strings.TrimSpace(issKey)
		if issKey == "" {
			return linPop, errors.New("issue map key is empty string")
		}
		lin, err := set.LineageValidateKey(issKey)
		if err != nil {
			if errors.Is(err, ErrLineageNotFound) {
				continue
			} else {
				panic(err)
			}
		} else {
			if len(lin) > 0 {
				linPop = append(linPop, lin[len(lin)-1])
			} else {
				panic("lineage empty - LineageTopKeysPopulated")
			}
		}
	}
	return stringsutil.SliceCondenseSpace(linPop, true, true), nil
}

func (set *IssuesSet) LineageTopKeysUnpopulated() ([]string, error) {
	var linUnpop []string
	issKeys := set.Keys()
	for _, issKey := range issKeys {
		issKey = strings.TrimSpace(issKey)
		if issKey == "" {
			return linUnpop, errors.New("issue map key is empty string")
		}
		lin, err := set.LineageValidateKey(issKey)
		if err != nil {
			if errors.Is(err, ErrLineageNotFound) {
				if len(lin) > 0 {
					linUnpop = append(linUnpop, lin[len(lin)-1])
				} else {
					panic("linage empty - LineageTopKeysUnpopulated")
				}
			} else {
				panic(err)
			}
		}
	}
	return stringsutil.SliceCondenseSpace(linUnpop, true, true), nil
}

// LineageValidateKey returns a lineage slice where the leaf key is in index position 0 (little-endian).
// This is done in case a parent cannot be found in which case the boolean returned is false.
func (set *IssuesSet) LineageValidateKey(key string) ([]string, error) {
	key = strings.TrimSpace(key)
	var lineage []string
	if key == "" {
		return lineage, errors.New("key not provided")
	}
	iss, ok := set.IssueOrParent(key)
	if !ok {
		return lineage, fmt.Errorf("key not found for (%s)", key)
	}
	im := NewIssueMore(iss)
	issKey := im.Key()
	if issKey != key {
		return lineage, fmt.Errorf("found key (%s) did not match request (%s)", issKey, key)
	} else {
		lineage = append(lineage, issKey)
	}
	parKey := im.ParentKey()
	for parKey != "" {
		lineage = append(lineage, parKey)
		parIss, ok := set.IssueOrParent(parKey)
		if !ok {
			return lineage, errorsutil.Wrapf(ErrLineageNotFound, "parent key not found (%s)", parKey)
		}
		parIssMore := NewIssueMore(parIss)
		parIssKey := parIssMore.Key()
		if parIssKey != parKey {
			return lineage, fmt.Errorf("found key (%s) did not match request (%s)", parIssKey, parKey)
		}
		parKey = parIssMore.ParentKey()
	}
	return lineage, nil
}
