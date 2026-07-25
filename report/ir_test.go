package report

import (
	"encoding/json"
	"testing"
)

func TestDashboardIRToJSON(t *testing.T) {
	dashboard := &DashboardIR{
		ID:          "test-dashboard",
		Title:       "Test Dashboard",
		Description: "A test dashboard",
		Layout:      LayoutIR{Columns: 12, Responsive: true},
		Widgets: []WidgetIR{
			{
				ID:    "widget1",
				Type:  "text",
				Title: "Welcome",
				Position: &PositionIR{
					X: 0,
					Y: 0,
					W: 12,
					H: 2,
				},
				Config: json.RawMessage(`{"content":"Hello World","format":"markdown"}`),
			},
		},
		DataSources: []DataSourceIR{},
	}

	data, err := dashboard.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Parse back and verify
	var parsed DashboardIR
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if parsed.ID != dashboard.ID {
		t.Errorf("ID = %q, want %q", parsed.ID, dashboard.ID)
	}
	if parsed.Title != dashboard.Title {
		t.Errorf("Title = %q, want %q", parsed.Title, dashboard.Title)
	}
	if len(parsed.Widgets) != 1 {
		t.Errorf("len(Widgets) = %d, want 1", len(parsed.Widgets))
	}
}

func TestNewInlineDataSource(t *testing.T) {
	data := []map[string]string{
		{"key": "TEST-1", "summary": "Test issue"},
		{"key": "TEST-2", "summary": "Another issue"},
	}

	ds, err := NewInlineDataSource("test-ds", data)
	if err != nil {
		t.Fatalf("NewInlineDataSource() error = %v", err)
	}

	if ds.ID != "test-ds" {
		t.Errorf("ID = %q, want %q", ds.ID, "test-ds")
	}
	if ds.Type != "inline" {
		t.Errorf("Type = %q, want %q", ds.Type, "inline")
	}
	if len(ds.Data) == 0 {
		t.Error("Data is empty")
	}
}

func TestNewTextWidget(t *testing.T) {
	config := &TextConfigIR{
		Content: "# Hello World",
		Format:  "markdown",
	}
	pos := &PositionIR{X: 0, Y: 0, W: 12, H: 2}

	widget, err := NewTextWidget("text-widget", "Welcome", pos, config)
	if err != nil {
		t.Fatalf("NewTextWidget() error = %v", err)
	}

	if widget.ID != "text-widget" {
		t.Errorf("ID = %q, want %q", widget.ID, "text-widget")
	}
	if widget.Type != "text" {
		t.Errorf("Type = %q, want %q", widget.Type, "text")
	}
	if widget.Title != "Welcome" {
		t.Errorf("Title = %q, want %q", widget.Title, "Welcome")
	}
}

func TestNewChartWidget(t *testing.T) {
	config := &ChartConfigIR{
		Type:    "bar",
		XField:  "sprint",
		YFields: []string{"completed"},
	}
	pos := &PositionIR{X: 0, Y: 0, W: 6, H: 4}

	widget, err := NewChartWidget("chart-widget", "Velocity", pos, "velocity-data", config)
	if err != nil {
		t.Fatalf("NewChartWidget() error = %v", err)
	}

	if widget.ID != "chart-widget" {
		t.Errorf("ID = %q, want %q", widget.ID, "chart-widget")
	}
	if widget.Type != "chart" {
		t.Errorf("Type = %q, want %q", widget.Type, "chart")
	}
	if widget.DataSourceID != "velocity-data" {
		t.Errorf("DataSourceID = %q, want %q", widget.DataSourceID, "velocity-data")
	}
}

func TestNewTableWidget(t *testing.T) {
	config := &TableConfigIR{
		Columns: []ColumnIR{
			{Field: "key", Header: "Key"},
			{Field: "summary", Header: "Summary"},
		},
		Sortable: true,
	}
	pos := &PositionIR{X: 0, Y: 0, W: 12, H: 6}

	widget, err := NewTableWidget("table-widget", "Issues", pos, "issues-data", config)
	if err != nil {
		t.Fatalf("NewTableWidget() error = %v", err)
	}

	if widget.Type != "table" {
		t.Errorf("Type = %q, want %q", widget.Type, "table")
	}
}

func TestNewMetricWidget(t *testing.T) {
	config := &MetricConfigIR{
		ValueField: "value",
		Format:     "number",
	}
	pos := &PositionIR{X: 0, Y: 0, W: 3, H: 2}

	widget, err := NewMetricWidget("metric-widget", "Total Issues", pos, "metric-data", config)
	if err != nil {
		t.Fatalf("NewMetricWidget() error = %v", err)
	}

	if widget.Type != "metric" {
		t.Errorf("Type = %q, want %q", widget.Type, "metric")
	}
}
