package mcpserver

import (
	"context"
	"fmt"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/grokify/go-atlassian/core"
	"github.com/grokify/go-atlassian/jira"
)

// CallTool dispatches a tool call to the appropriate handler.
func (s *Server) CallTool(ctx context.Context, name string, args map[string]any) (any, error) {
	switch name {
	case "jira_get_issue":
		return s.handleGetIssue(ctx, args)
	case "jira_search":
		return s.handleSearch(ctx, args)
	case "jira_update_issue":
		return s.handleUpdateIssue(ctx, args)
	case "jira_add_comment":
		return s.handleAddComment(ctx, args)
	case "jira_get_transitions":
		return s.handleGetTransitions(ctx, args)
	case "jira_transition_issue":
		return s.handleTransitionIssue(ctx, args)
	case "jira_get_comments":
		return s.handleGetComments(ctx, args)
	case "jira_get_projects":
		return s.handleGetProjects(ctx, args)
	case "jira_create_issue":
		return s.handleCreateIssue(ctx, args)
	case "jira_clone_issue":
		return s.handleCloneIssue(ctx, args)
	case "jira_bulk_update":
		return s.handleBulkUpdate(ctx, args)
	case "jira_get_boards":
		return s.handleGetBoards(ctx, args)
	case "jira_get_sprints":
		return s.handleGetSprints(ctx, args)
	case "jira_velocity_report":
		return s.handleVelocityReport(ctx, args)
	case "jira_burndown_report":
		return s.handleBurndownReport(ctx, args)
	case "jira_worklog_report":
		return s.handleWorklogReport(ctx, args)
	case "jira_cycle_time_report":
		return s.handleCycleTimeReport(ctx, args)
	case "jira_move_to_sprint":
		return s.handleMoveToSprint(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *Server) handleGetIssue(ctx context.Context, args map[string]any) (any, error) {
	key, ok := args["key"].(string)
	if !ok || key == "" {
		return nil, fmt.Errorf("key is required")
	}

	var opts *jira.GetQueryOptions
	if expand, ok := args["expand"].(string); ok && expand != "" {
		opts = &jira.GetQueryOptions{
			ExpandChangelog: expand == "changelog" || expand == "all",
		}
	}

	issue, err := s.client.IssueAPI.Issue(ctx, key, opts)
	if err != nil {
		return nil, fmt.Errorf("get issue %s: %w", key, err)
	}

	return jira.ToIssueOutput(issue), nil
}

func (s *Server) handleSearch(ctx context.Context, args map[string]any) (any, error) {
	jql, ok := args["jql"].(string)
	if !ok || jql == "" {
		return nil, fmt.Errorf("jql is required")
	}

	maxResults := 50
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
		if maxResults > 100 {
			maxResults = 100
		}
	}

	// Use the context-aware V3 API for search
	issues, err := s.client.IssueAPI.SearchIssuesAPIV3(ctx, jql, false)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Limit results
	if len(issues) > maxResults {
		issues = issues[:maxResults]
	}

	results := jira.ToIssueOutputs(issues)

	return map[string]any{
		"total":  len(results),
		"issues": results,
	}, nil
}

func (s *Server) handleUpdateIssue(ctx context.Context, args map[string]any) (any, error) {
	key, ok := args["key"].(string)
	if !ok || key == "" {
		return nil, fmt.Errorf("key is required")
	}

	// Build update request
	updateBody := jira.IssuePatchRequestBody{}
	hasUpdate := false

	// Handle label operations
	if addLabels, ok := args["add_labels"].([]any); ok && len(addLabels) > 0 {
		if updateBody.Update == nil {
			updateBody.Update = &jira.IssuePatchRequestBodyUpdate{}
		}
		for _, label := range addLabels {
			if labelStr, ok := label.(string); ok {
				labelCopy := labelStr
				updateBody.Update.Labels = append(updateBody.Update.Labels, jira.IssuePatchRequestBodyUpdateLabel{
					Add: &labelCopy,
				})
			}
		}
		hasUpdate = true
	}

	if removeLabels, ok := args["remove_labels"].([]any); ok && len(removeLabels) > 0 {
		if updateBody.Update == nil {
			updateBody.Update = &jira.IssuePatchRequestBodyUpdate{}
		}
		for _, label := range removeLabels {
			if labelStr, ok := label.(string); ok {
				labelCopy := labelStr
				updateBody.Update.Labels = append(updateBody.Update.Labels, jira.IssuePatchRequestBodyUpdateLabel{
					Remove: &labelCopy,
				})
			}
		}
		hasUpdate = true
	}

	// Handle summary and description through the SDK's update method
	if summary, ok := args["summary"].(string); ok && summary != "" {
		if updateBody.Fields == nil {
			updateBody.Fields = make(map[string]jira.IssuePatchRequestBodyField)
		}
		updateBody.Fields["summary"] = jira.IssuePatchRequestBodyField{Value: summary}
		hasUpdate = true
	}

	if description, ok := args["description"].(string); ok && description != "" {
		if updateBody.Fields == nil {
			updateBody.Fields = make(map[string]jira.IssuePatchRequestBodyField)
		}
		updateBody.Fields["description"] = jira.IssuePatchRequestBodyField{Value: description}
		hasUpdate = true
	}

	if !hasUpdate {
		return nil, fmt.Errorf("no update fields provided")
	}

	_, err := s.client.IssueAPI.IssuePatch(ctx, key, updateBody)
	if err != nil {
		return nil, fmt.Errorf("update issue %s: %w", key, err)
	}

	return map[string]any{
		"success": true,
		"key":     key,
		"message": "Issue updated successfully",
	}, nil
}

func (s *Server) handleAddComment(ctx context.Context, args map[string]any) (any, error) {
	key, ok := args["key"].(string)
	if !ok || key == "" {
		return nil, fmt.Errorf("key is required")
	}

	body, ok := args["body"].(string)
	if !ok || body == "" {
		return nil, fmt.Errorf("body is required")
	}

	comment := &gojira.Comment{
		Body: body,
	}

	addedComment, _, err := s.client.JiraClient.Issue.AddCommentWithContext(ctx, key, comment)
	if err != nil {
		return nil, fmt.Errorf("add comment to %s: %w", key, err)
	}

	return map[string]any{
		"success":    true,
		"key":        key,
		"comment_id": addedComment.ID,
		"message":    "Comment added successfully",
	}, nil
}

func (s *Server) handleGetTransitions(ctx context.Context, args map[string]any) (any, error) {
	key, ok := args["key"].(string)
	if !ok || key == "" {
		return nil, fmt.Errorf("key is required")
	}

	transitions, _, err := s.client.IssueAPI.GetTransitions(ctx, key, false)
	if err != nil {
		return nil, fmt.Errorf("get transitions for %s: %w", key, err)
	}

	results := make([]map[string]any, 0, len(transitions))
	for _, t := range transitions {
		results = append(results, map[string]any{
			"id":   t.ID,
			"name": t.Name,
			"to":   t.To.Name,
		})
	}

	return map[string]any{
		"key":         key,
		"transitions": results,
	}, nil
}

func (s *Server) handleTransitionIssue(ctx context.Context, args map[string]any) (any, error) {
	key, ok := args["key"].(string)
	if !ok || key == "" {
		return nil, fmt.Errorf("key is required")
	}

	// Support both "transition" (new) and "transition_id" (legacy) parameters
	transition, ok := args["transition"].(string)
	if !ok || transition == "" {
		// Try legacy parameter name
		transition, ok = args["transition_id"].(string)
		if !ok || transition == "" {
			return nil, fmt.Errorf("transition is required (name or ID)")
		}
	}

	var opts *jira.TransitionOptions
	if comment, ok := args["comment"].(string); ok && comment != "" {
		opts = &jira.TransitionOptions{
			Comment: comment,
		}
	}

	result, err := s.client.IssueAPI.TransitionIssue(ctx, key, transition, opts)
	if err != nil {
		return nil, fmt.Errorf("transition issue %s: %w", key, err)
	}

	return map[string]any{
		"success":     true,
		"key":         result.Key,
		"from_status": result.FromStatus,
		"to_status":   result.ToStatus,
		"transition":  result.Transition,
		"message":     fmt.Sprintf("Issue transitioned from %s to %s", result.FromStatus, result.ToStatus),
	}, nil
}

func (s *Server) handleGetComments(ctx context.Context, args map[string]any) (any, error) {
	key, ok := args["key"].(string)
	if !ok || key == "" {
		return nil, fmt.Errorf("key is required")
	}

	maxResults := 50
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// Use shared GetComments method
	response, err := s.client.GetComments(ctx, key, maxResults)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Server) handleGetProjects(ctx context.Context, _ map[string]any) (any, error) {
	projects, _, err := s.client.JiraClient.Project.GetListWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get projects: %w", err)
	}

	results := make([]map[string]any, 0, len(*projects))
	for _, p := range *projects {
		results = append(results, map[string]any{
			"key":  p.Key,
			"name": p.Name,
			"id":   p.ID,
		})
	}

	return map[string]any{
		"total":    len(results),
		"projects": results,
	}, nil
}

func (s *Server) handleCreateIssue(ctx context.Context, args map[string]any) (any, error) {
	// Build IssueInput from args
	input := &core.IssueInput{}

	// Required fields
	project, ok := args["project"].(string)
	if !ok || project == "" {
		return nil, fmt.Errorf("project is required")
	}
	input.Project = project

	issueType, ok := args["type"].(string)
	if !ok || issueType == "" {
		return nil, fmt.Errorf("type is required")
	}
	input.Type = issueType

	summary, ok := args["summary"].(string)
	if !ok || summary == "" {
		return nil, fmt.Errorf("summary is required")
	}
	input.Summary = summary

	// Optional fields
	if description, ok := args["description"].(string); ok {
		input.Description = description
	}

	if parent, ok := args["parent"].(string); ok {
		input.Parent = parent
	}

	if priority, ok := args["priority"].(string); ok {
		input.Priority = priority
	}

	if assignee, ok := args["assignee"].(string); ok {
		input.Assignee = assignee
	}

	// Handle labels array
	if labels, ok := args["labels"].([]any); ok {
		for _, label := range labels {
			if labelStr, ok := label.(string); ok {
				input.Labels = append(input.Labels, labelStr)
			}
		}
	}

	// Handle components array
	if components, ok := args["components"].([]any); ok {
		for _, comp := range components {
			if compStr, ok := comp.(string); ok {
				input.Components = append(input.Components, compStr)
			}
		}
	}

	// Handle custom fields
	if customFields, ok := args["custom_fields"].(map[string]any); ok {
		input.CustomFields = customFields
	}

	// Create the issue using core package
	result, err := core.CreateIssue(ctx, s.client, input)
	if err != nil {
		return nil, fmt.Errorf("create issue: %w", err)
	}

	return map[string]any{
		"success": true,
		"key":     result.Key,
		"id":      result.ID,
		"self":    result.Self,
		"summary": result.Summary,
		"message": fmt.Sprintf("Issue %s created successfully", result.Key),
	}, nil
}

func (s *Server) handleCloneIssue(ctx context.Context, args map[string]any) (any, error) {
	sourceKey, ok := args["source_key"].(string)
	if !ok || sourceKey == "" {
		return nil, fmt.Errorf("source_key is required")
	}

	opts := &jira.CloneOptions{}

	if targetProject, ok := args["target_project"].(string); ok {
		opts.TargetProject = targetProject
	}
	if targetType, ok := args["target_type"].(string); ok {
		opts.TargetType = targetType
	}
	if summaryPrefix, ok := args["summary_prefix"].(string); ok {
		opts.SummaryPrefix = summaryPrefix
	}
	if summarySuffix, ok := args["summary_suffix"].(string); ok {
		opts.SummarySuffix = summarySuffix
	}
	if linkToOriginal, ok := args["link_to_original"].(bool); ok {
		opts.LinkToOriginal = linkToOriginal
	}
	if linkType, ok := args["link_type"].(string); ok {
		opts.LinkType = linkType
	}
	if parent, ok := args["parent"].(string); ok {
		opts.Parent = parent
	}

	// Handle exclude_fields array
	if excludeFields, ok := args["exclude_fields"].([]any); ok {
		for _, field := range excludeFields {
			if fieldStr, ok := field.(string); ok {
				opts.ExcludeFields = append(opts.ExcludeFields, fieldStr)
			}
		}
	}

	result, err := s.client.IssueAPI.CloneIssue(ctx, sourceKey, opts)
	if err != nil {
		return nil, fmt.Errorf("clone issue %s: %w", sourceKey, err)
	}

	return map[string]any{
		"success":    true,
		"source_key": result.SourceKey,
		"cloned_key": result.ClonedKey,
		"linked":     result.Linked,
		"message":    fmt.Sprintf("Issue %s cloned to %s", result.SourceKey, result.ClonedKey),
	}, nil
}

func (s *Server) handleBulkUpdate(ctx context.Context, args map[string]any) (any, error) {
	var issueKeys []string

	// Get issue keys from direct array
	if keys, ok := args["issue_keys"].([]any); ok {
		for _, key := range keys {
			if keyStr, ok := key.(string); ok {
				issueKeys = append(issueKeys, keyStr)
			}
		}
	}

	// Get issue keys from JQL
	if jql, ok := args["jql"].(string); ok && jql != "" {
		issues, err := s.client.IssueAPI.SearchIssuesAPIV3(ctx, jql, false)
		if err != nil {
			return nil, fmt.Errorf("JQL search failed: %w", err)
		}
		for _, iss := range issues {
			issueKeys = append(issueKeys, iss.Key)
		}
	}

	if len(issueKeys) == 0 {
		return nil, fmt.Errorf("no issues to update (provide issue_keys or jql)")
	}

	opts := &jira.BulkUpdateOptions{}

	// Handle add_labels array
	if addLabels, ok := args["add_labels"].([]any); ok {
		for _, label := range addLabels {
			if labelStr, ok := label.(string); ok {
				opts.AddLabels = append(opts.AddLabels, labelStr)
			}
		}
	}

	// Handle remove_labels array
	if removeLabels, ok := args["remove_labels"].([]any); ok {
		for _, label := range removeLabels {
			if labelStr, ok := label.(string); ok {
				opts.RemoveLabels = append(opts.RemoveLabels, labelStr)
			}
		}
	}

	if comment, ok := args["comment"].(string); ok {
		opts.Comment = comment
	}

	// Validate that at least one update is requested
	if len(opts.AddLabels) == 0 && len(opts.RemoveLabels) == 0 && opts.Comment == "" {
		return nil, fmt.Errorf("no updates specified (use add_labels, remove_labels, or comment)")
	}

	result, err := s.client.IssueAPI.BulkUpdateIssues(ctx, issueKeys, opts)
	if err != nil {
		return nil, fmt.Errorf("bulk update failed: %w", err)
	}

	return map[string]any{
		"success":       true,
		"total":         result.Total,
		"success_count": result.SuccessCount,
		"fail_count":    result.FailCount,
		"successful":    result.Successful,
		"failed":        result.Failed,
		"message":       fmt.Sprintf("Updated %d of %d issues", result.SuccessCount, result.Total),
	}, nil
}

func (s *Server) handleGetBoards(ctx context.Context, args map[string]any) (any, error) {
	opts := &gojira.BoardListOptions{}

	if project, ok := args["project"].(string); ok {
		opts.ProjectKeyOrID = project
	}
	if boardType, ok := args["type"].(string); ok {
		opts.BoardType = boardType
	}

	boards, err := s.client.BoardAPI.GetBoards(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("get boards: %w", err)
	}

	results := make([]map[string]any, 0, len(boards))
	for _, b := range boards {
		results = append(results, map[string]any{
			"id":   b.ID,
			"name": b.Name,
			"type": b.Type,
		})
	}

	return map[string]any{
		"total":  len(results),
		"boards": results,
	}, nil
}

func (s *Server) handleGetSprints(ctx context.Context, args map[string]any) (any, error) {
	boardID, ok := args["board_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("board_id is required")
	}

	opts := &gojira.GetAllSprintsOptions{}
	if state, ok := args["state"].(string); ok {
		opts.State = state
	}

	sprints, err := s.client.BoardAPI.GetSprints(ctx, int(boardID), opts)
	if err != nil {
		return nil, fmt.Errorf("get sprints: %w", err)
	}

	results := make([]map[string]any, 0, len(sprints))
	for _, sp := range sprints {
		result := map[string]any{
			"id":    sp.ID,
			"name":  sp.Name,
			"state": sp.State,
		}
		if sp.StartDate != "" {
			result["start_date"] = sp.StartDate
		}
		if sp.EndDate != "" {
			result["end_date"] = sp.EndDate
		}
		results = append(results, result)
	}

	return map[string]any{
		"board_id": int(boardID),
		"total":    len(results),
		"sprints":  results,
	}, nil
}

func (s *Server) handleVelocityReport(ctx context.Context, args map[string]any) (any, error) {
	boardID, ok := args["board_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("board_id is required")
	}

	opts := &jira.VelocityOptions{}
	if sprintCount, ok := args["sprint_count"].(float64); ok {
		opts.SprintCount = int(sprintCount)
	}
	if includeActive, ok := args["include_active"].(bool); ok {
		opts.IncludeActive = includeActive
	}
	if storyPointsField, ok := args["story_points_field"].(string); ok {
		opts.StoryPointsFieldID = storyPointsField
	}

	report, err := s.client.BoardAPI.GetVelocityReport(ctx, int(boardID), opts)
	if err != nil {
		return nil, fmt.Errorf("get velocity report: %w", err)
	}

	sprintData := make([]map[string]any, 0, len(report.Sprints))
	for _, sp := range report.Sprints {
		sprintData = append(sprintData, map[string]any{
			"sprint_id":        sp.SprintID,
			"sprint_name":      sp.SprintName,
			"completed_points": sp.CompletedPts,
			"issue_count":      sp.IssueCount,
			"state":            sp.State,
		})
	}

	return map[string]any{
		"board_id":         int(boardID),
		"average_velocity": report.AveragePoints,
		"total_sprints":    report.SprintCount,
		"total_points":     report.TotalPoints,
		"sprints":          sprintData,
	}, nil
}

func (s *Server) handleBurndownReport(ctx context.Context, args map[string]any) (any, error) {
	sprintID, ok := args["sprint_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("sprint_id is required")
	}

	opts := &jira.BurndownOptions{}
	if storyPointsField, ok := args["story_points_field"].(string); ok {
		opts.StoryPointsFieldID = storyPointsField
	}

	report, err := s.client.BoardAPI.GetBurndownReport(ctx, int(sprintID), opts)
	if err != nil {
		return nil, fmt.Errorf("get burndown report: %w", err)
	}

	// Get summary from daily data (current state)
	var completedPoints, remainingPoints float64
	var completedIssues, remainingIssues int
	if len(report.DailyData) > 0 {
		current := report.DailyData[len(report.DailyData)-1]
		completedPoints = current.CompletedPoints
		remainingPoints = current.RemainingPoints
		completedIssues = current.CompletedIssues
		remainingIssues = current.RemainingIssues
	}

	return map[string]any{
		"sprint_id":        report.SprintID,
		"sprint_name":      report.SprintName,
		"total_points":     report.TotalPoints,
		"completed_points": completedPoints,
		"remaining_points": remainingPoints,
		"total_issues":     report.TotalIssues,
		"completed_issues": completedIssues,
		"remaining_issues": remainingIssues,
		"start_date":       report.StartDate,
		"end_date":         report.EndDate,
	}, nil
}

func (s *Server) handleWorklogReport(ctx context.Context, args map[string]any) (any, error) {
	opts := &jira.WorklogOptions{}

	if jql, ok := args["jql"].(string); ok {
		opts.JQL = jql
	}
	if sprintID, ok := args["sprint_id"].(float64); ok {
		opts.SprintID = int(sprintID)
	}

	if opts.JQL == "" && opts.SprintID == 0 {
		return nil, fmt.Errorf("either jql or sprint_id is required")
	}

	report, err := s.client.BoardAPI.GetWorklogReport(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("get worklog report: %w", err)
	}

	byAuthor := make([]map[string]any, 0, len(report.ByAuthor))
	for _, a := range report.ByAuthor {
		byAuthor = append(byAuthor, map[string]any{
			"author":          a.Author,
			"time_spent":      a.TimeSpentStr,
			"time_spent_secs": a.TimeSpent,
			"worklog_count":   a.WorklogCount,
		})
	}

	byIssue := make([]map[string]any, 0, len(report.ByIssue))
	for _, i := range report.ByIssue {
		byIssue = append(byIssue, map[string]any{
			"key":             i.Key,
			"summary":         i.Summary,
			"time_spent":      i.TimeSpentStr,
			"time_spent_secs": i.TimeSpent,
			"worklog_count":   i.WorklogCount,
		})
	}

	return map[string]any{
		"total_time_spent":      report.TotalTimeSpentStr,
		"total_time_spent_secs": report.TotalTimeSpent,
		"total_worklog_count":   report.WorklogCount,
		"issue_count":           report.IssueCount,
		"by_author":             byAuthor,
		"by_issue":              byIssue,
	}, nil
}

func (s *Server) handleCycleTimeReport(ctx context.Context, args map[string]any) (any, error) {
	opts := &jira.CycleTimeOptions{}

	if jql, ok := args["jql"].(string); ok {
		opts.JQL = jql
	}
	if sprintID, ok := args["sprint_id"].(float64); ok {
		opts.SprintID = int(sprintID)
	}

	if opts.JQL == "" && opts.SprintID == 0 {
		return nil, fmt.Errorf("either jql or sprint_id is required")
	}

	report, err := s.client.BoardAPI.GetCycleTimeReport(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("get cycle time report: %w", err)
	}

	issues := make([]map[string]any, 0, len(report.Issues))
	for _, i := range report.Issues {
		issues = append(issues, map[string]any{
			"key":             i.Key,
			"summary":         i.Summary,
			"type":            i.IssueType,
			"created":         i.Created,
			"resolved":        i.Resolved,
			"cycle_time_days": i.CycleTimeDays,
		})
	}

	return map[string]any{
		"average_days":   report.AverageCycleTime,
		"median_days":    report.MedianCycleTime,
		"min_days":       report.MinCycleTime,
		"max_days":       report.MaxCycleTime,
		"total_resolved": report.IssueCount,
		"issues":         issues,
	}, nil
}

func (s *Server) handleMoveToSprint(ctx context.Context, args map[string]any) (any, error) {
	sprintID, ok := args["sprint_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("sprint_id is required")
	}

	var issueKeys []string
	if keys, ok := args["issue_keys"].([]any); ok {
		for _, key := range keys {
			if keyStr, ok := key.(string); ok {
				issueKeys = append(issueKeys, keyStr)
			}
		}
	}

	if len(issueKeys) == 0 {
		return nil, fmt.Errorf("issue_keys is required")
	}

	result, err := s.client.BoardAPI.MoveIssuesToSprint(ctx, int(sprintID), issueKeys)
	if err != nil {
		return nil, fmt.Errorf("move to sprint: %w", err)
	}

	return map[string]any{
		"success":      true,
		"sprint_id":    result.SprintID,
		"issues_moved": result.IssuesMoved,
		"issue_count":  result.IssueCount,
		"message":      fmt.Sprintf("Moved %d issues to sprint %d", result.IssueCount, result.SprintID),
	}, nil
}
