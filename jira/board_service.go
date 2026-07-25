package jira

import (
	"context"
	"fmt"
	"strconv"
	"time"

	gojira "github.com/andygrunwald/go-jira"
)

// BoardService provides access to Jira Agile Board operations.
type BoardService struct {
	Client *Client
}

// NewBoardService creates a new BoardService.
func NewBoardService(client *Client) *BoardService {
	return &BoardService{Client: client}
}

// BoardOutput is a simplified view of a board for output.
type BoardOutput struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Self     string `json:"self,omitempty"`
	FilterID int    `json:"filter_id,omitempty"`
}

// SprintOutput is a simplified view of a sprint for output.
type SprintOutput struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	State         string `json:"state"`
	StartDate     string `json:"start_date,omitempty"`
	EndDate       string `json:"end_date,omitempty"`
	CompleteDate  string `json:"complete_date,omitempty"`
	OriginBoardID int    `json:"origin_board_id,omitempty"`
}

// GetBoards returns all boards, optionally filtered by project or type.
func (svc *BoardService) GetBoards(ctx context.Context, opts *gojira.BoardListOptions) ([]BoardOutput, error) {
	if svc.Client == nil || svc.Client.JiraClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	boardsList, resp, err := svc.Client.JiraClient.Board.GetAllBoardsWithContext(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("get boards: %w", err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("get boards: status %d", resp.StatusCode)
	}

	boards := make([]BoardOutput, 0, len(boardsList.Values))
	for _, b := range boardsList.Values {
		board := BoardOutput{
			ID:       b.ID,
			Name:     b.Name,
			Type:     b.Type,
			Self:     b.Self,
			FilterID: b.FilterID,
		}
		boards = append(boards, board)
	}

	return boards, nil
}

// GetBoard returns a single board by ID.
func (svc *BoardService) GetBoard(ctx context.Context, boardID int) (*BoardOutput, error) {
	if svc.Client == nil || svc.Client.JiraClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	board, resp, err := svc.Client.JiraClient.Board.GetBoardWithContext(ctx, boardID)
	if err != nil {
		return nil, fmt.Errorf("get board %d: %w", boardID, err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("get board %d: status %d", boardID, resp.StatusCode)
	}

	output := &BoardOutput{
		ID:       board.ID,
		Name:     board.Name,
		Type:     board.Type,
		Self:     board.Self,
		FilterID: board.FilterID,
	}

	return output, nil
}

// GetSprints returns all sprints for a board.
func (svc *BoardService) GetSprints(ctx context.Context, boardID int, opts *gojira.GetAllSprintsOptions) ([]SprintOutput, error) {
	if svc.Client == nil || svc.Client.JiraClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	sprintsList, resp, err := svc.Client.JiraClient.Board.GetAllSprintsWithOptionsWithContext(ctx, boardID, opts)
	if err != nil {
		return nil, fmt.Errorf("get sprints for board %d: %w", boardID, err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("get sprints for board %d: status %d", boardID, resp.StatusCode)
	}

	sprints := make([]SprintOutput, 0, len(sprintsList.Values))
	for _, s := range sprintsList.Values {
		sprint := SprintOutput{
			ID:            s.ID,
			Name:          s.Name,
			State:         s.State,
			OriginBoardID: s.OriginBoardID,
		}
		if s.StartDate != nil {
			sprint.StartDate = s.StartDate.Format("2006-01-02")
		}
		if s.EndDate != nil {
			sprint.EndDate = s.EndDate.Format("2006-01-02")
		}
		if s.CompleteDate != nil {
			sprint.CompleteDate = s.CompleteDate.Format("2006-01-02")
		}
		sprints = append(sprints, sprint)
	}

	return sprints, nil
}

// GetSprintIssues returns all issues in a sprint.
func (svc *BoardService) GetSprintIssues(ctx context.Context, sprintID int) (Issues, error) {
	if svc.Client == nil || svc.Client.JiraClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	// Use JQL to get sprint issues
	jql := fmt.Sprintf("Sprint = %d", sprintID)
	return svc.Client.IssueAPI.SearchIssuesAPIV3(ctx, jql, false)
}

// ParseBoardID parses a board ID from string.
func ParseBoardID(s string) (int, error) {
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid board ID %q: %w", s, err)
	}
	return id, nil
}

// MoveToSprintResult contains the result of a move to sprint operation.
type MoveToSprintResult struct {
	SprintID    int      `json:"sprint_id"`
	IssuesMoved []string `json:"issues_moved"`
	IssueCount  int      `json:"issue_count"`
}

// MoveIssuesToSprint moves issues to a sprint.
func (svc *BoardService) MoveIssuesToSprint(ctx context.Context, sprintID int, issueKeys []string) (*MoveToSprintResult, error) {
	if svc.Client == nil || svc.Client.JiraClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	if len(issueKeys) == 0 {
		return nil, fmt.Errorf("no issues to move")
	}

	resp, err := svc.Client.JiraClient.Sprint.MoveIssuesToSprintWithContext(ctx, sprintID, issueKeys)
	if err != nil {
		return nil, fmt.Errorf("move issues to sprint %d: %w", sprintID, err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("move issues to sprint %d: status %d", sprintID, resp.StatusCode)
	}

	return &MoveToSprintResult{
		SprintID:    sprintID,
		IssuesMoved: issueKeys,
		IssueCount:  len(issueKeys),
	}, nil
}

// RankIssuesInSprint reorders issues within a sprint.
// The first issue will be placed before the anchor issue (rankBeforeIssue).
func (svc *BoardService) RankIssuesInSprint(ctx context.Context, sprintID int, issueKeys []string, rankBeforeIssue string) error {
	if svc.Client == nil || svc.Client.JiraClient == nil {
		return ErrJiraRESTClientCannotBeNil
	}

	// Note: The Jira API for ranking requires the Agile API
	// This is a simplified implementation that just moves issues to the sprint
	// A full implementation would use the /rest/agile/1.0/sprint/{sprintId}/issue endpoint
	// with a rankBeforeIssue or rankAfterIssue parameter

	_, err := svc.MoveIssuesToSprint(ctx, sprintID, issueKeys)
	return err
}

// VelocityReport contains velocity metrics for sprints.
type VelocityReport struct {
	BoardID        int              `json:"board_id"`
	SprintCount    int              `json:"sprint_count"`
	TotalPoints    float64          `json:"total_points"`
	AveragePoints  float64          `json:"average_velocity"`
	Sprints        []SprintVelocity `json:"sprints"`
	StoryPointsKey string           `json:"story_points_key,omitempty"`
}

// SprintVelocity contains velocity metrics for a single sprint.
type SprintVelocity struct {
	SprintID     int     `json:"sprint_id"`
	SprintName   string  `json:"sprint_name"`
	State        string  `json:"state"`
	IssueCount   int     `json:"issue_count"`
	CompletedPts float64 `json:"completed_points"`
	StartDate    string  `json:"start_date,omitempty"`
	EndDate      string  `json:"end_date,omitempty"`
}

// VelocityOptions configures the velocity report.
type VelocityOptions struct {
	// SprintCount limits the number of sprints to include (0 = all closed sprints)
	SprintCount int
	// IncludeActive includes the active sprint in calculations
	IncludeActive bool
	// StoryPointsFieldID is the custom field ID for story points (e.g., "customfield_10016")
	StoryPointsFieldID string
}

// GetVelocityReport calculates velocity metrics for a board.
func (svc *BoardService) GetVelocityReport(ctx context.Context, boardID int, opts *VelocityOptions) (*VelocityReport, error) {
	if svc.Client == nil || svc.Client.JiraClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	// Get closed sprints
	sprintOpts := &gojira.GetAllSprintsOptions{State: "closed"}
	if opts != nil && opts.IncludeActive {
		sprintOpts.State = "closed,active"
	}

	sprints, err := svc.GetSprints(ctx, boardID, sprintOpts)
	if err != nil {
		return nil, fmt.Errorf("get sprints for velocity: %w", err)
	}

	// Limit sprint count if specified
	if opts != nil && opts.SprintCount > 0 && len(sprints) > opts.SprintCount {
		// Take the most recent N sprints (assumes sprints are ordered oldest to newest)
		sprints = sprints[len(sprints)-opts.SprintCount:]
	}

	// Determine story points field
	storyPointsKey := "customfield_10016" // Common default
	if opts != nil && opts.StoryPointsFieldID != "" {
		storyPointsKey = opts.StoryPointsFieldID
	}

	report := &VelocityReport{
		BoardID:        boardID,
		SprintCount:    len(sprints),
		Sprints:        make([]SprintVelocity, 0, len(sprints)),
		StoryPointsKey: storyPointsKey,
	}

	for _, sprint := range sprints {
		sv := SprintVelocity{
			SprintID:   sprint.ID,
			SprintName: sprint.Name,
			State:      sprint.State,
			StartDate:  sprint.StartDate,
			EndDate:    sprint.EndDate,
		}

		// Get issues in this sprint
		issues, err := svc.GetSprintIssues(ctx, sprint.ID)
		if err != nil {
			// Continue with zero points for this sprint
			report.Sprints = append(report.Sprints, sv)
			continue
		}

		sv.IssueCount = len(issues)

		// Sum story points for completed issues
		for _, iss := range issues {
			if iss.Fields == nil || iss.Fields.Status == nil {
				continue
			}
			// Only count issues in "Done" category
			if iss.Fields.Status.StatusCategory.Key != "done" {
				continue
			}

			// Extract story points from the unknowns map
			if iss.Fields.Unknowns != nil {
				if pts, ok := iss.Fields.Unknowns[storyPointsKey]; ok {
					switch v := pts.(type) {
					case float64:
						sv.CompletedPts += v
					case int:
						sv.CompletedPts += float64(v)
					}
				}
			}
		}

		report.TotalPoints += sv.CompletedPts
		report.Sprints = append(report.Sprints, sv)
	}

	if report.SprintCount > 0 {
		report.AveragePoints = report.TotalPoints / float64(report.SprintCount)
	}

	return report, nil
}

// WorklogReport contains worklog aggregation data.
type WorklogReport struct {
	TotalTimeSpent    int             `json:"total_time_spent_seconds"`
	TotalTimeSpentStr string          `json:"total_time_spent"`
	IssueCount        int             `json:"issue_count"`
	WorklogCount      int             `json:"worklog_count"`
	ByAuthor          []AuthorWorklog `json:"by_author"`
	ByIssue           []IssueWorklog  `json:"by_issue"`
}

// AuthorWorklog contains worklog data for an author.
type AuthorWorklog struct {
	Author       string `json:"author"`
	TimeSpent    int    `json:"time_spent_seconds"`
	TimeSpentStr string `json:"time_spent"`
	WorklogCount int    `json:"worklog_count"`
}

// IssueWorklog contains worklog data for an issue.
type IssueWorklog struct {
	Key          string `json:"key"`
	Summary      string `json:"summary"`
	TimeSpent    int    `json:"time_spent_seconds"`
	TimeSpentStr string `json:"time_spent"`
	WorklogCount int    `json:"worklog_count"`
}

// WorklogOptions configures the worklog report.
type WorklogOptions struct {
	// JQL to filter issues
	JQL string
	// SprintID to filter by sprint
	SprintID int
}

// GetWorklogReport generates a worklog summary report.
func (svc *BoardService) GetWorklogReport(ctx context.Context, opts *WorklogOptions) (*WorklogReport, error) {
	if svc.Client == nil || svc.Client.JiraClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	// Build JQL
	jql := ""
	if opts != nil {
		if opts.JQL != "" {
			jql = opts.JQL
		} else if opts.SprintID > 0 {
			jql = fmt.Sprintf("Sprint = %d", opts.SprintID)
		}
	}
	if jql == "" {
		return nil, fmt.Errorf("jql or sprint_id is required")
	}

	// Get issues
	issues, err := svc.Client.IssueAPI.SearchIssuesAPIV3(ctx, jql, false)
	if err != nil {
		return nil, fmt.Errorf("search issues: %w", err)
	}

	report := &WorklogReport{
		ByAuthor: []AuthorWorklog{},
		ByIssue:  []IssueWorklog{},
	}

	authorMap := make(map[string]*AuthorWorklog)

	for _, iss := range issues {
		if iss.Fields == nil {
			continue
		}

		// Get worklogs for this issue
		worklogs, _, err := svc.Client.JiraClient.Issue.GetWorklogsWithContext(ctx, iss.Key)
		if err != nil {
			continue // Skip issues we can't get worklogs for
		}
		if worklogs == nil || len(worklogs.Worklogs) == 0 {
			continue
		}

		issueWorklog := IssueWorklog{
			Key:     iss.Key,
			Summary: iss.Fields.Summary,
		}

		for _, wl := range worklogs.Worklogs {
			report.TotalTimeSpent += wl.TimeSpentSeconds
			report.WorklogCount++
			issueWorklog.TimeSpent += wl.TimeSpentSeconds
			issueWorklog.WorklogCount++

			// Aggregate by author
			authorName := "Unknown"
			if wl.Author != nil {
				authorName = wl.Author.DisplayName
				if authorName == "" {
					authorName = wl.Author.Name
				}
			}

			if _, ok := authorMap[authorName]; !ok {
				authorMap[authorName] = &AuthorWorklog{Author: authorName}
			}
			authorMap[authorName].TimeSpent += wl.TimeSpentSeconds
			authorMap[authorName].WorklogCount++
		}

		if issueWorklog.WorklogCount > 0 {
			issueWorklog.TimeSpentStr = formatDuration(issueWorklog.TimeSpent)
			report.ByIssue = append(report.ByIssue, issueWorklog)
			report.IssueCount++
		}
	}

	// Convert author map to slice
	for _, aw := range authorMap {
		aw.TimeSpentStr = formatDuration(aw.TimeSpent)
		report.ByAuthor = append(report.ByAuthor, *aw)
	}

	report.TotalTimeSpentStr = formatDuration(report.TotalTimeSpent)

	return report, nil
}

// formatDuration formats seconds into a human-readable string.
func formatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// CycleTimeReport contains cycle time analysis data.
type CycleTimeReport struct {
	IssueCount       int              `json:"issue_count"`
	AverageCycleTime float64          `json:"average_cycle_time_days"`
	MedianCycleTime  float64          `json:"median_cycle_time_days"`
	MinCycleTime     float64          `json:"min_cycle_time_days"`
	MaxCycleTime     float64          `json:"max_cycle_time_days"`
	Issues           []IssueCycleTime `json:"issues"`
}

// IssueCycleTime contains cycle time data for a single issue.
type IssueCycleTime struct {
	Key           string  `json:"key"`
	Summary       string  `json:"summary"`
	IssueType     string  `json:"type"`
	Created       string  `json:"created"`
	Resolved      string  `json:"resolved"`
	CycleTimeDays float64 `json:"cycle_time_days"`
}

// CycleTimeOptions configures the cycle time report.
type CycleTimeOptions struct {
	// JQL to filter issues
	JQL string
	// SprintID to filter by sprint
	SprintID int
}

// GetCycleTimeReport calculates cycle time metrics for resolved issues.
func (svc *BoardService) GetCycleTimeReport(ctx context.Context, opts *CycleTimeOptions) (*CycleTimeReport, error) {
	if svc.Client == nil || svc.Client.JiraClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	// Build JQL - only include resolved issues
	jql := "resolution IS NOT EMPTY"
	if opts != nil {
		if opts.JQL != "" {
			jql = opts.JQL + " AND resolution IS NOT EMPTY"
		} else if opts.SprintID > 0 {
			jql = fmt.Sprintf("Sprint = %d AND resolution IS NOT EMPTY", opts.SprintID)
		}
	}

	// Get issues
	issues, err := svc.Client.IssueAPI.SearchIssuesAPIV3(ctx, jql, false)
	if err != nil {
		return nil, fmt.Errorf("search issues: %w", err)
	}

	report := &CycleTimeReport{
		Issues: []IssueCycleTime{},
	}

	var cycleTimes []float64

	for _, iss := range issues {
		if iss.Fields == nil {
			continue
		}

		createdTime := time.Time(iss.Fields.Created)
		resolvedTime := time.Time(iss.Fields.Resolutiondate)
		if resolvedTime.IsZero() {
			continue // Skip if no resolution date
		}

		// Calculate cycle time in days
		cycleTime := resolvedTime.Sub(createdTime).Hours() / 24

		issueType := iss.Fields.Type.Name

		issueCycle := IssueCycleTime{
			Key:           iss.Key,
			Summary:       iss.Fields.Summary,
			IssueType:     issueType,
			Created:       createdTime.Format("2006-01-02"),
			Resolved:      resolvedTime.Format("2006-01-02"),
			CycleTimeDays: cycleTime,
		}

		report.Issues = append(report.Issues, issueCycle)
		cycleTimes = append(cycleTimes, cycleTime)

		// Update min/max
		if report.IssueCount == 0 || cycleTime < report.MinCycleTime {
			report.MinCycleTime = cycleTime
		}
		if cycleTime > report.MaxCycleTime {
			report.MaxCycleTime = cycleTime
		}

		report.IssueCount++
	}

	// Calculate average and median
	if len(cycleTimes) > 0 {
		var sum float64
		for _, ct := range cycleTimes {
			sum += ct
		}
		report.AverageCycleTime = sum / float64(len(cycleTimes))

		// Calculate median (simple approach - sort and take middle)
		sortedTimes := make([]float64, len(cycleTimes))
		copy(sortedTimes, cycleTimes)
		// Simple bubble sort for median
		for i := 0; i < len(sortedTimes)-1; i++ {
			for j := 0; j < len(sortedTimes)-i-1; j++ {
				if sortedTimes[j] > sortedTimes[j+1] {
					sortedTimes[j], sortedTimes[j+1] = sortedTimes[j+1], sortedTimes[j]
				}
			}
		}
		mid := len(sortedTimes) / 2
		if len(sortedTimes)%2 == 0 {
			report.MedianCycleTime = (sortedTimes[mid-1] + sortedTimes[mid]) / 2
		} else {
			report.MedianCycleTime = sortedTimes[mid]
		}
	}

	return report, nil
}

// BurndownReport contains burndown data for a sprint.
type BurndownReport struct {
	SprintID       int           `json:"sprint_id"`
	SprintName     string        `json:"sprint_name"`
	StartDate      string        `json:"start_date"`
	EndDate        string        `json:"end_date"`
	TotalPoints    float64       `json:"total_points"`
	TotalIssues    int           `json:"total_issues"`
	DailyData      []BurndownDay `json:"daily_data"`
	StoryPointsKey string        `json:"story_points_key,omitempty"`
}

// BurndownDay contains burndown data for a single day.
type BurndownDay struct {
	Date            string  `json:"date"`
	RemainingPoints float64 `json:"remaining_points"`
	RemainingIssues int     `json:"remaining_issues"`
	CompletedPoints float64 `json:"completed_points"`
	CompletedIssues int     `json:"completed_issues"`
}

// BurndownOptions configures the burndown report.
type BurndownOptions struct {
	// StoryPointsFieldID is the custom field ID for story points
	StoryPointsFieldID string
}

// GetBurndownReport calculates burndown data for a sprint.
func (svc *BoardService) GetBurndownReport(ctx context.Context, sprintID int, opts *BurndownOptions) (*BurndownReport, error) {
	if svc.Client == nil || svc.Client.JiraClient == nil {
		return nil, ErrJiraRESTClientCannotBeNil
	}

	// Get sprint details using JQL to find sprint info
	jql := fmt.Sprintf("Sprint = %d", sprintID)
	issues, err := svc.Client.IssueAPI.SearchIssuesAPIV3(ctx, jql, false)
	if err != nil {
		return nil, fmt.Errorf("get sprint issues: %w", err)
	}

	storyPointsKey := "customfield_10016"
	if opts != nil && opts.StoryPointsFieldID != "" {
		storyPointsKey = opts.StoryPointsFieldID
	}

	report := &BurndownReport{
		SprintID:       sprintID,
		TotalIssues:    len(issues),
		DailyData:      []BurndownDay{},
		StoryPointsKey: storyPointsKey,
	}

	// Calculate total points and current state
	var completedPoints, remainingPoints float64
	var completedIssues, remainingIssues int

	for _, iss := range issues {
		if iss.Fields == nil {
			continue
		}

		// Get story points
		var pts float64
		if iss.Fields.Unknowns != nil {
			if p, ok := iss.Fields.Unknowns[storyPointsKey]; ok {
				switch v := p.(type) {
				case float64:
					pts = v
				case int:
					pts = float64(v)
				}
			}
		}

		report.TotalPoints += pts

		// Check if issue is done
		if iss.Fields.Status != nil && iss.Fields.Status.StatusCategory.Key == "done" {
			completedPoints += pts
			completedIssues++
		} else {
			remainingPoints += pts
			remainingIssues++
		}
	}

	// Add current state as a single data point (simplified burndown)
	// A full implementation would query issue history for daily snapshots
	report.DailyData = append(report.DailyData, BurndownDay{
		Date:            "current",
		RemainingPoints: remainingPoints,
		RemainingIssues: remainingIssues,
		CompletedPoints: completedPoints,
		CompletedIssues: completedIssues,
	})

	return report, nil
}
