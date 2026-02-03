package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rgabriel/mcp-notion/notion"
)

// ListUsersHandler creates a handler for listing all users
func ListUsersHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Page size parameter
		pageSize := 100
		if ps, ok := args["page_size"].(float64); ok && ps > 0 {
			pageSize = int(ps)
		}

		// Start cursor for pagination
		startCursor := ""
		if sc, ok := args["start_cursor"].(string); ok {
			startCursor = sc
		}

		// List users
		result, err := client.ListUsers(ctx, pageSize, startCursor)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to list users: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"count":    len(result.Results),
			"users":    result.Results,
			"has_more": result.HasMore,
		}
		if result.NextCursor != "" {
			response["next_cursor"] = result.NextCursor
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// GetUserHandler creates a handler for getting a specific user
func GetUserHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract user ID
		userID, ok := args["user_id"].(string)
		if !ok || userID == "" {
			return mcp.NewToolResultError("user_id is required"), nil
		}

		// Get user
		user, err := client.GetUser(ctx, userID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get user: %v", err)), nil
		}

		// Format response
		jsonData, err := json.MarshalIndent(user, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// GetBotUserHandler creates a handler for getting the current bot user
func GetBotUserHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get bot user
		user, err := client.GetBotUser(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get bot user: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"id":   user.ID,
			"type": user.Type,
			"name": user.Name,
		}
		if user.AvatarURL != "" {
			response["avatar_url"] = user.AvatarURL
		}
		if user.Bot != nil {
			response["bot"] = user.Bot
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
