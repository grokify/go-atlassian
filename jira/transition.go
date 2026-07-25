package jira

import (
	"fmt"
	"sort"

	gojira "github.com/andygrunwald/go-jira"
)

type TransitionsFieldsMap map[string]TransitionField

func (m TransitionsFieldsMap) GetByName(fieldName string) (TransitionField, error) {
	for _, v := range m {
		if v.Name == fieldName {
			return v, nil
		}
	}
	return TransitionField{}, fmt.Errorf("field name not found (%s)", fieldName)
}

type Transition struct {
	ID     string               `json:"id" structs:"id"`
	Name   string               `json:"name" structs:"name"`
	To     gojira.Status        `json:"to" structs:"status"`
	Fields TransitionsFieldsMap `json:"fields" structs:"fields"`
}

type TransitionField struct {
	Name          string                        `json:"name"`
	Key           string                        `json:"key"` // e.g. "custom_12345"
	Required      bool                          `json:"required"`
	Schema        TransitionFieldSchema         `json:"schema"`
	Operations    []string                      `json:"operations"`
	AllowedValues []TransitionFieldAllowedValue `json:"allowedValues"`
}

type TransitionFieldSchema struct {
	Type     string `json:"type"`
	Custom   string `json:"custom"`
	CustomID int    `json:"customId" `
}

type TransitionFieldAllowedValue struct {
	Self  string `json:"self"`
	Value string `json:"value"`
	ID    string `json:"id"`
}

// RequiredFields returns the list of reuqired fields, but it may not be complete
// from analyzing the API response on update.
func (txn Transition) RequiredFields() []string {
	var out []string
	for fieldName, fieldInfo := range txn.Fields {
		if fieldInfo.Required {
			out = append(out, fieldName)
		}
	}
	sort.Strings(out)
	return out
}

type Transitions []Transition

func (txns Transitions) AddTransitionsSDK(txnsSDK []gojira.Transition) {
	for _, txnSDK := range txnsSDK {
		newTxn := Transition{
			ID:     txnSDK.ID,
			Name:   txnSDK.Name,
			To:     txnSDK.To,
			Fields: map[string]TransitionField{},
		}
		for k, v := range txnSDK.Fields {
			newTxn.Fields[k] = TransitionField{Required: v.Required}
		}
		txns = append(txns, newTxn)
	}
}

func (txns Transitions) GetByName(name string) (Transition, error) {
	for _, txn := range txns {
		if txn.Name == name {
			return txn, nil
		}
	}
	return Transition{}, fmt.Errorf("transition name not found (%s)", name)
}

func (txns Transitions) MapNameToID() map[string]string {
	out := map[string]string{}
	for _, txn := range txns {
		if _, ok := out[txn.Name]; ok {
			panic("txn name is not unique")
		}
		out[txn.Name] = txn.ID
	}
	return out
}

func (txns Transitions) Names() []string {
	var out []string
	for _, txn := range txns {
		out = append(out, txn.Name)
	}
	sort.Strings(out)
	return out
}
