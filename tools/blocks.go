package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rgabriel/mcp-notion/notion"
)

// GetBlockHandler creates a handler for getting a specific block
func GetBlockHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract block ID
		blockID, ok := args["block_id"].(string)
		if !ok || blockID == "" {
			return mcp.NewToolResultError("block_id is required"), nil
		}

		// Get block
		block, err := client.GetBlock(ctx, blockID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get block: %v", err)), nil
		}

		// Format response
		jsonData, err := json.MarshalIndent(block, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// GetBlockChildrenHandler creates a handler for getting children of a block
func GetBlockChildrenHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract block ID
		blockID, ok := args["block_id"].(string)
		if !ok || blockID == "" {
			return mcp.NewToolResultError("block_id is required"), nil
		}

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

		// Get block children
		result, err := client.GetBlockChildren(ctx, blockID, pageSize, startCursor)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get block children: %v", err)), nil
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

// AppendBlocksHandler creates a handler for appending blocks to a parent
func AppendBlocksHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract block ID (parent)
		blockID, ok := args["block_id"].(string)
		if !ok || blockID == "" {
			return mcp.NewToolResultError("block_id is required (parent block or page ID)"), nil
		}

		// Extract children blocks
		childrenData, ok := args["children"].([]any)
		if !ok || len(childrenData) == 0 {
			return mcp.NewToolResultError("children is required (array of block objects)"), nil
		}

		// Convert to Block array
		children := make([]notion.Block, 0, len(childrenData))
		for i, child := range childrenData {
			if childMap, ok := child.(map[string]any); ok {
				childJSON, err := json.Marshal(childMap)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("invalid block format at index %d: %v", i, err)), nil
				}
				var block notion.Block
				if err := json.Unmarshal(childJSON, &block); err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("invalid block structure at index %d: %v", i, err)), nil
				}
				children = append(children, block)
			}
		}

		// Append blocks
		appendReq := &notion.AppendBlockChildrenRequest{
			Children: children,
		}

		result, err := client.AppendBlockChildren(ctx, blockID, appendReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to append blocks: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"success": true,
			"count":   len(result.Results),
			"blocks":  result.Results,
			"message": fmt.Sprintf("Successfully appended %d block(s)", len(result.Results)),
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// UpdateBlockHandler creates a handler for updating a block
func UpdateBlockHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract block ID
		blockID, ok := args["block_id"].(string)
		if !ok || blockID == "" {
			return mcp.NewToolResultError("block_id is required"), nil
		}

		// Build block update - we need the block type-specific field
		// The args should contain the block type field with updated content
		blockData := make(map[string]any)
		for key, value := range args {
			if key != "block_id" {
				blockData[key] = value
			}
		}

		if len(blockData) == 0 {
			return mcp.NewToolResultError("at least one block property must be provided to update"), nil
		}

		// Convert to Block
		blockJSON, err := json.Marshal(blockData)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid block format: %v", err)), nil
		}

		var block notion.Block
		if err := json.Unmarshal(blockJSON, &block); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid block structure: %v", err)), nil
		}

		// Update block
		updated, err := client.UpdateBlock(ctx, blockID, &block)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to update block: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"success": true,
			"block":   updated,
			"message": "Block updated successfully",
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// DeleteBlockHandler creates a handler for deleting a block
func DeleteBlockHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract block ID
		blockID, ok := args["block_id"].(string)
		if !ok || blockID == "" {
			return mcp.NewToolResultError("block_id is required"), nil
		}

		// Delete block
		deleted, err := client.DeleteBlock(ctx, blockID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to delete block: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"success": true,
			"id":      deleted.ID,
			"message": "Block deleted successfully",
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
