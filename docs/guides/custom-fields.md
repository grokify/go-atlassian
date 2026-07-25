# Working with Custom Fields

Custom fields are user-defined fields in Jira that extend the default issue schema. This guide covers how to discover, use, and manage custom fields with go-atlassian.

## Understanding Custom Field IDs

Every custom field in Jira has a unique ID in the format `customfield_NNNNN` (e.g., `customfield_10001`). While Jira displays friendly names like "Story Points" or "Epic Link", the API uses these IDs internally.

**Important:** Multiple fields can share the same display name (e.g., after copying project schemes). Always verify field IDs when working with custom fields.

## Discovering Custom Fields

### List All Custom Fields

```bash
# List all custom fields
gojira fields

# Output as JSON for programmatic use
gojira fields --json
```

### Filter by Project

```bash
# Fields available in a specific project
gojira fields --project PROJ

# Fields for a specific issue type in a project
gojira fields --project PROJ --issue-type Story
```

### Find Required Fields

```bash
# Show only required fields for issue creation
gojira fields --project PROJ --issue-type Bug --required
```

### Identify Duplicate Names

```bash
# Find fields with duplicate names
gojira fields --duplicates
```

## Using Custom Fields in JQL

Custom fields in JQL can be referenced by:

1. **Clause name** (if defined): `"Epic Link" = PROJ-100`
2. **Field ID**: `cf[10001] = "value"`

```bash
# Search by Epic Link
gojira search "'Epic Link' = PROJ-100"

# Search by Sprint
gojira search "Sprint = 'Sprint 5'"

# Search by Story Points range
gojira search "'Story Points' >= 5 AND 'Story Points' <= 13"
```

## Custom Fields in Issue Creation

### YAML Format

When creating issues, use the exact field ID in your YAML file:

```yaml
project: PROJ
type: Story
summary: Add user authentication
description: |
  Implement OAuth2 login flow.

# Custom fields use their IDs
customfield_10001: 8              # Story Points
customfield_10002: PROJ-100       # Epic Link
customfield_10003:                # Sprint (array type)
  - name: "Sprint 5"
customfield_10004: "High"         # Priority dropdown
customfield_10005:                # Multi-select
  - value: "Option A"
  - value: "Option B"
```

### Validate Before Creating

Use `--validate` to check your custom fields against the Jira API:

```bash
gojira create -f story.yaml --validate
```

This will report:

- Missing required fields
- Unknown field IDs
- Available fields for the issue type

### Field Type Formats

Different custom field types require different value formats:

| Type | Format | Example |
|------|--------|---------|
| Text | String | `"Simple text"` |
| Number | Number | `8` or `3.5` |
| Select | String | `"Option Name"` |
| Multi-select | Array | `[{value: "A"}, {value: "B"}]` |
| User | Object | `{name: "username"}` |
| Cascading Select | Object | `{value: "Parent", child: {value: "Child"}}` |
| Date | String | `"2024-01-15"` |
| DateTime | String | `"2024-01-15T10:30:00.000+0000"` |

## SDK Usage

### Get Custom Fields

```go
import "github.com/grokify/go-atlassian/jira"

// Get all custom fields
fields, err := client.CustomFieldAPI.GetCustomFields()
if err != nil {
    return err
}

// Filter by name
epicLinkFields := fields.FilterByNames("Epic Link")

// Get by exact ID
field, err := client.CustomFieldAPI.GetCustomFieldByID("customfield_10001")
```

### Find Fields with Suggestions

```go
// When a field name doesn't match exactly, get suggestions
field, suggestions := fields.FindByNameWithSuggestions("Sprnt")
if field == nil {
    fmt.Println("Did you mean:")
    for _, s := range suggestions {
        fmt.Printf("  - %s (%s)\n", s.Name, s.ID)
    }
}
```

### Get Project-Specific Fields

```go
ctx := context.Background()

// Get fields available in a specific project
projectFields, err := client.CustomFieldAPI.GetCustomFieldsForProject(ctx, "PROJ")
if err != nil {
    return err
}

// Get required fields for an issue type
requiredFields, err := client.CreateMetaAPI.GetRequiredFields(ctx, "PROJ", "10001")
```

### Map Names to IDs

```go
// Handle duplicate field names
nameToIDs := fields.MapNameToIDs()
// nameToIDs["Sprint"] might return ["customfield_10001", "customfield_10004"]

// Find duplicates
duplicates := fields.DuplicateNames()
// ["Sprint", "Module"] - fields with multiple IDs
```

## Best Practices

### 1. Use Field IDs, Not Names

Field names can change or have duplicates. Always use IDs in automation:

```yaml
# Good - stable
customfield_10001: 8

# Risky - could break if renamed or duplicated
# "Story Points": 8
```

### 2. Validate Field Availability

Fields vary by project and issue type. Always check availability:

```bash
# Before creating a Bug in PROJ
gojira fields --project PROJ --issue-type Bug
```

### 3. Handle Duplicates Gracefully

```bash
# Check for duplicates before automation
gojira fields --duplicates

# If duplicates exist, use IDs explicitly
```

### 4. Cache Field Metadata

In SDK code, create the client with custom field caching:

```go
// Load custom fields once during client creation
client, err := jira.NewClientFromBasicAuth(url, user, token, true)
// client.CustomFieldSet is now populated
```

### 5. Document Your Field Mappings

Create a reference document mapping field names to IDs for your team:

| Display Name | Field ID | Type | Required |
|--------------|----------|------|----------|
| Story Points | customfield_10001 | Number | No |
| Epic Link | customfield_10002 | Issue Link | No |
| Sprint | customfield_10003 | Sprint | No |

## Common Issues

### "Field not found" Error

The field might not be available for the issue type or project:

```bash
# Check available fields
gojira fields --project PROJ --issue-type Story
```

### Multiple Fields with Same Name

Use the field ID instead of the name:

```bash
# Find the correct ID
gojira fields | grep -i "sprint"
# Use the ID in your YAML
```

### Wrong Value Format

Check the field type and format your value accordingly. Use `--validate` to catch errors before creating issues.

## Related Documentation

- [CLI Fields Command](../cli/fields.md)
- [Creating Issues](../cli/create.md)
- [JQL Examples](jql-examples.md)
- [Troubleshooting](troubleshooting.md)
