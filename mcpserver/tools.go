package mcpserver

// GetTools returns all available Jira tools.
func GetTools() []Tool {
	return []Tool{
		{
			Name:        "jira_get_issue",
			Description: "Get a Jira issue by key with all fields including description, status, assignee, and custom fields",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key": map[string]any{
						"type":        "string",
						"description": "Issue key (e.g., PROJ-123)",
					},
					"expand": map[string]any{
						"type":        "string",
						"description": "Comma-separated list of fields to expand (e.g., changelog,renderedFields)",
					},
				},
				"required": []string{"key"},
			},
		},
		{
			Name:        "jira_search",
			Description: "Search Jira issues using JQL (Jira Query Language). Returns matching issues with key fields.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"jql": map[string]any{
						"type":        "string",
						"description": "JQL query string (e.g., 'project = PROJ AND status = Open')",
					},
					"max_results": map[string]any{
						"type":        "integer",
						"description": "Maximum number of results to return (default: 50, max: 100)",
						"default":     50,
					},
					"fields": map[string]any{
						"type":        "string",
						"description": "Comma-separated list of fields to return (default: key,summary,status,assignee,created,updated)",
					},
				},
				"required": []string{"jql"},
			},
		},
		{
			Name:        "jira_update_issue",
			Description: "Update a Jira issue's fields such as summary, description, labels, or custom fields",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key": map[string]any{
						"type":        "string",
						"description": "Issue key (e.g., PROJ-123)",
					},
					"summary": map[string]any{
						"type":        "string",
						"description": "New summary/title for the issue",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "New description for the issue",
					},
					"labels": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Labels to set on the issue (replaces existing labels)",
					},
					"add_labels": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Labels to add to the issue (preserves existing labels)",
					},
					"remove_labels": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Labels to remove from the issue",
					},
				},
				"required": []string{"key"},
			},
		},
		{
			Name:        "jira_add_comment",
			Description: "Add a comment to a Jira issue",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key": map[string]any{
						"type":        "string",
						"description": "Issue key (e.g., PROJ-123)",
					},
					"body": map[string]any{
						"type":        "string",
						"description": "Comment body text",
					},
				},
				"required": []string{"key", "body"},
			},
		},
		{
			Name:        "jira_get_transitions",
			Description: "Get available status transitions for a Jira issue",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key": map[string]any{
						"type":        "string",
						"description": "Issue key (e.g., PROJ-123)",
					},
				},
				"required": []string{"key"},
			},
		},
		{
			Name:        "jira_transition_issue",
			Description: "Transition a Jira issue to a new status. Use either transition name or ID.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key": map[string]any{
						"type":        "string",
						"description": "Issue key (e.g., PROJ-123)",
					},
					"transition": map[string]any{
						"type":        "string",
						"description": "Transition name (e.g., 'In Progress', 'Done') or ID. Use jira_get_transitions to see available options.",
					},
					"comment": map[string]any{
						"type":        "string",
						"description": "Optional comment to add with the transition",
					},
				},
				"required": []string{"key", "transition"},
			},
		},
		{
			Name:        "jira_get_comments",
			Description: "Get comments on a Jira issue",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key": map[string]any{
						"type":        "string",
						"description": "Issue key (e.g., PROJ-123)",
					},
					"max_results": map[string]any{
						"type":        "integer",
						"description": "Maximum number of comments to return (default: 50)",
						"default":     50,
					},
				},
				"required": []string{"key"},
			},
		},
		{
			Name:        "jira_get_projects",
			Description: "List available Jira projects",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "jira_create_issue",
			Description: "Create a new Jira issue (Story, Bug, Task, etc.) with support for custom fields",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"project": map[string]any{
						"type":        "string",
						"description": "Project key (e.g., PROJ)",
					},
					"type": map[string]any{
						"type":        "string",
						"description": "Issue type (e.g., Story, Bug, Task, Epic)",
					},
					"summary": map[string]any{
						"type":        "string",
						"description": "Issue summary/title",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "Issue description (supports Jira markdown)",
					},
					"parent": map[string]any{
						"type":        "string",
						"description": "Parent issue key for subtasks or stories under epics (e.g., PROJ-100)",
					},
					"labels": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Labels to apply to the issue",
					},
					"priority": map[string]any{
						"type":        "string",
						"description": "Priority name (e.g., High, Medium, Low)",
					},
					"assignee": map[string]any{
						"type":        "string",
						"description": "Assignee username or email",
					},
					"components": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Component names",
					},
					"custom_fields": map[string]any{
						"type":        "object",
						"description": "Custom fields as key-value pairs (e.g., {\"customfield_12345\": \"value\"})",
					},
				},
				"required": []string{"project", "type", "summary"},
			},
		},
		{
			Name:        "jira_clone_issue",
			Description: "Clone a Jira issue with configurable field mapping. Creates a copy with optionally modified fields.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"source_key": map[string]any{
						"type":        "string",
						"description": "Source issue key to clone (e.g., PROJ-123)",
					},
					"target_project": map[string]any{
						"type":        "string",
						"description": "Target project key (default: same as source)",
					},
					"target_type": map[string]any{
						"type":        "string",
						"description": "Target issue type (default: same as source)",
					},
					"summary_prefix": map[string]any{
						"type":        "string",
						"description": "Prefix to add to cloned issue summary (e.g., '[Clone] ')",
					},
					"summary_suffix": map[string]any{
						"type":        "string",
						"description": "Suffix to add to cloned issue summary",
					},
					"link_to_original": map[string]any{
						"type":        "boolean",
						"description": "Create a link to the original issue (default: false)",
					},
					"link_type": map[string]any{
						"type":        "string",
						"description": "Link type name when linking to original (default: 'Cloners')",
					},
					"exclude_fields": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Fields to exclude from cloning (e.g., labels, components)",
					},
					"parent": map[string]any{
						"type":        "string",
						"description": "Parent issue key for subtasks (e.g., PROJ-100)",
					},
				},
				"required": []string{"source_key"},
			},
		},
		{
			Name:        "jira_bulk_update",
			Description: "Update multiple Jira issues with the same field changes",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"issue_keys": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Issue keys to update (e.g., [\"PROJ-1\", \"PROJ-2\"])",
					},
					"jql": map[string]any{
						"type":        "string",
						"description": "JQL query to select issues (alternative to issue_keys)",
					},
					"add_labels": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Labels to add to all issues",
					},
					"remove_labels": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Labels to remove from all issues",
					},
					"comment": map[string]any{
						"type":        "string",
						"description": "Comment to add to all issues",
					},
				},
			},
		},
		{
			Name:        "jira_get_boards",
			Description: "List Jira agile boards with optional filtering",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"project": map[string]any{
						"type":        "string",
						"description": "Filter boards by project key",
					},
					"type": map[string]any{
						"type":        "string",
						"description": "Filter boards by type (scrum, kanban)",
					},
				},
			},
		},
		{
			Name:        "jira_get_sprints",
			Description: "List sprints for a Jira agile board",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"board_id": map[string]any{
						"type":        "integer",
						"description": "Board ID to get sprints for",
					},
					"state": map[string]any{
						"type":        "string",
						"description": "Filter by sprint state (active, closed, future)",
					},
				},
				"required": []string{"board_id"},
			},
		},
		{
			Name:        "jira_velocity_report",
			Description: "Get velocity report for a Jira agile board showing story points per sprint",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"board_id": map[string]any{
						"type":        "integer",
						"description": "Board ID to get velocity for",
					},
					"sprint_count": map[string]any{
						"type":        "integer",
						"description": "Number of sprints to include (default: all closed sprints)",
					},
					"include_active": map[string]any{
						"type":        "boolean",
						"description": "Include active sprint in calculations",
					},
					"story_points_field": map[string]any{
						"type":        "string",
						"description": "Custom field ID for story points (default: customfield_10016)",
					},
				},
				"required": []string{"board_id"},
			},
		},
		{
			Name:        "jira_burndown_report",
			Description: "Get burndown report for a sprint showing remaining vs completed work",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"sprint_id": map[string]any{
						"type":        "integer",
						"description": "Sprint ID to get burndown for",
					},
					"story_points_field": map[string]any{
						"type":        "string",
						"description": "Custom field ID for story points (default: customfield_10016)",
					},
				},
				"required": []string{"sprint_id"},
			},
		},
		{
			Name:        "jira_worklog_report",
			Description: "Get worklog summary report showing time spent by author and issue",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"jql": map[string]any{
						"type":        "string",
						"description": "JQL query to filter issues",
					},
					"sprint_id": map[string]any{
						"type":        "integer",
						"description": "Sprint ID to filter issues (alternative to jql)",
					},
				},
			},
		},
		{
			Name:        "jira_cycle_time_report",
			Description: "Get cycle time analysis for resolved issues showing time from creation to resolution",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"jql": map[string]any{
						"type":        "string",
						"description": "JQL query to filter issues",
					},
					"sprint_id": map[string]any{
						"type":        "integer",
						"description": "Sprint ID to filter issues (alternative to jql)",
					},
				},
			},
		},
		{
			Name:        "jira_move_to_sprint",
			Description: "Move issues to a sprint for sprint planning",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"sprint_id": map[string]any{
						"type":        "integer",
						"description": "Target sprint ID",
					},
					"issue_keys": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Issue keys to move to the sprint",
					},
				},
				"required": []string{"sprint_id", "issue_keys"},
			},
		},
	}
}
