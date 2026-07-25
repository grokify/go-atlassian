# GoJira SDK Refactor Tasks

Breaking changes are acceptable. Goal: simplify structure, maintain dependency separation, clearer API.

## Design Principles

1. **Dependency Separation**: Root `gojira/` package stays lightweight with no go-jira SDK dependency
2. **Keep Split Files**: Files like `issues_set__*.go` remain split for easier human searching
3. **Symmetric Naming**: Use `rest/` and `xml/` as subpackages

## Target Structure

```
gojira/
├── jql.go                 # JQL builder (no deps)
├── jqls.go                # JQL helpers (no deps)
├── config.go              # Configuration (no deps)
├── constants.go           # Constants (no deps)
├── status.go              # Status types (no deps)
├── stage.go               # Stage types (no deps)
├── customfield.go         # Custom field types (no deps)
├── keys.go                # Issue key helpers (no deps)
├── rest/                  # REST API client (renamed from jirarest/)
│   ├── client.go          # Main client
│   ├── issue.go           # Issue types
│   ├── issue_more.go      # IssueMore wrapper
│   ├── issue_meta.go      # IssueMeta type
│   ├── issue_service.go   # Issue operations
│   ├── issue_service__*.go # Issue service methods (keep split)
│   ├── issues.go          # Issues slice type
│   ├── issues_set.go      # IssuesSet type
│   ├── issues_set__*.go   # IssuesSet methods (keep split)
│   ├── customfield.go     # Custom field service
│   ├── customfield_set.go # CustomFieldSet
│   ├── backlog_service.go # Backlog operations
│   ├── transition.go      # Transition types
│   ├── transition_api.go  # Transition API
│   ├── errors.go          # Error types
│   └── apiv3/             # V3 API models
│       └── ...
├── xml/                   # XML parsing (renamed from jiraxml/)
│   └── ...
├── web/                   # URL helpers (renamed from jiraweb/)
│   └── url.go
└── cmd/gojira/            # CLI (keep as-is)
```

## Phase 1: Rename Packages

- [x] Rename `jirarest/` to `rest/`
- [x] Update all imports from `gojira/jirarest` to `gojira/rest`
- [x] Rename `jiraxml/` to `xml/`
- [x] Update all imports from `gojira/jiraxml` to `gojira/xml`
- [x] Rename `jiraweb/` to `web/`
- [x] Update all imports from `gojira/jiraweb` to `gojira/web`

## Phase 2: Update CLI and cmd/ directories

- [x] Update `cmd/gojira/` imports
- [x] Update all other `cmd/*/` imports

## Phase 3: Cleanup

- [x] Run `go mod tidy`
- [x] Run full test suite
- [x] Run linter
- [x] Update documentation

## Migration Guide (for users)

```go
// Before
import "github.com/grokify/go-atlassian/jirarest"
client, err := jirajira.NewClientFromBasicAuth(...)

// After
import "github.com/grokify/go-atlassian/rest"
client, err := jira.NewClientFromBasicAuth(...)
```

```go
// Before
import "github.com/grokify/go-atlassian/jiraxml"

// After
import "github.com/grokify/go-atlassian/xml"
```

```go
// Before
import "github.com/grokify/go-atlassian/jiraweb"

// After
import "github.com/grokify/go-atlassian/web"
```

## Notes

- Breaking changes are acceptable
- Root `gojira/` package remains lightweight (no go-jira dependency)
- Split files (e.g., `issues_set__*.go`) are kept for easier human searching
- `rest/apiv3/` stays as subpackage (V3 models are distinct)
- `xml/` stays as subpackage (optional functionality for non-API access)
