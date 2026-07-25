# JQL Builder

The root package provides a JQL builder for constructing Jira Query Language queries programmatically.

## Basic Usage

```go
import "github.com/grokify/go-atlassian"

jql := jira.JQL{
    ProjectsIncl: [][]string{{"FOO"}},
    StatusesIncl: [][]string{{"Open", "In Progress"}},
}

query := jql.String()
// Result: "project = 'FOO' AND status IN ('Open', 'In Progress')"
```

## JQL Structure

The `JQL` struct supports include and exclude conditions for various fields:

```go
type JQL struct {
    // Date conditions
    CreatedGT       *time.Time
    CreatedGTE      *time.Time
    CreatedLT       *time.Time
    CreatedLTE      *time.Time
    DueGT           *time.Time
    DueGTE          *time.Time
    DueLT           *time.Time
    DueLTE          *time.Time
    UpdatedGT       *time.Time
    UpdatedGTE      *time.Time
    UpdatedLT       *time.Time
    UpdatedLTE      *time.Time

    // String field conditions (outer = AND, inner = IN)
    FiltersIncl     [][]string
    FiltersExcl     [][]string
    IssuesIncl      [][]string
    IssuesExcl      [][]string
    KeysIncl        [][]string
    KeysExcl        [][]string
    LabelsIncl      [][]string
    LabelsExcl      [][]string
    ParentsIncl     [][]string
    ParentsExcl     [][]string
    ProjectsIncl    [][]string
    ProjectsExcl    [][]string
    ResolutionIncl  [][]string
    ResolutionExcl  [][]string
    StatusesIncl    [][]string
    StatusesExcl    [][]string
    TypesIncl       [][]string
    TypesExcl       [][]string

    // Text search
    SummaryLike     []string
    SummaryNotLike  []string
    TextLike        []string
    TextNotLike     []string

    // Custom fields
    CustomFieldIncl map[string][]string
    CustomFieldExcl map[string][]string

    // Raw JQL fragments
    Raw             []string
}
```

## Examples

### Single Project, Single Status

```go
jql := jira.JQL{
    ProjectsIncl: [][]string{{"FOO"}},
    StatusesIncl: [][]string{{"Open"}},
}
// project = 'FOO' AND status = 'Open'
```

### Multiple Projects

```go
jql := jira.JQL{
    ProjectsIncl: [][]string{{"FOO", "BAR", "BAZ"}},
}
// project IN ('FOO', 'BAR', 'BAZ')
```

### Multiple Status Groups (AND)

```go
jql := jira.JQL{
    StatusesIncl: [][]string{
        {"Open", "In Progress"},  // First condition
        {"Blocked"},               // AND this condition
    },
}
// status IN ('Open', 'In Progress') AND status = 'Blocked'
```

### Exclude Conditions

```go
jql := jira.JQL{
    ProjectsIncl: [][]string{{"FOO"}},
    StatusesExcl: [][]string{{"Done", "Closed"}},
}
// project = 'FOO' AND status NOT IN ('Done', 'Closed')
```

### Date Filtering

```go
import "time"

sevenDaysAgo := time.Now().AddDate(0, 0, -7)
jql := jira.JQL{
    ProjectsIncl: [][]string{{"FOO"}},
    CreatedGTE:   &sevenDaysAgo,
}
// project = 'FOO' AND createdDate >= 2024-01-08
```

### Text Search

```go
jql := jira.JQL{
    ProjectsIncl: [][]string{{"FOO"}},
    TextLike:     []string{"error"},
}
// project = 'FOO' AND text ~ "error"
```

### Summary Search

```go
jql := jira.JQL{
    SummaryLike: []string{"login", "authentication"},
}
// summary ~ "login" AND summary ~ "authentication"
```

### Custom Fields

```go
jql := jira.JQL{
    CustomFieldIncl: map[string][]string{
        "customfield_10001": {"Team A", "Team B"},
    },
}
// cf[10001] IN ('Team A', 'Team B')
```

### Raw JQL

For complex conditions not supported by the builder:

```go
jql := jira.JQL{
    ProjectsIncl: [][]string{{"FOO"}},
    Raw: []string{
        "fixVersion = '1.0'",
        "sprint in openSprints()",
    },
}
// project = 'FOO' AND fixVersion = '1.0' AND sprint in openSprints()
```

### Labels

```go
jql := jira.JQL{
    LabelsIncl: [][]string{{"bug", "urgent"}},
}
// labels IN ('bug', 'urgent')
```

## Query String

Generate URL-encoded query string:

```go
jql := jira.JQL{
    ProjectsIncl: [][]string{{"FOO"}},
}

qs := jql.QueryString()
// jql=project%20%3D%20%27FOO%27
```

## JQL Metadata

Store query metadata for tracking:

```go
jql := jira.JQL{
    Meta: jira.JQLMeta{
        Name:        "Open Bugs",
        Description: "All open bugs in FOO project",
        QueryTime:   time.Now(),
    },
    ProjectsIncl: [][]string{{"FOO"}},
    TypesIncl:    [][]string{{"Bug"}},
    StatusesIncl: [][]string{{"Open"}},
}
```

## Splitting Long Queries

For very long value lists that exceed Jira's limits:

```go
values := []string{"KEY-1", "KEY-2", ..., "KEY-1000"}

jqls := jira.JQLStringsSimple(
    "key",    // Field name
    false,    // Exclude (false = include)
    values,   // Values
    0,        // Max length (0 = default)
)

// Returns multiple JQL strings, each under the length limit
for _, jql := range jqls {
    issues, _ := client.IssueAPI.SearchIssues(jql, false)
    // Process issues...
}
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/grokify/go-atlassian"
    "github.com/grokify/go-atlassian/jira"
)

func main() {
    // Build JQL query
    oneMonthAgo := time.Now().AddDate(0, -1, 0)

    jql := jira.JQL{
        ProjectsIncl: [][]string{{"FOO", "BAR"}},
        TypesIncl:    [][]string{{"Bug", "Task"}},
        StatusesExcl: [][]string{{"Done", "Closed"}},
        CreatedGTE:   &oneMonthAgo,
    }

    query := jql.String()
    fmt.Printf("JQL: %s\n", query)

    // Use with client
    client, err := jira.NewClientFromBasicAuth(
        "https://your-instance.atlassian.net",
        "your-email@example.com",
        "your-api-token",
        false,
    )
    if err != nil {
        log.Fatal(err)
    }

    issues, err := client.IssueAPI.SearchIssues(query, true)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d issues\n", len(issues))
}
```
