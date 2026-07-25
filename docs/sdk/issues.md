# Issues

The SDK provides several types for working with Jira issues.

## Issue Types

### jira.Issues

A slice of `jira.Issue` from the go-jira library:

```go
issues, err := client.IssueAPI.SearchIssues("project = FOO", false)
// issues is jira.Issues ([]jira.Issue)

for _, issue := range issues {
    fmt.Printf("%s: %s\n", issue.Key, issue.Fields.Summary)
}
```

### jira.IssueMore

A wrapper around `jira.Issue` with convenience methods:

```go
import "github.com/grokify/go-atlassian/jira"

issue := issues[0]
im := jira.NewIssueMore(&issue)

// Convenience accessors
key := im.Key()
summary := im.Summary()
status := im.StatusName()
typeName := im.TypeName()
assignee := im.AssigneeName()
resolution := im.Resolution()
```

### jira.IssuesSet

A structured set of issues with aggregation and filtering:

```go
// Create from Issues
issuesSet, err := issues.IssuesSet(nil)
if err != nil {
    log.Fatal(err)
}

// Get counts
countsByStatus := issuesSet.CountsByStatus()
countsByType := issuesSet.CountsByType(true, true)
countsByProject := issuesSet.CountsByProject()

// Filter
filtered := issuesSet.FilterByStatus("Open")
```

## Searching Issues

### Basic Search

```go
issues, err := client.IssueAPI.SearchIssues(
    "project = FOO AND status = Open", // JQL query
    false,                              // Retrieve all (paginate)
)
```

### Search with Pagination

```go
// Retrieve all matching issues (handles pagination internally)
issues, err := client.IssueAPI.SearchIssues(jql, true)
```

### Search to IssuesSet

```go
issuesSet, err := client.IssueAPI.SearchIssuesSet("project = FOO")
```

## Getting Issues

### Single Issue

```go
ctx := context.Background()
issue, err := client.IssueAPI.Issue(ctx, "FOO-123", nil)
```

### With Options

```go
opts := &jira.GetQueryOptions{
    ExpandChangelog: true,  // Include issue history
}
issue, err := client.IssueAPI.Issue(ctx, "FOO-123", opts)
```

### Multiple Issues

```go
keys := []string{"FOO-123", "FOO-456", "FOO-789"}
issues, err := client.IssueAPI.Issues(ctx, keys, nil)
```

## IssuesSet Operations

### Aggregations

```go
issuesSet, _ := issues.IssuesSet(nil)

// Count by standard fields
byStatus := issuesSet.CountsByStatus()     // map[string]uint
byType := issuesSet.CountsByType(true, true)
byProject := issuesSet.CountsByProject()

// Count by custom field ID
byCF, err := issuesSet.CountsByCustomFieldValues("customfield_10001")

// Count by custom field name (handles duplicate field names)
// Requires a CustomFieldSet for name resolution
byName, err := issuesSet.CountsByCustomFieldName("Module", client.CustomFieldSet, true)
```

### Filtering

```go
// Filter by status
open := issuesSet.FilterByStatus("Open")

// Filter by type
bugs := issuesSet.FilterByType("Bug")

// Filter by custom criteria
filtered := issuesSet.Filter(func(im *jira.IssueMore) bool {
    return im.Priority() == "High"
})
```

### Export to Table

```go
tbl, err := issuesSet.TableDefault(nil, true, "Initiative", []string{})
if err != nil {
    log.Fatal(err)
}

// Write to file
err = tbl.WriteXLSX("output.xlsx", "Issues")
```

### Parent Issues

Include parent issues (epics, initiatives):

```go
err := client.IssueAPI.IssuesSetAddParents(issuesSet)
if err != nil {
    log.Fatal(err)
}

// Access parents
if issuesSet.Parents != nil {
    parentKeys := issuesSet.Parents.Keys()
}
```

## Reading from Files

### From JSON

```go
issuesSet, err := jira.IssuesSetReadFileJSON("backup.json")
if err != nil {
    log.Fatal(err)
}
```

## Issue Fields

Access fields through the `jira.Issue.Fields` struct:

```go
issue := issues[0]

// Standard fields
summary := issue.Fields.Summary
description := issue.Fields.Description
status := issue.Fields.Status.Name
issueType := issue.Fields.Type.Name
priority := issue.Fields.Priority.Name
assignee := issue.Fields.Assignee.DisplayName
created := issue.Fields.Created
updated := issue.Fields.Updated

// Custom fields (via Unknowns map)
if val, ok := issue.Fields.Unknowns["customfield_10001"]; ok {
    // Handle custom field value
}
```

## Example: Issue Report

```go
package main

import (
    "fmt"
    "log"

    "github.com/grokify/go-atlassian/jira"
)

func main() {
    client, err := jira.NewClientFromBasicAuth(
        "https://your-instance.atlassian.net",
        "your-email@example.com",
        "your-api-token",
        false,
    )
    if err != nil {
        log.Fatal(err)
    }

    // Search issues
    issues, err := client.IssueAPI.SearchIssues("project = FOO", true)
    if err != nil {
        log.Fatal(err)
    }

    // Convert to IssuesSet for analysis
    issuesSet, err := issues.IssuesSet(nil)
    if err != nil {
        log.Fatal(err)
    }

    // Print status breakdown
    fmt.Println("Issues by Status:")
    for status, count := range issuesSet.CountsByStatus() {
        fmt.Printf("  %s: %d\n", status, count)
    }

    // Print type breakdown
    fmt.Println("\nIssues by Type:")
    for issueType, count := range issuesSet.CountsByType(true, true) {
        fmt.Printf("  %s: %d\n", issueType, count)
    }
}
```
