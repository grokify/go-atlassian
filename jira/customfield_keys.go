package jira

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const customfieldPrefix = "customfield_"

var (
	ErrInvalidCustomFieldFormat = errors.New("invalid customfield format")
	rxCustomFieldBrackets       = regexp.MustCompile(`^cf\[([0-9]+)\]$`)
	rxCustomFieldPrefix         = regexp.MustCompile(`^customfield_([0-9]+)$`)
	rxCustomFieldDigits         = regexp.MustCompile(`^[0-9]+$`)
)

// CustomFieldLabelToID converts a custom field string to `customfield_12345`.
func CustomFieldLabelToID(key string) (CustomFieldID, error) {
	key = strings.ToLower(strings.TrimSpace(key))
	if rxCustomFieldDigits.MatchString(key) {
		if keyID, err := strconv.Atoi(key); err != nil {
			return 0, err
		} else {
			return CustomFieldID(keyID), nil
		}
	} else if m := rxCustomFieldPrefix.FindAllStringSubmatch(key, -1); len(m) > 0 {
		n := m[0]
		if len(n) > 1 {
			keyIDStr := n[1]
			if keyID, err := strconv.Atoi(keyIDStr); err != nil {
				return 0, err
			} else {
				return CustomFieldID(keyID), nil
			}
		}
	} else if m := rxCustomFieldBrackets.FindAllStringSubmatch(key, -1); len(m) > 0 {
		n := m[0]
		if len(n) > 1 {
			keyIDStr := n[1]
			if keyID, err := strconv.Atoi(keyIDStr); err != nil {
				return 0, err
			} else {
				return CustomFieldID(keyID), nil
			}
		}
	}
	return 0, ErrInvalidCustomFieldFormat
}

func CustomFieldKeyAnyToBrackets(key string) (string, error) {
	if cfID, err := CustomFieldLabelToID(key); err != nil {
		return "", err
	} else {
		return cfID.StringBrackets(), nil
	}
}

type CustomFieldID int

// StringBrackets returns a string in the format of `cf[12345]`.
// This is used in JQL queries.
func (cfid CustomFieldID) StringBrackets() string {
	return fmt.Sprintf("cf[%d]", cfid)
}

// StringPrefix returns a string in the format of `customfield_12345`.
// This is used in API responses.
func (cfid CustomFieldID) StringPrefix() string {
	return fmt.Sprintf("customfield_%d", cfid)
}

// CustomFieldKeyCanonical converts a custom field string to `customfield_12345`.
func CustomFieldKeyCanonical(key string) (string, error) {
	key = strings.ToLower(strings.TrimSpace(key))
	if rxCustomFieldPrefix.MatchString(key) {
		return key, nil
	} else if rxCustomFieldDigits.MatchString(key) {
		return customfieldPrefix + key, nil
	} else if m := rxCustomFieldBrackets.FindAllStringSubmatch(key, -1); len(m) > 0 {
		n := m[0]
		if len(n) > 1 {
			return customfieldPrefix + n[1], nil
		}
	}
	return "", ErrInvalidCustomFieldFormat
}

func IsCustomFieldKey(key string) (string, bool) {
	if can, err := CustomFieldKeyCanonical(key); err != nil {
		return key, false
	} else {
		return can, true
	}
}
