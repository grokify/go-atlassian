package export

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grokify/go-atlassian/report"
)

func TestHTMLExporter(t *testing.T) {
	dashboard := &report.DashboardIR{
		ID:          "test-dashboard",
		Title:       "Test Dashboard",
		Description: "A test dashboard for unit tests",
		Layout:      report.LayoutIR{Columns: 12, Responsive: true},
		Widgets: []report.WidgetIR{
			{
				ID:    "welcome",
				Type:  "text",
				Title: "Welcome",
				Position: &report.PositionIR{
					X: 0, Y: 0, W: 12, H: 2,
				},
				Config: json.RawMessage(`{"content":"# Hello World","format":"markdown"}`),
			},
			{
				ID:    "metric",
				Type:  "metric",
				Title: "Total Issues",
				Position: &report.PositionIR{
					X: 0, Y: 2, W: 4, H: 2,
				},
				Config: json.RawMessage(`{"valueField":"value","format":"number"}`),
			},
		},
		DataSources: []report.DataSourceIR{},
	}

	exporter := NewHTMLExporter()
	html, err := exporter.Export(dashboard)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	htmlStr := string(html)

	// Check that basic elements are present
	if !strings.Contains(htmlStr, "Test Dashboard") {
		t.Error("HTML does not contain title")
	}
	if !strings.Contains(htmlStr, "Welcome") {
		t.Error("HTML does not contain widget title")
	}
	if !strings.Contains(htmlStr, "dashboardData") {
		t.Error("HTML does not contain embedded dashboard data")
	}
}

func TestHTMLExporterDarkMode(t *testing.T) {
	dashboard := &report.DashboardIR{
		ID:    "dark-dashboard",
		Title: "Dark Dashboard",
		Theme: &report.ThemeIR{
			Mode: "dark",
		},
		Layout: report.LayoutIR{Columns: 12},
		Widgets: []report.WidgetIR{
			{
				ID:   "text",
				Type: "text",
				Position: &report.PositionIR{
					X: 0, Y: 0, W: 12, H: 2,
				},
				Config: json.RawMessage(`{"content":"Dark mode test"}`),
			},
		},
	}

	exporter := NewHTMLExporter()
	html, err := exporter.Export(dashboard)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	htmlStr := string(html)

	// Dark mode should include different CSS variables
	if !strings.Contains(htmlStr, "--bg-color: #1a1a2e") {
		t.Error("HTML does not contain dark mode styles")
	}
}

func TestHTMLExporterNoEmbedData(t *testing.T) {
	dashboard := &report.DashboardIR{
		ID:     "simple-dashboard",
		Title:  "Simple",
		Layout: report.LayoutIR{Columns: 12},
		Widgets: []report.WidgetIR{
			{
				ID:   "text",
				Type: "text",
				Position: &report.PositionIR{
					X: 0, Y: 0, W: 12, H: 2,
				},
				Config: json.RawMessage(`{"content":"Hello"}`),
			},
		},
	}

	exporter := NewHTMLExporter()
	exporter.EmbedData = false

	html, err := exporter.Export(dashboard)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	htmlStr := string(html)

	// Should not contain embedded data
	if strings.Contains(htmlStr, "dashboardData") {
		t.Error("HTML should not contain embedded dashboard data when EmbedData is false")
	}
}
