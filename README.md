# Notion MCP Server

A Model Context Protocol (MCP) server that connects AI assistants to Notion workspaces via the official REST API. Search, read, create, and manage pages, databases, blocks, comments, and users — all through a standardized tool interface.

## Features

- **Search & Discovery** — Full-text search across pages and databases, list all databases
- **Page Management** — Get, create, update, move, and delete pages with full property support
- **Database Operations** — Query with complex filters and sorts, create and update database schemas
- **Block Management** — Read, append, update, and delete content blocks
- **Comments** — Read and create comments on pages and discussion threads
- **User Management** — List workspace users, get user details, identify bot user
- **Helper Tools** — Simplified page creation with markdown-like syntax, in-database search
- **Batch Operations** — Create, update, or delete multiple pages with rate limiting and error recovery
- **Templates** — Create pages from built-in templates (meeting notes, daily log, project brief, sprint planning, retrospective)
- **Export** — Export pages as Markdown (with optional frontmatter) and databases as CSV
- **Smart Queries** — Recently edited pages, task finder, relation follower
- **Circuit Breaker** — Automatic failure detection with three-state circuit breaker (closed/open/half-open) to prevent cascading failures
- **Secret Redaction** — API tokens and secrets are automatically redacted from all log output
- **Audit Logging** — Destructive operations (create, update, delete) are logged with full argument details
- **Concurrency Control** — Configurable semaphore limits concurrent tool executions
- **Hardened Transport** — TLS 1.2 minimum, granular connection timeouts, response body size limits
- **Cross-Platform** — Linux, macOS, and Windows

## Prerequisites

- **Go 1.25+** — [Install Go](https://go.dev/doc/install)
- **Notion Account** with access to create integrations
- **Notion Integration Token** — Create one at [notion.so/my-integrations](https://www.notion.so/my-integrations)

## Getting Your Notion Integration Token

1. Go to [notion.so/my-integrations](https://www.notion.so/my-integrations)
2. Click **New integration**
3. Give it a name (e.g., "MCP Server")
4. Select the workspace you want to connect
5. Under **Capabilities**, enable the permissions your use case requires:
   - **Read content** — required for all read operations
   - **Update content** — required for create, update, and delete operations
   - **Insert content** — required for creating pages and appending blocks
   - **Read comments** — required for reading comments
   - **Insert comments** — required for creating comments
   - **Read user information** — required for user-related tools
6. Click **Submit** to create the integration
7. Copy the **Internal Integration Secret** (starts with `secret_`)
8. Share pages and databases with the integration: open a page in Notion, click the **...** menu, select **Add connections**, and choose your integration

> **Important:** The integration can only access pages and databases that have been explicitly shared with it. If you get "object not found" errors, make sure you've shared the relevant pages.

## Installation

```bash
# Clone and build
git clone https://github.com/rgabriel/mcp-notion.git
cd mcp-notion
make build

# Or install directly
go install github.com/rgabriel/mcp-notion@latest
```

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `NOTION_API_TOKEN` | Yes | — | Integration token from [notion.so/my-integrations](https://www.notion.so/my-integrations) (starts with `secret_`) |
| `NOTION_API_VERSION` | No | `2022-06-28` | Notion API version |
| `NOTION_TIMEOUT` | No | `30` | HTTP request timeout in seconds |

Copy `.env.example` to `.env` for local development:

```bash
cp .env.example .env
# Edit .env and set NOTION_API_TOKEN
```

## Usage with Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "notion": {
      "command": "/path/to/mcp-notion",
      "env": {
        "NOTION_API_TOKEN": "secret_your_token_here"
      }
    }
  }
}
```

**Config file locations:**
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Linux: `~/.config/claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

## Usage with Claude Code

Add to your `~/.claude.json` or project `.claude/settings.json`:

```json
{
  "mcpServers": {
    "notion": {
      "command": "/path/to/mcp-notion",
      "env": {
        "NOTION_API_TOKEN": "secret_your_token_here"
      }
    }
  }
}
```

Or launch via CLI:

```bash
NOTION_API_TOKEN="secret_..." claude --mcp-server "notion:/path/to/mcp-notion"
```

## Available Tools

### Search & Discovery

| Tool | Description | Required Parameters |
|---|---|---|
| `search` | Search across all pages and databases | — |
| `list_databases` | List all accessible databases | — |
| `get_database` | Get database details including schema | `database_id` |

### Pages

| Tool | Description | Required Parameters |
|---|---|---|
| `get_page` | Get page metadata and properties | `page_id` |
| `get_page_content` | Get page content as blocks | `page_id` |
| `create_page` | Create a page in a database or under a page | `parent`, `properties` |
| `update_page` | Update page properties, icon, or cover | `page_id` |
| `move_page` | Move a page to a new parent | `page_id`, `parent` |
| `delete_page` | Archive (delete) a page | `page_id` |

### Databases

| Tool | Description | Required Parameters |
|---|---|---|
| `query_database` | Query with filters, sorts, pagination | `database_id` |
| `create_database` | Create a new database | `parent`, `title`, `properties` |
| `update_database` | Update database title, description, schema | `database_id` |

### Blocks

| Tool | Description | Required Parameters |
|---|---|---|
| `get_block` | Get a specific block | `block_id` |
| `get_block_children` | Get child blocks of a block or page | `block_id` |
| `append_blocks` | Append blocks to a page or block | `block_id`, `children` |
| `update_block` | Update block content | `block_id` |
| `delete_block` | Delete a block and its children | `block_id` |

### Comments

| Tool | Description | Required Parameters |
|---|---|---|
| `get_comments` | Get comments on a page or block | `block_id` or `page_id` |
| `create_comment` | Create a comment or reply | `parent`, `rich_text` |

### Users

| Tool | Description | Required Parameters |
|---|---|---|
| `list_users` | List all workspace users | — |
| `get_user` | Get user details | `user_id` |
| `get_bot_user` | Get current bot user info | — |

### Helper Tools

| Tool | Description | Required Parameters |
|---|---|---|
| `create_simple_page` | Create a page with markdown content | `title` |
| `append_to_page` | Append markdown content to a page | `page_id`, `content` |
| `search_in_database` | Search within a specific database | `database_id`, `query` |

### Batch Operations

| Tool | Description | Required Parameters |
|---|---|---|
| `batch_create_pages` | Create multiple pages | `pages` |
| `batch_update_pages` | Update multiple pages | `page_ids` |
| `batch_delete_pages` | Archive multiple pages | `page_ids` |

### Templates

| Tool | Description | Required Parameters |
|---|---|---|
| `create_page_from_template` | Create page from template | `template_name`, `title` |
| `list_templates` | List available templates | — |

Available templates: `meeting_notes`, `daily_log`, `project_brief`, `sprint_planning`, `retrospective`

### Export

| Tool | Description | Required Parameters |
|---|---|---|
| `export_page_as_markdown` | Export page as Markdown | `page_id` |
| `export_database_as_csv` | Export database as CSV | `database_id` |

### Smart Queries

| Tool | Description | Required Parameters |
|---|---|---|
| `get_recently_edited` | Get recently edited pages | — |
| `get_my_tasks` | Find tasks assigned to bot user | — |
| `get_related_pages` | Follow page relations | `page_id` |

## Notion-Specific Features

### Property Types

Notion databases support many property types. When creating or updating pages, use the appropriate property format:

- **Title** — `{"title": [{"text": {"content": "Page name"}}]}`
- **Rich Text** — `{"rich_text": [{"text": {"content": "Text value"}}]}`
- **Number** — `{"number": 42}`
- **Select** — `{"select": {"name": "Option"}}`
- **Multi-Select** — `{"multi_select": [{"name": "Tag1"}, {"name": "Tag2"}]}`
- **Date** — `{"date": {"start": "2024-01-15"}}`
- **Checkbox** — `{"checkbox": true}`
- **URL** — `{"url": "https://example.com"}`
- **Email** — `{"email": "user@example.com"}`
- **Phone** — `{"phone_number": "+1-555-0123"}`

### Filter Syntax

Use `query_database` with Notion's filter format:

```json
{
  "filter": {
    "and": [
      {"property": "Status", "select": {"equals": "In Progress"}},
      {"property": "Priority", "select": {"equals": "High"}}
    ]
  },
  "sorts": [
    {"property": "Due Date", "direction": "ascending"}
  ]
}
```

## Rate Limiting

The server enforces a 3 requests/second rate limit against the Notion API with automatic retries and exponential backoff. Batch operations add 350ms delays between individual requests to stay within limits.

If the circuit breaker detects 5 consecutive API failures (5xx errors or connection failures), it temporarily blocks outgoing requests for 30 seconds before allowing a probe request to test recovery.

## Development

```bash
# Run tests
make test

# Run tests with coverage
make cover

# Lint
make lint

# Format code
make fmt

# Build with version
make build

# Run locally
NOTION_API_TOKEN="secret_..." make run
```

## Troubleshooting

### "Object not found" error
The page/database hasn't been shared with your integration. Open it in Notion, click the **...** menu, select **Add connections**, and choose your integration.

### Authentication failed
Verify `NOTION_API_TOKEN` is set correctly and starts with `secret_`. Regenerate at [notion.so/my-integrations](https://www.notion.so/my-integrations) if needed.

### Invalid UUID format
Extract the 32-character ID from the Notion URL. Both dash and no-dash formats are accepted. Example: `https://notion.so/Page-Title-a1b2c3d4e5f6...` — the ID is the hex string at the end.

### Rate limiting
The server automatically handles rate limits with retries. If you're seeing persistent rate limit errors, reduce the frequency of batch operations or increase delays.

### Timeout errors
Increase `NOTION_TIMEOUT` (default: 30 seconds). Check your network connection.

### Property not found
Use `get_database` to see exact property names — they are case-sensitive.

### Circuit breaker open
If you see "circuit breaker is open" errors, the Notion API has been returning consecutive failures. The circuit breaker will automatically retry after 30 seconds.

## Architecture

The server is structured as three packages:

- **`main`** — MCP server setup, middleware composition (concurrency control, observability/audit logging, panic recovery), and tool registration
- **`notion`** — Notion API client with rate limiting (3 req/sec with token bucket), circuit breaker (5-failure threshold, 30s reset), retry with exponential backoff, and hardened HTTP transport
- **`tools`** — MCP tool handler implementations for all 30+ tools, organized by resource type (pages, databases, blocks, comments, users, batch, templates, export, smart queries)
- **`config`** — Environment-based configuration loading with `.env` file support

Middleware is applied in this order (outermost to innermost):
1. **Concurrency limiter** — caps parallel tool executions
2. **Observability** — logs tool calls with timing; adds audit details for destructive operations
3. **Recovery** — catches panics and returns error results

## Dependencies

| Package | Purpose |
|---|---|
| [mcp-go](https://github.com/mark3labs/mcp-go) | MCP protocol SDK for Go |
| [resty](https://github.com/go-resty/resty) | HTTP client with retry support |
| [godotenv](https://github.com/joho/godotenv) | `.env` file loading |
| [uuid](https://github.com/google/uuid) | UUID generation for request tracing |

## Security Considerations

- **Token handling** — The Notion API token is loaded from environment variables, never hardcoded. The `.env` file is gitignored.
- **Secret redaction** — All log output passes through a redacting handler that replaces API tokens with `[REDACTED]`.
- **Audit logging** — Destructive operations (create, update, delete) are logged with their arguments for accountability.
- **Transport hardening** — TLS 1.2 minimum version, 5-second TLS handshake timeout, 5-second dial timeout, 10 MB response body size limit.
- **Input validation** — All tool handlers validate required parameters before making API calls. UUID formats are normalized and validated.
- **No credential storage** — The server is stateless and does not persist credentials.

## License

MIT License — see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome. Please open an issue to discuss changes before submitting a PR.

## Support

For bug reports and feature requests, please [open an issue](https://github.com/rgabriel/mcp-notion/issues).

## Acknowledgments

- [Model Context Protocol](https://modelcontextprotocol.io) specification
- [mcp-go](https://github.com/mark3labs/mcp-go) SDK
- [Notion API](https://developers.notion.com) documentation
