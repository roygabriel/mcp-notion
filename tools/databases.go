package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rgabriel/mcp-notion/notion"
)

// QueryDatabaseHandler creates a handler for querying a database
func QueryDatabaseHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract database ID
		databaseID, ok := args["database_id"].(string)
		if !ok || databaseID == "" {
			return mcp.NewToolResultError("database_id is required"), nil
		}

		// Build query request
		queryReq := &notion.QueryDatabaseRequest{}

		// Filter parameter
		if filter, ok := args["filter"].(map[string]any); ok && len(filter) > 0 {
			queryReq.Filter = filter
		}

		// Sorts parameter
		if sortsData, ok := args["sorts"].([]any); ok && len(sortsData) > 0 {
			sorts := make([]map[string]any, 0, len(sortsData))
			for _, sortData := range sortsData {
				if sortMap, ok := sortData.(map[string]any); ok {
					sorts = append(sorts, sortMap)
				}
			}
			queryReq.Sorts = sorts
		}

		// Page size parameter
		if pageSize, ok := args["page_size"].(float64); ok && pageSize > 0 {
			queryReq.PageSize = int(pageSize)
		}

		// Start cursor for pagination
		if startCursor, ok := args["start_cursor"].(string); ok && startCursor != "" {
			queryReq.StartCursor = startCursor
		}

		// Query database
		result, err := client.QueryDatabase(ctx, databaseID, queryReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to query database: %v", err)), nil
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

// CreateDatabaseHandler creates a handler for creating a new database
func CreateDatabaseHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract parent
		parentMap, ok := args["parent"].(map[string]any)
		if !ok || len(parentMap) == 0 {
			return mcp.NewToolResultError("parent is required (must be object with page_id)"), nil
		}

		parent := notion.Parent{}
		if pageID, ok := parentMap["page_id"].(string); ok && pageID != "" {
			parent.Type = "page_id"
			parent.PageID = pageID
		} else {
			return mcp.NewToolResultError("parent must contain page_id"), nil
		}

		// Extract title
		titleData, ok := args["title"].([]any)
		if !ok || len(titleData) == 0 {
			return mcp.NewToolResultError("title is required (array of rich text objects)"), nil
		}

		// Convert title to RichText array
		titleJSON, err := json.Marshal(titleData)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid title format: %v", err)), nil
		}
		var title []notion.RichText
		if err := json.Unmarshal(titleJSON, &title); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid title structure: %v", err)), nil
		}

		// Extract properties (database schema)
		propertiesData, ok := args["properties"].(map[string]any)
		if !ok || len(propertiesData) == 0 {
			return mcp.NewToolResultError("properties is required (database schema definition)"), nil
		}

		// Convert properties to PropertyDef map
		propertiesJSON, err := json.Marshal(propertiesData)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid properties format: %v", err)), nil
		}
		var properties map[string]notion.PropertyDef
		if err := json.Unmarshal(propertiesJSON, &properties); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid properties structure: %v", err)), nil
		}

		// Build create request
		createReq := &notion.CreateDatabaseRequest{
			Parent:     parent,
			Title:      title,
			Properties: properties,
		}

		// Optional icon
		if iconData, ok := args["icon"].(map[string]any); ok && len(iconData) > 0 {
			iconJSON, _ := json.Marshal(iconData)
			var icon notion.Icon
			if err := json.Unmarshal(iconJSON, &icon); err == nil {
				createReq.Icon = &icon
			}
		}

		// Optional cover
		if coverData, ok := args["cover"].(map[string]any); ok && len(coverData) > 0 {
			coverJSON, _ := json.Marshal(coverData)
			var cover notion.Cover
			if err := json.Unmarshal(coverJSON, &cover); err == nil {
				createReq.Cover = &cover
			}
		}

		// Create database
		database, err := client.CreateDatabase(ctx, createReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create database: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"success": true,
			"id":      database.ID,
			"url":     database.URL,
			"message": "Database created successfully",
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// UpdateDatabaseHandler creates a handler for updating a database
func UpdateDatabaseHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract database ID
		databaseID, ok := args["database_id"].(string)
		if !ok || databaseID == "" {
			return mcp.NewToolResultError("database_id is required"), nil
		}

		// Build update request
		updateReq := &notion.UpdateDatabaseRequest{}

		// Title to update
		if titleData, ok := args["title"].([]any); ok && len(titleData) > 0 {
			titleJSON, err := json.Marshal(titleData)
			if err == nil {
				var title []notion.RichText
				if err := json.Unmarshal(titleJSON, &title); err == nil {
					updateReq.Title = title
				}
			}
		}

		// Description to update
		if descData, ok := args["description"].([]any); ok && len(descData) > 0 {
			descJSON, err := json.Marshal(descData)
			if err == nil {
				var description []notion.RichText
				if err := json.Unmarshal(descJSON, &description); err == nil {
					updateReq.Description = description
				}
			}
		}

		// Properties to update (add/modify schema)
		if propertiesData, ok := args["properties"].(map[string]any); ok && len(propertiesData) > 0 {
			propertiesJSON, err := json.Marshal(propertiesData)
			if err == nil {
				var properties map[string]notion.PropertyDef
				if err := json.Unmarshal(propertiesJSON, &properties); err == nil {
					updateReq.Properties = properties
				}
			}
		}

		// Archived status
		if archived, ok := args["archived"].(bool); ok {
			updateReq.Archived = &archived
		}

		// Update database
		database, err := client.UpdateDatabase(ctx, databaseID, updateReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to update database: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"success": true,
			"id":      database.ID,
			"url":     database.URL,
			"message": "Database updated successfully",
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
