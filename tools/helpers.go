package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rgabriel/mcp-notion/notion"
)

// CreateSimplePageHandler creates a handler for creating a simple page with markdown-like content
func CreateSimplePageHandler(client notion.NotionClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract title
		title, ok := args["title"].(string)
		if !ok || title == "" {
			return mcp.NewToolResultError("title is required"), nil
		}

		// Extract content
		content, ok := args["content"].(string)
		if !ok {
			content = ""
		}

		// Determine parent
		parent := notion.Parent{}
		if parentPageID, ok := args["parent_page_id"].(string); ok && parentPageID != "" {
			parent.Type = "page_id"
			parent.PageID = parentPageID
		} else if parentDatabaseID, ok := args["parent_database_id"].(string); ok && parentDatabaseID != "" {
			parent.Type = "database_id"
			parent.DatabaseID = parentDatabaseID
		} else {
			// Default to workspace
			parent.Type = "page_id"
			parent.Workspace = true
		}

		// Build properties based on parent type
		properties := make(map[string]any)
		if parent.Type == "database_id" {
			// For database pages, use Name property
			properties["Name"] = notion.NewTitleProperty(title)
		} else {
			// For regular pages, use title property
			properties["title"] = notion.NewTitleProperty(title)
		}

		// Convert markdown content to blocks
		var children []notion.Block
		if content != "" {
			children = notion.MarkdownToBlocks(content)
		}

		// Create page request
		createReq := &notion.CreatePageRequest{
			Parent:     parent,
			Properties: properties,
			Children:   children,
		}

		// Create page
		page, err := client.CreatePage(ctx, createReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create simple page: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"success": true,
			"id":      page.ID,
			"url":     page.URL,
			"message": fmt.Sprintf("Page '%s' created successfully", title),
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// AppendToPageHandler creates a handler for appending content to an existing page
func AppendToPageHandler(client notion.NotionClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract page ID
		pageID, ok := args["page_id"].(string)
		if !ok || pageID == "" {
			return mcp.NewToolResultError("page_id is required"), nil
		}

		// Extract content
		content, ok := args["content"].(string)
		if !ok || content == "" {
			return mcp.NewToolResultError("content is required"), nil
		}

		// Convert markdown content to blocks
		blocks := notion.MarkdownToBlocks(content)

		if len(blocks) == 0 {
			return mcp.NewToolResultError("no valid blocks generated from content"), nil
		}

		// Append blocks
		appendReq := &notion.AppendBlockChildrenRequest{
			Children: blocks,
		}

		result, err := client.AppendBlockChildren(ctx, pageID, appendReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to append to page: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"success": true,
			"count":   len(result.Results),
			"message": fmt.Sprintf("Successfully appended %d block(s) to page", len(result.Results)),
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// SearchInDatabaseHandler creates a handler for searching within a specific database
func SearchInDatabaseHandler(client notion.NotionClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract database ID
		databaseID, ok := args["database_id"].(string)
		if !ok || databaseID == "" {
			return mcp.NewToolResultError("database_id is required"), nil
		}

		// Extract query
		query, ok := args["query"].(string)
		if !ok || query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}

		// Build query request with text filter
		queryReq := &notion.QueryDatabaseRequest{
			PageSize: 100,
		}

		// Optional property filters (simplified)
		if propertyFilters, ok := args["property_filters"].(map[string]any); ok && len(propertyFilters) > 0 {
			// Build a simple filter from property_filters
			// This is a simplified version - users can use query_database for complex filters
			filters := make([]map[string]any, 0)
			for propName, propValue := range propertyFilters {
				// Try to infer filter type from value
				switch v := propValue.(type) {
				case string:
					// Assume text/select filter
					filters = append(filters, map[string]any{
						"property": propName,
						"rich_text": map[string]any{
							"contains": v,
						},
					})
				case bool:
					// Checkbox filter
					filters = append(filters, map[string]any{
						"property": propName,
						"checkbox": map[string]any{
							"equals": v,
						},
					})
				case float64:
					// Number filter
					filters = append(filters, map[string]any{
						"property": propName,
						"number": map[string]any{
							"equals": v,
						},
					})
				}
			}

			if len(filters) > 0 {
				if len(filters) == 1 {
					queryReq.Filter = filters[0]
				} else {
					queryReq.Filter = map[string]any{
						"and": filters,
					}
				}
			}
		}

		// Query database
		result, err := client.QueryDatabase(ctx, databaseID, queryReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search in database: %v", err)), nil
		}

		// Filter results by query text (search in title and other text properties)
		matchedResults := make([]notion.Page, 0)
		for _, page := range result.Results {
			// Search in properties for the query text
			matches := false
			for _, prop := range page.Properties {
				if propMap, ok := prop.(map[string]any); ok {
					// Check title property
					if titleArray, ok := propMap["title"].([]any); ok {
						for _, titleItem := range titleArray {
							if titleMap, ok := titleItem.(map[string]any); ok {
								if textMap, ok := titleMap["text"].(map[string]any); ok {
									if content, ok := textMap["content"].(string); ok {
										if containsIgnoreCase(content, query) {
											matches = true
											break
										}
									}
								}
								if plainText, ok := titleMap["plain_text"].(string); ok {
									if containsIgnoreCase(plainText, query) {
										matches = true
										break
									}
								}
							}
						}
					}
					// Check rich_text property
					if richTextArray, ok := propMap["rich_text"].([]any); ok {
						for _, rtItem := range richTextArray {
							if rtMap, ok := rtItem.(map[string]any); ok {
								if plainText, ok := rtMap["plain_text"].(string); ok {
									if containsIgnoreCase(plainText, query) {
										matches = true
										break
									}
								}
							}
						}
					}
				}
			}
			if matches {
				matchedResults = append(matchedResults, page)
			}
		}

		// If no matches with text search, return all results from filter
		if len(matchedResults) == 0 {
			matchedResults = result.Results
		}

		// Format response
		response := map[string]any{
			"count":   len(matchedResults),
			"results": matchedResults,
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// containsIgnoreCase is a case-insensitive substring check helper.
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
