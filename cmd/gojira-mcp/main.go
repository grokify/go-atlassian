// gojira-mcp is a Model Context Protocol (MCP) server for Jira.
// It enables AI assistants like Claude to interact with Jira through
// standardized tool calls over JSON-RPC.
//
// Configuration via environment variables:
//
//	JIRA_BASE_URL    - Jira server URL (e.g., https://company.atlassian.net)
//	JIRA_USERNAME    - Jira username or email
//	JIRA_API_TOKEN   - Jira API token or password
//
// Usage with Claude Code:
//
//	{
//	  "mcpServers": {
//	    "jira": {
//	      "command": "gojira-mcp",
//	      "env": {
//	        "JIRA_BASE_URL": "https://company.atlassian.net",
//	        "JIRA_USERNAME": "user@example.com",
//	        "JIRA_API_TOKEN": "your-api-token"
//	      }
//	    }
//	  }
//	}
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/grokify/go-atlassian/jira"
	"github.com/grokify/go-atlassian/mcpserver"
)

func main() {
	// Configure logging to stderr (stdout is for JSON-RPC)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: getLogLevel(),
	}))
	slog.SetDefault(logger)

	// Read configuration from environment
	baseURL := os.Getenv("JIRA_BASE_URL")
	username := os.Getenv("JIRA_USERNAME")
	apiToken := os.Getenv("JIRA_API_TOKEN")

	if baseURL == "" || username == "" || apiToken == "" {
		logger.Error("missing required environment variables",
			"JIRA_BASE_URL", baseURL != "",
			"JIRA_USERNAME", username != "",
			"JIRA_API_TOKEN", apiToken != "")
		fmt.Fprintln(os.Stderr, "Error: JIRA_BASE_URL, JIRA_USERNAME, and JIRA_API_TOKEN are required")
		os.Exit(1)
	}

	// Create Jira client
	client, err := jira.NewClientFromBasicAuth(baseURL, username, apiToken, false)
	if err != nil {
		logger.Error("failed to create Jira client", "error", err)
		fmt.Fprintf(os.Stderr, "Error creating Jira client: %v\n", err)
		os.Exit(1)
	}

	// Create MCP server
	server := mcpserver.NewServer(client, logger)

	logger.Info("gojira-mcp server started", "base_url", baseURL)

	// Process JSON-RPC messages from stdin
	scanner := bufio.NewScanner(os.Stdin)
	// Increase buffer size for large requests
	const maxScanTokenSize = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var req mcpserver.JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			logger.Error("failed to parse request", "error", err, "line", line)
			writeError(nil, mcpserver.ErrorCodeParseError, fmt.Sprintf("parse error: %v", err))
			continue
		}

		logger.Debug("received request", "method", req.Method, "id", req.ID)

		resp := server.Handle(context.Background(), req)

		respBytes, err := json.Marshal(resp)
		if err != nil {
			logger.Error("failed to marshal response", "error", err)
			writeError(req.ID, mcpserver.ErrorCodeInternalError, fmt.Sprintf("marshal error: %v", err))
			continue
		}

		fmt.Println(string(respBytes))
	}

	if err := scanner.Err(); err != nil {
		logger.Error("scanner error", "error", err)
		os.Exit(1)
	}
}

func writeError(id any, code int, message string) {
	resp := mcpserver.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &mcpserver.JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
	respBytes, _ := json.Marshal(resp)
	fmt.Println(string(respBytes))
}

func getLogLevel() slog.Level {
	level := os.Getenv("GOJIRA_MCP_LOG_LEVEL")
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
