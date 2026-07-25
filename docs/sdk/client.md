# Client

The `jira.Client` provides access to the Jira REST API. It wraps the [go-jira](https://github.com/andygrunwald/go-jira) library with additional utilities.

## Creating a Client

### From Direct Credentials

```go
import "github.com/grokify/go-atlassian/jira"

client, err := jira.NewClientFromBasicAuth(
    "https://your-instance.atlassian.net", // Server URL
    "your-email@example.com",              // Username
    "your-api-token",                       // API token/password
    false,                                   // Load custom fields on init
)
if err != nil {
    log.Fatal(err)
}
```

### From goauth Credentials File

```go
client, err := jira.NewClientGoauthBasicAuthFile(
    "~/.config/goauth/credentials.json", // Credentials file path
    "jira-prod",                          // Account key
    false,                                 // Load custom fields
)
```

### From Environment Variables

Create a helper to read from environment:

```go
import (
    "os"
    "github.com/grokify/go-atlassian/jira"
)

func NewClientFromEnv() (*jira.Client, error) {
    return jira.NewClientFromBasicAuth(
        os.Getenv("JIRA_URL"),
        os.Getenv("JIRA_USER"),
        os.Getenv("JIRA_TOKEN"),
        false,
    )
}
```

### Interactive Selection

Use goauth CLI for interactive account selection:

```go
client, err := jira.NewClientFromGoauthCLI(
    true,  // Include accounts on error
    false, // Load custom fields
)
```

## Client Structure

```go
type Client struct {
    Config         *jira.Config
    HTTPClient     *http.Client
    JiraClient     *jira.Client
    Logger         *slog.Logger
    BacklogAPI     *BacklogService
    CreateMetaAPI  *CreateMetaService
    CustomFieldAPI *CustomFieldService
    IssueAPI       *IssueService
    CustomFieldSet *CustomFieldSet
}
```

## Service APIs

### IssueAPI

The most commonly used service for issue operations:

```go
// Search issues
issues, err := client.IssueAPI.SearchIssues("project = FOO", false)

// Get single issue
ctx := context.Background()
issue, err := client.IssueAPI.Issue(ctx, "FOO-123", nil)

// Get multiple issues
issues, err := client.IssueAPI.Issues(ctx, []string{"FOO-1", "FOO-2"}, nil)

// Update issue
resp, err := client.IssueAPI.IssuePatch(ctx, "FOO-123", patchBody)
```

### CustomFieldAPI

Access custom field definitions:

```go
// Get all custom fields
fields, err := client.CustomFieldAPI.GetCustomFields()

// Get Epic Link field
epicField, err := client.CustomFieldAPI.GetCustomFieldEpicLink()

// Get field by exact ID
field, err := client.CustomFieldAPI.GetCustomFieldByID("customfield_10001")

// Get all fields matching a name (handles duplicates)
fields, err := client.CustomFieldAPI.GetCustomFieldsByName("Module")

// Get custom fields available in a specific project
ctx := context.Background()
projectFields, err := client.CustomFieldAPI.GetCustomFieldsForProject(ctx, "ABC")
```

### CreateMetaAPI

Discover which fields are available for issue creation in specific projects:

```go
ctx := context.Background()

// Get issue types for a project
issueTypes, err := client.CreateMetaAPI.GetIssueTypes(ctx, "ABC")

// Get fields for a specific issue type
fields, err := client.CreateMetaAPI.GetFields(ctx, "ABC", "10001")

// Get all fields across all issue types in a project
allFields, err := client.CreateMetaAPI.GetAllFieldsForProject(ctx, "ABC")
```

### BacklogAPI

Backlog operations:

```go
// Get backlog for a board
backlog, err := client.BacklogAPI.Backlog(boardID)
```

## Loading Custom Fields

Custom fields can be loaded during client initialization or later:

```go
// Load during init
client, err := jira.NewClientFromBasicAuth(url, user, token, true)

// Or load later
err := client.LoadCustomFields()

// Access loaded fields
if client.CustomFieldSet != nil {
    name, err := client.CustomFieldSet.IDToName("customfield_10001")
}
```

## Handling Duplicate Custom Field Names

Jira allows multiple custom fields to share the same display name (common when copying schemes or reinstalling apps). The SDK provides tools to handle this:

### Detecting Duplicates

```go
fields, err := client.CustomFieldAPI.GetCustomFields()
if err != nil {
    log.Fatal(err)
}

// Find which names appear more than once
duplicates := fields.DuplicateNames()
// Returns: ["Module", "Sprint"] (sorted)

// Map names to all their IDs
nameToIDs := fields.MapNameToIDs()
// Returns: {"Module": ["customfield_123", "customfield_456"], ...}
```

### Getting All Fields by Name

When a name might have duplicates, use `GetCustomFieldsByName()`:

```go
// Returns all fields named "Module" (may be multiple)
fields, err := client.CustomFieldAPI.GetCustomFieldsByName("Module")
if err != nil {
    log.Fatal(err)
}

for _, f := range fields {
    fmt.Printf("ID: %s, Name: %s\n", f.ID, f.Name)
}
```

### Resolving Values from Issues

Use `CustomFieldSet` to extract values when field names are ambiguous:

```go
// Load custom fields
err := client.LoadCustomFields()
cfSet := client.CustomFieldSet

// Get all values for fields named "Module" from an issue
values := cfSet.IssueCustomFieldsByName(issue, "Module")
for _, v := range values {
    fmt.Printf("Field %s (%s): %s\n", v.FieldName, v.FieldID, v.Value)
}

// Get only populated values (recommended for ambiguous names)
// The populated one is typically the active field
populated := cfSet.IssueCustomFieldsByNamePopulated(issue, "Module")
if len(populated) == 1 {
    fmt.Printf("Active Module field: %s\n", populated[0].Value)
}
```

### Counting by Custom Field Name

When aggregating issues, use name-based counting:

```go
issuesSet, _ := client.IssueAPI.SearchIssuesSet("project = FOO")

// Count by field name, handling duplicates
counts, err := issuesSet.CountsByCustomFieldName("Module", cfSet, true)
// true = only count issues with populated values
```

## Configuration

The client includes configuration from the root package:

```go
// Access config
serverURL := client.Config.ServerURL

// Modify config
client.Config.ServerURL = "https://new-instance.atlassian.net"
```

## Logging

Set a logger for debug output:

```go
import "log/slog"

client.Logger = slog.Default()
```

## Low-Level Access

For operations not covered by service APIs, access the underlying clients:

```go
// go-jira client
jiraClient := client.JiraClient

// Raw HTTP client
httpClient := client.HTTPClient
```

## Example: Complete Workflow

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/grokify/go-atlassian/jira"
)

func main() {
    // Create client
    client, err := jira.NewClientFromBasicAuth(
        "https://your-instance.atlassian.net",
        "your-email@example.com",
        "your-api-token",
        true, // Load custom fields
    )
    if err != nil {
        log.Fatal(err)
    }

    // Search issues
    issues, err := client.IssueAPI.SearchIssues("project = FOO AND status = Open", false)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d issues\n", len(issues))

    // Get details for first issue
    if len(issues) > 0 {
        ctx := context.Background()
        issue, err := client.IssueAPI.Issue(ctx, issues[0].Key, nil)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Issue: %s - %s\n", issue.Key, issue.Fields.Summary)
    }
}
```
