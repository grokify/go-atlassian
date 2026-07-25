package report

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/grokify/go-atlassian/jira"
)

// Engine executes report definitions and produces Dashforge IR.
type Engine struct {
	client     *jira.Client
	cache      *Cache
	processors map[SectionType]SectionProcessor
	mu         sync.RWMutex
}

// NewEngine creates a new report engine.
func NewEngine(client *jira.Client) *Engine {
	e := &Engine{
		client:     client,
		cache:      NewCache(5 * time.Minute),
		processors: make(map[SectionType]SectionProcessor),
	}

	// Register built-in processors
	e.RegisterProcessor(SectionTypeMarkdown, &MarkdownProcessor{})
	e.RegisterProcessor(SectionTypeJQL, &JQLProcessor{client: client})
	e.RegisterProcessor(SectionTypeVelocity, &VelocityProcessor{client: client})
	e.RegisterProcessor(SectionTypeBurndown, &BurndownProcessor{client: client})
	e.RegisterProcessor(SectionTypeWorklog, &WorklogProcessor{client: client})
	e.RegisterProcessor(SectionTypeCycleTime, &CycleTimeProcessor{client: client})
	e.RegisterProcessor(SectionTypeMetric, &MetricProcessor{client: client})
	e.RegisterProcessor(SectionTypeTable, &TableProcessor{client: client})

	return e
}

// RegisterProcessor registers a custom section processor.
func (e *Engine) RegisterProcessor(sectionType SectionType, processor SectionProcessor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.processors[sectionType] = processor
}

// Execute runs the report and returns a Dashforge Dashboard IR.
func (e *Engine) Execute(ctx context.Context, def *Definition, vars map[string]any) (*DashboardIR, error) {
	if def == nil {
		return nil, fmt.Errorf("definition is nil")
	}

	// Resolve variables
	resolvedVars := e.resolveVariables(def.Variables, vars)

	// Create execution context
	execCtx := &ExecutionContext{
		Context:   ctx,
		Variables: resolvedVars,
		Cache:     e.cache,
	}

	// Process sections
	widgets := make([]WidgetIR, 0, len(def.Sections))
	dataSources := make([]DataSourceIR, 0)

	currentY := 0
	for _, section := range def.Sections {
		e.mu.RLock()
		processor, ok := e.processors[section.Type]
		e.mu.RUnlock()

		if !ok {
			return nil, fmt.Errorf("unknown section type: %s", section.Type)
		}

		result, err := processor.Process(execCtx, &section)
		if err != nil {
			return nil, fmt.Errorf("process section %s: %w", section.ID, err)
		}

		// Add data source if present
		if result.DataSource != nil {
			dataSources = append(dataSources, *result.DataSource)
		}

		// Add widget with auto-positioning if not specified
		widget := result.Widget
		if widget.Position == nil {
			widget.Position = &PositionIR{
				X: 0,
				Y: currentY,
				W: 12,
				H: 4,
			}
			currentY += 4
		} else {
			// Update currentY based on widget position
			widgetBottom := widget.Position.Y + widget.Position.H
			if widgetBottom > currentY {
				currentY = widgetBottom
			}
		}
		widgets = append(widgets, widget)
	}

	// Convert variables to IR format
	variablesIR := make([]VariableIR, 0, len(def.Variables))
	for _, v := range def.Variables {
		variablesIR = append(variablesIR, VariableIR{
			ID:      v.ID,
			Type:    string(v.Type),
			Label:   v.Label,
			Default: v.Default,
			Options: v.Options,
		})
	}

	// Build dashboard IR
	dashboard := &DashboardIR{
		ID:          def.ID,
		Title:       def.Title,
		Description: def.Description,
		Version:     def.Version,
		Layout:      LayoutIR{Columns: 12, Responsive: true},
		DataSources: dataSources,
		Widgets:     widgets,
		Variables:   variablesIR,
		Theme:       e.convertTheme(def.Theme),
	}

	return dashboard, nil
}

// resolveVariables merges default values with provided values.
func (e *Engine) resolveVariables(defs []Variable, provided map[string]any) map[string]any {
	resolved := make(map[string]any)

	// Apply defaults
	for _, v := range defs {
		if v.Default != nil {
			resolved[v.ID] = v.Default
		}
	}

	// Apply provided values
	for k, v := range provided {
		resolved[k] = v
	}

	return resolved
}

// convertTheme converts theme to IR format.
func (e *Engine) convertTheme(theme *Theme) *ThemeIR {
	if theme == nil {
		return nil
	}

	ir := &ThemeIR{}
	if theme.DarkMode {
		ir.Mode = "dark"
	} else {
		ir.Mode = "light"
	}

	if theme.Colors != nil {
		ir.Colors = &ThemeColorsIR{
			Primary:   theme.Colors.Primary,
			Secondary: theme.Colors.Secondary,
			Success:   theme.Colors.Success,
			Warning:   theme.Colors.Warning,
			Danger:    theme.Colors.Danger,
		}
	}

	return ir
}

// SectionProcessor processes a section and produces widget IR.
type SectionProcessor interface {
	Process(ctx *ExecutionContext, section *Section) (*SectionResult, error)
}

// ExecutionContext provides context for section processing.
type ExecutionContext struct {
	Context   context.Context
	Variables map[string]any
	Cache     *Cache
}

// GetVariable returns a variable value.
func (c *ExecutionContext) GetVariable(name string) (any, bool) {
	v, ok := c.Variables[name]
	return v, ok
}

// GetStringVariable returns a string variable value.
func (c *ExecutionContext) GetStringVariable(name string) string {
	v, ok := c.Variables[name]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// GetIntVariable returns an int variable value.
func (c *ExecutionContext) GetIntVariable(name string) int {
	v, ok := c.Variables[name]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	default:
		return 0
	}
}

// SectionResult contains the output of processing a section.
type SectionResult struct {
	Widget     WidgetIR
	DataSource *DataSourceIR
}

// Cache provides caching for Jira API responses.
type Cache struct {
	data map[string]cacheEntry
	mu   sync.RWMutex
	ttl  time.Duration
}

type cacheEntry struct {
	value     any
	expiresAt time.Time
}

// NewCache creates a new cache with the given TTL.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		data: make(map[string]cacheEntry),
		ttl:  ttl,
	}
}

// Get retrieves a value from the cache.
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.data[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.value, true
}

// Set stores a value in the cache.
func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Clear clears the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheEntry)
}
