# Troubleshooting Guide

Common issues and solutions when using Go-Atlassian.

## Authentication Issues

### Error: "authentication failed"

**Cause:** Invalid credentials or missing environment variables.

**Solutions:**

1. Verify environment variables are set:

   ```bash
   echo $JIRA_URL $JIRA_USER $JIRA_TOKEN
   ```

2. Check credentials are correct:

   ```bash
   # Test with curl
   curl -u "user@example.com:api-token" \
     "https://company.atlassian.net/rest/api/3/myself"
   ```

3. For Jira Cloud, ensure you're using an API token (not password):

   - Generate at: https://id.atlassian.com/manage-profile/security/api-tokens

### Error: "401 Unauthorized"

**Cause:** API token expired or revoked.

**Solution:** Generate a new API token and update your credentials.

### Error: "403 Forbidden"

**Cause:** Insufficient permissions for the operation.

**Solutions:**

1. Check your Jira permissions for the project
2. Verify the API token has the required scopes
3. Contact your Jira administrator

## Issue Operations

### Error: "issue not found" for valid issue key

**Causes:**

- Issue was deleted or moved
- No permission to view the issue
- Wrong Jira instance

**Solutions:**

1. Verify the issue exists in the Jira web UI
2. Check you're connected to the correct Jira instance
3. Verify your permissions on the project

### Error: "field not found" when creating issues

**Cause:** Custom field name doesn't match or field isn't available for the issue type.

**Solutions:**

1. List available fields for your project:

   ```bash
   gojira fields --project PROJ --issue-type Story
   ```

2. Use `--validate` to check before creating:

   ```bash
   gojira create -f issue.yaml --validate
   ```

3. Use field ID instead of name:

   ```yaml
   customfield_10001: value
   ```

### Error: "multiple custom fields found with name"

**Cause:** Duplicate field names in Jira (common after copying schemes).

**Solutions:**

1. Use field ID instead of name:

   ```bash
   gojira fields --project PROJ
   # Find the correct ID, then use it in your YAML
   ```

2. Check for duplicate fields:

   ```bash
   gojira fields --duplicates
   ```

## JQL Issues

### Error: "JQL parse error"

**Cause:** Invalid JQL syntax.

**Common mistakes:**

```bash
# Wrong: unquoted project name with spaces
gojira search "project = My Project"

# Correct: quote values with spaces
gojira search "project = 'My Project'"

# Wrong: invalid operator
gojira search "status <> Done"

# Correct: use != for not equal
gojira search "status != Done"
```

### Error: "field does not exist"

**Cause:** Using a field name that Jira doesn't recognize.

**Solutions:**

1. Check field names in Jira UI
2. Use clause names from `gojira fields`:

   ```bash
   gojira fields | grep -i sprint
   ```

3. Custom fields may need quotes:

   ```bash
   gojira search "'Epic Link' = PROJ-100"
   ```

## Performance Issues

### Slow search queries

**Causes:**

- Unbounded queries returning too many results
- Complex JQL with multiple OR conditions

**Solutions:**

1. Add filters to reduce results:

   ```bash
   # Add project filter
   gojira search "assignee = currentUser()" --project PROJ

   # Add date filter
   gojira search "project = PROJ AND updated >= -7d"
   ```

2. Use `--max` to limit results:

   ```bash
   gojira search "project = PROJ" --max 100
   ```

### Timeout errors

**Cause:** Large result sets or slow Jira server.

**Solutions:**

1. Add more specific filters to reduce results
2. Use pagination:

   ```bash
   gojira search "project = PROJ" --max 50
   ```

## Custom Fields

### Finding the correct custom field ID

```bash
# List all custom fields
gojira fields

# Filter by project
gojira fields --project PROJ

# Search for specific field
gojira fields | grep -i "epic"
```

### Custom field not appearing in results

**Cause:** Field not in the default fields list.

**Solution:** Specify the field in your JQL or use `--expand`:

```bash
gojira get PROJ-123 --expand fields
```

## MCP Server Issues

### Error: "missing required environment variables"

**Cause:** MCP server not configured with Jira credentials.

**Solution:** Set all required environment variables in your MCP config:

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

### MCP server not responding

**Solutions:**

1. Test the server manually:

   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | gojira-mcp
   ```

2. Check logs (stderr) for errors

3. Verify the binary is in your PATH:

   ```bash
   which gojira-mcp
   ```

## Getting Help

If you're still experiencing issues:

1. Enable debug logging:

   ```bash
   export GOJIRA_LOG_LEVEL=debug
   gojira search "project = PROJ"
   ```

2. Check the [GitHub Issues](https://github.com/grokify/go-atlassian/issues)

3. Open a new issue with:

   - gojira version (`gojira version`)
   - Command that failed
   - Full error message
   - Relevant environment info
