# Installation

## CLI Installation

Install the `gojira` command-line tool:

```bash
go install github.com/grokify/go-atlassian/cmd/gojira@latest
```

Verify the installation:

```bash
gojira version
```

## SDK Installation

Add Go-Atlassian as a dependency to your Go project:

```bash
go get github.com/grokify/go-atlassian
```

Import the packages you need:

```go
import (
    "github.com/grokify/go-atlassian"       // JQL builder, config
    "github.com/grokify/go-atlassian/jira"  // REST API client
    "github.com/grokify/go-atlassian/xml"   // XML parser
)
```

## Requirements

- Go 1.21 or later
- Jira Cloud or Jira Server instance
- API token (for Jira Cloud) or username/password (for Jira Server)

## Getting an API Token

For Jira Cloud, you need an API token instead of your password:

1. Go to [Atlassian API Tokens](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Click **Create API token**
3. Give it a descriptive name (e.g., "Go-Atlassian CLI")
4. Copy the token - you won't see it again

!!! warning "Keep your token secure"
    API tokens provide full access to your Jira account. Never commit them to version control or share them publicly.

## Next Steps

- [Quick Start](quickstart.md) - Configure authentication and run your first commands
- [Authentication](cli/authentication.md) - Detailed authentication options
