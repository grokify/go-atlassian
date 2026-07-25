# JQL Examples

Common JQL (Jira Query Language) patterns for use with the Go-Atlassian CLI and SDK.

## Basic Queries

### By Project

```bash
# Single project
gojira search --jql "project = FOO"

# Multiple projects
gojira search --jql "project in (FOO, BAR, BAZ)"
```

### By Status

```bash
# Single status
gojira search --jql "status = Open"

# Multiple statuses
gojira search --jql "status in (Open, 'In Progress', 'In Review')"

# Exclude statuses
gojira search --jql "status not in (Done, Closed)"
```

### By Type

```bash
# Bugs only
gojira search --jql "issuetype = Bug"

# Multiple types
gojira search --jql "issuetype in (Bug, Task, Story)"
```

### By Assignee

```bash
# Current user
gojira search --jql "assignee = currentUser()"

# Specific user
gojira search --jql "assignee = 'john.doe'"

# Unassigned
gojira search --jql "assignee is EMPTY"
```

## Combining Conditions

### AND

```bash
gojira search --jql "project = FOO AND status = Open AND type = Bug"
```

### OR

```bash
gojira search --jql "project = FOO AND (status = Open OR status = 'In Progress')"
```

## Date Queries

### Relative Dates

```bash
# Created in last 7 days
gojira search --jql "created >= -7d"

# Updated in last 24 hours
gojira search --jql "updated >= -1d"

# Due in next 3 days
gojira search --jql "due <= 3d"
```

### Absolute Dates

```bash
# Created after date
gojira search --jql "created >= '2024-01-01'"

# Created between dates
gojira search --jql "created >= '2024-01-01' AND created <= '2024-01-31'"
```

### Date Functions

```bash
# Created this week
gojira search --jql "created >= startOfWeek()"

# Due this month
gojira search --jql "due >= startOfMonth() AND due <= endOfMonth()"

# Updated this year
gojira search --jql "updated >= startOfYear()"
```

## Text Search

### Full Text

```bash
# Search all text fields
gojira search --jql "text ~ 'error'"

# Exact phrase
gojira search --jql 'text ~ "connection timeout"'
```

### Summary Only

```bash
gojira search --jql "summary ~ 'login'"
```

### Description Only

```bash
gojira search --jql "description ~ 'reproduction steps'"
```

## Priority and Resolution

### By Priority

```bash
# High priority
gojira search --jql "priority = High"

# High or Critical
gojira search --jql "priority in (High, Highest, Critical)"
```

### By Resolution

```bash
# Unresolved
gojira search --jql "resolution is EMPTY"

# Resolved
gojira search --jql "resolution is not EMPTY"

# Won't Fix
gojira search --jql "resolution = 'Won\\'t Fix'"
```

## Sprint and Version

### Sprint

```bash
# Current sprint
gojira search --jql "sprint in openSprints()"

# Specific sprint
gojira search --jql "sprint = 'Sprint 42'"

# Future sprints
gojira search --jql "sprint in futureSprints()"
```

### Fix Version

```bash
# Specific version
gojira search --jql "fixVersion = '1.0'"

# No fix version
gojira search --jql "fixVersion is EMPTY"

# Released versions
gojira search --jql "fixVersion in releasedVersions()"
```

## Labels

```bash
# Has label
gojira search --jql "labels = bug"

# Has any of these labels
gojira search --jql "labels in (bug, urgent, production)"

# No labels
gojira search --jql "labels is EMPTY"
```

## Components

```bash
# In component
gojira search --jql "component = Backend"

# Multiple components
gojira search --jql "component in (Backend, API)"
```

## Custom Fields

```bash
# By custom field ID
gojira search --jql "cf[10001] = 'Value'"

# By custom field name (if unique)
gojira search --jql "'Story Points' >= 5"

# Custom field is empty
gojira search --jql "cf[10001] is EMPTY"
```

## Parent/Child Relationships

### Epics

```bash
# Issues in an epic
gojira search --jql "'Epic Link' = FOO-100"

# Issues with no epic
gojira search --jql "'Epic Link' is EMPTY"
```

### Sub-tasks

```bash
# Sub-tasks of an issue
gojira search --jql "parent = FOO-123"

# All sub-tasks in project
gojira search --jql "project = FOO AND issuetype = Sub-task"
```

## Ordering and Limiting

### ORDER BY

```bash
# Order by priority (descending)
gojira search --jql "project = FOO ORDER BY priority DESC"

# Order by created date
gojira search --jql "project = FOO ORDER BY created ASC"

# Multiple order fields
gojira search --jql "project = FOO ORDER BY priority DESC, created ASC"
```

### Limiting Results

```bash
# First 10 results
gojira search --jql "project = FOO" --max 10

# All results
gojira search --jql "project = FOO" --all
```

## Complex Examples

### Open High-Priority Bugs

```bash
gojira search --jql "project = FOO AND type = Bug AND priority in (High, Highest) AND status != Done"
```

### My Overdue Tasks

```bash
gojira search --jql "assignee = currentUser() AND due < now() AND resolution is EMPTY"
```

### Recently Updated in Sprint

```bash
gojira search --jql "sprint in openSprints() AND updated >= -1d ORDER BY updated DESC"
```

### Blockers and Critical Issues

```bash
gojira search --jql "project = FOO AND (priority = Blocker OR labels = critical) AND resolution is EMPTY"
```

### Stale Issues

```bash
# Not updated in 30 days, still open
gojira search --jql "project = FOO AND updated <= -30d AND status not in (Done, Closed)"
```

## JQL Builder (SDK)

Use the SDK's JQL builder for programmatic query construction:

```go
import "github.com/grokify/go-atlassian"

jql := jira.JQL{
    ProjectsIncl: [][]string{{"FOO"}},
    TypesIncl:    [][]string{{"Bug"}},
    StatusesExcl: [][]string{{"Done", "Closed"}},
}

query := jql.String()
// project = 'FOO' AND type = 'Bug' AND status NOT IN ('Done', 'Closed')
```

See [JQL Builder](../sdk/jql.md) for more details.
