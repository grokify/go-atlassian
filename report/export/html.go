// Package export provides report export functionality.
package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"

	"github.com/grokify/go-atlassian/report"
)

// HTMLExporter exports dashboards to HTML.
type HTMLExporter struct {
	// Template is an optional custom HTML template
	Template *template.Template
	// EmbedData embeds the dashboard data in the HTML when true
	EmbedData bool
}

// NewHTMLExporter creates a new HTML exporter with default settings.
func NewHTMLExporter() *HTMLExporter {
	return &HTMLExporter{
		EmbedData: true,
	}
}

// Export exports a dashboard to HTML.
func (e *HTMLExporter) Export(dashboard *report.DashboardIR) ([]byte, error) {
	tmpl := e.Template
	if tmpl == nil {
		var err error
		tmpl, err = template.New("dashboard").Parse(defaultHTMLTemplate)
		if err != nil {
			return nil, fmt.Errorf("parse template: %w", err)
		}
	}

	// Prepare data for template
	data := templateData{
		Title:       dashboard.Title,
		Description: dashboard.Description,
		DarkMode:    dashboard.Theme != nil && dashboard.Theme.Mode == "dark",
	}

	if e.EmbedData {
		jsonData, err := json.Marshal(dashboard)
		if err != nil {
			return nil, fmt.Errorf("marshal dashboard: %w", err)
		}
		data.DashboardJSON = template.JS(jsonData) //nolint:gosec // JSON from trusted dashboard data
	}

	// Render widgets
	for _, widget := range dashboard.Widgets {
		rendered := e.renderWidget(&widget)
		data.Widgets = append(data.Widgets, rendered)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportToFile exports a dashboard to an HTML file.
func (e *HTMLExporter) ExportToFile(dashboard *report.DashboardIR, path string) error {
	data, err := e.Export(dashboard)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// ExportToWriter exports a dashboard to a writer.
func (e *HTMLExporter) ExportToWriter(dashboard *report.DashboardIR, w io.Writer) error {
	data, err := e.Export(dashboard)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

type templateData struct {
	Title         string
	Description   string
	DarkMode      bool
	DashboardJSON template.JS
	Widgets       []renderedWidget
}

type renderedWidget struct {
	ID       string
	Type     string
	Title    string
	X        int
	Y        int
	W        int
	H        int
	Content  template.HTML
	DataJSON template.JS
}

func (e *HTMLExporter) renderWidget(widget *report.WidgetIR) renderedWidget {
	rendered := renderedWidget{
		ID:    widget.ID,
		Type:  widget.Type,
		Title: widget.Title,
	}

	if widget.Position != nil {
		rendered.X = widget.Position.X
		rendered.Y = widget.Position.Y
		rendered.W = widget.Position.W
		rendered.H = widget.Position.H
	}

	// Render content based on widget type
	switch widget.Type {
	case "text":
		var config report.TextConfigIR
		if err := json.Unmarshal(widget.Config, &config); err == nil {
			// For static HTML, render markdown as-is (would need markdown parser for full rendering)
			// Content is escaped via HTMLEscapeString, so this is safe
			rendered.Content = template.HTML("<div class=\"markdown\">" + template.HTMLEscapeString(config.Content) + "</div>") //nolint:gosec // HTML-escaped content
		}
	case "metric":
		var config report.MetricConfigIR
		if err := json.Unmarshal(widget.Config, &config); err == nil {
			// ValueField is a field name from config, not user input
			rendered.Content = template.HTML(fmt.Sprintf("<div class=\"metric-value\">%s</div>", template.HTMLEscapeString(config.ValueField))) //nolint:gosec // HTML-escaped content
		}
	case "table":
		// For tables and charts, embed config as JSON for client-side rendering
		rendered.DataJSON = template.JS(widget.Config) //nolint:gosec // JSON from trusted widget config
	case "chart":
		rendered.DataJSON = template.JS(widget.Config) //nolint:gosec // JSON from trusted widget config
	}

	return rendered
}

const defaultHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        :root {
            --bg-color: #ffffff;
            --text-color: #1a1a1a;
            --card-bg: #f8f9fa;
            --border-color: #e9ecef;
            --primary-color: #6366f1;
        }
        {{if .DarkMode}}
        :root {
            --bg-color: #1a1a2e;
            --text-color: #eaeaea;
            --card-bg: #16213e;
            --border-color: #0f3460;
            --primary-color: #818cf8;
        }
        {{end}}
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-color);
            line-height: 1.6;
            padding: 24px;
        }
        .dashboard-header {
            margin-bottom: 24px;
            padding-bottom: 16px;
            border-bottom: 1px solid var(--border-color);
        }
        .dashboard-header h1 {
            font-size: 24px;
            font-weight: 600;
            margin-bottom: 8px;
        }
        .dashboard-header p {
            color: var(--text-color);
            opacity: 0.8;
        }
        .dashboard-grid {
            display: grid;
            grid-template-columns: repeat(12, 1fr);
            gap: 16px;
        }
        .widget {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 16px;
            overflow: hidden;
        }
        .widget-title {
            font-size: 14px;
            font-weight: 600;
            margin-bottom: 12px;
            color: var(--text-color);
        }
        .widget-content {
            font-size: 14px;
        }
        .metric-value {
            font-size: 36px;
            font-weight: 700;
            color: var(--primary-color);
        }
        .markdown {
            white-space: pre-wrap;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 8px 12px;
            text-align: left;
            border-bottom: 1px solid var(--border-color);
        }
        th {
            font-weight: 600;
            background: rgba(0,0,0,0.02);
        }
        .chart-placeholder {
            background: linear-gradient(135deg, var(--card-bg) 0%, var(--border-color) 100%);
            height: 200px;
            display: flex;
            align-items: center;
            justify-content: center;
            border-radius: 4px;
            color: var(--text-color);
            opacity: 0.6;
        }
    </style>
</head>
<body>
    <div class="dashboard-header">
        <h1>{{.Title}}</h1>
        {{if .Description}}<p>{{.Description}}</p>{{end}}
    </div>
    <div class="dashboard-grid">
        {{range .Widgets}}
        <div class="widget" style="grid-column: span {{.W}}; grid-row: span {{.H}};">
            {{if .Title}}<div class="widget-title">{{.Title}}</div>{{end}}
            <div class="widget-content">
                {{if eq .Type "text"}}
                    {{.Content}}
                {{else if eq .Type "metric"}}
                    {{.Content}}
                {{else if eq .Type "table"}}
                    <div class="table-container" data-config="{{.DataJSON}}">
                        <p>Table data loaded from configuration</p>
                    </div>
                {{else if eq .Type "chart"}}
                    <div class="chart-placeholder" data-config="{{.DataJSON}}">
                        Chart visualization
                    </div>
                {{else}}
                    <p>Unknown widget type: {{.Type}}</p>
                {{end}}
            </div>
        </div>
        {{end}}
    </div>
    {{if .DashboardJSON}}
    <script>
        window.dashboardData = {{.DashboardJSON}};
    </script>
    {{end}}
</body>
</html>`
