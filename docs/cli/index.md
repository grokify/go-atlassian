# CLI Reference

The `gojira` command-line interface provides commands for searching, retrieving, and managing Jira issues.

## Installation

```bash
go install github.com/grokify/go-atlassian/cmd/gojira@latest
```

## Commands

| Command | Description |
|---------|-------------|
| [search](search.md) | Search issues with JQL |
| [get](get.md) | Get one or more issues by key |
| [comments](comments.md) | Get comments for an issue |
| [patch](patch.md) | Update issue fields |
| [export](export.md) | Export issues to JSON or XLSX |
| [fields](fields.md) | List and filter custom fields |
| [stats](stats.md) | Show issue statistics grouped by field |
| version | Show version information |

## Global Flags

These flags are available on all commands:

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output as JSON (default) |
| `--table` | `-t` | Output as human-readable table |
| `--toon` | | Output as TOON (Token-Optimized Object Notation) |
| `--creds-file` | | Path to goauth credentials file |
| `--account` | | Account key in credentials file |
| `--quiet` | `-q` | Suppress non-essential output |

## Output Formats

### JSON (default)

Machine-readable JSON output, ideal for piping to `jq` or programmatic processing:

```bash
gojira search --jql "project = FOO" --json | jq '.[].key'
```

### Table

Human-readable ASCII table:

```bash
gojira search --jql "project = FOO" --table
```

Output:

```
KEY        TYPE    STATUS    SUMMARY
FOO-123    Bug     Open      Fix login issue
FOO-124    Story   Done      Add user profile
```

### TOON

Token-Optimized Object Notation - compact format optimized for LLM consumption:

```bash
gojira search --jql "project = FOO" --toon
```

TOON format uses abbreviated keys and is approximately 8x more token-efficient than JSON.

## Authentication

The CLI authenticates in this order:

1. **CLI flags**: `--creds-file` and `--account`
2. **Environment variables**: `JIRA_URL`, `JIRA_USER`, `JIRA_TOKEN`
3. **goauth file**: `~/.config/goauth/credentials.json`

See [Authentication](authentication.md) for details.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (authentication failure, API error, invalid arguments) |

## Examples

```bash
# Search with environment variables
export JIRA_URL=https://your-instance.atlassian.net
export JIRA_USER=your-email@example.com
export JIRA_TOKEN=your-api-token
gojira search --jql "project = FOO"

# Search with credentials file
gojira search --jql "project = FOO" --creds-file ~/.config/goauth/creds.json --account myaccount

# Quiet mode (suppress progress messages)
gojira search --jql "project = FOO" -q
```
