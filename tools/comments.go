package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rgabriel/mcp-notion/notion"
)

// GetCommentsHandler creates a handler for getting comments
func GetCommentsHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract block_id or page_id (at least one is required)
		blockID, _ := args["block_id"].(string)
		pageID, _ := args["page_id"].(string)

		if blockID == "" && pageID == "" {
			return mcp.NewToolResultError("either block_id or page_id is required"), nil
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

		// Get comments
		result, err := client.GetComments(ctx, blockID, pageID, pageSize, startCursor)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get comments: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"count":    len(result.Results),
			"comments": result.Results,
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

// CreateCommentHandler creates a handler for creating a comment
func CreateCommentHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract parent
		parentMap, ok := args["parent"].(map[string]any)
		if !ok || len(parentMap) == 0 {
			return mcp.NewToolResultError("parent is required (object with page_id)"), nil
		}

		parent := notion.Parent{}
		if pageID, ok := parentMap["page_id"].(string); ok && pageID != "" {
			parent.Type = "page_id"
			parent.PageID = pageID
		} else {
			return mcp.NewToolResultError("parent must contain page_id"), nil
		}

		// Extract rich_text
		richTextData, ok := args["rich_text"].([]any)
		if !ok || len(richTextData) == 0 {
			return mcp.NewToolResultError("rich_text is required (array of rich text objects)"), nil
		}

		// Convert to RichText array
		richTextJSON, err := json.Marshal(richTextData)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid rich_text format: %v", err)), nil
		}
		var richText []notion.RichText
		if err := json.Unmarshal(richTextJSON, &richText); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid rich_text structure: %v", err)), nil
		}

		// Build create comment request
		createReq := &notion.CreateCommentRequest{
			Parent:   parent,
			RichText: richText,
		}

		// Optional discussion_id (for replies)
		if discussionID, ok := args["discussion_id"].(string); ok && discussionID != "" {
			createReq.DiscussionID = discussionID
		}

		// Create comment
		comment, err := client.CreateComment(ctx, createReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create comment: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"success":       true,
			"id":            comment.ID,
			"discussion_id": comment.DiscussionID,
			"message":       "Comment created successfully",
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
