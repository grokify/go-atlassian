package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gojira "github.com/andygrunwald/go-jira"
)

func TestNewBoardService(t *testing.T) {
	client := &Client{}
	svc := NewBoardService(client)

	if svc == nil {
		t.Fatal("NewBoardService() returned nil")
	}
	if svc.Client != client {
		t.Error("NewBoardService() client mismatch")
	}
}

func TestBoardServiceGetBoardsNilClient(t *testing.T) {
	svc := &BoardService{Client: nil}

	_, err := svc.GetBoards(context.Background(), nil)
	if err == nil {
		t.Error("GetBoards() with nil client should return error")
	}
}

func TestBoardServiceGetBoards(t *testing.T) {
	// Create a mock server that returns valid boards
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := gojira.BoardsList{
			Values: []gojira.Board{
				{ID: 1, Name: "Test Board", Type: "scrum", FilterID: 100},
				{ID: 2, Name: "Kanban Board", Type: "kanban", FilterID: 200},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	jiraClient, err := gojira.NewClient(nil, server.URL)
	if err != nil {
		t.Fatalf("failed to create jira client: %v", err)
	}

	client := &Client{JiraClient: jiraClient}
	svc := NewBoardService(client)

	boards, err := svc.GetBoards(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetBoards() error = %v", err)
	}

	if len(boards) != 2 {
		t.Errorf("GetBoards() returned %d boards, want 2", len(boards))
	}

	if boards[0].ID != 1 || boards[0].Name != "Test Board" {
		t.Errorf("GetBoards() first board = %+v, want ID=1, Name=Test Board", boards[0])
	}
}

func TestBoardServiceGetBoardNilClient(t *testing.T) {
	svc := &BoardService{Client: nil}

	_, err := svc.GetBoard(context.Background(), 1)
	if err == nil {
		t.Error("GetBoard() with nil client should return error")
	}
}

func TestBoardServiceGetBoard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		board := gojira.Board{
			ID:       1,
			Name:     "Test Board",
			Type:     "scrum",
			FilterID: 100,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(board); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	jiraClient, err := gojira.NewClient(nil, server.URL)
	if err != nil {
		t.Fatalf("failed to create jira client: %v", err)
	}

	client := &Client{JiraClient: jiraClient}
	svc := NewBoardService(client)

	board, err := svc.GetBoard(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetBoard() error = %v", err)
	}

	if board.ID != 1 {
		t.Errorf("GetBoard() ID = %d, want 1", board.ID)
	}
	if board.Name != "Test Board" {
		t.Errorf("GetBoard() Name = %q, want Test Board", board.Name)
	}
}

func TestBoardServiceGetSprintsNilClient(t *testing.T) {
	svc := &BoardService{Client: nil}

	_, err := svc.GetSprints(context.Background(), 1, nil)
	if err == nil {
		t.Error("GetSprints() with nil client should return error")
	}
}

func TestBoardServiceMoveIssuesToSprintNilClient(t *testing.T) {
	svc := &BoardService{Client: nil}

	_, err := svc.MoveIssuesToSprint(context.Background(), 1, []string{"TEST-1"})
	if err == nil {
		t.Error("MoveIssuesToSprint() with nil client should return error")
	}
}

func TestBoardServiceMoveIssuesToSprintEmptyIssues(t *testing.T) {
	client := &Client{}
	svc := NewBoardService(client)

	_, err := svc.MoveIssuesToSprint(context.Background(), 1, []string{})
	if err == nil {
		t.Error("MoveIssuesToSprint() with empty issues should return error")
	}
}

func TestParseBoardID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:    "valid ID",
			input:   "123",
			want:    123,
			wantErr: false,
		},
		{
			name:    "invalid ID",
			input:   "abc",
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBoardID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBoardID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseBoardID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVelocityOptionsDefaults(t *testing.T) {
	opts := &VelocityOptions{}

	if opts.SprintCount != 0 {
		t.Errorf("VelocityOptions default SprintCount = %d, want 0", opts.SprintCount)
	}
	if opts.IncludeActive {
		t.Error("VelocityOptions default IncludeActive should be false")
	}
	if opts.StoryPointsFieldID != "" {
		t.Errorf("VelocityOptions default StoryPointsFieldID = %q, want empty", opts.StoryPointsFieldID)
	}
}

func TestWorklogOptionsValidation(t *testing.T) {
	client := &Client{JiraClient: nil}
	svc := NewBoardService(client)

	// Neither JQL nor SprintID provided
	_, err := svc.GetWorklogReport(context.Background(), &WorklogOptions{})
	if err == nil {
		t.Error("GetWorklogReport() with no JQL or SprintID should return error")
	}
}

func TestCycleTimeOptionsValidation(t *testing.T) {
	client := &Client{JiraClient: nil}
	svc := NewBoardService(client)

	// Neither JQL nor SprintID provided
	_, err := svc.GetCycleTimeReport(context.Background(), &CycleTimeOptions{})
	if err == nil {
		t.Error("GetCycleTimeReport() with no JQL or SprintID should return error")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name    string
		seconds int
		want    string
	}{
		{
			name:    "minutes only",
			seconds: 1800, // 30 minutes
			want:    "30m",
		},
		{
			name:    "hours and minutes",
			seconds: 3960, // 1h 6m
			want:    "1h 6m",
		},
		{
			name:    "multiple hours",
			seconds: 7200, // 2 hours
			want:    "2h 0m",
		},
		{
			name:    "zero",
			seconds: 0,
			want:    "0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.seconds)
			if got != tt.want {
				t.Errorf("formatDuration(%d) = %q, want %q", tt.seconds, got, tt.want)
			}
		})
	}
}
