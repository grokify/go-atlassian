# Quick Start

Get up and running with go-atlassian in minutes.

## Prerequisites

- Go 1.21 or later installed
- Access to a Jira instance (Cloud or Server)
- API token or credentials

## 1. Install the CLI

```bash
go install github.com/grokify/go-atlassian/cmd/gojira@latest
```

## 2. Configure Authentication

Set environment variables for your Jira instance:

```bash
export JIRA_URL=https://your-instance.atlassian.net
export JIRA_USER=your-email@example.com
export JIRA_TOKEN=your-api-token
```

!!! tip "Add to your shell profile"
    Add these exports to `~/.bashrc` or `~/.zshrc` to persist them.

## 3. Search for Issues

```bash
# Search for open issues in a project
gojira search --jql "project = FOO AND status = Open"

# Get the first 10 results
gojira search --jql "project = FOO" --max 10

# Output as human-readable table
gojira search --jql "assignee = currentUser()" --table
```

## 4. Get Issue Details

```bash
# Get a single issue
gojira get ISSUE-123

# Get multiple issues
gojira get ISSUE-123 ISSUE-456

# Include changelog
gojira get ISSUE-123 --expand
```

## 5. View Statistics

```bash
# Count issues by status
gojira stats --jql "project = FOO" --by status --format table

# Count by assignee
gojira stats --jql "project = FOO" --by assignee
```

## Output Formats

go-atlassian supports three output formats:

| Format | Flag | Use Case |
|--------|------|----------|
| JSON | `--json` | Default, machine-readable |
| Table | `--table` | Human-readable |
| TOON | `--toon` | Token-optimized for LLMs |

Example:

```bash
# JSON output (default)
gojira search --jql "project = FOO" --json

# Table output
gojira search --jql "project = FOO" --table

# TOON output (efficient for AI agents)
gojira search --jql "project = FOO" --toon
```

## Next Steps

- [Authentication](cli/authentication.md) - Configure goauth credentials
- [CLI Reference](cli/index.md) - Full command documentation
- [JQL Examples](guides/jql-examples.md) - Common JQL patterns
