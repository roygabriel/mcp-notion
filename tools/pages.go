package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rgabriel/mcp-notion/notion"
)

// GetPageHandler creates a handler for getting page metadata
func GetPageHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract page ID
		pageID, ok := args["page_id"].(string)
		if !ok || pageID == "" {
			return mcp.NewToolResultError("page_id is required"), nil
		}

		// Get page
		page, err := client.GetPage(ctx, pageID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get page: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"id":               page.ID,
			"url":              page.URL,
			"properties":       page.Properties,
			"parent":           page.Parent,
			"archived":         page.Archived,
			"created_time":     page.CreatedTime,
			"last_edited_time": page.LastEditedTime,
		}

		if page.Icon != nil {
			response["icon"] = page.Icon
		}
		if page.Cover != nil {
			response["cover"] = page.Cover
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// GetPageContentHandler creates a handler for getting page content (blocks)
func GetPageContentHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract page ID
		pageID, ok := args["page_id"].(string)
		if !ok || pageID == "" {
			return mcp.NewToolResultError("page_id is required"), nil
		}

		// Page size parameter
		pageSize := 100
		if ps, ok := args["page_size"].(float64); ok && ps > 0 {
			pageSize = int(ps)
		}

		// Get blocks
		result, err := client.GetBlockChildren(ctx, pageID, pageSize, "")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get page content: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"count":    len(result.Results),
			"blocks":   result.Results,
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

// CreatePageHandler creates a handler for creating a new page
func CreatePageHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract parent
		parentMap, ok := args["parent"].(map[string]any)
		if !ok || len(parentMap) == 0 {
			return mcp.NewToolResultError("parent is required (must be object with page_id or database_id)"), nil
		}

		parent := notion.Parent{}
		if pageID, ok := parentMap["page_id"].(string); ok && pageID != "" {
			parent.Type = "page_id"
			parent.PageID = pageID
		} else if dbID, ok := parentMap["database_id"].(string); ok && dbID != "" {
			parent.Type = "database_id"
			parent.DatabaseID = dbID
		} else {
			return mcp.NewToolResultError("parent must contain either page_id or database_id"), nil
		}

		// Extract properties
		properties, ok := args["properties"].(map[string]any)
		if !ok || len(properties) == 0 {
			return mcp.NewToolResultError("properties is required"), nil
		}

		// Build request
		createReq := &notion.CreatePageRequest{
			Parent:     parent,
			Properties: properties,
		}

		// Optional children (blocks)
		if childrenData, ok := args["children"].([]any); ok && len(childrenData) > 0 {
			children := make([]notion.Block, 0, len(childrenData))
			for _, child := range childrenData {
				if childMap, ok := child.(map[string]any); ok {
					// Convert to JSON and back to Block
					childJSON, err := json.Marshal(childMap)
					if err != nil {
						return mcp.NewToolResultError(fmt.Sprintf("invalid child block format: %v", err)), nil
					}
					var block notion.Block
					if err := json.Unmarshal(childJSON, &block); err != nil {
						return mcp.NewToolResultError(fmt.Sprintf("invalid child block structure: %v", err)), nil
					}
					children = append(children, block)
				}
			}
			createReq.Children = children
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

		// Create page
		page, err := client.CreatePage(ctx, createReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create page: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"success": true,
			"id":      page.ID,
			"url":     page.URL,
			"message": "Page created successfully",
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// UpdatePageHandler creates a handler for updating a page
func UpdatePageHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract page ID
		pageID, ok := args["page_id"].(string)
		if !ok || pageID == "" {
			return mcp.NewToolResultError("page_id is required"), nil
		}

		// Build update request
		updateReq := &notion.UpdatePageRequest{}

		// Properties to update
		if properties, ok := args["properties"].(map[string]any); ok && len(properties) > 0 {
			updateReq.Properties = properties
		}

		// Archived status
		if archived, ok := args["archived"].(bool); ok {
			updateReq.Archived = &archived
		}

		// Icon
		if iconData, ok := args["icon"].(map[string]any); ok && len(iconData) > 0 {
			iconJSON, _ := json.Marshal(iconData)
			var icon notion.Icon
			if err := json.Unmarshal(iconJSON, &icon); err == nil {
				updateReq.Icon = &icon
			}
		}

		// Cover
		if coverData, ok := args["cover"].(map[string]any); ok && len(coverData) > 0 {
			coverJSON, _ := json.Marshal(coverData)
			var cover notion.Cover
			if err := json.Unmarshal(coverJSON, &cover); err == nil {
				updateReq.Cover = &cover
			}
		}

		// Update page
		page, err := client.UpdatePage(ctx, pageID, updateReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to update page: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"success": true,
			"id":      page.ID,
			"url":     page.URL,
			"message": "Page updated successfully",
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// DeletePageHandler creates a handler for deleting (archiving) a page
func DeletePageHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract page ID
		pageID, ok := args["page_id"].(string)
		if !ok || pageID == "" {
			return mcp.NewToolResultError("page_id is required"), nil
		}

		// Archive the page
		archived := true
		updateReq := &notion.UpdatePageRequest{
			Archived: &archived,
		}

		page, err := client.UpdatePage(ctx, pageID, updateReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to delete page: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"success": true,
			"id":      page.ID,
			"message": "Page archived successfully",
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// MovePageHandler creates a handler for moving pages to new parent locations
func MovePageHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract page ID
		pageID, ok := args["page_id"].(string)
		if !ok || pageID == "" {
			return mcp.NewToolResultError("page_id is required"), nil
		}

		// Extract parent
		parentData, ok := args["parent"].(map[string]any)
		if !ok || len(parentData) == 0 {
			return mcp.NewToolResultError("parent is required"), nil
		}

		// Build parent object
		parent := notion.Parent{}
		if parentType, ok := parentData["type"].(string); ok {
			parent.Type = parentType

			switch parentType {
			case "page_id":
				if pageID, ok := parentData["page_id"].(string); ok && pageID != "" {
					parent.PageID = pageID
				} else {
					return mcp.NewToolResultError("page_id is required in parent when type is 'page_id'"), nil
				}
			case "data_source_id":
				if dataSourceID, ok := parentData["data_source_id"].(string); ok && dataSourceID != "" {
					parent.DataSourceID = dataSourceID
				} else {
					return mcp.NewToolResultError("data_source_id is required in parent when type is 'data_source_id'"), nil
				}
			default:
				return mcp.NewToolResultError(fmt.Sprintf("invalid parent type '%s', must be 'page_id' or 'data_source_id'", parentType)), nil
			}
		} else {
			return mcp.NewToolResultError("parent.type is required (must be 'page_id' or 'data_source_id')"), nil
		}

		// Create move request
		moveReq := &notion.MovePageRequest{
			Parent: parent,
		}

		// Move page
		page, err := client.MovePage(ctx, pageID, moveReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to move page: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"success":          true,
			"id":               page.ID,
			"url":              page.URL,
			"parent":           page.Parent,
			"message":          "Page moved successfully",
			"archived":         page.Archived,
			"last_edited_time": page.LastEditedTime,
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
