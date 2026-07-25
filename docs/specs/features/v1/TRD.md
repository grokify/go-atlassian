# Go-Atlassian v1.0 Technical Requirements Document

## Overview

This document provides technical details and implementation guidance for the Go-Atlassian v1.0 release.

## Current State Analysis

### API Review Summary

| Aspect | Status | Notes |
|--------|--------|-------|
| ID/URL naming consistency | ✅ Good | All use `ID` not `Id`, proper camelCase |
| Pointer vs value fields | ✅ Good | Pointers for optional, values for required |
| Method receivers | ✅ Excellent | Pointer for services, value for collections |
| Service naming (API vs Service) | ⚠️ Inconsistent | Mixed terminology |
| Interface definitions | ❌ Missing | No interfaces defined |
| JSON tag format | ✅ Good | Consistent format |

### Codebase Statistics

- **Total Go files**: ~50 files
- **Total lines**: ~12,000 lines (non-test)
- **Packages**: root, jira, jira/apiv3, core, xml, mcpserver, web, cmd/*

## Technical Debt Items

### 1. Naming Inconsistencies

#### Service/API Terminology

**Current state**:

```go
// rest/client.go - Field uses "API" suffix
type Client struct {
    CustomFieldAPI *CustomFieldService  // Inconsistent naming
}

// File naming inconsistency
rest/transition_api.go      // Uses _api.go
rest/backlog_service.go     // Uses _service.go
rest/issue_service.go       // Uses _service.go
```

**Recommended fix**:

```go
// Option A: Rename field to match type
type Client struct {
    CustomFieldService *CustomFieldService
}

// Option B: Rename all to use API suffix (not recommended)
```

**Files to update**:

- `rest/client.go`: Rename `CustomFieldAPI` field
- `rest/transition_api.go`: Consider renaming to `rest/transition_service.go`

#### Config Struct Verbosity

**Current state** (`stage.go`):

```go
type StageConfig struct {
    StageNamePlanning            string
    StageNameDesign              string
    StageNameDevelopment         string
    MetaStageInPlanning          string
    MetaStageReadyForPlanning    string
    // ... 20+ fields with repetitive prefixes
}
```

**Recommended refactor** (P2 - post v1):

```go
type StageConfig struct {
    Stages struct {
        Planning    string
        Design      string
        Development string
        Testing     string
        Done        string
    }
    MetaStages struct {
        InPlanning          string
        ReadyForPlanning    string
        // ...
    }
    Prefixes struct {
        In       string
        ReadyFor string
    }
}
```

### 2. Missing Interfaces

**Current state**: No interfaces defined in any package.

**Recommended interfaces** (P2 - post v1):

```go
// rest/interfaces.go

// IssueGetter defines issue retrieval operations.
type IssueGetter interface {
    Issue(ctx context.Context, key string, opts *GetQueryOptions) (*jira.Issue, error)
    Issues(ctx context.Context, keys []string, opts *GetQueryOptions) (Issues, error)
}

// IssueSearcher defines issue search operations.
type IssueSearcher interface {
    SearchIssues(jql string, retrieveAll bool) (Issues, error)
    SearchIssuesPages(jql string, startAt, maxResults, maxPages int) (Issues, error)
}

// CommentGetter defines comment retrieval operations.
type CommentGetter interface {
    GetComments(ctx context.Context, issueKey string, maxResults int) (*CommentsResponse, error)
}
```

**Benefits**:

- Enables unit testing with mocks
- Supports dependency injection
- Documents expected behavior

### 3. Custom Field Type Clarity

**Current representations**:

| Type | Package | Purpose |
|------|---------|---------|
| `CustomField` | rest | Field definition (id, name, schema) |
| `CustomFields` | rest | Slice of CustomField with helper methods |
| `IssueCustomField` | rest | Field value on an issue (id, self, value) |
| `CustomFieldOption` | rest/apiv3 | Dropdown/select option |
| `CustomFieldID` | root | Type alias for int (legacy?) |

**Recommendation**: Add documentation comments clarifying when to use each type.

### 4. String Literal Bug

**Location**: `rest/backlog_service.go:63`

```go
// Current (buggy)
u.Add(ParamValidateQuery, "false) ")

// Should be
u.Add(ParamValidateQuery, "false")
```

## Implementation Plan

### Phase 1: Bug Fixes (v0.32.0)

1. Fix `backlog_service.go` string literal
2. Rename `CustomFieldAPI` → `CustomFieldService` field
3. Add documentation for custom field types

### Phase 2: Consistency (v0.33.0)

1. Rename `transition_api.go` → `transition_service.go`
2. Audit and fix any remaining naming inconsistencies
3. Add package-level documentation

### Phase 3: v1.0.0 Release

1. Final API review
2. Update CHANGELOG with migration guide
3. Tag v1.0.0

### Phase 4: Post-v1 Improvements

1. Define service interfaces
2. Refactor verbose config structs
3. Increase test coverage

## Breaking Changes Summary

| Change | Migration |
|--------|-----------|
| `ReadyforPlanningName()` → `ReadyForPlanningName()` | Find/replace in calling code |
| `ReadyForDevlopmentName()` → `ReadyForDevelopmentName()` | Find/replace in calling code |
| `TransitionsAPIReponse` → `TransitionsAPIResponse` | Find/replace in calling code |
| `Client.CustomFieldAPI` → `Client.CustomFieldService` | Find/replace in calling code |

## Testing Requirements

Before v1.0:

- [ ] All existing tests pass
- [ ] Build succeeds on Go 1.21+
- [ ] Linting passes (`golangci-lint run`)
- [ ] Example code compiles

## References

- [PRD.md](PRD.md) - Product Requirements Document
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
