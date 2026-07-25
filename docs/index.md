# Go-Atlassian

Go-Atlassian is a Go SDK and CLI for Jira that provides:

- **REST API client** (`rest/`) - wrapper around [go-jira](https://github.com/andygrunwald/go-jira) with additional utilities
- **XML parser** (`xml/`) - parse Jira XML exports when API access is unavailable
- **JQL builder** (root package) - programmatically construct JQL queries
- **CLI tool** (`cmd/gojira/`) - command-line interface optimized for AI agents and humans

## Features

- **Multiple output formats**: JSON, Table, and TOON (Token-Optimized Object Notation)
- **Flexible authentication**: Environment variables, goauth credential files, or CLI flags
- **AI-agent friendly**: Non-interactive mode, structured output, consistent exit codes
- **Lightweight core**: Root package has no external dependencies

## Package Structure

| Package | Description | Dependencies |
|---------|-------------|--------------|
| root | JQL builder, config, constants | None (lightweight) |
| `go-atlassian/jira` | REST API client | go-jira SDK |
| `go-atlassian/xml` | XML export parser | None |
| `go-atlassian/web` | URL helpers | None |

## Quick Example

=== "CLI"

    ```bash
    # Search issues
    gojira search --jql "project = FOO AND status = Open"

    # Get issue details
    gojira get ISSUE-123

    # Show statistics
    gojira stats --jql "project = FOO" --by status --format table
    ```

=== "SDK"

    ```go
    import "github.com/grokify/go-atlassian/jira"

    client, err := jira.NewClientFromBasicAuth(
        "https://your-instance.atlassian.net",
        "your-email@example.com",
        "your-api-token",
        false,
    )

    issues, err := client.IssueAPI.SearchIssues("project = FOO", false)
    ```

## Use Cases

1. **Automate Jira operations** - Search, update, and export issues via CLI
2. **Build custom integrations** - Use the SDK to integrate Jira with other systems
3. **Generate reports** - Export issues to JSON/XLSX for analysis
4. **AI agent workflows** - Token-optimized output for LLM consumption
5. **Parse XML exports** - Access Jira data when API is unavailable

## Getting Started

- [Installation](installation.md) - Install the CLI and SDK
- [Quick Start](quickstart.md) - Get up and running in minutes
- [CLI Reference](cli/index.md) - Full command documentation
- [SDK Guide](sdk/index.md) - Using Go-Atlassian as a library
