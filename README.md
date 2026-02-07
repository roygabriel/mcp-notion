# Notion MCP Server

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A Model Context Protocol (MCP) server that connects AI assistants to Notion workspaces via the official REST API. Search, read, create, and manage pages, databases, blocks, comments, and users.

## Features

- **Search & Discovery** — Search across pages and databases, list all databases
- **Page Management** — Get, create, update, move, and delete pages with full property support
- **Database Operations** — Query with complex filters, create and update database schemas
- **Block Management** — Read, append, update, and delete content blocks
- **Comments** — Read and create comments on pages and blocks
- **User Management** — List workspace users, get user details
- **Helper Tools** — Simplified page creation with markdown-like syntax
- **Batch Operations** — Create, update, or delete multiple pages with rate limiting
- **Templates** — Create pages from built-in templates (meeting notes, daily log, etc.)
- **Export** — Export pages as Markdown and databases as CSV
- **Smart Queries** — Recently edited pages, task finder, relation follower
- **Cross-Platform** — Linux, macOS, and Windows

## Prerequisites

- **Go 1.25+** — [Install Go](https://go.dev/doc/install)
- **Notion Account** with access to create integrations
- **Notion Integration Token** — Create one at [notion.so/my-integrations](https://www.notion.so/my-integrations)

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

**Important:** You must share each page/database with your integration via the "..." menu → "Add connections" in Notion. Without this, you'll get "object not found" errors.

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

## Examples

```
"Search for all pages about project planning"

"Create a meeting notes page titled 'Team Sync - Feb 3' with sections for discussion and action items"

"Query the Tasks database for all items where Status is 'In Progress' and Priority is 'High'"

"Export the Q1 Planning page as Markdown with frontmatter"

"Show me pages I edited in the last 3 days"
```

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
The page/database hasn't been shared with your integration. Open it in Notion → "..." menu → "Add connections" → select your integration.

### Authentication failed
Verify `NOTION_API_TOKEN` is set correctly and starts with `secret_`. Regenerate at [notion.so/my-integrations](https://www.notion.so/my-integrations) if needed.

### Invalid UUID format
Extract the 32-character ID from the Notion URL. Both dash and no-dash formats are accepted.

### Rate limiting
The server enforces 3 req/sec with automatic retries and exponential backoff. Batch operations add 350ms delays between requests.

### Timeout errors
Increase `NOTION_TIMEOUT` (default: 30 seconds). Check your network connection.

### Property not found
Use `get_database` to see exact property names — they are case-sensitive.

## License

MIT License — see [LICENSE](LICENSE) for details.
