// Package sections provides section processors for report generation.
package sections

import (
	"github.com/grokify/go-atlassian/jira"
	"github.com/grokify/go-atlassian/report"
)

// Processor is an alias for report.SectionProcessor for convenience.
type Processor = report.SectionProcessor

// Context is an alias for report.ExecutionContext for convenience.
type Context = report.ExecutionContext

// Result is an alias for report.SectionResult for convenience.
type Result = report.SectionResult

// Section is an alias for report.Section for convenience.
type Section = report.Section

// NewMarkdownProcessor creates a markdown processor.
func NewMarkdownProcessor() *report.MarkdownProcessor {
	return &report.MarkdownProcessor{}
}

// NewJQLProcessor creates a JQL processor.
func NewJQLProcessor(client *jira.Client) *report.JQLProcessor {
	return &report.JQLProcessor{Client: client}
}

// NewVelocityProcessor creates a velocity processor.
func NewVelocityProcessor(client *jira.Client) *report.VelocityProcessor {
	return &report.VelocityProcessor{Client: client}
}

// NewBurndownProcessor creates a burndown processor.
func NewBurndownProcessor(client *jira.Client) *report.BurndownProcessor {
	return &report.BurndownProcessor{Client: client}
}

// NewWorklogProcessor creates a worklog processor.
func NewWorklogProcessor(client *jira.Client) *report.WorklogProcessor {
	return &report.WorklogProcessor{Client: client}
}

// NewCycleTimeProcessor creates a cycle time processor.
func NewCycleTimeProcessor(client *jira.Client) *report.CycleTimeProcessor {
	return &report.CycleTimeProcessor{Client: client}
}

// NewMetricProcessor creates a metric processor.
func NewMetricProcessor(client *jira.Client) *report.MetricProcessor {
	return &report.MetricProcessor{Client: client}
}

// NewTableProcessor creates a table processor.
func NewTableProcessor(client *jira.Client) *report.TableProcessor {
	return &report.TableProcessor{Client: client}
}
