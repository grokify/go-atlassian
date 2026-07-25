package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseFile parses a report definition from a file.
// The format is determined by the file extension (.yaml, .yml, or .json).
func ParseFile(path string) (*Definition, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return ParseYAML(f)
	case ".json":
		return ParseJSON(f)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}

// ParseYAML parses a report definition from YAML.
func ParseYAML(r io.Reader) (*Definition, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read yaml: %w", err)
	}

	var def Definition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if err := def.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	return &def, nil
}

// ParseJSON parses a report definition from JSON.
func ParseJSON(r io.Reader) (*Definition, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read json: %w", err)
	}

	var def Definition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	if err := def.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	return &def, nil
}

// ParseYAMLString parses a report definition from a YAML string.
func ParseYAMLString(s string) (*Definition, error) {
	return ParseYAML(strings.NewReader(s))
}

// ParseJSONString parses a report definition from a JSON string.
func ParseJSONString(s string) (*Definition, error) {
	return ParseJSON(strings.NewReader(s))
}

// ParseSectionConfig parses a section's config into the appropriate type.
func ParseSectionConfig[T any](section *Section) (*T, error) {
	if len(section.Config) == 0 {
		return nil, nil
	}

	var config T
	if err := json.Unmarshal(section.Config, &config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &config, nil
}

// GetJQLConfig extracts JQL config from a section.
func (s *Section) GetJQLConfig() (*JQLConfig, error) {
	return ParseSectionConfig[JQLConfig](s)
}

// GetVelocityConfig extracts velocity config from a section.
func (s *Section) GetVelocityConfig() (*VelocityConfig, error) {
	return ParseSectionConfig[VelocityConfig](s)
}

// GetBurndownConfig extracts burndown config from a section.
func (s *Section) GetBurndownConfig() (*BurndownConfig, error) {
	return ParseSectionConfig[BurndownConfig](s)
}

// GetWorklogConfig extracts worklog config from a section.
func (s *Section) GetWorklogConfig() (*WorklogConfig, error) {
	return ParseSectionConfig[WorklogConfig](s)
}

// GetCycleTimeConfig extracts cycle time config from a section.
func (s *Section) GetCycleTimeConfig() (*CycleTimeConfig, error) {
	return ParseSectionConfig[CycleTimeConfig](s)
}

// GetMetricConfig extracts metric config from a section.
func (s *Section) GetMetricConfig() (*MetricConfig, error) {
	return ParseSectionConfig[MetricConfig](s)
}

// GetTableConfig extracts table config from a section.
func (s *Section) GetTableConfig() (*TableConfig, error) {
	return ParseSectionConfig[TableConfig](s)
}

// ToJSON converts the definition to JSON.
func (d *Definition) ToJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}

// ToYAML converts the definition to YAML.
func (d *Definition) ToYAML() ([]byte, error) {
	return yaml.Marshal(d)
}
