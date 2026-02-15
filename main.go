package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rgabriel/mcp-notion/config"
	"github.com/rgabriel/mcp-notion/notion"
	"github.com/rgabriel/mcp-notion/tools"
)

// Set via -ldflags at build time.
var version = "dev"

// destructiveTools is the set of tool names that modify data and require audit logging.
var destructiveTools = map[string]bool{
	"delete_page":              true,
	"delete_block":             true,
	"update_page":              true,
	"update_block":             true,
	"move_page":                true,
	"create_page":              true,
	"create_database":          true,
	"update_database":          true,
	"append_blocks":            true,
	"create_comment":           true,
	"batch_create_pages":       true,
	"batch_update_pages":       true,
	"batch_delete_pages":       true,
	"create_simple_page":       true,
	"append_to_page":           true,
	"create_page_from_template": true,
}

func main() {
	// Load configuration first so we can redact secrets in logs.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Configure structured logging to stderr with secret redaction.
	baseHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(newRedactingHandler(baseHandler, cfg.NotionAPIToken))
	slog.SetDefault(logger)

	// Create Notion client
	notionClient := notion.NewClient(cfg)

	// Test API connection by getting bot user info
	ctx := context.Background()
	botUser, err := notionClient.GetBotUser(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to Notion API (check your token): %v", err)
	}

	// Create MCP server with middleware stack (outermost first).
	s := server.NewMCPServer(
		"Notion MCP Server",
		version,
		server.WithToolCapabilities(false),
		server.WithToolHandlerMiddleware(concurrencyMiddleware(10)),
		server.WithToolHandlerMiddleware(observabilityMiddleware()),
		server.WithRecovery(),
	)

	// Register Search & Discovery tools
	registerSearchTools(s, notionClient)

	// Register Page tools
	registerPageTools(s, notionClient)

	// Register Database tools
	registerDatabaseTools(s, notionClient)

	// Register Block tools
	registerBlockTools(s, notionClient)

	// Register Comment tools
	registerCommentTools(s, notionClient)

	// Register User tools
	registerUserTools(s, notionClient)

	// Register Helper tools
	registerHelperTools(s, notionClient)

	// Register Batch tools
	registerBatchTools(s, notionClient)

	// Register Template tools
	registerTemplateTools(s, notionClient)

	// Register Export tools
	registerExportTools(s, notionClient)

	// Register Smart Query tools
	registerSmartQueryTools(s, notionClient)

	// Log startup
	slog.Info("server starting",
		"name", "Notion MCP Server",
		"version", version,
		"bot_name", botUser.Name,
		"bot_id", botUser.ID,
		"api_version", cfg.NotionAPIVersion,
		"timeout_seconds", cfg.NotionTimeout,
	)

	// Start the stdio server with graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	stdioServer := server.NewStdioServer(s)
	if err := stdioServer.Listen(ctx, os.Stdin, os.Stdout); err != nil {
		if ctx.Err() == nil {
			log.Fatalf("Server error: %v", err)
		}
		slog.Info("server shutting down gracefully")
	}
}

func observabilityMiddleware() server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			requestID := uuid.New().String()[:8]
			toolName := req.Params.Name
			start := time.Now()

			attrs := []slog.Attr{
				slog.String("tool", toolName),
				slog.String("request_id", requestID),
			}

			// Audit logging for destructive operations
			if destructiveTools[toolName] {
				attrs = append(attrs, slog.Bool("audit", true))
				if argsJSON, err := json.Marshal(req.GetArguments()); err == nil {
					attrs = append(attrs, slog.String("arguments", string(argsJSON)))
				}
			}

			slog.LogAttrs(ctx, slog.LevelInfo, "tool call started", attrs...)

			result, err := next(ctx, req)

			duration := time.Since(start)
			if err != nil {
				slog.Info("tool call failed",
					"tool", toolName,
					"request_id", requestID,
					"duration_ms", duration.Milliseconds(),
					"error", err.Error(),
				)
			} else {
				slog.Info("tool call completed",
					"tool", toolName,
					"request_id", requestID,
					"duration_ms", duration.Milliseconds(),
				)
			}

			return result, err
		}
	}
}

func concurrencyMiddleware(maxConcurrent int) server.ToolHandlerMiddleware {
	sem := make(chan struct{}, maxConcurrent)
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
				return next(ctx, req)
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
}

func registerSearchTools(s *server.MCPServer, client notion.NotionClient) {
	// search - Search across all pages and databases
	searchTool := mcp.NewTool("search",
		mcp.WithDescription("Search across all pages and databases with optional filters and sorting"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("query",
			mcp.Description("Search term to query for"),
		),
		mcp.WithObject("filter",
			mcp.Description("Optional filter object (e.g., {\"property\": \"object\", \"value\": \"page\"})"),
		),
		mcp.WithObject("sort",
			mcp.Description("Optional sort object (e.g., {\"direction\": \"descending\", \"timestamp\": \"last_edited_time\"})"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Number of results per page (default: 100, max: 100)"),
			mcp.Min(1), mcp.Max(100), mcp.DefaultNumber(100),
		),
	)
	s.AddTool(searchTool, tools.SearchHandler(client))

	// list_databases - List all accessible databases
	listDatabasesTool := mcp.NewTool("list_databases",
		mcp.WithDescription("List all accessible databases with their IDs, titles, and metadata"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(listDatabasesTool, tools.ListDatabasesHandler(client))

	// get_database - Get database details including schema
	getDatabaseTool := mcp.NewTool("get_database",
		mcp.WithDescription("Get detailed information about a database including its properties schema"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("database_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Database ID (with or without dashes)"),
		),
	)
	s.AddTool(getDatabaseTool, tools.GetDatabaseHandler(client))
}

func registerPageTools(s *server.MCPServer, client notion.NotionClient) {
	// get_page - Get page metadata
	getPageTool := mcp.NewTool("get_page",
		mcp.WithDescription("Get page metadata including properties, parent, and timestamps"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("page_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Page ID (with or without dashes)"),
		),
	)
	s.AddTool(getPageTool, tools.GetPageHandler(client))

	// get_page_content - Get page content (blocks)
	getPageContentTool := mcp.NewTool("get_page_content",
		mcp.WithDescription("Get the content of a page as an array of blocks"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("page_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Page ID (with or without dashes)"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Number of blocks per page (default: 100, max: 100)"),
			mcp.Min(1), mcp.Max(100), mcp.DefaultNumber(100),
		),
	)
	s.AddTool(getPageContentTool, tools.GetPageContentHandler(client))

	// create_page - Create a new page
	createPageTool := mcp.NewTool("create_page",
		mcp.WithDescription("Create a new page in a database or as a child of another page"),
		mcp.WithObject("parent",
			mcp.Required(),
			mcp.Description("Parent object with either page_id or database_id"),
		),
		mcp.WithObject("properties",
			mcp.Required(),
			mcp.Description("Page properties (must include title for pages, or database-specific properties)"),
		),
		mcp.WithArray("children",
			mcp.Description("Optional array of block objects to include as page content"),
		),
		mcp.WithObject("icon",
			mcp.Description("Optional page icon (emoji or external URL)"),
		),
		mcp.WithObject("cover",
			mcp.Description("Optional page cover image (external URL)"),
		),
	)
	s.AddTool(createPageTool, tools.CreatePageHandler(client))

	// update_page - Update page properties
	updatePageTool := mcp.NewTool("update_page",
		mcp.WithDescription("Update page properties, icon, cover, or archive status"),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("page_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Page ID to update"),
		),
		mcp.WithObject("properties",
			mcp.Description("Properties to update"),
		),
		mcp.WithBoolean("archived",
			mcp.Description("Archive or unarchive the page"),
		),
		mcp.WithObject("icon",
			mcp.Description("Update page icon"),
		),
		mcp.WithObject("cover",
			mcp.Description("Update page cover"),
		),
	)
	s.AddTool(updatePageTool, tools.UpdatePageHandler(client))

	// move_page - Move page to new parent location
	movePageTool := mcp.NewTool("move_page",
		mcp.WithDescription("Move a page to a new parent location (another page, database, or workspace)"),
		mcp.WithString("page_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Page ID to move (with or without dashes)"),
		),
		mcp.WithObject("parent",
			mcp.Required(),
			mcp.Description("New parent location with 'type' and corresponding ID field"),
		),
	)
	s.AddTool(movePageTool, tools.MovePageHandler(client))

	// delete_page - Archive/delete a page
	deletePageTool := mcp.NewTool("delete_page",
		mcp.WithDescription("Archive (delete) a page by setting its archived property to true"),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithString("page_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Page ID to delete"),
		),
	)
	s.AddTool(deletePageTool, tools.DeletePageHandler(client))
}

func registerDatabaseTools(s *server.MCPServer, client notion.NotionClient) {
	// query_database - Query database with filters and sorts
	queryDatabaseTool := mcp.NewTool("query_database",
		mcp.WithDescription("Query a database with optional filters, sorts, and pagination"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("database_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Database ID to query"),
		),
		mcp.WithObject("filter",
			mcp.Description("Optional filter object for complex queries"),
		),
		mcp.WithArray("sorts",
			mcp.Description("Optional array of sort objects"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Number of results per page (default: 100, max: 100)"),
			mcp.Min(1), mcp.Max(100), mcp.DefaultNumber(100),
		),
		mcp.WithString("start_cursor",
			mcp.Description("Pagination cursor from previous response"),
		),
	)
	s.AddTool(queryDatabaseTool, tools.QueryDatabaseHandler(client))

	// create_database - Create a new database
	createDatabaseTool := mcp.NewTool("create_database",
		mcp.WithDescription("Create a new database as a child of a page"),
		mcp.WithObject("parent",
			mcp.Required(),
			mcp.Description("Parent object with page_id"),
		),
		mcp.WithArray("title",
			mcp.Required(),
			mcp.Description("Database title as array of rich text objects"),
		),
		mcp.WithObject("properties",
			mcp.Required(),
			mcp.Description("Database schema - property definitions"),
		),
		mcp.WithObject("icon",
			mcp.Description("Optional database icon"),
		),
		mcp.WithObject("cover",
			mcp.Description("Optional database cover"),
		),
	)
	s.AddTool(createDatabaseTool, tools.CreateDatabaseHandler(client))

	// update_database - Update database properties
	updateDatabaseTool := mcp.NewTool("update_database",
		mcp.WithDescription("Update database title, description, or add/modify properties"),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("database_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Database ID to update"),
		),
		mcp.WithArray("title",
			mcp.Description("New database title"),
		),
		mcp.WithArray("description",
			mcp.Description("New database description"),
		),
		mcp.WithObject("properties",
			mcp.Description("Properties to add or modify in schema"),
		),
		mcp.WithBoolean("archived",
			mcp.Description("Archive or unarchive the database"),
		),
	)
	s.AddTool(updateDatabaseTool, tools.UpdateDatabaseHandler(client))
}

func registerBlockTools(s *server.MCPServer, client notion.NotionClient) {
	// get_block - Get a specific block
	getBlockTool := mcp.NewTool("get_block",
		mcp.WithDescription("Get details of a specific block by ID"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("block_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Block ID"),
		),
	)
	s.AddTool(getBlockTool, tools.GetBlockHandler(client))

	// get_block_children - Get children of a block
	getBlockChildrenTool := mcp.NewTool("get_block_children",
		mcp.WithDescription("Get child blocks of a parent block or page"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("block_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Parent block or page ID"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Number of blocks per page (default: 100, max: 100)"),
			mcp.Min(1), mcp.Max(100), mcp.DefaultNumber(100),
		),
		mcp.WithString("start_cursor",
			mcp.Description("Pagination cursor from previous response"),
		),
	)
	s.AddTool(getBlockChildrenTool, tools.GetBlockChildrenHandler(client))

	// append_blocks - Append blocks to a page or block
	appendBlocksTool := mcp.NewTool("append_blocks",
		mcp.WithDescription("Append child blocks to a page or block"),
		mcp.WithString("block_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Parent block or page ID to append to"),
		),
		mcp.WithArray("children",
			mcp.Required(),
			mcp.Description("Array of block objects to append"),
		),
	)
	s.AddTool(appendBlocksTool, tools.AppendBlocksHandler(client))

	// update_block - Update a block
	updateBlockTool := mcp.NewTool("update_block",
		mcp.WithDescription("Update the content of a block"),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("block_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Block ID to update"),
		),
	)
	s.AddTool(updateBlockTool, tools.UpdateBlockHandler(client))

	// delete_block - Delete a block
	deleteBlockTool := mcp.NewTool("delete_block",
		mcp.WithDescription("Delete a block and its children"),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithString("block_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Block ID to delete"),
		),
	)
	s.AddTool(deleteBlockTool, tools.DeleteBlockHandler(client))
}

func registerCommentTools(s *server.MCPServer, client notion.NotionClient) {
	// get_comments - Get comments on a page or block
	getCommentsTool := mcp.NewTool("get_comments",
		mcp.WithDescription("Get comments on a page or block"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("block_id",
			mcp.Description("Block ID to get comments for"),
		),
		mcp.WithString("page_id",
			mcp.Description("Page ID to get comments for (provide either block_id or page_id)"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Number of comments per page (default: 100, max: 100)"),
			mcp.Min(1), mcp.Max(100), mcp.DefaultNumber(100),
		),
		mcp.WithString("start_cursor",
			mcp.Description("Pagination cursor from previous response"),
		),
	)
	s.AddTool(getCommentsTool, tools.GetCommentsHandler(client))

	// create_comment - Create a comment
	createCommentTool := mcp.NewTool("create_comment",
		mcp.WithDescription("Create a comment on a page or reply to a discussion"),
		mcp.WithObject("parent",
			mcp.Required(),
			mcp.Description("Parent object with page_id"),
		),
		mcp.WithArray("rich_text",
			mcp.Required(),
			mcp.Description("Comment content as array of rich text objects"),
		),
		mcp.WithString("discussion_id",
			mcp.Description("Optional discussion ID to reply to existing comment thread"),
		),
	)
	s.AddTool(createCommentTool, tools.CreateCommentHandler(client))
}

func registerUserTools(s *server.MCPServer, client notion.NotionClient) {
	// list_users - List all users
	listUsersTool := mcp.NewTool("list_users",
		mcp.WithDescription("List all users in the workspace"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithNumber("page_size",
			mcp.Description("Number of users per page (default: 100, max: 100)"),
			mcp.Min(1), mcp.Max(100), mcp.DefaultNumber(100),
		),
		mcp.WithString("start_cursor",
			mcp.Description("Pagination cursor from previous response"),
		),
	)
	s.AddTool(listUsersTool, tools.ListUsersHandler(client))

	// get_user - Get a specific user
	getUserTool := mcp.NewTool("get_user",
		mcp.WithDescription("Get details of a specific user by ID"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("user_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("User ID"),
		),
	)
	s.AddTool(getUserTool, tools.GetUserHandler(client))

	// get_bot_user - Get current bot user info
	getBotUserTool := mcp.NewTool("get_bot_user",
		mcp.WithDescription("Get information about the current bot/integration user"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(getBotUserTool, tools.GetBotUserHandler(client))
}

func registerHelperTools(s *server.MCPServer, client notion.NotionClient) {
	// create_simple_page - Simplified page creation
	createSimplePageTool := mcp.NewTool("create_simple_page",
		mcp.WithDescription("Create a simple page with markdown-like content (easier than create_page)"),
		mcp.WithString("parent_page_id",
			mcp.Description("Optional parent page ID (uses workspace root if not provided)"),
		),
		mcp.WithString("parent_database_id",
			mcp.Description("Optional parent database ID"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Page title"),
		),
		mcp.WithString("content",
			mcp.Description("Page content in markdown-like format (supports headings, lists, code blocks, etc.)"),
		),
	)
	s.AddTool(createSimplePageTool, tools.CreateSimplePageHandler(client))

	// append_to_page - Simplified content appending
	appendToPageTool := mcp.NewTool("append_to_page",
		mcp.WithDescription("Append markdown-like content to an existing page (easier than append_blocks)"),
		mcp.WithString("page_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Page ID to append content to"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Content to append in markdown-like format"),
		),
	)
	s.AddTool(appendToPageTool, tools.AppendToPageHandler(client))

	// search_in_database - Database-specific search
	searchInDatabaseTool := mcp.NewTool("search_in_database",
		mcp.WithDescription("Search for entries within a specific database with simple text query"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("database_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Database ID to search in"),
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Search query text"),
		),
		mcp.WithObject("property_filters",
			mcp.Description("Optional simplified property filters (e.g., {\"Status\": \"Done\", \"Priority\": \"High\"})"),
		),
	)
	s.AddTool(searchInDatabaseTool, tools.SearchInDatabaseHandler(client))
}

func registerBatchTools(s *server.MCPServer, client notion.NotionClient) {
	// batch_create_pages - Create multiple pages
	batchCreatePagesTool := mcp.NewTool("batch_create_pages",
		mcp.WithDescription("Create multiple pages in a single batch operation with rate limiting"),
		mcp.WithArray("pages",
			mcp.Required(),
			mcp.Description("Array of page objects to create, each with parent, properties, children, icon, cover"),
		),
		mcp.WithBoolean("continue_on_error",
			mcp.Description("If true, continue creating remaining pages even if one fails (default: false)"),
		),
	)
	s.AddTool(batchCreatePagesTool, tools.BatchCreatePagesHandler(client))

	// batch_update_pages - Update multiple pages
	batchUpdatePagesTool := mcp.NewTool("batch_update_pages",
		mcp.WithDescription("Update multiple pages with the same property changes in a batch operation"),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithArray("page_ids",
			mcp.Required(),
			mcp.Description("Array of page IDs to update"),
		),
		mcp.WithObject("properties",
			mcp.Description("Properties to update on all pages"),
		),
		mcp.WithBoolean("archived",
			mcp.Description("Archive or unarchive all pages"),
		),
		mcp.WithObject("icon",
			mcp.Description("Update icon on all pages"),
		),
		mcp.WithObject("cover",
			mcp.Description("Update cover on all pages"),
		),
		mcp.WithBoolean("continue_on_error",
			mcp.Description("If true, continue updating remaining pages even if one fails (default: false)"),
		),
	)
	s.AddTool(batchUpdatePagesTool, tools.BatchUpdatePagesHandler(client))

	// batch_delete_pages - Delete multiple pages
	batchDeletePagesTool := mcp.NewTool("batch_delete_pages",
		mcp.WithDescription("Archive (delete) multiple pages in a batch operation"),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithArray("page_ids",
			mcp.Required(),
			mcp.Description("Array of page IDs to archive/delete"),
		),
		mcp.WithBoolean("continue_on_error",
			mcp.Description("If true, continue deleting remaining pages even if one fails (default: false)"),
		),
	)
	s.AddTool(batchDeletePagesTool, tools.BatchDeletePagesHandler(client))
}

func registerTemplateTools(s *server.MCPServer, client notion.NotionClient) {
	// create_page_from_template - Create page from template
	createPageFromTemplateTool := mcp.NewTool("create_page_from_template",
		mcp.WithDescription("Create a new page using a predefined template with variable substitution"),
		mcp.WithString("template_name",
			mcp.Required(),
			mcp.Enum("meeting_notes", "daily_log", "project_brief", "sprint_planning", "retrospective"),
			mcp.Description("Template name"),
		),
		mcp.WithString("parent_page_id",
			mcp.Description("Parent page ID (uses workspace root if not provided)"),
		),
		mcp.WithString("parent_database_id",
			mcp.Description("Parent database ID"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Page title (supports variables: {date}, {datetime}, {time}, and custom variables)"),
		),
		mcp.WithObject("variables",
			mcp.Description("Custom variables for template substitution (e.g., {project_name: 'Alpha'})"),
		),
	)
	s.AddTool(createPageFromTemplateTool, tools.CreatePageFromTemplateHandler(client))

	// list_templates - List available templates
	listTemplatesTool := mcp.NewTool("list_templates",
		mcp.WithDescription("List all available page templates with their structure and supported variables"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(listTemplatesTool, tools.ListTemplatesHandler(client))
}

func registerExportTools(s *server.MCPServer, client notion.NotionClient) {
	// export_page_as_markdown - Export page as Markdown
	exportPageAsMarkdownTool := mcp.NewTool("export_page_as_markdown",
		mcp.WithDescription("Export a Notion page to clean Markdown format with optional frontmatter"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("page_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Page ID to export"),
		),
		mcp.WithBoolean("include_children",
			mcp.Description("Recursively export child blocks (default: false)"),
		),
		mcp.WithBoolean("frontmatter",
			mcp.Description("Include YAML frontmatter with metadata (default: false)"),
		),
	)
	s.AddTool(exportPageAsMarkdownTool, tools.ExportPageAsMarkdownHandler(client))

	// export_database_as_csv - Export database as CSV
	exportDatabaseAsCSVTool := mcp.NewTool("export_database_as_csv",
		mcp.WithDescription("Export database entries as CSV with all properties"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("database_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Database ID to export"),
		),
		mcp.WithObject("filter",
			mcp.Description("Optional filter to limit entries exported"),
		),
		mcp.WithBoolean("include_archived",
			mcp.Description("Include archived entries (default: false)"),
		),
	)
	s.AddTool(exportDatabaseAsCSVTool, tools.ExportDatabaseAsCSVHandler(client))
}

func registerSmartQueryTools(s *server.MCPServer, client notion.NotionClient) {
	// get_recently_edited - Get recently edited pages
	getRecentlyEditedTool := mcp.NewTool("get_recently_edited",
		mcp.WithDescription("Get pages edited within a specified time window"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithNumber("days",
			mcp.Description("Number of days to look back (default: 7)"),
			mcp.Min(1), mcp.DefaultNumber(7),
		),
		mcp.WithString("database_id",
			mcp.Description("Optional: limit to specific database"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results (default: 50)"),
			mcp.Min(1), mcp.DefaultNumber(50),
		),
	)
	s.AddTool(getRecentlyEditedTool, tools.GetRecentlyEditedHandler(client))

	// get_my_tasks - Smart task finder
	getMyTasksTool := mcp.NewTool("get_my_tasks",
		mcp.WithDescription("Smart task finder for the current user with flexible property detection"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("database_id",
			mcp.Description("Specific tasks database (searches for task databases if not provided)"),
		),
		mcp.WithArray("statuses",
			mcp.Description("Status values to include (default: ['In Progress', 'To Do', 'Not Started'])"),
		),
		mcp.WithBoolean("include_overdue",
			mcp.Description("Include overdue tasks (default: true)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of tasks to return (default: 50)"),
			mcp.Min(1), mcp.DefaultNumber(50),
		),
	)
	s.AddTool(getMyTasksTool, tools.GetMyTasksHandler(client))

	// get_related_pages - Follow page relations
	getRelatedPagesTool := mcp.NewTool("get_related_pages",
		mcp.WithDescription("Follow page relations and return connected pages"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("page_id",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Page ID to get relations for"),
		),
		mcp.WithString("relation_property",
			mcp.Description("Specific relation property name (returns all relations if not specified)"),
		),
		mcp.WithBoolean("include_properties",
			mcp.Description("Include full page properties for related pages (default: false)"),
		),
	)
	s.AddTool(getRelatedPagesTool, tools.GetRelatedPagesHandler(client))
}
