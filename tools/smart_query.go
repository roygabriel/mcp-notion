package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rgabriel/mcp-notion/notion"
)

// GetRecentlyEditedHandler creates a handler for getting recently edited pages
func GetRecentlyEditedHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract parameters
		days := 7
		if daysFloat, ok := args["days"].(float64); ok && daysFloat > 0 {
			days = int(daysFloat)
		}

		limit := 50
		if limitFloat, ok := args["limit"].(float64); ok && limitFloat > 0 {
			limit = int(limitFloat)
		}

		databaseID := ""
		if dbID, ok := args["database_id"].(string); ok {
			databaseID = dbID
		}

		// Calculate cutoff time
		cutoffTime := time.Now().AddDate(0, 0, -days)

		// Search with sort by last_edited_time
		searchReq := &notion.SearchRequest{
			Sort: map[string]any{
				"direction": "descending",
				"timestamp": "last_edited_time",
			},
			PageSize: 100,
		}

		// Search
		result, err := client.Search(ctx, searchReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to search: %v", err)), nil
		}

		// Collect results
		allResults := result.Results

		// Handle pagination (get more if needed)
		for result.HasMore && result.NextCursor != "" && len(allResults) < limit*2 {
			searchReq.StartCursor = result.NextCursor
			nextResult, err := client.Search(ctx, searchReq)
			if err != nil {
				break
			}
			allResults = append(allResults, nextResult.Results...)
			result = nextResult
		}

		// Filter by time and database
		filtered := make([]notion.SearchResult, 0)
		for _, item := range allResults {
			// Check time filter
			if !item.LastEditedTime.IsZero() && item.LastEditedTime.Before(cutoffTime) {
				continue
			}

			// Check database filter
			if databaseID != "" && item.Parent != nil {
				if item.Parent.DatabaseID != databaseID {
					continue
				}
			}

			filtered = append(filtered, item)

			// Stop if we have enough
			if len(filtered) >= limit {
				break
			}
		}

		// Format response
		response := map[string]any{
			"count":        len(filtered),
			"cutoff_date":  cutoffTime.Format("2006-01-02"),
			"pages":        filtered,
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// GetMyTasksHandler creates a handler for getting tasks assigned to current user
func GetMyTasksHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Get bot user
		botUser, err := client.GetBotUser(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get bot user: %v", err)), nil
		}

		// Extract parameters
		databaseID := ""
		if dbID, ok := args["database_id"].(string); ok && dbID != "" {
			databaseID = dbID
		}

		statuses := []string{"In Progress", "To Do", "Not Started"}
		if statusesData, ok := args["statuses"].([]any); ok && len(statusesData) > 0 {
			statuses = make([]string, 0, len(statusesData))
			for _, status := range statusesData {
				if statusStr, ok := status.(string); ok {
					statuses = append(statuses, statusStr)
				}
			}
		}

		includeOverdue := true
		if flag, ok := args["include_overdue"].(bool); ok {
			includeOverdue = flag
		}

		limit := 50
		if limitFloat, ok := args["limit"].(float64); ok && limitFloat > 0 {
			limit = int(limitFloat)
		}

		// If no database_id, search for task databases
		if databaseID == "" {
			searchReq := &notion.SearchRequest{
				Filter: map[string]any{
					"property": "object",
					"value":    "database",
				},
				PageSize: 20,
			}

			searchResult, err := client.Search(ctx, searchReq)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to search for databases: %v", err)), nil
			}

			// Look for databases with "task" in the name
			for _, result := range searchResult.Results {
				title := notion.ExtractPlainText(result.Title)
				if strings.Contains(strings.ToLower(title), "task") {
					databaseID = result.ID
					break
				}
			}

			if databaseID == "" {
				return mcp.NewToolResultError("no database_id provided and no task database found. Please specify a database_id."), nil
			}
		}

		// Get database to understand schema
		database, err := client.GetDatabase(ctx, databaseID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get database: %v", err)), nil
		}

		// Find property names
		statusPropName := findPropertyName(database.Properties, []string{"Status", "State", "Progress", "Stage"})
		assigneePropName := findPropertyName(database.Properties, []string{"Assigned To", "Assignee", "Owner", "Assigned"})
		dueDatePropName := findPropertyName(database.Properties, []string{"Due Date", "Due", "Deadline"})

		// Build query filter
		var filters []map[string]any

		// Status filter
		if statusPropName != "" && len(statuses) > 0 {
			if len(statuses) == 1 {
				filters = append(filters, map[string]any{
					"property": statusPropName,
					"select": map[string]any{
						"equals": statuses[0],
					},
				})
			} else {
				statusFilters := make([]map[string]any, 0, len(statuses))
				for _, status := range statuses {
					statusFilters = append(statusFilters, map[string]any{
						"property": statusPropName,
						"select": map[string]any{
							"equals": status,
						},
					})
				}
				filters = append(filters, map[string]any{
					"or": statusFilters,
				})
			}
		}

		// Assignee filter (if property exists)
		if assigneePropName != "" {
			filters = append(filters, map[string]any{
				"property": assigneePropName,
				"people": map[string]any{
					"contains": botUser.ID,
				},
			})
		}

		// Build query request
		queryReq := &notion.QueryDatabaseRequest{
			PageSize: 100,
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

		// Filter by due date if needed
		tasks := make([]map[string]any, 0)
		now := time.Now()

		for _, page := range allPages {
			task := map[string]any{
				"id":  page.ID,
				"url": page.URL,
			}

			// Extract title
			if titleProp, ok := page.Properties["Name"]; ok {
				task["title"] = extractPropertyValue(titleProp)
			} else if titleProp, ok := page.Properties["Title"]; ok {
				task["title"] = extractPropertyValue(titleProp)
			}

			// Extract status
			if statusPropName != "" {
				if statusProp, ok := page.Properties[statusPropName]; ok {
					task["status"] = extractPropertyValue(statusProp)
				}
			}

			// Extract and check due date
			if dueDatePropName != "" {
				if dueDateProp, ok := page.Properties[dueDatePropName]; ok {
					dueDateStr := extractPropertyValue(dueDateProp)
					task["due_date"] = dueDateStr

					// Check if overdue
					if includeOverdue && dueDateStr != "" {
						if dueDate, err := time.Parse("2006-01-02", dueDateStr[:10]); err == nil {
							if dueDate.Before(now) {
								task["overdue"] = true
							}
						}
					}
				}
			}

			tasks = append(tasks, task)

			// Limit results
			if len(tasks) >= limit {
				break
			}
		}

		// Format response
		response := map[string]any{
			"count":       len(tasks),
			"database_id": databaseID,
			"tasks":       tasks,
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// GetRelatedPagesHandler creates a handler for getting related pages
func GetRelatedPagesHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract page ID
		pageID, ok := args["page_id"].(string)
		if !ok || pageID == "" {
			return mcp.NewToolResultError("page_id is required"), nil
		}

		// Extract options
		relationProperty := ""
		if prop, ok := args["relation_property"].(string); ok {
			relationProperty = prop
		}

		includeProperties := false
		if flag, ok := args["include_properties"].(bool); ok {
			includeProperties = flag
		}

		// Get page
		page, err := client.GetPage(ctx, pageID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get page: %v", err)), nil
		}

		// Find relation properties
		relations := make(map[string]any)

		for propName, propValue := range page.Properties {
			// Skip if specific property requested and this isn't it
			if relationProperty != "" && propName != relationProperty {
				continue
			}

			// Check if it's a relation property
			if propMap, ok := propValue.(map[string]any); ok {
				if relArray, ok := propMap["relation"].([]any); ok && len(relArray) > 0 {
					// Extract related page IDs
					relatedPages := make([]map[string]any, 0)

					for _, rel := range relArray {
						if relMap, ok := rel.(map[string]any); ok {
							if relID, ok := relMap["id"].(string); ok {
								relatedPage := map[string]any{
									"id": relID,
								}

								// Get full page details if requested
								if includeProperties {
									fullPage, err := client.GetPage(ctx, relID)
									if err == nil {
										// Extract title
										title := extractPageTitle(fullPage.Properties)
										relatedPage["title"] = title
										relatedPage["url"] = fullPage.URL
										relatedPage["properties"] = fullPage.Properties
									} else {
										relatedPage["title"] = "Unknown"
									}
								}

								relatedPages = append(relatedPages, relatedPage)

								// Rate limiting for multiple page fetches
								if includeProperties && len(relatedPages) < len(relArray) {
									time.Sleep(350 * time.Millisecond)
								}
							}
						}
					}

					if len(relatedPages) > 0 {
						relations[propName] = relatedPages
					}
				}
			}
		}

		// Format response
		response := map[string]any{
			"page_id":   pageID,
			"relations": relations,
			"count":     len(relations),
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// findPropertyName finds a property by checking multiple possible names
func findPropertyName(properties map[string]notion.PropertyDef, possibleNames []string) string {
	for _, name := range possibleNames {
		if _, exists := properties[name]; exists {
			return name
		}
	}
	return ""
}
