# Go-Atlassian Implementation Plan

This document contains detailed implementation specifications for planned features.

## Phase 1: CLI Commands

### 1.1 `gojira issue-types` Command

List available issue types for a project using the createmeta API.

**Usage:**

```bash
# List issue types for a project
gojira issue-types --project ABC

# Output as JSON
gojira issue-types --project ABC --json

# Output as table (default)
gojira issue-types --project ABC --table
```

**Output (table):**

```
ID      NAME        SUBTASK
10001   Bug         false
10002   Story       false
10003   Task        false
10004   Sub-task    true
```

**Output (JSON):**

```json
[
  {"id": "10001", "name": "Bug", "subtask": false},
  {"id": "10002", "name": "Story", "subtask": false}
]
```

**Implementation:**

- File: `cmd/gojira/issue_types.go`
- Uses: `client.CreateMetaAPI.GetIssueTypes()`

### 1.2 `gojira transitions` Command

Show available workflow transitions for an issue.

**Usage:**

```bash
# Show transitions for an issue
gojira transitions ISSUE-123

# Output as JSON
gojira transitions ISSUE-123 --json
```

**Output (table):**

```
ID    NAME              TO STATUS
11    Start Progress    In Progress
21    Resolve           Done
31    Close             Closed
```

**Implementation:**

- File: `cmd/gojira/transitions.go`
- Uses: `client.IssueAPI.GetTransitions()` (already exists in rest package)

### 1.3 `gojira link` Command

Create links between issues.

**Usage:**

```bash
# Link issues
gojira link ISSUE-1 blocks ISSUE-2
gojira link ISSUE-1 "is blocked by" ISSUE-2
gojira link ISSUE-1 relates-to ISSUE-2

# List available link types
gojira link --list-types
```

**Implementation:**

- File: `cmd/gojira/link.go`
- New: `rest/link_service.go` for link operations
- API: `POST /rest/api/3/issueLink`

### 1.4 `gojira create --validate` Flag

Validate issue YAML against createmeta before creating.

**Usage:**

```bash
# Validate without creating
gojira create -f issue.yaml --validate

# Dry run (existing) shows what would be created
gojira create -f issue.yaml --dry-run

# Validate then create
gojira create -f issue.yaml --validate --create
```

**Validation checks:**

- Project exists
- Issue type valid for project
- Required fields present
- Custom field IDs valid

**Implementation:**

- Modify: `cmd/gojira/create.go`
- New: `core/validate.go` for validation logic

## Phase 2: SDK Features

### 2.1 `GetRequiredFieldsForProject()`

Return only required fields for issue creation.

**API:**

```go
// Get required fields for any issue type in project
fields, err := client.CreateMetaAPI.GetRequiredFieldsForProject(ctx, "ABC")

// Get required fields for specific issue type
fields, err := client.CreateMetaAPI.GetRequiredFields(ctx, "ABC", "10001")
```

**Implementation:**

- Add to: `rest/createmeta_service.go`
- Filter `CreateMetaFields` by `Required: true`

### 2.2 Field Name Suggestions

Suggest similar field names when a field is not found.

**API:**

```go
field, suggestions, err := client.CustomFieldAPI.GetCustomFieldWithSuggestions("Modul")
// err: field not found
// suggestions: ["Module", "Module Type", "Modules"]
```

**Implementation:**

- Add to: `rest/customfield_service.go`
- Use Levenshtein distance or prefix matching
- Return top 3-5 suggestions

### 2.3 Custom Field Caching

Cache field metadata to reduce API calls.

**API:**

```go
// Enable caching (TTL: 5 minutes)
client.CustomFieldAPI.EnableCache(5 * time.Minute)

// Force refresh
client.CustomFieldAPI.RefreshCache()
```

**Implementation:**

- Add to: `rest/customfield_service.go`
- In-memory cache with TTL
- Thread-safe with sync.RWMutex

### 2.4 Bulk Transitions

Transition multiple issues at once.

**API:**

```go
results, err := client.IssueAPI.DoTransitionsBulk(ctx, []string{"FOO-1", "FOO-2"}, "Done")
// results contains success/failure per issue
```

**Implementation:**

- Add to: `rest/issue_service__transitions.go`
- Parallel execution with configurable concurrency
- Return detailed results per issue

## Phase 3: Testing

### 3.1 CustomFields Helper Tests

**File:** `rest/customfield_test.go`

```go
func TestCustomFields_MapNameToIDs(t *testing.T)
func TestCustomFields_MapIDToName(t *testing.T)
func TestCustomFields_DuplicateNames(t *testing.T)
func TestCustomFields_FilterByIDs(t *testing.T)
func TestCustomFields_FilterByNames(t *testing.T)
```

### 3.2 CreateMetaService Tests

**File:** `rest/createmeta_service_test.go`

```go
func TestCreateMetaService_GetIssueTypes(t *testing.T)
func TestCreateMetaService_GetFields(t *testing.T)
func TestCreateMetaService_GetAllFieldsForProject(t *testing.T)
func TestCreateMetaFields_CustomOnly(t *testing.T)
func TestCreateMetaFields_RequiredOnly(t *testing.T)
```

### 3.3 CustomFieldService Tests

**File:** `rest/customfield_service_test.go`

```go
func TestCustomFieldService_GetCustomFieldsByName(t *testing.T)
func TestCustomFieldService_GetCustomFieldByID(t *testing.T)
func TestCustomFieldService_GetCustomField_NotFound(t *testing.T)
func TestCustomFieldService_GetCustomField_MultiplFound(t *testing.T)
```

## Phase 4: Documentation

### 4.1 Troubleshooting Guide

**File:** `docs/guides/troubleshooting.md`

Topics:

- Authentication errors
- Permission denied errors
- Custom field not found
- Rate limiting
- SSL/TLS issues

### 4.2 Custom Fields Guide

**File:** `docs/guides/custom-fields.md`

Topics:

- Understanding custom field IDs
- Handling duplicate field names
- Project-specific fields
- Reading custom field values
- Writing custom field values
- Common custom field types

## Implementation Order

1. **High Priority (v0.33.0):**
   - `gojira issue-types` command
   - `gojira transitions` command
   - Unit tests for CustomFields helpers
   - Unit tests for CreateMetaService

2. **Medium Priority (v0.33.0):**
   - `gojira link` command
   - `gojira create --validate`
   - `GetRequiredFieldsForProject()`
   - Field name suggestions
   - Documentation guides

3. **Low Priority (v0.34.0+):**
   - Custom field caching
   - Bulk transitions
