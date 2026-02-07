package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rgabriel/mcp-notion/notion"
)

// SearchHandler creates a handler for searching across pages and databases
func SearchHandler(client notion.NotionClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Build search request
		searchReq := &notion.SearchRequest{}

		// Query parameter
		if query, ok := args["query"].(string); ok && query != "" {
			searchReq.Query = query
		}

		// Filter parameter (object with property/value or type property)
		if filter, ok := args["filter"].(map[string]any); ok && len(filter) > 0 {
			searchReq.Filter = filter
		}

		// Sort parameter (object with direction and timestamp)
		if sort, ok := args["sort"].(map[string]any); ok && len(sort) > 0 {
			searchReq.Sort = sort
		}

		// Page size parameter
		if pageSize, ok := args["page_size"].(float64); ok && pageSize > 0 {
			searchReq.PageSize = int(pageSize)
		}

		// Execute search
		result, err := client.Search(ctx, searchReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"count":    len(result.Results),
			"results":  result.Results,
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

// ListDatabasesHandler creates a handler for listing all databases
func ListDatabasesHandler(client notion.NotionClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Search with database filter
		searchReq := &notion.SearchRequest{
			Filter: map[string]any{
				"property": "object",
				"value":    "database",
			},
			PageSize: 100,
		}

		result, err := client.Search(ctx, searchReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to list databases: %v", err)), nil
		}

		// Extract database info
		databases := make([]map[string]any, 0, len(result.Results))
		for _, item := range result.Results {
			if item.Object == "database" {
				db := map[string]any{
					"id":    item.ID,
					"url":   item.URL,
					"title": notion.ExtractPlainText(item.Title),
				}
				if item.Icon != nil {
					db["icon"] = item.Icon
				}
				if len(item.Description) > 0 {
					db["description"] = notion.ExtractPlainText(item.Description)
				}
				databases = append(databases, db)
			}
		}

		// Format response
		response := map[string]any{
			"count":     len(databases),
			"databases": databases,
			"has_more":  result.HasMore,
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

// GetDatabaseHandler creates a handler for getting database details
func GetDatabaseHandler(client notion.NotionClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract database ID
		databaseID, ok := args["database_id"].(string)
		if !ok || databaseID == "" {
			return mcp.NewToolResultError("database_id is required"), nil
		}

		// Get database
		database, err := client.GetDatabase(ctx, databaseID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get database: %v", err)), nil
		}

		// Format response with database details
		response := map[string]any{
			"id":               database.ID,
			"title":            notion.ExtractPlainText(database.Title),
			"url":              database.URL,
			"properties":       database.Properties,
			"is_inline":        database.IsInline,
			"archived":         database.Archived,
			"created_time":     database.CreatedTime,
			"last_edited_time": database.LastEditedTime,
		}

		if database.Icon != nil {
			response["icon"] = database.Icon
		}
		if len(database.Description) > 0 {
			response["description"] = notion.ExtractPlainText(database.Description)
		}
		if database.Parent != nil {
			response["parent"] = database.Parent
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
