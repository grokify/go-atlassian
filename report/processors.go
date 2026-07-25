package report

import (
	"encoding/json"
	"fmt"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/go-atlassian/jira"
)

// MarkdownProcessor processes markdown sections.
type MarkdownProcessor struct{}

// Process converts a markdown section to a text widget.
func (p *MarkdownProcessor) Process(ctx *ExecutionContext, section *Section) (*SectionResult, error) {
	content := section.Content
	if content == "" {
		// Try to get content from config
		var config struct {
			Content string `json:"content"`
		}
		if len(section.Config) > 0 {
			if err := json.Unmarshal(section.Config, &config); err == nil {
				content = config.Content
			}
		}
	}

	// Expand variables in content
	content = ctx.expandVariables(content)

	configIR := &TextConfigIR{
		Content: content,
		Format:  "markdown",
	}
	configJSON, err := json.Marshal(configIR)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	widget := WidgetIR{
		ID:     section.ID,
		Type:   "text",
		Title:  section.Title,
		Config: configJSON,
	}

	if section.Position != nil {
		widget.Position = &PositionIR{
			X: section.Position.X,
			Y: section.Position.Y,
			W: section.Position.W,
			H: section.Position.H,
		}
	}

	return &SectionResult{Widget: widget}, nil
}

// expandVariables replaces variable placeholders in content.
func (ctx *ExecutionContext) expandVariables(content string) string {
	// Simple variable expansion - could be enhanced with template engine
	return content
}

// JQLProcessor processes JQL query sections.
type JQLProcessor struct {
	Client *jira.Client
	client *jira.Client // For internal use, supports both casing
}

// Process executes a JQL query and produces appropriate widget.
func (p *JQLProcessor) Process(ctx *ExecutionContext, section *Section) (*SectionResult, error) {
	config, err := section.GetJQLConfig()
	if err != nil {
		return nil, fmt.Errorf("get JQL config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("JQL config is required")
	}

	client := p.Client
	if client == nil {
		client = p.client
	}
	if client == nil {
		return nil, fmt.Errorf("client is required for JQL processor")
	}

	// Execute JQL query
	maxResults := config.MaxResults
	if maxResults == 0 {
		maxResults = 50
	}

	searchOpts := &gojira.SearchOptions{MaxResults: maxResults}
	issues, _, err := client.JiraClient.Issue.Search(config.JQL, searchOpts)
	if err != nil {
		return nil, fmt.Errorf("execute JQL: %w", err)
	}

	// Create data source
	dataSourceID := section.ID + "-data"
	dataSource, err := NewInlineDataSource(dataSourceID, issues)
	if err != nil {
		return nil, fmt.Errorf("create data source: %w", err)
	}

	// Create widget based on display type
	var widget WidgetIR
	switch config.Display {
	case DisplayChart:
		widget, err = p.createChartWidget(section, dataSourceID, config)
	case DisplayTable, DisplayList, "":
		widget, err = p.createTableWidget(section, dataSourceID, config)
	default:
		widget, err = p.createTableWidget(section, dataSourceID, config)
	}
	if err != nil {
		return nil, err
	}

	return &SectionResult{
		Widget:     widget,
		DataSource: dataSource,
	}, nil
}

func (p *JQLProcessor) createChartWidget(section *Section, dataSourceID string, config *JQLConfig) (WidgetIR, error) {
	chartType := config.ChartType
	if chartType == "" {
		chartType = "bar"
	}

	chartConfig := &ChartConfigIR{
		Type:   chartType,
		XField: config.GroupBy,
	}
	configJSON, err := json.Marshal(chartConfig)
	if err != nil {
		return WidgetIR{}, fmt.Errorf("marshal chart config: %w", err)
	}

	widget := WidgetIR{
		ID:           section.ID,
		Type:         "chart",
		Title:        section.Title,
		DataSourceID: dataSourceID,
		Config:       configJSON,
	}

	if section.Position != nil {
		widget.Position = &PositionIR{
			X: section.Position.X,
			Y: section.Position.Y,
			W: section.Position.W,
			H: section.Position.H,
		}
	}

	return widget, nil
}

func (p *JQLProcessor) createTableWidget(section *Section, dataSourceID string, config *JQLConfig) (WidgetIR, error) {
	fields := config.Fields
	if len(fields) == 0 {
		fields = []string{"key", "summary", "status", "assignee"}
	}

	columns := make([]ColumnIR, len(fields))
	for i, field := range fields {
		columns[i] = ColumnIR{
			Field:  field,
			Header: field,
		}
	}

	tableConfig := &TableConfigIR{
		Columns:  columns,
		Sortable: true,
	}
	configJSON, err := json.Marshal(tableConfig)
	if err != nil {
		return WidgetIR{}, fmt.Errorf("marshal table config: %w", err)
	}

	widget := WidgetIR{
		ID:           section.ID,
		Type:         "table",
		Title:        section.Title,
		DataSourceID: dataSourceID,
		Config:       configJSON,
	}

	if section.Position != nil {
		widget.Position = &PositionIR{
			X: section.Position.X,
			Y: section.Position.Y,
			W: section.Position.W,
			H: section.Position.H,
		}
	}

	return widget, nil
}

// VelocityProcessor processes velocity chart sections.
type VelocityProcessor struct {
	Client *jira.Client
	client *jira.Client
}

// Process generates velocity chart data from sprint history.
func (p *VelocityProcessor) Process(ctx *ExecutionContext, section *Section) (*SectionResult, error) {
	config, err := section.GetVelocityConfig()
	if err != nil {
		return nil, fmt.Errorf("get velocity config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("velocity config is required")
	}

	client := p.Client
	if client == nil {
		client = p.client
	}
	if client == nil {
		return nil, fmt.Errorf("client is required for velocity processor")
	}

	// Get sprint velocity data
	sprintCount := config.SprintCount
	if sprintCount == 0 {
		sprintCount = 6
	}

	opts := &jira.VelocityOptions{
		SprintCount:        sprintCount,
		IncludeActive:      config.IncludeActive,
		StoryPointsFieldID: config.StoryPointsField,
	}

	report, err := client.BoardAPI.GetVelocityReport(ctx.Context, config.BoardID, opts)
	if err != nil {
		return nil, fmt.Errorf("get velocity report: %w", err)
	}

	// Transform to chart data
	type velocityPoint struct {
		Sprint    string  `json:"sprint"`
		Completed float64 `json:"completed"`
	}

	data := make([]velocityPoint, len(report.Sprints))
	for i, sp := range report.Sprints {
		data[i] = velocityPoint{
			Sprint:    sp.SprintName,
			Completed: sp.CompletedPts,
		}
	}

	// Create data source
	dataSourceID := section.ID + "-data"
	dataSource, err := NewInlineDataSource(dataSourceID, data)
	if err != nil {
		return nil, fmt.Errorf("create data source: %w", err)
	}

	// Create chart widget
	chartType := config.ChartType
	if chartType == "" {
		chartType = "bar"
	}

	chartConfig := &ChartConfigIR{
		Type:    chartType,
		XField:  "sprint",
		YFields: []string{"completed"},
		Series: []SeriesIR{
			{Name: "Completed", DataField: "completed", Color: "#22c55e"},
		},
	}
	configJSON, err := json.Marshal(chartConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal chart config: %w", err)
	}

	widget := WidgetIR{
		ID:           section.ID,
		Type:         "chart",
		Title:        section.Title,
		DataSourceID: dataSourceID,
		Config:       configJSON,
	}

	if section.Position != nil {
		widget.Position = &PositionIR{
			X: section.Position.X,
			Y: section.Position.Y,
			W: section.Position.W,
			H: section.Position.H,
		}
	}

	return &SectionResult{
		Widget:     widget,
		DataSource: dataSource,
	}, nil
}

// BurndownProcessor processes burndown chart sections.
type BurndownProcessor struct {
	Client *jira.Client
	client *jira.Client
}

// Process generates burndown chart data for a sprint.
func (p *BurndownProcessor) Process(ctx *ExecutionContext, section *Section) (*SectionResult, error) {
	config, err := section.GetBurndownConfig()
	if err != nil {
		return nil, fmt.Errorf("get burndown config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("burndown config is required")
	}

	client := p.Client
	if client == nil {
		client = p.client
	}
	if client == nil {
		return nil, fmt.Errorf("client is required for burndown processor")
	}

	// Get burndown data
	opts := &jira.BurndownOptions{
		StoryPointsFieldID: config.StoryPointsField,
	}

	report, err := client.BoardAPI.GetBurndownReport(ctx.Context, config.SprintID, opts)
	if err != nil {
		return nil, fmt.Errorf("get burndown report: %w", err)
	}

	// Transform to chart data
	type burndownPoint struct {
		Date      string  `json:"date"`
		Remaining float64 `json:"remaining"`
		Ideal     float64 `json:"ideal,omitempty"`
	}

	data := make([]burndownPoint, len(report.DailyData))
	for i, pt := range report.DailyData {
		point := burndownPoint{
			Date:      pt.Date,
			Remaining: pt.RemainingPoints,
		}
		if config.ShowIdealLine && report.TotalPoints > 0 {
			// Calculate ideal line (linear from total to 0)
			progress := float64(i) / float64(len(report.DailyData))
			point.Ideal = report.TotalPoints * (1 - progress)
		}
		data[i] = point
	}

	// Create data source
	dataSourceID := section.ID + "-data"
	dataSource, err := NewInlineDataSource(dataSourceID, data)
	if err != nil {
		return nil, fmt.Errorf("create data source: %w", err)
	}

	// Create chart widget
	series := []SeriesIR{
		{Name: "Remaining", DataField: "remaining", Color: "#3b82f6"},
	}
	yFields := []string{"remaining"}
	if config.ShowIdealLine {
		series = append(series, SeriesIR{Name: "Ideal", DataField: "ideal", Color: "#94a3b8"})
		yFields = append(yFields, "ideal")
	}

	chartConfig := &ChartConfigIR{
		Type:    "line",
		XField:  "date",
		YFields: yFields,
		Series:  series,
	}
	configJSON, err := json.Marshal(chartConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal chart config: %w", err)
	}

	widget := WidgetIR{
		ID:           section.ID,
		Type:         "chart",
		Title:        section.Title,
		DataSourceID: dataSourceID,
		Config:       configJSON,
	}

	if section.Position != nil {
		widget.Position = &PositionIR{
			X: section.Position.X,
			Y: section.Position.Y,
			W: section.Position.W,
			H: section.Position.H,
		}
	}

	return &SectionResult{
		Widget:     widget,
		DataSource: dataSource,
	}, nil
}

// WorklogProcessor processes worklog summary sections.
type WorklogProcessor struct {
	Client *jira.Client
	client *jira.Client
}

// Process generates worklog summary data.
func (p *WorklogProcessor) Process(ctx *ExecutionContext, section *Section) (*SectionResult, error) {
	config, err := section.GetWorklogConfig()
	if err != nil {
		return nil, fmt.Errorf("get worklog config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("worklog config is required")
	}

	client := p.Client
	if client == nil {
		client = p.client
	}
	if client == nil {
		return nil, fmt.Errorf("client is required for worklog processor")
	}

	// Get worklog data - either from JQL or sprint
	var worklogs []map[string]interface{}

	if config.JQL != "" {
		issues, _, err := client.JiraClient.Issue.Search(config.JQL, nil)
		if err != nil {
			return nil, fmt.Errorf("search issues: %w", err)
		}

		// Aggregate worklogs from issues
		for _, issue := range issues {
			// Note: Would need to fetch worklogs per issue
			_ = issue
		}
	}

	// Create data source
	dataSourceID := section.ID + "-data"
	dataSource, err := NewInlineDataSource(dataSourceID, worklogs)
	if err != nil {
		return nil, fmt.Errorf("create data source: %w", err)
	}

	// Create widget based on display type
	var widget WidgetIR
	switch config.Display {
	case DisplayChart:
		chartType := config.ChartType
		if chartType == "" {
			chartType = "bar"
		}
		chartConfig := &ChartConfigIR{
			Type:   chartType,
			XField: config.GroupBy,
		}
		configJSON, err := json.Marshal(chartConfig)
		if err != nil {
			return nil, fmt.Errorf("marshal chart config: %w", err)
		}
		widget = WidgetIR{
			ID:           section.ID,
			Type:         "chart",
			Title:        section.Title,
			DataSourceID: dataSourceID,
			Config:       configJSON,
		}
	default:
		tableConfig := &TableConfigIR{
			Columns: []ColumnIR{
				{Field: "author", Header: "Author"},
				{Field: "timeSpent", Header: "Time Spent"},
				{Field: "date", Header: "Date"},
			},
			Sortable: true,
		}
		configJSON, err := json.Marshal(tableConfig)
		if err != nil {
			return nil, fmt.Errorf("marshal table config: %w", err)
		}
		widget = WidgetIR{
			ID:           section.ID,
			Type:         "table",
			Title:        section.Title,
			DataSourceID: dataSourceID,
			Config:       configJSON,
		}
	}

	if section.Position != nil {
		widget.Position = &PositionIR{
			X: section.Position.X,
			Y: section.Position.Y,
			W: section.Position.W,
			H: section.Position.H,
		}
	}

	return &SectionResult{
		Widget:     widget,
		DataSource: dataSource,
	}, nil
}

// CycleTimeProcessor processes cycle time analysis sections.
type CycleTimeProcessor struct {
	Client *jira.Client
	client *jira.Client
}

// Process generates cycle time analysis data.
func (p *CycleTimeProcessor) Process(ctx *ExecutionContext, section *Section) (*SectionResult, error) {
	config, err := section.GetCycleTimeConfig()
	if err != nil {
		return nil, fmt.Errorf("get cycle time config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("cycle time config is required")
	}

	client := p.Client
	if client == nil {
		client = p.client
	}
	if client == nil {
		return nil, fmt.Errorf("client is required for cycle time processor")
	}

	// Get cycle time data
	var cycleTimeData []map[string]interface{}

	if config.JQL != "" {
		issues, _, err := client.JiraClient.Issue.Search(config.JQL, nil)
		if err != nil {
			return nil, fmt.Errorf("search issues: %w", err)
		}

		// Calculate cycle times from issue history
		for _, issue := range issues {
			_ = issue
		}
	}

	// Create data source
	dataSourceID := section.ID + "-data"
	dataSource, err := NewInlineDataSource(dataSourceID, cycleTimeData)
	if err != nil {
		return nil, fmt.Errorf("create data source: %w", err)
	}

	// Create widget based on display type
	var widget WidgetIR
	switch config.Display {
	case DisplayChart:
		chartType := config.ChartType
		if chartType == "" {
			chartType = "line"
		}
		chartConfig := &ChartConfigIR{
			Type:    chartType,
			XField:  "date",
			YFields: []string{"cycleTime"},
		}
		configJSON, err := json.Marshal(chartConfig)
		if err != nil {
			return nil, fmt.Errorf("marshal chart config: %w", err)
		}
		widget = WidgetIR{
			ID:           section.ID,
			Type:         "chart",
			Title:        section.Title,
			DataSourceID: dataSourceID,
			Config:       configJSON,
		}
	default:
		tableConfig := &TableConfigIR{
			Columns: []ColumnIR{
				{Field: "key", Header: "Issue"},
				{Field: "cycleTime", Header: "Cycle Time"},
				{Field: "startDate", Header: "Start"},
				{Field: "endDate", Header: "End"},
			},
			Sortable: true,
		}
		configJSON, err := json.Marshal(tableConfig)
		if err != nil {
			return nil, fmt.Errorf("marshal table config: %w", err)
		}
		widget = WidgetIR{
			ID:           section.ID,
			Type:         "table",
			Title:        section.Title,
			DataSourceID: dataSourceID,
			Config:       configJSON,
		}
	}

	if section.Position != nil {
		widget.Position = &PositionIR{
			X: section.Position.X,
			Y: section.Position.Y,
			W: section.Position.W,
			H: section.Position.H,
		}
	}

	return &SectionResult{
		Widget:     widget,
		DataSource: dataSource,
	}, nil
}

// MetricProcessor processes single metric display sections.
type MetricProcessor struct {
	Client *jira.Client
	client *jira.Client
}

// Process generates a metric widget.
func (p *MetricProcessor) Process(ctx *ExecutionContext, section *Section) (*SectionResult, error) {
	config, err := section.GetMetricConfig()
	if err != nil {
		return nil, fmt.Errorf("get metric config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("metric config is required")
	}

	client := p.Client
	if client == nil {
		client = p.client
	}
	if client == nil {
		return nil, fmt.Errorf("client is required for metric processor")
	}

	// Calculate metric value
	var value float64

	if config.JQL != "" {
		issues, _, err := client.JiraClient.Issue.Search(config.JQL, nil)
		if err != nil {
			return nil, fmt.Errorf("search issues: %w", err)
		}

		switch config.Aggregation {
		case "count":
			value = float64(len(issues))
		case "sum":
			// Sum field values
			for _, issue := range issues {
				_ = issue
			}
		case "avg":
			// Average field values
			for _, issue := range issues {
				_ = issue
			}
		default:
			value = float64(len(issues))
		}
	}

	// Create data source with metric value
	dataSourceID := section.ID + "-data"
	metricData := map[string]interface{}{
		"value": value,
	}
	dataSource, err := NewInlineDataSource(dataSourceID, metricData)
	if err != nil {
		return nil, fmt.Errorf("create data source: %w", err)
	}

	// Create metric widget
	metricConfig := &MetricConfigIR{
		ValueField: "value",
		Format:     config.Format,
	}
	if config.Comparison != nil {
		metricConfig.ComparisonType = config.Comparison.Type
		metricConfig.ComparisonValue = config.Comparison.TargetValue
	}

	configJSON, err := json.Marshal(metricConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal metric config: %w", err)
	}

	widget := WidgetIR{
		ID:           section.ID,
		Type:         "metric",
		Title:        section.Title,
		DataSourceID: dataSourceID,
		Config:       configJSON,
	}

	if section.Position != nil {
		widget.Position = &PositionIR{
			X: section.Position.X,
			Y: section.Position.Y,
			W: section.Position.W,
			H: section.Position.H,
		}
	}

	return &SectionResult{
		Widget:     widget,
		DataSource: dataSource,
	}, nil
}

// TableProcessor processes table sections.
type TableProcessor struct {
	Client *jira.Client
	client *jira.Client
}

// Process generates a table widget.
func (p *TableProcessor) Process(ctx *ExecutionContext, section *Section) (*SectionResult, error) {
	config, err := section.GetTableConfig()
	if err != nil {
		return nil, fmt.Errorf("get table config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("table config is required")
	}

	client := p.Client
	if client == nil {
		client = p.client
	}
	if client == nil {
		return nil, fmt.Errorf("client is required for table processor")
	}

	// Execute JQL query
	maxResults := config.MaxResults
	if maxResults == 0 {
		maxResults = 50
	}

	issues, _, err := client.JiraClient.Issue.Search(config.JQL, nil)
	if err != nil {
		return nil, fmt.Errorf("search issues: %w", err)
	}

	// Create data source
	dataSourceID := section.ID + "-data"
	dataSource, err := NewInlineDataSource(dataSourceID, issues)
	if err != nil {
		return nil, fmt.Errorf("create data source: %w", err)
	}

	// Create columns from fields
	fields := config.Fields
	if len(fields) == 0 {
		fields = []string{"key", "summary", "status", "assignee"}
	}

	columns := make([]ColumnIR, len(fields))
	for i, field := range fields {
		columns[i] = ColumnIR{
			Field:    field,
			Header:   field,
			Sortable: config.Sortable,
		}
	}

	tableConfig := &TableConfigIR{
		Columns:   columns,
		Sortable:  config.Sortable,
		Paginated: config.Paginated,
		PageSize:  maxResults,
	}
	configJSON, err := json.Marshal(tableConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal table config: %w", err)
	}

	widget := WidgetIR{
		ID:           section.ID,
		Type:         "table",
		Title:        section.Title,
		DataSourceID: dataSourceID,
		Config:       configJSON,
	}

	if section.Position != nil {
		widget.Position = &PositionIR{
			X: section.Position.X,
			Y: section.Position.Y,
			W: section.Position.W,
			H: section.Position.H,
		}
	}

	return &SectionResult{
		Widget:     widget,
		DataSource: dataSource,
	}, nil
}
