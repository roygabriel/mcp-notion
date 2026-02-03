package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rgabriel/mcp-notion/config"
	"github.com/rgabriel/mcp-notion/notion"
	"github.com/rgabriel/mcp-notion/tools"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create Notion client
	notionClient := notion.NewClient(cfg)

	// Test API connection by getting bot user info
	ctx := context.Background()
	botUser, err := notionClient.GetBotUser(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to Notion API (check your token): %v", err)
	}

	// Create MCP server
	s := server.NewMCPServer(
		"Notion MCP Server",
		"1.0.0",
		server.WithToolCapabilities(false),
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

	// Log startup
	fmt.Fprintf(os.Stderr, "Notion MCP Server v1.0.0 starting...\n")
	fmt.Fprintf(os.Stderr, "Connected to Notion as: %s\n", botUser.Name)
	fmt.Fprintf(os.Stderr, "Bot ID: %s\n", botUser.ID)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func registerSearchTools(s *server.MCPServer, client *notion.Client) {
	// search - Search across all pages and databases
	searchTool := mcp.NewTool("search",
		mcp.WithDescription("Search across all pages and databases with optional filters and sorting"),
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
		),
	)
	s.AddTool(searchTool, tools.SearchHandler(client))

	// list_databases - List all accessible databases
	listDatabasesTool := mcp.NewTool("list_databases",
		mcp.WithDescription("List all accessible databases with their IDs, titles, and metadata"),
	)
	s.AddTool(listDatabasesTool, tools.ListDatabasesHandler(client))

	// get_database - Get database details including schema
	getDatabaseTool := mcp.NewTool("get_database",
		mcp.WithDescription("Get detailed information about a database including its properties schema"),
		mcp.WithString("database_id",
			mcp.Required(),
			mcp.Description("Database ID (with or without dashes)"),
		),
	)
	s.AddTool(getDatabaseTool, tools.GetDatabaseHandler(client))
}

func registerPageTools(s *server.MCPServer, client *notion.Client) {
	// get_page - Get page metadata
	getPageTool := mcp.NewTool("get_page",
		mcp.WithDescription("Get page metadata including properties, parent, and timestamps"),
		mcp.WithString("page_id",
			mcp.Required(),
			mcp.Description("Page ID (with or without dashes)"),
		),
	)
	s.AddTool(getPageTool, tools.GetPageHandler(client))

	// get_page_content - Get page content (blocks)
	getPageContentTool := mcp.NewTool("get_page_content",
		mcp.WithDescription("Get the content of a page as an array of blocks"),
		mcp.WithString("page_id",
			mcp.Required(),
			mcp.Description("Page ID (with or without dashes)"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Number of blocks per page (default: 100, max: 100)"),
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
		mcp.WithString("page_id",
			mcp.Required(),
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

	// delete_page - Archive/delete a page
	deletePageTool := mcp.NewTool("delete_page",
		mcp.WithDescription("Archive (delete) a page by setting its archived property to true"),
		mcp.WithString("page_id",
			mcp.Required(),
			mcp.Description("Page ID to delete"),
		),
	)
	s.AddTool(deletePageTool, tools.DeletePageHandler(client))
}

func registerDatabaseTools(s *server.MCPServer, client *notion.Client) {
	// query_database - Query database with filters and sorts
	queryDatabaseTool := mcp.NewTool("query_database",
		mcp.WithDescription("Query a database with optional filters, sorts, and pagination"),
		mcp.WithString("database_id",
			mcp.Required(),
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
		mcp.WithString("database_id",
			mcp.Required(),
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

func registerBlockTools(s *server.MCPServer, client *notion.Client) {
	// get_block - Get a specific block
	getBlockTool := mcp.NewTool("get_block",
		mcp.WithDescription("Get details of a specific block by ID"),
		mcp.WithString("block_id",
			mcp.Required(),
			mcp.Description("Block ID"),
		),
	)
	s.AddTool(getBlockTool, tools.GetBlockHandler(client))

	// get_block_children - Get children of a block
	getBlockChildrenTool := mcp.NewTool("get_block_children",
		mcp.WithDescription("Get child blocks of a parent block or page"),
		mcp.WithString("block_id",
			mcp.Required(),
			mcp.Description("Parent block or page ID"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Number of blocks per page (default: 100, max: 100)"),
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
		mcp.WithString("block_id",
			mcp.Required(),
			mcp.Description("Block ID to update"),
		),
	)
	s.AddTool(updateBlockTool, tools.UpdateBlockHandler(client))

	// delete_block - Delete a block
	deleteBlockTool := mcp.NewTool("delete_block",
		mcp.WithDescription("Delete a block and its children"),
		mcp.WithString("block_id",
			mcp.Required(),
			mcp.Description("Block ID to delete"),
		),
	)
	s.AddTool(deleteBlockTool, tools.DeleteBlockHandler(client))
}

func registerCommentTools(s *server.MCPServer, client *notion.Client) {
	// get_comments - Get comments on a page or block
	getCommentsTool := mcp.NewTool("get_comments",
		mcp.WithDescription("Get comments on a page or block"),
		mcp.WithString("block_id",
			mcp.Description("Block ID to get comments for"),
		),
		mcp.WithString("page_id",
			mcp.Description("Page ID to get comments for (provide either block_id or page_id)"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Number of comments per page (default: 100, max: 100)"),
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

func registerUserTools(s *server.MCPServer, client *notion.Client) {
	// list_users - List all users
	listUsersTool := mcp.NewTool("list_users",
		mcp.WithDescription("List all users in the workspace"),
		mcp.WithNumber("page_size",
			mcp.Description("Number of users per page (default: 100, max: 100)"),
		),
		mcp.WithString("start_cursor",
			mcp.Description("Pagination cursor from previous response"),
		),
	)
	s.AddTool(listUsersTool, tools.ListUsersHandler(client))

	// get_user - Get a specific user
	getUserTool := mcp.NewTool("get_user",
		mcp.WithDescription("Get details of a specific user by ID"),
		mcp.WithString("user_id",
			mcp.Required(),
			mcp.Description("User ID"),
		),
	)
	s.AddTool(getUserTool, tools.GetUserHandler(client))

	// get_bot_user - Get current bot user info
	getBotUserTool := mcp.NewTool("get_bot_user",
		mcp.WithDescription("Get information about the current bot/integration user"),
	)
	s.AddTool(getBotUserTool, tools.GetBotUserHandler(client))
}

func registerHelperTools(s *server.MCPServer, client *notion.Client) {
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
			mcp.Description("Page ID to append content to"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Content to append in markdown-like format"),
		),
	)
	s.AddTool(appendToPageTool, tools.AppendToPageHandler(client))

	// search_in_database - Database-specific search
	searchInDatabaseTool := mcp.NewTool("search_in_database",
		mcp.WithDescription("Search for entries within a specific database with simple text query"),
		mcp.WithString("database_id",
			mcp.Required(),
			mcp.Description("Database ID to search in"),
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query text"),
		),
		mcp.WithObject("property_filters",
			mcp.Description("Optional simplified property filters (e.g., {\"Status\": \"Done\", \"Priority\": \"High\"})"),
		),
	)
	s.AddTool(searchInDatabaseTool, tools.SearchInDatabaseHandler(client))
}
