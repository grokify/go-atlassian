# MCP Server

Go-Atlassian includes an MCP (Model Context Protocol) server that enables AI assistants like Claude to interact with Jira directly.

## Overview

The MCP server (`gojira-mcp`) provides a stdio-based JSON-RPC interface that AI tools can use to:

- Search and retrieve Jira issues
- Create new issues with custom fields
- Update issue fields and labels
- Add comments and transition issue status
- List projects and available transitions

## Installation

```bash
go install github.com/grokify/go-atlassian/cmd/gojira-mcp@latest
```

## Configuration

### Environment Variables

The MCP server requires these environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `JIRA_BASE_URL` | Jira server URL | `https://company.atlassian.net` |
| `JIRA_USERNAME` | Jira username or email | `user@example.com` |
| `JIRA_API_TOKEN` | Jira API token | `your-api-token` |

Optional:

| Variable | Description | Default |
|----------|-------------|---------|
| `GOJIRA_MCP_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |

### Claude Code Setup

Add to your Claude Code MCP settings (`~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "jira": {
      "command": "gojira-mcp",
      "env": {
        "JIRA_BASE_URL": "https://company.atlassian.net",
        "JIRA_USERNAME": "user@example.com",
        "JIRA_API_TOKEN": "your-api-token"
      }
    }
  }
}
```

### Claude Desktop Setup

Add to your Claude Desktop config:

=== "macOS"

    `~/Library/Application Support/Claude/claude_desktop_config.json`

=== "Windows"

    `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "jira": {
      "command": "gojira-mcp",
      "env": {
        "JIRA_BASE_URL": "https://company.atlassian.net",
        "JIRA_USERNAME": "user@example.com",
        "JIRA_API_TOKEN": "your-api-token"
      }
    }
  }
}
```

## Available Tools

### jira_get_issue

Get a Jira issue by key with all fields.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `key` | string | Yes | Issue key (e.g., PROJ-123) |
| `expand` | string | No | Fields to expand (e.g., changelog, renderedFields) |

**Example:**

```json
{
  "key": "PROJ-123",
  "expand": "changelog"
}
```

### jira_search

Search Jira issues using JQL.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `jql` | string | Yes | JQL query string |
| `max_results` | integer | No | Maximum results (default: 50, max: 100) |
| `fields` | string | No | Comma-separated fields to return |

**Example:**

```json
{
  "jql": "project = PROJ AND status = 'In Progress'",
  "max_results": 20
}
```

### jira_create_issue

Create a new Jira issue.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `project` | string | Yes | Project key |
| `type` | string | Yes | Issue type (Story, Bug, Task, Epic) |
| `summary` | string | Yes | Issue summary/title |
| `description` | string | No | Issue description |
| `parent` | string | No | Parent issue key |
| `labels` | array | No | Labels to apply |
| `priority` | string | No | Priority name |
| `assignee` | string | No | Assignee username |
| `components` | array | No | Component names |
| `custom_fields` | object | No | Custom field values |

**Example:**

```json
{
  "project": "PROJ",
  "type": "Story",
  "summary": "Add user authentication",
  "description": "Implement OAuth2 login flow",
  "labels": ["auth", "mvp"],
  "priority": "High",
  "custom_fields": {
    "customfield_12345": "Given A, When B, Then C"
  }
}
```

### jira_update_issue

Update a Jira issue's fields.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `key` | string | Yes | Issue key |
| `summary` | string | No | New summary |
| `description` | string | No | New description |
| `labels` | array | No | Labels to set (replaces existing) |
| `add_labels` | array | No | Labels to add |
| `remove_labels` | array | No | Labels to remove |

**Example:**

```json
{
  "key": "PROJ-123",
  "summary": "Updated title",
  "add_labels": ["reviewed"]
}
```

### jira_add_comment

Add a comment to a Jira issue.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `key` | string | Yes | Issue key |
| `body` | string | Yes | Comment body text |

**Example:**

```json
{
  "key": "PROJ-123",
  "body": "Implementation complete. Ready for review."
}
```

### jira_get_transitions

Get available status transitions for an issue.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `key` | string | Yes | Issue key |

**Response:**

```json
{
  "key": "PROJ-123",
  "transitions": [
    {"id": "21", "name": "Start Progress", "to": "In Progress"},
    {"id": "31", "name": "Done", "to": "Done"}
  ]
}
```

### jira_transition_issue

Transition a Jira issue to a new status.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `key` | string | Yes | Issue key |
| `transition_id` | string | Yes | Transition ID from jira_get_transitions |
| `comment` | string | No | Comment to add with transition |

**Example:**

```json
{
  "key": "PROJ-123",
  "transition_id": "21",
  "comment": "Starting work on this issue"
}
```

### jira_get_comments

Get comments on a Jira issue.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `key` | string | Yes | Issue key |
| `max_results` | integer | No | Maximum comments (default: 50) |

### jira_get_projects

List available Jira projects.

**Parameters:** None

**Response:**

```json
{
  "total": 3,
  "projects": [
    {"key": "PROJ", "name": "Project Name", "id": "10001"},
    {"key": "DEV", "name": "Development", "id": "10002"}
  ]
}
```

## Protocol

The MCP server uses JSON-RPC 2.0 over stdio. It implements the standard MCP methods:

- `initialize` - Initialize the server
- `tools/list` - List available tools
- `tools/call` - Execute a tool

## Troubleshooting

### Connection Issues

1. Verify environment variables are set correctly
2. Test credentials with the CLI: `gojira get PROJ-123`
3. Check the log level: `GOJIRA_MCP_LOG_LEVEL=debug gojira-mcp`

### Permission Errors

Ensure your API token has sufficient permissions for the operations you're attempting. Some operations (like transitions) may require specific project permissions.
