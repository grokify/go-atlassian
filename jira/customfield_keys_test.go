package jira

import (
	"testing"
)

const customFieldTestValue12345 = "customfield_12345"

var customFieldKeyCanonicalTests = []struct {
	v    string
	want string
}{
	{"customfield_12345", customFieldTestValue12345},
	{"CustomField_12345", customFieldTestValue12345},
	{"cf[12345]", customFieldTestValue12345},
	{"CF[12345]", customFieldTestValue12345},
	{"12345", customFieldTestValue12345},
	{"  12345  ", customFieldTestValue12345},
}

func TestCustomFieldKeyCanonical(t *testing.T) {
	for _, tt := range customFieldKeyCanonicalTests {
		try, err := CustomFieldKeyCanonical(tt.v)
		if err != nil {
			t.Errorf("rest.CustomFieldKeyCanonical(\"%s\"): want (%s), error (%s)", tt.v, tt.want, err.Error())
		}
		if try != tt.want {
			t.Errorf("rest.CustomFieldKeyCanonical(\"%s\"): want (%s), got (%s)", tt.v, tt.want, try)
		}
	}
}
