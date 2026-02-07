package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rgabriel/mcp-notion/notion"
)

// BatchCreatePagesHandler creates a handler for batch page creation
func BatchCreatePagesHandler(client notion.NotionClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract pages array
		pagesData, ok := args["pages"].([]any)
		if !ok || len(pagesData) == 0 {
			return mcp.NewToolResultError("pages array is required and must not be empty"), nil
		}

		// Extract continue_on_error flag
		continueOnError := false
		if flag, ok := args["continue_on_error"].(bool); ok {
			continueOnError = flag
		}

		// Track results
		created := make([]map[string]string, 0)
		failed := make([]map[string]any, 0)

		// Process each page sequentially
		for i, pageData := range pagesData {
			pageMap, ok := pageData.(map[string]any)
			if !ok {
				failureEntry := map[string]any{
					"index": i,
					"error": "invalid page object format",
				}
				failed = append(failed, failureEntry)
				if !continueOnError {
					break
				}
				continue
			}

			// Build CreatePageRequest
			createReq, err := buildCreatePageRequest(pageMap)
			if err != nil {
				failureEntry := map[string]any{
					"index": i,
					"error": fmt.Sprintf("invalid page data: %v", err),
				}
				failed = append(failed, failureEntry)
				if !continueOnError {
					break
				}
				continue
			}

			// Create the page
			page, err := client.CreatePage(ctx, createReq)
			if err != nil {
				failureEntry := map[string]any{
					"index": i,
					"error": fmt.Sprintf("failed to create page: %v", err),
				}
				failed = append(failed, failureEntry)
				if !continueOnError {
					break
				}
				continue
			}

			// Record success
			created = append(created, map[string]string{
				"id":  page.ID,
				"url": page.URL,
			})

			// Rate limiting: wait 350ms between requests
			if i < len(pagesData)-1 {
				time.Sleep(350 * time.Millisecond)
			}
		}

		// Build response
		response := map[string]any{
			"created": created,
			"failed":  failed,
			"summary": map[string]int{
				"total":     len(pagesData),
				"succeeded": len(created),
				"failed":    len(failed),
			},
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// BatchUpdatePagesHandler creates a handler for batch page updates
func BatchUpdatePagesHandler(client notion.NotionClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract page_ids array
		pageIDsData, ok := args["page_ids"].([]any)
		if !ok || len(pageIDsData) == 0 {
			return mcp.NewToolResultError("page_ids array is required and must not be empty"), nil
		}

		// Convert to string array
		pageIDs := make([]string, 0, len(pageIDsData))
		for _, id := range pageIDsData {
			if idStr, ok := id.(string); ok && idStr != "" {
				pageIDs = append(pageIDs, idStr)
			}
		}

		if len(pageIDs) == 0 {
			return mcp.NewToolResultError("page_ids must contain valid page ID strings"), nil
		}

		// Extract continue_on_error flag
		continueOnError := false
		if flag, ok := args["continue_on_error"].(bool); ok {
			continueOnError = flag
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

		// Track results
		updated := make([]map[string]string, 0)
		failed := make([]map[string]any, 0)

		// Process each page sequentially
		for i, pageID := range pageIDs {
			page, err := client.UpdatePage(ctx, pageID, updateReq)
			if err != nil {
				failureEntry := map[string]any{
					"page_id": pageID,
					"error":   fmt.Sprintf("failed to update page: %v", err),
				}
				failed = append(failed, failureEntry)
				if !continueOnError {
					break
				}
				continue
			}

			// Record success
			updated = append(updated, map[string]string{
				"id":  page.ID,
				"url": page.URL,
			})

			// Rate limiting: wait 350ms between requests
			if i < len(pageIDs)-1 {
				time.Sleep(350 * time.Millisecond)
			}
		}

		// Build response
		response := map[string]any{
			"updated": updated,
			"failed":  failed,
			"summary": map[string]int{
				"total":     len(pageIDs),
				"succeeded": len(updated),
				"failed":    len(failed),
			},
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// BatchDeletePagesHandler creates a handler for batch page deletion (archiving)
func BatchDeletePagesHandler(client notion.NotionClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract page_ids array
		pageIDsData, ok := args["page_ids"].([]any)
		if !ok || len(pageIDsData) == 0 {
			return mcp.NewToolResultError("page_ids array is required and must not be empty"), nil
		}

		// Convert to string array
		pageIDs := make([]string, 0, len(pageIDsData))
		for _, id := range pageIDsData {
			if idStr, ok := id.(string); ok && idStr != "" {
				pageIDs = append(pageIDs, idStr)
			}
		}

		if len(pageIDs) == 0 {
			return mcp.NewToolResultError("page_ids must contain valid page ID strings"), nil
		}

		// Extract continue_on_error flag
		continueOnError := false
		if flag, ok := args["continue_on_error"].(bool); ok {
			continueOnError = flag
		}

		// Track results
		deleted := make([]string, 0)
		failed := make([]map[string]any, 0)

		// Archive pages sequentially
		archived := true
		updateReq := &notion.UpdatePageRequest{
			Archived: &archived,
		}

		for i, pageID := range pageIDs {
			_, err := client.UpdatePage(ctx, pageID, updateReq)
			if err != nil {
				failureEntry := map[string]any{
					"page_id": pageID,
					"error":   fmt.Sprintf("failed to delete page: %v", err),
				}
				failed = append(failed, failureEntry)
				if !continueOnError {
					break
				}
				continue
			}

			// Record success
			deleted = append(deleted, pageID)

			// Rate limiting: wait 350ms between requests
			if i < len(pageIDs)-1 {
				time.Sleep(350 * time.Millisecond)
			}
		}

		// Build response
		response := map[string]any{
			"deleted": deleted,
			"failed":  failed,
			"summary": map[string]int{
				"total":     len(pageIDs),
				"succeeded": len(deleted),
				"failed":    len(failed),
			},
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// buildCreatePageRequest constructs a CreatePageRequest from a map
func buildCreatePageRequest(pageMap map[string]any) (*notion.CreatePageRequest, error) {
	// Extract parent
	parentMap, ok := pageMap["parent"].(map[string]any)
	if !ok || len(parentMap) == 0 {
		return nil, fmt.Errorf("parent is required")
	}

	parent := notion.Parent{}
	if pageID, ok := parentMap["page_id"].(string); ok && pageID != "" {
		parent.Type = "page_id"
		parent.PageID = pageID
	} else if dbID, ok := parentMap["database_id"].(string); ok && dbID != "" {
		parent.Type = "database_id"
		parent.DatabaseID = dbID
	} else {
		return nil, fmt.Errorf("parent must contain either page_id or database_id")
	}

	// Extract properties
	properties, ok := pageMap["properties"].(map[string]any)
	if !ok || len(properties) == 0 {
		return nil, fmt.Errorf("properties is required")
	}

	// Build request
	createReq := &notion.CreatePageRequest{
		Parent:     parent,
		Properties: properties,
	}

	// Optional children (blocks)
	if childrenData, ok := pageMap["children"].([]any); ok && len(childrenData) > 0 {
		children := make([]notion.Block, 0, len(childrenData))
		for _, child := range childrenData {
			if childMap, ok := child.(map[string]any); ok {
				childJSON, err := json.Marshal(childMap)
				if err != nil {
					continue
				}
				var block notion.Block
				if err := json.Unmarshal(childJSON, &block); err == nil {
					children = append(children, block)
				}
			}
		}
		if len(children) > 0 {
			createReq.Children = children
		}
	}

	// Optional icon
	if iconData, ok := pageMap["icon"].(map[string]any); ok && len(iconData) > 0 {
		iconJSON, _ := json.Marshal(iconData)
		var icon notion.Icon
		if err := json.Unmarshal(iconJSON, &icon); err == nil {
			createReq.Icon = &icon
		}
	}

	// Optional cover
	if coverData, ok := pageMap["cover"].(map[string]any); ok && len(coverData) > 0 {
		coverJSON, _ := json.Marshal(coverData)
		var cover notion.Cover
		if err := json.Unmarshal(coverJSON, &cover); err == nil {
			createReq.Cover = &cover
		}
	}

	return createReq, nil
}
