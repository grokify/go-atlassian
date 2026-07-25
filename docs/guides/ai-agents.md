# AI Agents Guide

go-atlassian is designed to work well with AI agents like Claude Code. This guide covers best practices for AI-agent workflows.

## Why go-atlassian for AI Agents?

1. **Non-interactive operation**: All auth via environment variables or files
2. **Structured output**: JSON and TOON formats for reliable parsing
3. **Consistent exit codes**: 0 for success, 1 for errors
4. **Token efficiency**: TOON format reduces token consumption by ~8x
5. **Quiet mode**: Suppress progress messages for cleaner output

## TOON Format

TOON (Token-Optimized Object Notation) is a compact format designed for LLM consumption:

```bash
# JSON output (verbose)
gojira search --jql "project = FOO" --json

# TOON output (compact)
gojira search --jql "project = FOO" --toon
```

### TOON Example

JSON:
```json
{
  "field": "status",
  "total": 150,
  "results": [
    {"value": "Done", "count": 75},
    {"value": "Open", "count": 45}
  ]
}
```

TOON:
```
f:status|t:150|r:[{v:Done,n:75},{v:Open,n:45}]
```

The TOON format uses abbreviated keys (`f` for field, `t` for total, etc.) and minimal punctuation.

## Environment Setup

Set credentials via environment variables for non-interactive operation:

```bash
export JIRA_URL=https://your-instance.atlassian.net
export JIRA_USER=your-email@example.com
export JIRA_TOKEN=your-api-token
```

## Common Workflows

### 1. Search and Parse Issues

```bash
# Search for issues (quiet mode, TOON output)
gojira search --jql "project = FOO AND status = Open" --toon -q

# Search with JSON for detailed parsing
gojira search --jql "project = FOO" --json -q | jq '.[].key'
```

### 2. Get Issue Details

```bash
# Get single issue
gojira get FOO-123 --toon -q

# Get multiple issues
gojira get FOO-123 FOO-456 --json -q
```

### 3. Analyze Issue Distribution

```bash
# Status breakdown
gojira stats --jql "project = FOO" --by status --format json -q

# Type breakdown
gojira stats --jql "project = FOO" --by type --format json -q
```

### 4. Update Issues

```bash
# Add label
gojira patch FOO-123 --add-label reviewed -q

# Set field
gojira patch FOO-123 --set priority=High -q
```

### 5. Export for Analysis

```bash
# Export to JSON
gojira export --jql "project = FOO" --json issues.json -q

# Export to XLSX
gojira export --jql "project = FOO" --xlsx report.xlsx -q
```

## Error Handling

Check exit codes for automation:

```bash
if gojira search --jql "project = FOO" -q > /dev/null 2>&1; then
    echo "Search succeeded"
else
    echo "Search failed"
fi
```

## Token Optimization Tips

1. **Use TOON format**: `--toon` uses ~8x fewer tokens than JSON
2. **Use quiet mode**: `-q` suppresses progress messages
3. **Limit results**: `--max 10` for quick checks
4. **Use stats**: Get summaries instead of raw issues

### Example: Efficient Sprint Status Check

```bash
# Inefficient (returns full issue data)
gojira search --jql "sprint = 'Sprint 42'" --json

# Efficient (returns summary counts)
gojira stats --jql "sprint = 'Sprint 42'" --by status --format toon -q
```

## SDK Integration

For programmatic AI agent integration:

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/grokify/go-atlassian/jira"
)

func main() {
    // Create client from environment
    client, err := jira.NewClientFromBasicAuth(
        os.Getenv("JIRA_URL"),
        os.Getenv("JIRA_USER"),
        os.Getenv("JIRA_TOKEN"),
        false,
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    // Search and output JSON
    issues, err := client.IssueAPI.SearchIssues("project = FOO", false)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    // Output as JSON for LLM parsing
    data, _ := json.Marshal(issues)
    fmt.Println(string(data))
}
```

## Security Considerations

1. **Use environment variables** for credentials, not CLI flags
2. **Use goauth files** with appropriate file permissions
3. **Never log credentials** in AI agent outputs
4. **Limit API token scope** when possible
