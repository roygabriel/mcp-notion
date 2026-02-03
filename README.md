# Notion MCP Server

A Model Context Protocol (MCP) server that connects to Notion using their official REST API. This server enables AI assistants like Claude to interact with your Notion workspace - search, read, create, and manage pages, databases, and content.

Built with Go using the official [mcp-go SDK](https://mcp-go.dev) and works on all operating systems (Linux, Windows, macOS).

## Features

- **Search & Discovery** - Search across pages and databases, list databases
- **Page Management** - Get, create, update, and delete pages with full property support
- **Database Operations** - Query databases with complex filters, create and update database schemas
- **Content Management** - Work with blocks (paragraphs, headings, lists, code, etc.)
- **Comments** - Read and create comments on pages and blocks
- **User Management** - List and get user information
- **Helper Tools** - Simplified tools for common tasks with markdown-like syntax
- **Cross-Platform** - Works on Linux, Windows, and macOS
- **Secure** - Uses Notion Integration tokens with proper rate limiting

## Prerequisites

- **Go 1.21 or higher** - [Install Go](https://go.dev/doc/install)
- **Notion Account** with access to create integrations
- **Notion Integration Token** - Required for API access (see setup below)

## Notion Integration Setup

Notion requires you to create an integration to access their API. Follow these steps:

### 1. Create an Integration

1. Go to [Notion Integrations](https://www.notion.so/my-integrations)
2. Click **"New integration"**
3. Fill in the details:
   - **Name**: "MCP Notion Server" (or your preferred name)
   - **Associated workspace**: Select your workspace
   - **Type**: Internal Integration
4. Click **"Submit"**
5. Copy the **Internal Integration Token** (starts with `secret_`)
6. Save this token securely - you'll need it for configuration

### 2. Grant Integration Permissions

The integration needs the following capabilities (set during creation):

- ✅ Read content
- ✅ Update content  
- ✅ Insert content
- ✅ Read comments
- ✅ Insert comments
- ✅ Read user information (including email)

### 3. Share Pages/Databases with Integration

**CRITICAL**: The integration can only access pages and databases that are explicitly shared with it.

For each page or database you want the integration to access:

1. Open the page/database in Notion
2. Click the **"..."** menu in the top right
3. Select **"Add connections"**
4. Find and select your integration
5. Click **"Confirm"**

Without this step, you'll get "object not found" errors even with a valid token!

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/rgabriel/mcp-notion.git
cd mcp-notion

# Build the server
go build -o mcp-notion

# Optional: Install to your PATH
go install
```

### Using go install

```bash
go install github.com/rgabriel/mcp-notion@latest
```

## Configuration

Create a `.env` file in the same directory as the server executable (for local testing):

```bash
cp .env.example .env
```

Edit `.env` and add your integration token:

```bash
NOTION_API_TOKEN=secret_your_integration_token_here
NOTION_API_VERSION=2022-06-28
NOTION_TIMEOUT=30
```

**Environment Variables:**

- `NOTION_API_TOKEN` (required) - Your Notion Integration Token from notion.so/my-integrations
- `NOTION_API_VERSION` (optional) - API version, default: "2022-06-28"
- `NOTION_TIMEOUT` (optional) - Request timeout in seconds, default: 30

## Usage with Claude Desktop

Add this server to your Claude Desktop configuration file:

### macOS

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "notion": {
      "command": "/path/to/mcp-notion",
      "env": {
        "NOTION_API_TOKEN": "secret_your_integration_token_here"
      }
    }
  }
}
```

### Windows

Edit `%APPDATA%\Claude\claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "notion": {
      "command": "C:\\path\\to\\mcp-notion.exe",
      "env": {
        "NOTION_API_TOKEN": "secret_your_integration_token_here"
      }
    }
  }
}
```

### Linux

Edit `~/.config/claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "notion": {
      "command": "/path/to/mcp-notion",
      "env": {
        "NOTION_API_TOKEN": "secret_your_integration_token_here"
      }
    }
  }
}
```

After adding the configuration, restart Claude Desktop.

## Available Tools

### Search & Discovery

#### 1. search

Search across all pages and databases with optional filters and sorting.

**Parameters:**
- `query` (optional) - Search term
- `filter` (optional) - Filter object (e.g., `{"property": "object", "value": "page"}`)
- `sort` (optional) - Sort object (e.g., `{"direction": "descending", "timestamp": "last_edited_time"}`)
- `page_size` (optional) - Results per page (default: 100, max: 100)

**Example:**
```json
{
  "query": "project planning",
  "filter": {
    "property": "object",
    "value": "page"
  }
}
```

#### 2. list_databases

List all accessible databases with their IDs, titles, and metadata.

**Parameters:** None

**Example Response:**
```json
{
  "count": 3,
  "databases": [
    {
      "id": "abc123...",
      "title": "Tasks Database",
      "url": "https://notion.so/...",
      "icon": {"type": "emoji", "emoji": "✅"}
    }
  ]
}
```

#### 3. get_database

Get detailed information about a database including its properties schema.

**Parameters:**
- `database_id` (required) - Database ID (with or without dashes)

### Page Operations

#### 4. get_page

Get page metadata including properties, parent, and timestamps.

**Parameters:**
- `page_id` (required) - Page ID

#### 5. get_page_content

Get the content of a page as an array of blocks.

**Parameters:**
- `page_id` (required) - Page ID
- `page_size` (optional) - Blocks per page (default: 100)

#### 6. create_page

Create a new page in a database or as a child of another page.

**Parameters:**
- `parent` (required) - Parent object with `page_id` or `database_id`
- `properties` (required) - Page properties (must include title)
- `children` (optional) - Array of block objects for content
- `icon` (optional) - Page icon (emoji or external URL)
- `cover` (optional) - Page cover image

**Example:**
```json
{
  "parent": {"page_id": "abc123..."},
  "properties": {
    "title": [{
      "text": {"content": "My New Page"}
    }]
  },
  "children": [
    {
      "object": "block",
      "type": "paragraph",
      "paragraph": {
        "rich_text": [{
          "text": {"content": "This is the first paragraph."}
        }]
      }
    }
  ]
}
```

#### 7. update_page

Update page properties, icon, cover, or archive status.

**Parameters:**
- `page_id` (required) - Page ID
- `properties` (optional) - Properties to update
- `archived` (optional) - Archive/unarchive the page
- `icon` (optional) - Update page icon
- `cover` (optional) - Update page cover

#### 8. delete_page

Archive (delete) a page.

**Parameters:**
- `page_id` (required) - Page ID to delete

### Database Operations

#### 9. query_database

Query a database with optional filters, sorts, and pagination.

**Parameters:**
- `database_id` (required) - Database ID
- `filter` (optional) - Filter object for complex queries
- `sorts` (optional) - Array of sort objects
- `page_size` (optional) - Results per page
- `start_cursor` (optional) - Pagination cursor

**Filter Example:**
```json
{
  "database_id": "abc123...",
  "filter": {
    "and": [
      {
        "property": "Status",
        "select": {"equals": "In Progress"}
      },
      {
        "property": "Priority",
        "select": {"equals": "High"}
      }
    ]
  },
  "sorts": [
    {
      "property": "Due Date",
      "direction": "ascending"
    }
  ]
}
```

#### 10. create_database

Create a new database as a child of a page.

**Parameters:**
- `parent` (required) - Parent object with `page_id`
- `title` (required) - Database title as array of rich text
- `properties` (required) - Database schema (property definitions)
- `icon` (optional) - Database icon
- `cover` (optional) - Database cover

**Example:**
```json
{
  "parent": {"page_id": "abc123..."},
  "title": [{"text": {"content": "My Database"}}],
  "properties": {
    "Name": {"title": {}},
    "Status": {
      "select": {
        "options": [
          {"name": "Not Started"},
          {"name": "In Progress"},
          {"name": "Done"}
        ]
      }
    },
    "Due Date": {"date": {}}
  }
}
```

#### 11. update_database

Update database title, description, or add/modify properties.

**Parameters:**
- `database_id` (required) - Database ID
- `title` (optional) - New title
- `description` (optional) - New description
- `properties` (optional) - Properties to add/modify
- `archived` (optional) - Archive status

### Block Operations

#### 12. get_block

Get details of a specific block by ID.

**Parameters:**
- `block_id` (required) - Block ID

#### 13. get_block_children

Get child blocks of a parent block or page.

**Parameters:**
- `block_id` (required) - Parent block or page ID
- `page_size` (optional) - Blocks per page
- `start_cursor` (optional) - Pagination cursor

#### 14. append_blocks

Append child blocks to a page or block.

**Parameters:**
- `block_id` (required) - Parent block or page ID
- `children` (required) - Array of block objects

**Example:**
```json
{
  "block_id": "abc123...",
  "children": [
    {
      "object": "block",
      "type": "heading_2",
      "heading_2": {
        "rich_text": [{"text": {"content": "New Section"}}]
      }
    },
    {
      "object": "block",
      "type": "paragraph",
      "paragraph": {
        "rich_text": [{"text": {"content": "Section content here."}}]
      }
    }
  ]
}
```

#### 15. update_block

Update the content of a block.

**Parameters:**
- `block_id` (required) - Block ID
- Additional parameters depend on block type

#### 16. delete_block

Delete a block and its children.

**Parameters:**
- `block_id` (required) - Block ID

### Comment Operations

#### 17. get_comments

Get comments on a page or block.

**Parameters:**
- `block_id` (optional) - Block ID
- `page_id` (optional) - Page ID (provide either block_id or page_id)
- `page_size` (optional) - Comments per page
- `start_cursor` (optional) - Pagination cursor

#### 18. create_comment

Create a comment on a page or reply to a discussion.

**Parameters:**
- `parent` (required) - Parent object with `page_id`
- `rich_text` (required) - Comment content as array of rich text
- `discussion_id` (optional) - Reply to existing discussion

### User Operations

#### 19. list_users

List all users in the workspace.

**Parameters:**
- `page_size` (optional) - Users per page
- `start_cursor` (optional) - Pagination cursor

#### 20. get_user

Get details of a specific user by ID.

**Parameters:**
- `user_id` (required) - User ID

#### 21. get_bot_user

Get information about the current bot/integration user.

**Parameters:** None

### Helper Tools (Simplified)

#### 22. create_simple_page

Create a simple page with markdown-like content (easier than `create_page`).

**Parameters:**
- `parent_page_id` (optional) - Parent page ID (uses workspace root if not provided)
- `parent_database_id` (optional) - Parent database ID
- `title` (required) - Page title
- `content` (optional) - Markdown-like content

**Example:**
```json
{
  "title": "Meeting Notes",
  "content": "# Discussion Points\n\n- Project timeline\n- Budget review\n- Next steps\n\n## Action Items\n\n- [ ] Review proposal\n- [ ] Schedule follow-up"
}
```

**Supported Markdown Syntax:**
- `# Heading 1`, `## Heading 2`, `### Heading 3`
- `- Bullet list` or `* Bullet list`
- `1. Numbered list`
- `- [ ] To-do item` or `- [x] Completed to-do`
- `` ```language `` for code blocks
- `> Quote`
- `---` for divider

#### 23. append_to_page

Append markdown-like content to an existing page.

**Parameters:**
- `page_id` (required) - Page ID
- `content` (required) - Markdown-like content

#### 24. search_in_database

Search for entries within a specific database with simple text query.

**Parameters:**
- `database_id` (required) - Database ID
- `query` (required) - Search query text
- `property_filters` (optional) - Simplified filters (e.g., `{"Status": "Done"}`)

## Advanced Tools

### Batch Operations

Perform operations on multiple pages at once with automatic rate limiting.

#### 25. batch_create_pages

Create multiple pages in a single batch operation.

**Parameters:**
- `pages` (required) - Array of page objects, each containing:
  - `parent` - Parent object with `database_id` or `page_id`
  - `properties` - Page properties
  - `children` (optional) - Block content
  - `icon` (optional) - Page icon
- `continue_on_error` (optional, boolean) - Continue on failures (default: false)

**Returns:**
```json
{
  "created": [
    {"id": "page-id-1", "url": "https://notion.so/..."},
    {"id": "page-id-2", "url": "https://notion.so/..."}
  ],
  "failed": [
    {"index": 2, "error": "invalid properties"}
  ],
  "summary": {
    "total": 3,
    "succeeded": 2,
    "failed": 1
  }
}
```

**Rate Limiting:** Automatically adds 350ms delay between operations.

#### 26. batch_update_pages

Update multiple pages with the same property changes.

**Parameters:**
- `page_ids` (required) - Array of page IDs to update
- `properties` (optional) - Properties to update on all pages
- `archived` (optional) - Archive/unarchive status
- `icon` (optional) - Update icon on all pages
- `cover` (optional) - Update cover on all pages
- `continue_on_error` (optional, boolean) - Continue on failures

**Example:**
```json
{
  "page_ids": ["id1", "id2", "id3"],
  "properties": {
    "Status": {
      "select": {"name": "Complete"}
    }
  },
  "continue_on_error": true
}
```

#### 27. batch_delete_pages

Archive (delete) multiple pages at once.

**Parameters:**
- `page_ids` (required) - Array of page IDs to archive
- `continue_on_error` (optional, boolean) - Continue on failures

### Template System

Create pages from predefined templates with variable substitution.

#### 28. create_page_from_template

Create a new page using a built-in template.

**Parameters:**
- `template_name` (required) - Template to use (see available templates below)
- `parent_page_id` or `parent_database_id` (required) - Where to create the page
- `title` (required) - Page title (supports variable placeholders)
- `variables` (optional) - Custom variable values for substitution

**Available Templates:**

1. **meeting_notes** - Meeting notes with attendees, agenda, discussion, decisions, and action items
2. **daily_log** - Daily log with morning priorities, accomplishments, blockers, and tomorrow's plan
3. **project_brief** - Project brief with overview, goals, timeline, stakeholders, resources, and risks
4. **sprint_planning** - Sprint planning with goal, capacity, stories, dependencies, and risks
5. **retrospective** - Retrospective with what went well, improvements, and action items

**Supported Variables:**
- `{date}` - Current date (YYYY-MM-DD)
- `{datetime}` - Current date and time
- `{time}` - Current time (HH:MM)
- Custom variables from the `variables` parameter

**Example:**
```json
{
  "template_name": "meeting_notes",
  "parent_page_id": "abc123...",
  "title": "Team Sync - {date}",
  "variables": {
    "team_name": "Engineering"
  }
}
```

#### 29. list_templates

List all available templates with their structure.

**Parameters:** None

**Returns:** Array of templates with names, descriptions, and structure previews.

### Export & Backup

Export Notion content to portable formats.

#### 30. export_page_as_markdown

Export a Notion page to clean Markdown format.

**Parameters:**
- `page_id` (required) - Page ID to export
- `include_children` (optional, boolean) - Recursively export child blocks (default: false)
- `frontmatter` (optional, boolean) - Include YAML frontmatter with metadata (default: false)

**Block to Markdown Mapping:**
- `paragraph` → plain text
- `heading_1/2/3` → `#`, `##`, `###` + text
- `bulleted_list_item` → `- ` + text
- `numbered_list_item` → `1. ` + text
- `to_do` → `- [ ]` or `- [x]` + text
- `code` → ` ```language\ncode\n``` `
- `quote` → `> ` + text
- `divider` → `---`
- `callout` → `> 💡 ` + text
- `child_page` → `[Page Title](URL)`

**Rich Text Formatting:**
- Bold: `**text**`
- Italic: `*text*`
- Code: `` `text` ``
- Link: `[text](url)`
- Strikethrough: `~~text~~`

**Example Response:**
```json
{
  "markdown": "# Page Title\n\n## Section\n\nContent here...",
  "metadata": {
    "title": "Page Title",
    "notion_url": "https://notion.so/...",
    "created_time": "2024-01-15T10:30:00Z",
    "last_edited_time": "2024-01-20T14:22:00Z"
  }
}
```

**With Frontmatter:**
```markdown
---
title: Page Title
notion_url: https://notion.so/...
created: 2024-01-15T10:30:00Z
last_edited: 2024-01-20T14:22:00Z
---

# Page Title

Content here...
```

#### 31. export_database_as_csv

Export database entries as CSV.

**Parameters:**
- `database_id` (required) - Database ID to export
- `filter` (optional) - Filter entries to export
- `include_archived` (optional, boolean) - Include archived entries (default: false)

**Returns:**
```json
{
  "csv": "Name,Status,Due Date\nTask 1,Done,2024-01-15\nTask 2,In Progress,2024-01-20",
  "row_count": 2
}
```

**Property Type Handling:**
- `title/rich_text` → plain text
- `number` → numeric value
- `select` → option name
- `multi_select` → names joined with semicolon
- `date` → start date (or date range)
- `checkbox` → "true"/"false"
- `url/email/phone` → direct value
- `people` → names joined with semicolon

### Smart Queries

Intelligent query helpers for common use cases.

#### 32. get_recently_edited

Get pages edited within a time window.

**Parameters:**
- `days` (optional, number) - Days to look back (default: 7)
- `database_id` (optional) - Limit to specific database
- `limit` (optional, number) - Max results (default: 50)

**Example:**
```json
{
  "days": 3,
  "limit": 20
}
```

**Use Cases:**
- "Show me pages I worked on this week"
- "What changed in the last 24 hours?"
- "Find recently updated project pages"

#### 33. get_my_tasks

Smart task finder for the current user.

**Parameters:**
- `database_id` (optional) - Specific tasks database (auto-searches if not provided)
- `statuses` (optional, array) - Status values to include (default: ["In Progress", "To Do", "Not Started"])
- `include_overdue` (optional, boolean) - Include overdue tasks (default: true)
- `limit` (optional, number) - Max tasks (default: 50)

**Property Detection:**
Automatically detects property names including:
- **Status**: "Status", "State", "Progress", "Stage"
- **Assignee**: "Assigned To", "Assignee", "Owner", "Assigned"
- **Due Date**: "Due Date", "Due", "Deadline"

**Example:**
```json
{
  "statuses": ["In Progress", "Blocked"],
  "include_overdue": true,
  "limit": 25
}
```

**Returns:**
```json
{
  "count": 5,
  "database_id": "abc123...",
  "tasks": [
    {
      "id": "task-id-1",
      "url": "https://notion.so/...",
      "title": "Fix login bug",
      "status": "In Progress",
      "due_date": "2024-01-20",
      "overdue": true
    }
  ]
}
```

#### 34. get_related_pages

Follow page relations and return connected pages.

**Parameters:**
- `page_id` (required) - Page ID to get relations for
- `relation_property` (optional) - Specific relation property name (returns all if not specified)
- `include_properties` (optional, boolean) - Include full page properties (default: false)

**Example:**
```json
{
  "page_id": "abc123...",
  "include_properties": true
}
```

**Returns:**
```json
{
  "page_id": "abc123...",
  "relations": {
    "Related Projects": [
      {
        "id": "proj-1",
        "title": "Project Alpha",
        "url": "https://notion.so/...",
        "properties": {...}
      }
    ],
    "Blocked By": [
      {
        "id": "task-2",
        "title": "Task 123"
      }
    ]
  },
  "count": 2
}
```

**Use Cases:**
- "Show all projects related to this task"
- "Find pages blocking this issue"
- "Get all linked resources"

## Notion Concepts

### Pages

Pages are the basic unit of content in Notion. They can:
- Exist standalone in the workspace
- Be children of other pages
- Be entries in a database
- Contain properties (metadata)
- Contain blocks (content)

### Databases

Databases are collections of pages with:
- A defined schema (properties)
- Property types (title, text, select, date, etc.)
- Views and filters
- Can be full-page or inline

### Blocks

Blocks are content units within pages:
- Paragraphs, headings, lists
- To-do items, toggles
- Code blocks, quotes, callouts
- Child pages, dividers
- Can be nested (blocks can have children)

### Properties

Properties are typed fields that store metadata:
- **Title**: Required for pages, primary text
- **Rich Text**: Multi-line text with formatting
- **Number**: Numeric values
- **Select**: Single choice from options
- **Multi-select**: Multiple choices
- **Date**: Date or date range
- **Checkbox**: Boolean true/false
- **URL**: Web link
- **Email**: Email address
- **People**: User references
- **Files**: File attachments

## Finding Page/Database IDs

IDs can be extracted from Notion URLs:

### From URL

```
https://notion.so/workspace/Page-Title-12345678901234567890123456789012
```

The ID is the 32-character string at the end: `12345678901234567890123456789012`

### Normalize to API Format

The API requires UUIDs with dashes. This server handles both formats:
- Without dashes: `12345678901234567890123456789012`
- With dashes: `12345678-9012-3456-7890-123456789012` (API format)

You can provide either format - the server will normalize them automatically.

## Filter and Sort Examples

### Filter Examples

**Single condition:**
```json
{
  "property": "Status",
  "select": {
    "equals": "In Progress"
  }
}
```

**Multiple conditions (AND):**
```json
{
  "and": [
    {
      "property": "Status",
      "select": {"equals": "Done"}
    },
    {
      "property": "Priority",
      "select": {"equals": "High"}
    }
  ]
}
```

**Date filters:**
```json
{
  "property": "Due Date",
  "date": {
    "on_or_before": "2024-12-31"
  }
}
```

**Text search:**
```json
{
  "property": "Description",
  "rich_text": {
    "contains": "important"
  }
}
```

### Sort Examples

```json
{
  "sorts": [
    {
      "property": "Priority",
      "direction": "ascending"
    },
    {
      "property": "Due Date",
      "direction": "descending"
    }
  ]
}
```

## Block Type Examples

### Paragraph
```json
{
  "object": "block",
  "type": "paragraph",
  "paragraph": {
    "rich_text": [{"text": {"content": "Plain text"}}]
  }
}
```

### Heading
```json
{
  "object": "block",
  "type": "heading_2",
  "heading_2": {
    "rich_text": [{"text": {"content": "Section Title"}}]
  }
}
```

### To-Do
```json
{
  "object": "block",
  "type": "to_do",
  "to_do": {
    "rich_text": [{"text": {"content": "Task item"}}],
    "checked": false
  }
}
```

### Code Block
```json
{
  "object": "block",
  "type": "code",
  "code": {
    "rich_text": [{"text": {"content": "console.log('hello');"}}],
    "language": "javascript"
  }
}
```

### Bulleted List
```json
{
  "object": "block",
  "type": "bulleted_list_item",
  "bulleted_list_item": {
    "rich_text": [{"text": {"content": "List item"}}]
  }
}
```

## Common Use Cases

### Create a Meeting Notes Page

```
"Create a simple page titled 'Team Sync - Feb 3' with sections for Discussion, Decisions, and Action Items"
```

### Search for All Tasks

```
"Search for all pages in the Tasks database where Status is 'In Progress'"
```

### Add Notes to Existing Page

```
"Append these notes to today's daily log: [your notes]"
```

### Create a Database

```
"Create a new database called 'Reading List' with properties for Title, Author, Status (Not Started/Reading/Done), and Rating (1-5)"
```

### Query Database with Filters

```
"Find all entries in the Projects database where Status is 'Active' and Priority is 'High', sorted by due date"
```

### Update Page Properties

```
"Update the Status property of page [id] to 'Complete' and set the Completion Date to today"
```

## Troubleshooting

### "Object not found" Error

**Problem:** "object_not_found: Could not find page/database"

**Solutions:**
1. Make sure the page/database is **shared with your integration**:
   - Open the page in Notion
   - Click "..." → "Add connections"
   - Select your integration
2. Verify the page/database ID is correct
3. Check that the integration has the necessary permissions

### Authentication Failed

**Problem:** "Failed to connect to Notion API (check your token)"

**Solutions:**
- Verify your `NOTION_API_TOKEN` is correct
- Make sure the token starts with `secret_`
- Regenerate the token at notion.so/my-integrations if needed
- Check that your integration hasn't been deleted

### Invalid UUID Format

**Problem:** "invalid UUID format" or "invalid UUID length"

**Solutions:**
- Extract the 32-character ID from the Notion URL
- The server accepts both formats (with or without dashes)
- Make sure you're using the page/database ID, not the workspace ID

### Rate Limiting

**Problem:** API requests are slow or failing

**Solutions:**
- The server automatically handles rate limiting (3 requests/sec average)
- If you hit rate limits repeatedly, add delays between operations
- The server implements exponential backoff on 429 errors

### Network Timeouts

**Problem:** "request timeout" errors

**Solutions:**
- Check your internet connection
- Increase `NOTION_TIMEOUT` environment variable (default: 30 seconds)
- Notion API may be temporarily unavailable - retry after a moment

### Property Not Found

**Problem:** "property does not exist" when querying/updating

**Solutions:**
- Use `get_database` to see the exact property names and types
- Property names are case-sensitive
- Check the property type matches what you're trying to do

## Development

### Running Locally

```bash
# Set environment variables
export NOTION_API_TOKEN="secret_your_token_here"
export NOTION_API_VERSION="2022-06-28"

# Run the server
go run main.go
```

### Building

```bash
# Build for your current platform
go build -o mcp-notion

# Build for specific platforms
GOOS=linux GOARCH=amd64 go build -o mcp-notion-linux
GOOS=darwin GOARCH=arm64 go build -o mcp-notion-macos
GOOS=windows GOARCH=amd64 go build -o mcp-notion-windows.exe
```

### Testing with MCP Inspector

Use the [MCP Inspector](https://github.com/modelcontextprotocol/inspector) to test the server:

```bash
npx @modelcontextprotocol/inspector mcp-notion
```

## Architecture

The server consists of:

- **Configuration** (`config/config.go`) - Environment variable loading and validation
- **Notion Client** (`notion/client.go`) - HTTP client wrapper with rate limiting and error handling
- **Type Definitions** (`notion/types.go`) - Go structs for Notion API objects
- **Helper Functions** (`notion/helpers.go`) - Block/property builders and markdown converter
- **Tool Handlers** (`tools/*.go`) - MCP tool implementations for each operation
- **Main Server** (`main.go`) - MCP server initialization and tool registration

## Dependencies

- [mcp-go](https://github.com/mark3labs/mcp-go) v0.43.2+ - Official MCP Go SDK
- [resty](https://github.com/go-resty/resty) v2.16.4+ - HTTP client
- [uuid](https://github.com/google/uuid) v1.6.0+ - UUID handling
- [godotenv](https://github.com/joho/godotenv) v1.5.1 - Environment variable loading

## Security Considerations

- Never commit your `.env` file to version control
- Store your integration token securely
- Integration tokens can be revoked at notion.so/my-integrations
- The server runs locally and doesn't send data to third parties
- All communication with Notion uses HTTPS (TLS encryption)
- Use integration permissions to limit access scope

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues, questions, or feature requests, please open an issue on [GitHub](https://github.com/rgabriel/mcp-notion/issues).

## Acknowledgments

- Built with the official [mcp-go SDK](https://mcp-go.dev)
- Uses [Notion REST API v1](https://developers.notion.com/reference/intro)
- Follows the [Model Context Protocol](https://modelcontextprotocol.io) specification
