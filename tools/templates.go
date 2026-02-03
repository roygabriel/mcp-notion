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

// Template represents a page template
type Template struct {
	Name        string
	Description string
	Blocks      func() []notion.Block
}

// GetTemplates returns all available templates
func GetTemplates() map[string]Template {
	return map[string]Template{
		"meeting_notes": {
			Name:        "meeting_notes",
			Description: "Meeting notes template with sections for attendees, agenda, discussion, decisions, and action items",
			Blocks:      meetingNotesTemplate,
		},
		"daily_log": {
			Name:        "daily_log",
			Description: "Daily log template for tracking priorities, accomplishments, blockers, and plans",
			Blocks:      dailyLogTemplate,
		},
		"project_brief": {
			Name:        "project_brief",
			Description: "Project brief template with overview, goals, timeline, stakeholders, resources, and risks",
			Blocks:      projectBriefTemplate,
		},
		"sprint_planning": {
			Name:        "sprint_planning",
			Description: "Sprint planning template with sprint goal, capacity, stories, dependencies, and risks",
			Blocks:      sprintPlanningTemplate,
		},
		"retrospective": {
			Name:        "retrospective",
			Description: "Retrospective template with sections for what went well, improvements, and action items",
			Blocks:      retrospectiveTemplate,
		},
	}
}

// meetingNotesTemplate generates meeting notes blocks
func meetingNotesTemplate() []notion.Block {
	return []notion.Block{
		notion.NewHeadingBlock(2, "Attendees"),
		notion.NewParagraphBlock(""),
		notion.NewHeadingBlock(2, "Agenda"),
		notion.NewBulletedListItemBlock("Topic 1"),
		notion.NewBulletedListItemBlock("Topic 2"),
		notion.NewHeadingBlock(2, "Discussion"),
		notion.NewParagraphBlock(""),
		notion.NewHeadingBlock(2, "Decisions"),
		notion.NewBulletedListItemBlock("Decision 1"),
		notion.NewHeadingBlock(2, "Action Items"),
		notion.NewToDoBlock("Action item 1", false),
		notion.NewToDoBlock("Action item 2", false),
	}
}

// dailyLogTemplate generates daily log blocks
func dailyLogTemplate() []notion.Block {
	return []notion.Block{
		notion.NewHeadingBlock(2, "Morning Priorities"),
		notion.NewToDoBlock("Priority 1", false),
		notion.NewToDoBlock("Priority 2", false),
		notion.NewToDoBlock("Priority 3", false),
		notion.NewHeadingBlock(2, "What I Did Today"),
		notion.NewParagraphBlock(""),
		notion.NewHeadingBlock(2, "Blockers/Issues"),
		notion.NewParagraphBlock(""),
		notion.NewHeadingBlock(2, "Tomorrow's Plan"),
		notion.NewParagraphBlock(""),
	}
}

// projectBriefTemplate generates project brief blocks
func projectBriefTemplate() []notion.Block {
	return []notion.Block{
		notion.NewHeadingBlock(2, "Overview"),
		notion.NewParagraphBlock("Brief description of the project..."),
		notion.NewHeadingBlock(2, "Goals"),
		notion.NewBulletedListItemBlock("Goal 1"),
		notion.NewBulletedListItemBlock("Goal 2"),
		notion.NewHeadingBlock(2, "Timeline"),
		notion.NewParagraphBlock("Key milestones and deadlines..."),
		notion.NewHeadingBlock(2, "Stakeholders"),
		notion.NewBulletedListItemBlock("Stakeholder 1"),
		notion.NewBulletedListItemBlock("Stakeholder 2"),
		notion.NewHeadingBlock(2, "Resources"),
		notion.NewParagraphBlock("Required resources and budget..."),
		notion.NewHeadingBlock(2, "Risks"),
		notion.NewBulletedListItemBlock("Risk 1: Description and mitigation"),
		notion.NewBulletedListItemBlock("Risk 2: Description and mitigation"),
	}
}

// sprintPlanningTemplate generates sprint planning blocks
func sprintPlanningTemplate() []notion.Block {
	return []notion.Block{
		notion.NewHeadingBlock(2, "Sprint Goal"),
		notion.NewParagraphBlock("Clear, concise sprint goal..."),
		notion.NewHeadingBlock(2, "Capacity"),
		notion.NewParagraphBlock("Team capacity for this sprint..."),
		notion.NewHeadingBlock(2, "Planned Stories"),
		notion.NewToDoBlock("Story 1: Description", false),
		notion.NewToDoBlock("Story 2: Description", false),
		notion.NewToDoBlock("Story 3: Description", false),
		notion.NewHeadingBlock(2, "Dependencies"),
		notion.NewBulletedListItemBlock("Dependency 1"),
		notion.NewHeadingBlock(2, "Risks"),
		notion.NewBulletedListItemBlock("Risk 1"),
	}
}

// retrospectiveTemplate generates retrospective blocks
func retrospectiveTemplate() []notion.Block {
	return []notion.Block{
		notion.NewHeadingBlock(2, "What Went Well"),
		notion.NewBulletedListItemBlock("Success 1"),
		notion.NewBulletedListItemBlock("Success 2"),
		notion.NewHeadingBlock(2, "What Could Be Improved"),
		notion.NewBulletedListItemBlock("Improvement area 1"),
		notion.NewBulletedListItemBlock("Improvement area 2"),
		notion.NewHeadingBlock(2, "Action Items"),
		notion.NewToDoBlock("Action to improve 1", false),
		notion.NewToDoBlock("Action to improve 2", false),
	}
}

// CreatePageFromTemplateHandler creates a handler for creating pages from templates
func CreatePageFromTemplateHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		// Extract template name
		templateName, ok := args["template_name"].(string)
		if !ok || templateName == "" {
			return mcp.NewToolResultError("template_name is required"), nil
		}

		// Extract title
		title, ok := args["title"].(string)
		if !ok || title == "" {
			return mcp.NewToolResultError("title is required"), nil
		}

		// Extract variables for replacement
		variables := make(map[string]string)
		if varsData, ok := args["variables"].(map[string]any); ok {
			for k, v := range varsData {
				if vStr, ok := v.(string); ok {
					variables[k] = vStr
				}
			}
		}

		// Get template
		templates := GetTemplates()
		template, exists := templates[templateName]
		if !exists {
			return mcp.NewToolResultError(fmt.Sprintf("template '%s' not found. Available templates: %s",
				templateName, strings.Join(getTemplateNames(), ", "))), nil
		}

		// Replace variables in title
		processedTitle := replaceVariables(title, variables)

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
			properties["Name"] = notion.NewTitleProperty(processedTitle)
		} else {
			// For regular pages, use title property
			properties["title"] = notion.NewTitleProperty(processedTitle)
		}

		// Generate blocks from template
		blocks := template.Blocks()

		// Replace variables in blocks
		processedBlocks := make([]notion.Block, 0, len(blocks))
		for _, block := range blocks {
			processedBlock := replaceVariablesInBlock(block, variables)
			processedBlocks = append(processedBlocks, processedBlock)
		}

		// Create page request
		createReq := &notion.CreatePageRequest{
			Parent:     parent,
			Properties: properties,
			Children:   processedBlocks,
		}

		// Create page
		page, err := client.CreatePage(ctx, createReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create page from template: %v", err)), nil
		}

		// Format response
		response := map[string]any{
			"success":  true,
			"id":       page.ID,
			"url":      page.URL,
			"template": templateName,
			"message":  fmt.Sprintf("Page '%s' created successfully from template '%s'", processedTitle, templateName),
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// ListTemplatesHandler creates a handler for listing available templates
func ListTemplatesHandler(client *notion.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templates := GetTemplates()

		// Build response with template details
		templateList := make([]map[string]any, 0, len(templates))
		for name, template := range templates {
			blocks := template.Blocks()
			
			// Extract structure preview
			structure := make([]string, 0, len(blocks))
			for _, block := range blocks {
				structureItem := getBlockStructurePreview(block)
				if structureItem != "" {
					structure = append(structure, structureItem)
				}
			}

			templateList = append(templateList, map[string]any{
				"name":        name,
				"description": template.Description,
				"structure":   structure,
			})
		}

		// Format response
		response := map[string]any{
			"count":     len(templateList),
			"templates": templateList,
			"variables": map[string]string{
				"{date}":     "Current date (YYYY-MM-DD)",
				"{datetime}": "Current date and time",
				"{time}":     "Current time (HH:MM)",
				"custom":     "Any custom variables passed in the 'variables' parameter",
			},
		}

		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// replaceVariables replaces placeholders in text with actual values
func replaceVariables(text string, variables map[string]string) string {
	result := text

	// Replace built-in variables
	now := time.Now()
	result = strings.ReplaceAll(result, "{date}", now.Format("2006-01-02"))
	result = strings.ReplaceAll(result, "{datetime}", now.Format("2006-01-02 15:04:05"))
	result = strings.ReplaceAll(result, "{time}", now.Format("15:04"))

	// Replace custom variables
	for key, value := range variables {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// replaceVariablesInBlock replaces variables in block content
func replaceVariablesInBlock(block notion.Block, variables map[string]string) notion.Block {
	// Replace in different block types
	if block.Paragraph != nil && len(block.Paragraph.RichText) > 0 {
		for i := range block.Paragraph.RichText {
			if block.Paragraph.RichText[i].Text != nil {
				block.Paragraph.RichText[i].Text.Content = replaceVariables(block.Paragraph.RichText[i].Text.Content, variables)
			}
		}
	}

	if block.Heading1 != nil && len(block.Heading1.RichText) > 0 {
		for i := range block.Heading1.RichText {
			if block.Heading1.RichText[i].Text != nil {
				block.Heading1.RichText[i].Text.Content = replaceVariables(block.Heading1.RichText[i].Text.Content, variables)
			}
		}
	}

	if block.Heading2 != nil && len(block.Heading2.RichText) > 0 {
		for i := range block.Heading2.RichText {
			if block.Heading2.RichText[i].Text != nil {
				block.Heading2.RichText[i].Text.Content = replaceVariables(block.Heading2.RichText[i].Text.Content, variables)
			}
		}
	}

	if block.Heading3 != nil && len(block.Heading3.RichText) > 0 {
		for i := range block.Heading3.RichText {
			if block.Heading3.RichText[i].Text != nil {
				block.Heading3.RichText[i].Text.Content = replaceVariables(block.Heading3.RichText[i].Text.Content, variables)
			}
		}
	}

	if block.BulletedListItem != nil && len(block.BulletedListItem.RichText) > 0 {
		for i := range block.BulletedListItem.RichText {
			if block.BulletedListItem.RichText[i].Text != nil {
				block.BulletedListItem.RichText[i].Text.Content = replaceVariables(block.BulletedListItem.RichText[i].Text.Content, variables)
			}
		}
	}

	if block.NumberedListItem != nil && len(block.NumberedListItem.RichText) > 0 {
		for i := range block.NumberedListItem.RichText {
			if block.NumberedListItem.RichText[i].Text != nil {
				block.NumberedListItem.RichText[i].Text.Content = replaceVariables(block.NumberedListItem.RichText[i].Text.Content, variables)
			}
		}
	}

	if block.ToDo != nil && len(block.ToDo.RichText) > 0 {
		for i := range block.ToDo.RichText {
			if block.ToDo.RichText[i].Text != nil {
				block.ToDo.RichText[i].Text.Content = replaceVariables(block.ToDo.RichText[i].Text.Content, variables)
			}
		}
	}

	return block
}

// getBlockStructurePreview returns a preview string for a block
func getBlockStructurePreview(block notion.Block) string {
	switch block.Type {
	case "heading_1":
		if block.Heading1 != nil && len(block.Heading1.RichText) > 0 {
			return "# " + notion.ExtractPlainText(block.Heading1.RichText)
		}
	case "heading_2":
		if block.Heading2 != nil && len(block.Heading2.RichText) > 0 {
			return "## " + notion.ExtractPlainText(block.Heading2.RichText)
		}
	case "heading_3":
		if block.Heading3 != nil && len(block.Heading3.RichText) > 0 {
			return "### " + notion.ExtractPlainText(block.Heading3.RichText)
		}
	case "paragraph":
		if block.Paragraph != nil && len(block.Paragraph.RichText) > 0 {
			return "Paragraph: " + notion.ExtractPlainText(block.Paragraph.RichText)
		}
	case "bulleted_list_item":
		if block.BulletedListItem != nil && len(block.BulletedListItem.RichText) > 0 {
			return "• " + notion.ExtractPlainText(block.BulletedListItem.RichText)
		}
	case "numbered_list_item":
		if block.NumberedListItem != nil && len(block.NumberedListItem.RichText) > 0 {
			return "1. " + notion.ExtractPlainText(block.NumberedListItem.RichText)
		}
	case "to_do":
		if block.ToDo != nil && len(block.ToDo.RichText) > 0 {
			checkbox := "☐"
			if block.ToDo.Checked {
				checkbox = "☑"
			}
			return checkbox + " " + notion.ExtractPlainText(block.ToDo.RichText)
		}
	}
	return ""
}

// getTemplateNames returns a slice of all template names
func getTemplateNames() []string {
	templates := GetTemplates()
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return names
}
