package jira

import (
	"errors"
	"strconv"
	"strings"

	gojira "github.com/andygrunwald/go-jira"
)

type EpicsSet struct {
	EpicsMap map[string]gojira.Epic
}

func NewEpicsSet() EpicsSet {
	return EpicsSet{EpicsMap: map[string]gojira.Epic{}}
}

func (es *EpicsSet) GetKeys(jclient *gojira.Client, epicKeys []string) error {
	if jclient == nil {
		return errors.New("jclient cannot be nil")
	}
	c := Client{JiraClient: jclient}
	newEpics, err := c.IssueAPI.GetIssuesSetForKeys(epicKeys)
	if err != nil {
		return err
	}
	err = es.AddIssues(newEpics.Issues())
	return err
}

func (es *EpicsSet) AddIssues(issues []gojira.Issue) error {
	if es.EpicsMap == nil {
		es.EpicsMap = map[string]gojira.Epic{}
	}
	for _, iss := range issues {
		epic, err := IssueToEpic(iss)
		if err != nil {
			return err
		}
		es.EpicsMap[epic.Key] = *epic
	}
	return nil
}

func IssueToEpic(iss gojira.Issue) (*gojira.Epic, error) {
	idInt, err := strconv.Atoi(iss.ID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(iss.Key) == "" {
		return nil, errors.New("key is empty")
	}
	epic := &gojira.Epic{
		ID:  idInt,
		Key: iss.Key,
	}
	if iss.Fields != nil {
		epic.Summary = iss.Fields.Summary
	}
	return epic, nil
}
