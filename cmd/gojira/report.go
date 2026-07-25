package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grokify/go-atlassian/report"
	"github.com/grokify/go-atlassian/report/export"
	"github.com/spf13/cobra"
)

var (
	reportOutputFile string
	reportVars       map[string]string
)

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.AddCommand(reportValidateCmd)
	reportCmd.AddCommand(reportRunCmd)
	reportCmd.AddCommand(reportExportCmd)

	// Run command flags
	reportRunCmd.Flags().StringToStringVar(&reportVars, "var", nil, "Variable values (e.g., --var project=FOO)")

	// Export command flags
	reportExportCmd.Flags().StringVarP(&reportOutputFile, "output", "o", "", "Output file path (required)")
	reportExportCmd.Flags().StringToStringVar(&reportVars, "var", nil, "Variable values (e.g., --var project=FOO)")
	if err := reportExportCmd.MarkFlagRequired("output"); err != nil {
		panic(err)
	}
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Report generation commands",
	Long: `Commands for working with Jira reports.

Reports are defined in YAML or JSON files and can be:
- Validated for correctness
- Executed to produce Dashforge Dashboard IR
- Exported to HTML for viewing

Example report definition (YAML):
  id: sprint-report
  title: Sprint Progress Report
  sections:
    - id: summary
      type: markdown
      content: "# Sprint Progress"
    - id: velocity
      type: velocity
      config:
        board_id: 123
        sprint_count: 6`,
}

var reportValidateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate a report definition file",
	Long: `Validates a report definition file for correct syntax and schema.

The file can be YAML (.yaml, .yml) or JSON (.json).

Examples:
  gojira report validate sprint-report.yaml
  gojira report validate report.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		def, err := report.ParseFile(filePath)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		result := struct {
			Valid    bool   `json:"valid"`
			ID       string `json:"id"`
			Title    string `json:"title"`
			Sections int    `json:"sections"`
		}{
			Valid:    true,
			ID:       def.ID,
			Title:    def.Title,
			Sections: len(def.Sections),
		}

		return outputResult(cmd, result)
	},
}

var reportRunCmd = &cobra.Command{
	Use:   "run <file>",
	Short: "Execute a report and output Dashforge IR",
	Long: `Executes a report definition and outputs the Dashforge Dashboard IR.

The IR can be used by Dashforge or other tools to render the report.

Examples:
  gojira report run sprint-report.yaml
  gojira report run report.yaml --var project=FOO --var sprint=10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Parse the report definition
		def, err := report.ParseFile(filePath)
		if err != nil {
			return fmt.Errorf("parse report: %w", err)
		}

		// Get Jira client
		client, err := NewClientFromOptions(getAuthOptions())
		if err != nil {
			return fmt.Errorf("connect to Jira: %w", err)
		}

		// Create engine
		engine := report.NewEngine(client)

		// Convert string vars to any
		vars := make(map[string]any)
		for k, v := range reportVars {
			vars[k] = v
		}

		// Execute report
		dashboard, err := engine.Execute(context.Background(), def, vars)
		if err != nil {
			return fmt.Errorf("execute report: %w", err)
		}

		// Output as JSON (Dashforge IR is always JSON)
		data, err := json.MarshalIndent(dashboard, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal output: %w", err)
		}
		fmt.Println(string(data))
		return nil
	},
}

var reportExportCmd = &cobra.Command{
	Use:   "export <file>",
	Short: "Export a report to HTML",
	Long: `Executes a report and exports it to an HTML file.

The HTML file can be viewed in any web browser.

Examples:
  gojira report export sprint-report.yaml -o report.html
  gojira report export report.yaml --var project=FOO -o output.html`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Parse the report definition
		def, err := report.ParseFile(filePath)
		if err != nil {
			return fmt.Errorf("parse report: %w", err)
		}

		// Get Jira client
		client, err := NewClientFromOptions(getAuthOptions())
		if err != nil {
			return fmt.Errorf("connect to Jira: %w", err)
		}

		// Create engine
		engine := report.NewEngine(client)

		// Convert string vars to any
		vars := make(map[string]any)
		for k, v := range reportVars {
			vars[k] = v
		}

		// Execute report
		dashboard, err := engine.Execute(context.Background(), def, vars)
		if err != nil {
			return fmt.Errorf("execute report: %w", err)
		}

		// Export to HTML
		exporter := export.NewHTMLExporter()
		if err := exporter.ExportToFile(dashboard, reportOutputFile); err != nil {
			return fmt.Errorf("export to HTML: %w", err)
		}

		result := struct {
			Success    bool   `json:"success"`
			OutputFile string `json:"output_file"`
			Title      string `json:"title"`
			Widgets    int    `json:"widgets"`
		}{
			Success:    true,
			OutputFile: reportOutputFile,
			Title:      dashboard.Title,
			Widgets:    len(dashboard.Widgets),
		}

		return outputResult(cmd, result)
	},
}
