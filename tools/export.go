package tools

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rgabriel/mcp-notion/notion"
)

// ExportPageAsMarkdownHandler creates a handler for exporting pages as markdown
func ExportPageAsMarkdownHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract page ID
		pageID, ok := args["page_id"].(string)
		if !ok || pageID == "" {
			return mcp.NewToolResultError("page_id is required"), nil
		}

		// Extract options
		includeFrontmatter := false
		if flag, ok := args["frontmatter"].(bool); ok {
			includeFrontmatter = flag
		}

		includeChildren := false
		if flag, ok := args["include_children"].(bool); ok {
			includeChildren = flag
		}

		// Get page metadata
		page, err := client.GetPage(ctx, pageID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get page: %v", err)), nil
		}

		// Extract page title
		title := extractPageTitle(page.Properties)

		// Get page content
		blocks, err := client.GetBlockChildren(ctx, pageID, 100, "")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get page content: %v", err)), nil
		}

		// Build markdown
		var markdown strings.Builder

		// Add frontmatter if requested
		if includeFrontmatter {
			markdown.WriteString("---\n")
			markdown.WriteString(fmt.Sprintf("title: %s\n", title))
			markdown.WriteString(fmt.Sprintf("notion_url: %s\n", page.URL))
			markdown.WriteString(fmt.Sprintf("created: %s\n", page.CreatedTime.Format(time.RFC3339)))
			markdown.WriteString(fmt.Sprintf("last_edited: %s\n", page.LastEditedTime.Format(time.RFC3339)))
			markdown.WriteString("---\n\n")
		}

		// Add title as H1
		markdown.WriteString("# " + title + "\n\n")

		// Convert blocks to markdown
		allBlocks := blocks.Results

		// Handle pagination
		for blocks.HasMore && blocks.NextCursor != "" {
			nextBlocks, err := client.GetBlockChildren(ctx, pageID, 100, blocks.NextCursor)
			if err != nil {
				break
			}
			allBlocks = append(allBlocks, nextBlocks.Results...)
			blocks = nextBlocks
		}

		// Convert blocks
		for _, block := range allBlocks {
			markdown.WriteString(notion.BlockToMarkdown(&block, 0))

			// Recursively get children if requested and block has children
			if includeChildren && block.HasChildren {
				childMarkdown, err := exportBlockChildren(ctx, client, block.ID, 1)
				if err == nil {
					markdown.WriteString(childMarkdown)
				}
			}
		}

		// Build response
		response := map[string]any{
			"markdown": markdown.String(),
			"metadata": map[string]any{
				"title":            title,
				"notion_url":       page.URL,
				"created_time":     page.CreatedTime.Format(time.RFC3339),
				"last_edited_time": page.LastEditedTime.Format(time.RFC3339),
			},
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// ExportDatabaseAsCSVHandler creates a handler for exporting databases as CSV
func ExportDatabaseAsCSVHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract database ID
		databaseID, ok := args["database_id"].(string)
		if !ok || databaseID == "" {
			return mcp.NewToolResultError("database_id is required"), nil
		}

		// Extract options
		includeArchived := false
		if flag, ok := args["include_archived"].(bool); ok {
			includeArchived = flag
		}

		// Build query request
		queryReq := &notion.QueryDatabaseRequest{
			PageSize: 100,
		}

		// Apply filter if provided
		if filterData, ok := args["filter"].(map[string]any); ok && len(filterData) > 0 {
			queryReq.Filter = filterData
		}

		// Get database to understand schema
		database, err := client.GetDatabase(ctx, databaseID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get database: %v", err)), nil
		}

		// Query database
		result, err := client.QueryDatabase(ctx, databaseID, queryReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to query database: %v", err)), nil
		}

		// Collect all pages
		allPages := result.Results

		// Handle pagination
		for result.HasMore && result.NextCursor != "" {
			queryReq.StartCursor = result.NextCursor
			nextResult, err := client.QueryDatabase(ctx, databaseID, queryReq)
			if err != nil {
				break
			}
			allPages = append(allPages, nextResult.Results...)
			result = nextResult
		}

		// Filter out archived if needed
		if !includeArchived {
			filteredPages := make([]notion.Page, 0)
			for _, page := range allPages {
				if !page.Archived {
					filteredPages = append(filteredPages, page)
				}
			}
			allPages = filteredPages
		}

		// Get property names from database schema
		propertyNames := make([]string, 0)
		for propName := range database.Properties {
			propertyNames = append(propertyNames, propName)
		}

		// Create CSV buffer
		var csvBuffer bytes.Buffer
		csvWriter := csv.NewWriter(&csvBuffer)

		// Write header row
		if err := csvWriter.Write(propertyNames); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to write CSV header: %v", err)), nil
		}

		// Write data rows
		for _, page := range allPages {
			row := make([]string, len(propertyNames))
			for i, propName := range propertyNames {
				row[i] = extractPropertyValue(page.Properties[propName])
			}
			if err := csvWriter.Write(row); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to write CSV row: %v", err)), nil
			}
		}

		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("CSV writer error: %v", err)), nil
		}

		// Build response
		response := map[string]any{
			"csv":       csvBuffer.String(),
			"row_count": len(allPages),
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// exportBlockChildren recursively exports child blocks
func exportBlockChildren(ctx context.Context, client *notion.Client, blockID string, indent int) (string, error) {
	blocks, err := client.GetBlockChildren(ctx, blockID, 100, "")
	if err != nil {
		return "", err
	}

	var markdown strings.Builder

	allBlocks := blocks.Results

	// Handle pagination
	for blocks.HasMore && blocks.NextCursor != "" {
		nextBlocks, err := client.GetBlockChildren(ctx, blockID, 100, blocks.NextCursor)
		if err != nil {
			break
		}
		allBlocks = append(allBlocks, nextBlocks.Results...)
		blocks = nextBlocks
	}

	// Convert blocks
	for _, block := range allBlocks {
		markdown.WriteString(notion.BlockToMarkdown(&block, indent))

		// Recursively get children
		if block.HasChildren {
			childMarkdown, err := exportBlockChildren(ctx, client, block.ID, indent+1)
			if err == nil {
				markdown.WriteString(childMarkdown)
			}
		}
	}

	return markdown.String(), nil
}

// extractPageTitle extracts the title from page properties
func extractPageTitle(properties map[string]any) string {
	// Try common title property names
	titleNames := []string{"title", "Title", "Name", "name"}

	for _, titleName := range titleNames {
		if prop, ok := properties[titleName]; ok {
			if propMap, ok := prop.(map[string]any); ok {
				// Title property
				if titleArray, ok := propMap["title"].([]any); ok {
					var titleParts []string
					for _, item := range titleArray {
						if itemMap, ok := item.(map[string]any); ok {
							if plainText, ok := itemMap["plain_text"].(string); ok {
								titleParts = append(titleParts, plainText)
							}
						}
					}
					if len(titleParts) > 0 {
						return strings.Join(titleParts, "")
					}
				}
			}
		}
	}

	return "Untitled"
}

// extractPropertyValue extracts a string value from any property type
func extractPropertyValue(property any) string {
	if property == nil {
		return ""
	}

	propMap, ok := property.(map[string]any)
	if !ok {
		return ""
	}

	// Title property
	if titleArray, ok := propMap["title"].([]any); ok {
		return extractRichTextArray(titleArray)
	}

	// Rich text property
	if richTextArray, ok := propMap["rich_text"].([]any); ok {
		return extractRichTextArray(richTextArray)
	}

	// Number property
	if number, ok := propMap["number"].(float64); ok {
		return fmt.Sprintf("%v", number)
	}

	// Select property
	if selectMap, ok := propMap["select"].(map[string]any); ok {
		if name, ok := selectMap["name"].(string); ok {
			return name
		}
	}

	// Multi-select property
	if multiSelectArray, ok := propMap["multi_select"].([]any); ok {
		var names []string
		for _, item := range multiSelectArray {
			if itemMap, ok := item.(map[string]any); ok {
				if name, ok := itemMap["name"].(string); ok {
					names = append(names, name)
				}
			}
		}
		return strings.Join(names, "; ")
	}

	// Date property
	if dateMap, ok := propMap["date"].(map[string]any); ok {
		if start, ok := dateMap["start"].(string); ok {
			if end, ok := dateMap["end"].(string); ok && end != "" {
				return start + " → " + end
			}
			return start
		}
	}

	// Checkbox property
	if checkbox, ok := propMap["checkbox"].(bool); ok {
		if checkbox {
			return "true"
		}
		return "false"
	}

	// URL property
	if url, ok := propMap["url"].(string); ok {
		return url
	}

	// Email property
	if email, ok := propMap["email"].(string); ok {
		return email
	}

	// Phone number property
	if phone, ok := propMap["phone_number"].(string); ok {
		return phone
	}

	// People property
	if peopleArray, ok := propMap["people"].([]any); ok {
		var names []string
		for _, person := range peopleArray {
			if personMap, ok := person.(map[string]any); ok {
				if name, ok := personMap["name"].(string); ok {
					names = append(names, name)
				}
			}
		}
		return strings.Join(names, "; ")
	}

	// Files property
	if filesArray, ok := propMap["files"].([]any); ok {
		var fileNames []string
		for _, file := range filesArray {
			if fileMap, ok := file.(map[string]any); ok {
				if name, ok := fileMap["name"].(string); ok {
					fileNames = append(fileNames, name)
				}
			}
		}
		return strings.Join(fileNames, "; ")
	}

	// Created/edited time
	if createdTime, ok := propMap["created_time"].(string); ok {
		return createdTime
	}
	if lastEditedTime, ok := propMap["last_edited_time"].(string); ok {
		return lastEditedTime
	}

	return ""
}

// extractRichTextArray extracts plain text from a rich text array
func extractRichTextArray(richTextArray []any) string {
	var parts []string
	for _, item := range richTextArray {
		if itemMap, ok := item.(map[string]any); ok {
			if plainText, ok := itemMap["plain_text"].(string); ok {
				parts = append(parts, plainText)
			}
		}
	}
	return strings.Join(parts, "")
}
