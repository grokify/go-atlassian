package report

import "encoding/json"

// DashboardIR represents a Dashforge Dashboard.
// This matches the Dashforge schema for compatibility.
type DashboardIR struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Version     string         `json:"version,omitempty"`
	Layout      LayoutIR       `json:"layout"`
	DataSources []DataSourceIR `json:"dataSources"`
	Widgets     []WidgetIR     `json:"widgets"`
	Variables   []VariableIR   `json:"variables,omitempty"`
	Theme       *ThemeIR       `json:"theme,omitempty"`
}

// LayoutIR defines the dashboard grid layout.
type LayoutIR struct {
	Columns    int  `json:"columns"`
	RowHeight  int  `json:"rowHeight,omitempty"`
	Responsive bool `json:"responsive,omitempty"`
}

// DataSourceIR defines a data source in the dashboard.
type DataSourceIR struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"` // "inline", "url", "derived"
	Data      json.RawMessage `json:"data,omitempty"`
	URL       string          `json:"url,omitempty"`
	Transform []TransformIR   `json:"transform,omitempty"`
}

// TransformIR defines a data transformation.
type TransformIR struct {
	Type   string          `json:"type"` // "filter", "aggregate", "sort", etc.
	Config json.RawMessage `json:"config"`
}

// WidgetIR defines a widget in the dashboard.
type WidgetIR struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"` // "chart", "table", "metric", "text"
	Title        string          `json:"title,omitempty"`
	Position     *PositionIR     `json:"position"`
	DataSourceID string          `json:"dataSourceId,omitempty"`
	Config       json.RawMessage `json:"config"`
}

// PositionIR defines the grid position of a widget.
type PositionIR struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// VariableIR defines a variable in the dashboard.
type VariableIR struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"`
	Label   string         `json:"label"`
	Default any            `json:"default,omitempty"`
	Options []SelectOption `json:"options,omitempty"`
}

// ThemeIR defines the dashboard theme.
type ThemeIR struct {
	Mode   string         `json:"mode,omitempty"` // "light", "dark"
	Colors *ThemeColorsIR `json:"colors,omitempty"`
}

// ThemeColorsIR defines the theme color palette.
type ThemeColorsIR struct {
	Primary   string `json:"primary,omitempty"`
	Secondary string `json:"secondary,omitempty"`
	Success   string `json:"success,omitempty"`
	Warning   string `json:"warning,omitempty"`
	Danger    string `json:"danger,omitempty"`
}

// ChartConfigIR defines chart configuration.
type ChartConfigIR struct {
	Type    string          `json:"type"` // "bar", "line", "pie", "area"
	XField  string          `json:"xField,omitempty"`
	YFields []string        `json:"yFields,omitempty"`
	Series  []SeriesIR      `json:"series,omitempty"`
	Options json.RawMessage `json:"options,omitempty"`
}

// SeriesIR defines a data series in a chart.
type SeriesIR struct {
	Name      string `json:"name"`
	DataField string `json:"dataField"`
	Color     string `json:"color,omitempty"`
}

// TableConfigIR defines table configuration.
type TableConfigIR struct {
	Columns   []ColumnIR `json:"columns"`
	Sortable  bool       `json:"sortable,omitempty"`
	Paginated bool       `json:"paginated,omitempty"`
	PageSize  int        `json:"pageSize,omitempty"`
}

// ColumnIR defines a table column.
type ColumnIR struct {
	Field    string `json:"field"`
	Header   string `json:"header"`
	Width    string `json:"width,omitempty"`
	Sortable bool   `json:"sortable,omitempty"`
	Format   string `json:"format,omitempty"`
}

// MetricConfigIR defines metric widget configuration.
type MetricConfigIR struct {
	ValueField      string  `json:"valueField"`
	Format          string  `json:"format,omitempty"` // "number", "percent", "duration"
	Prefix          string  `json:"prefix,omitempty"`
	Suffix          string  `json:"suffix,omitempty"`
	ComparisonValue float64 `json:"comparisonValue,omitempty"`
	ComparisonType  string  `json:"comparisonType,omitempty"` // "increase", "decrease"
}

// TextConfigIR defines text widget configuration.
type TextConfigIR struct {
	Content string `json:"content"`
	Format  string `json:"format,omitempty"` // "markdown", "html", "plain"
}

// ToJSON converts the dashboard IR to JSON.
func (d *DashboardIR) ToJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}

// NewInlineDataSource creates an inline data source with the given data.
func NewInlineDataSource(id string, data any) (*DataSourceIR, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &DataSourceIR{
		ID:   id,
		Type: "inline",
		Data: jsonData,
	}, nil
}

// NewChartWidget creates a chart widget.
func NewChartWidget(id, title string, pos *PositionIR, dataSourceID string, config *ChartConfigIR) (*WidgetIR, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &WidgetIR{
		ID:           id,
		Type:         "chart",
		Title:        title,
		Position:     pos,
		DataSourceID: dataSourceID,
		Config:       configJSON,
	}, nil
}

// NewTableWidget creates a table widget.
func NewTableWidget(id, title string, pos *PositionIR, dataSourceID string, config *TableConfigIR) (*WidgetIR, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &WidgetIR{
		ID:           id,
		Type:         "table",
		Title:        title,
		Position:     pos,
		DataSourceID: dataSourceID,
		Config:       configJSON,
	}, nil
}

// NewMetricWidget creates a metric widget.
func NewMetricWidget(id, title string, pos *PositionIR, dataSourceID string, config *MetricConfigIR) (*WidgetIR, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &WidgetIR{
		ID:           id,
		Type:         "metric",
		Title:        title,
		Position:     pos,
		DataSourceID: dataSourceID,
		Config:       configJSON,
	}, nil
}

// NewTextWidget creates a text widget.
func NewTextWidget(id, title string, pos *PositionIR, config *TextConfigIR) (*WidgetIR, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &WidgetIR{
		ID:       id,
		Type:     "text",
		Title:    title,
		Position: pos,
		Config:   configJSON,
	}, nil
}
