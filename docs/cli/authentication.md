# Authentication

go-atlassian supports multiple authentication methods to connect to your Jira instance.

## Authentication Priority

The CLI checks authentication sources in this order:

1. **CLI flags** (`--creds-file`, `--account`)
2. **Environment variables** (`JIRA_URL`, `JIRA_USER`, `JIRA_TOKEN`)
3. **goauth credentials file** (`~/.config/goauth/credentials.json`)

## Method 1: Environment Variables

Set these environment variables:

```bash
export JIRA_URL=https://your-instance.atlassian.net
export JIRA_USER=your-email@example.com
export JIRA_TOKEN=your-api-token
```

This is the recommended method for:

- Local development
- CI/CD pipelines
- Docker containers
- AI agent integrations

!!! tip "Non-interactive"
    Environment variables enable fully non-interactive operation, ideal for automation.

## Method 2: goauth Credentials File

go-atlassian integrates with [goauth](https://github.com/grokify/goauth) for credential management.

### Default Location

```
~/.config/goauth/credentials.json
```

### File Format

```json
{
  "credentials": {
    "jira-prod": {
      "service": "jira",
      "type": "basic",
      "basic": {
        "serverURL": "https://your-instance.atlassian.net",
        "username": "your-email@example.com",
        "password": "your-api-token"
      }
    },
    "jira-staging": {
      "service": "jira",
      "type": "basic",
      "basic": {
        "serverURL": "https://staging.atlassian.net",
        "username": "your-email@example.com",
        "password": "your-staging-token"
      }
    }
  }
}
```

### Using a Specific Account

```bash
# Use account key from default credentials file
gojira search --jql "project = FOO" --account jira-prod

# Use custom credentials file
gojira search --jql "project = FOO" --creds-file ~/my-creds.json --account myaccount
```

### Interactive Selection

If no `--account` is specified, goauth will interactively prompt you to select an account.

## Getting an API Token

### Jira Cloud

1. Go to [Atlassian API Tokens](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Click **Create API token**
3. Give it a descriptive name (e.g., "gojira CLI")
4. Copy the token immediately

### Jira Server / Data Center

Use your Jira password or create a personal access token in your Jira Server settings.

## Security Best Practices

!!! warning "Keep credentials secure"

    - Never commit credentials to version control
    - Use environment variables in CI/CD
    - Restrict API token permissions when possible
    - Rotate tokens periodically

### Using .envrc with direnv

For project-specific credentials, use [direnv](https://direnv.net/):

```bash
# .envrc (add to .gitignore)
export JIRA_URL=https://your-instance.atlassian.net
export JIRA_USER=your-email@example.com
export JIRA_TOKEN=your-api-token
```

Then run:

```bash
direnv allow
```

## Troubleshooting

### "authentication failed" Error

1. Verify your credentials are correct
2. Check that `JIRA_URL` includes `https://`
3. Ensure your API token hasn't expired
4. Verify you have access to the Jira project

### Testing Authentication

```bash
# Test with a simple query
gojira search --jql "project = FOO" --max 1

# Check version (no auth required)
gojira version
```
