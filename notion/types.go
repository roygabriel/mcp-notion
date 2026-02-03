package notion

import "time"

// RichText represents Notion's rich text object
type RichText struct {
	Type        string            `json:"type"`
	Text        *TextContent      `json:"text,omitempty"`
	Annotations *Annotations      `json:"annotations,omitempty"`
	PlainText   string            `json:"plain_text,omitempty"`
	Href        string            `json:"href,omitempty"`
	Mention     map[string]any    `json:"mention,omitempty"`
	Equation    map[string]string `json:"equation,omitempty"`
}

// TextContent represents the text content within a RichText object
type TextContent struct {
	Content string      `json:"content"`
	Link    *LinkObject `json:"link,omitempty"`
}

// LinkObject represents a link in rich text
type LinkObject struct {
	URL string `json:"url"`
}

// Annotations represents text formatting
type Annotations struct {
	Bold          bool   `json:"bold"`
	Italic        bool   `json:"italic"`
	Strikethrough bool   `json:"strikethrough"`
	Underline     bool   `json:"underline"`
	Code          bool   `json:"code"`
	Color         string `json:"color"`
}

// Parent represents a parent object (page or database)
type Parent struct {
	Type       string `json:"type"`
	PageID     string `json:"page_id,omitempty"`
	DatabaseID string `json:"database_id,omitempty"`
	Workspace  bool   `json:"workspace,omitempty"`
}

// Icon represents a page or database icon
type Icon struct {
	Type     string  `json:"type"`
	Emoji    string  `json:"emoji,omitempty"`
	External *File   `json:"external,omitempty"`
	File     *File   `json:"file,omitempty"`
}

// Cover represents a page cover image
type Cover struct {
	Type     string `json:"type"`
	External *File  `json:"external,omitempty"`
	File     *File  `json:"file,omitempty"`
}

// File represents a file object
type File struct {
	URL        string     `json:"url,omitempty"`
	ExpiryTime *time.Time `json:"expiry_time,omitempty"`
}

// Page represents a Notion page
type Page struct {
	Object         string         `json:"object"`
	ID             string         `json:"id"`
	CreatedTime    time.Time      `json:"created_time"`
	LastEditedTime time.Time      `json:"last_edited_time"`
	CreatedBy      *User          `json:"created_by,omitempty"`
	LastEditedBy   *User          `json:"last_edited_by,omitempty"`
	Parent         *Parent        `json:"parent"`
	Archived       bool           `json:"archived"`
	Properties     map[string]any `json:"properties"`
	Icon           *Icon          `json:"icon,omitempty"`
	Cover          *Cover         `json:"cover,omitempty"`
	URL            string         `json:"url"`
}

// Database represents a Notion database
type Database struct {
	Object         string                  `json:"object"`
	ID             string                  `json:"id"`
	CreatedTime    time.Time               `json:"created_time"`
	LastEditedTime time.Time               `json:"last_edited_time"`
	CreatedBy      *User                   `json:"created_by,omitempty"`
	LastEditedBy   *User                   `json:"last_edited_by,omitempty"`
	Title          []RichText              `json:"title"`
	Description    []RichText              `json:"description,omitempty"`
	Icon           *Icon                   `json:"icon,omitempty"`
	Cover          *Cover                  `json:"cover,omitempty"`
	Properties     map[string]PropertyDef  `json:"properties"`
	Parent         *Parent                 `json:"parent"`
	Archived       bool                    `json:"archived"`
	IsInline       bool                    `json:"is_inline"`
	URL            string                  `json:"url"`
}

// PropertyDef represents a database property schema definition
type PropertyDef struct {
	ID          string         `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Type        string         `json:"type"`
	Title       map[string]any `json:"title,omitempty"`
	RichText    map[string]any `json:"rich_text,omitempty"`
	Number      map[string]any `json:"number,omitempty"`
	Select      map[string]any `json:"select,omitempty"`
	MultiSelect map[string]any `json:"multi_select,omitempty"`
	Date        map[string]any `json:"date,omitempty"`
	People      map[string]any `json:"people,omitempty"`
	Files       map[string]any `json:"files,omitempty"`
	Checkbox    map[string]any `json:"checkbox,omitempty"`
	URL         map[string]any `json:"url,omitempty"`
	Email       map[string]any `json:"email,omitempty"`
	PhoneNumber map[string]any `json:"phone_number,omitempty"`
	Formula     map[string]any `json:"formula,omitempty"`
	Relation    map[string]any `json:"relation,omitempty"`
	Rollup      map[string]any `json:"rollup,omitempty"`
	CreatedTime map[string]any `json:"created_time,omitempty"`
	CreatedBy   map[string]any `json:"created_by,omitempty"`
	LastEditedTime map[string]any `json:"last_edited_time,omitempty"`
	LastEditedBy   map[string]any `json:"last_edited_by,omitempty"`
}

// Block represents a Notion block
type Block struct {
	Object         string     `json:"object"`
	ID             string     `json:"id,omitempty"`
	Type           string     `json:"type"`
	CreatedTime    time.Time  `json:"created_time,omitempty"`
	LastEditedTime time.Time  `json:"last_edited_time,omitempty"`
	CreatedBy      *User      `json:"created_by,omitempty"`
	LastEditedBy   *User      `json:"last_edited_by,omitempty"`
	HasChildren    bool       `json:"has_children,omitempty"`
	Archived       bool       `json:"archived,omitempty"`
	
	// Block type-specific fields
	Paragraph         *RichTextBlock       `json:"paragraph,omitempty"`
	Heading1          *RichTextBlock       `json:"heading_1,omitempty"`
	Heading2          *RichTextBlock       `json:"heading_2,omitempty"`
	Heading3          *RichTextBlock       `json:"heading_3,omitempty"`
	BulletedListItem  *RichTextBlock       `json:"bulleted_list_item,omitempty"`
	NumberedListItem  *RichTextBlock       `json:"numbered_list_item,omitempty"`
	ToDo              *ToDoBlock           `json:"to_do,omitempty"`
	Toggle            *RichTextBlock       `json:"toggle,omitempty"`
	Code              *CodeBlock           `json:"code,omitempty"`
	Quote             *RichTextBlock       `json:"quote,omitempty"`
	Callout           *CalloutBlock        `json:"callout,omitempty"`
	Divider           map[string]any       `json:"divider,omitempty"`
	TableOfContents   map[string]any       `json:"table_of_contents,omitempty"`
	Bookmark          *BookmarkBlock       `json:"bookmark,omitempty"`
	LinkToPage        *LinkToPageBlock     `json:"link_to_page,omitempty"`
	ChildPage         *ChildPageBlock      `json:"child_page,omitempty"`
	ChildDatabase     *ChildDatabaseBlock  `json:"child_database,omitempty"`
}

// RichTextBlock represents blocks that contain rich text
type RichTextBlock struct {
	RichText []RichText `json:"rich_text"`
	Color    string     `json:"color,omitempty"`
	Children []Block    `json:"children,omitempty"`
}

// ToDoBlock represents a to-do block
type ToDoBlock struct {
	RichText []RichText `json:"rich_text"`
	Checked  bool       `json:"checked"`
	Color    string     `json:"color,omitempty"`
	Children []Block    `json:"children,omitempty"`
}

// CodeBlock represents a code block
type CodeBlock struct {
	RichText []RichText `json:"rich_text"`
	Caption  []RichText `json:"caption,omitempty"`
	Language string     `json:"language"`
}

// CalloutBlock represents a callout block
type CalloutBlock struct {
	RichText []RichText `json:"rich_text"`
	Icon     *Icon      `json:"icon,omitempty"`
	Color    string     `json:"color,omitempty"`
	Children []Block    `json:"children,omitempty"`
}

// BookmarkBlock represents a bookmark block
type BookmarkBlock struct {
	URL     string     `json:"url"`
	Caption []RichText `json:"caption,omitempty"`
}

// LinkToPageBlock represents a link to page block
type LinkToPageBlock struct {
	Type       string `json:"type"`
	PageID     string `json:"page_id,omitempty"`
	DatabaseID string `json:"database_id,omitempty"`
}

// ChildPageBlock represents a child page block
type ChildPageBlock struct {
	Title string `json:"title"`
}

// ChildDatabaseBlock represents a child database block
type ChildDatabaseBlock struct {
	Title string `json:"title"`
}

// User represents a Notion user
type User struct {
	Object    string  `json:"object"`
	ID        string  `json:"id"`
	Type      string  `json:"type,omitempty"`
	Name      string  `json:"name,omitempty"`
	AvatarURL string  `json:"avatar_url,omitempty"`
	Person    *Person `json:"person,omitempty"`
	Bot       *Bot    `json:"bot,omitempty"`
}

// Person represents a person user
type Person struct {
	Email string `json:"email,omitempty"`
}

// Bot represents a bot user
type Bot struct {
	Owner         map[string]any `json:"owner,omitempty"`
	WorkspaceName string         `json:"workspace_name,omitempty"`
}

// Comment represents a Notion comment
type Comment struct {
	Object         string     `json:"object"`
	ID             string     `json:"id"`
	Parent         *Parent    `json:"parent"`
	DiscussionID   string     `json:"discussion_id"`
	RichText       []RichText `json:"rich_text"`
	CreatedTime    time.Time  `json:"created_time"`
	LastEditedTime time.Time  `json:"last_edited_time"`
	CreatedBy      *User      `json:"created_by"`
}

// SearchResponse represents a search API response
type SearchResponse struct {
	Object     string         `json:"object"`
	Results    []SearchResult `json:"results"`
	HasMore    bool           `json:"has_more"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

// SearchResult represents a search result (can be page or database)
type SearchResult struct {
	Object         string         `json:"object"`
	ID             string         `json:"id"`
	CreatedTime    time.Time      `json:"created_time,omitempty"`
	LastEditedTime time.Time      `json:"last_edited_time,omitempty"`
	Parent         *Parent        `json:"parent,omitempty"`
	Archived       bool           `json:"archived,omitempty"`
	URL            string         `json:"url,omitempty"`
	Properties     map[string]any `json:"properties,omitempty"`
	Title          []RichText     `json:"title,omitempty"`
	Description    []RichText     `json:"description,omitempty"`
	Icon           *Icon          `json:"icon,omitempty"`
	Cover          *Cover         `json:"cover,omitempty"`
}

// QueryDatabaseResponse represents a database query response
type QueryDatabaseResponse struct {
	Object     string `json:"object"`
	Results    []Page `json:"results"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// BlockListResponse represents a list of blocks response
type BlockListResponse struct {
	Object     string  `json:"object"`
	Results    []Block `json:"results"`
	HasMore    bool    `json:"has_more"`
	NextCursor string  `json:"next_cursor,omitempty"`
}

// CommentListResponse represents a list of comments response
type CommentListResponse struct {
	Object     string    `json:"object"`
	Results    []Comment `json:"results"`
	HasMore    bool      `json:"has_more"`
	NextCursor string    `json:"next_cursor,omitempty"`
}

// UserListResponse represents a list of users response
type UserListResponse struct {
	Object     string `json:"object"`
	Results    []User `json:"results"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// NotionError represents an error response from Notion API
type NotionError struct {
	Object  string `json:"object"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *NotionError) Error() string {
	return e.Message
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query  string         `json:"query,omitempty"`
	Filter map[string]any `json:"filter,omitempty"`
	Sort   map[string]any `json:"sort,omitempty"`
	PageSize int          `json:"page_size,omitempty"`
	StartCursor string    `json:"start_cursor,omitempty"`
}

// QueryDatabaseRequest represents a database query request
type QueryDatabaseRequest struct {
	Filter      map[string]any   `json:"filter,omitempty"`
	Sorts       []map[string]any `json:"sorts,omitempty"`
	PageSize    int              `json:"page_size,omitempty"`
	StartCursor string           `json:"start_cursor,omitempty"`
}

// CreatePageRequest represents a create page request
type CreatePageRequest struct {
	Parent     Parent         `json:"parent"`
	Properties map[string]any `json:"properties"`
	Children   []Block        `json:"children,omitempty"`
	Icon       *Icon          `json:"icon,omitempty"`
	Cover      *Cover         `json:"cover,omitempty"`
}

// UpdatePageRequest represents an update page request
type UpdatePageRequest struct {
	Properties map[string]any `json:"properties,omitempty"`
	Archived   *bool          `json:"archived,omitempty"`
	Icon       *Icon          `json:"icon,omitempty"`
	Cover      *Cover         `json:"cover,omitempty"`
}

// CreateDatabaseRequest represents a create database request
type CreateDatabaseRequest struct {
	Parent     Parent                 `json:"parent"`
	Title      []RichText             `json:"title"`
	Properties map[string]PropertyDef `json:"properties"`
	Icon       *Icon                  `json:"icon,omitempty"`
	Cover      *Cover                 `json:"cover,omitempty"`
}

// UpdateDatabaseRequest represents an update database request
type UpdateDatabaseRequest struct {
	Title      []RichText             `json:"title,omitempty"`
	Description []RichText            `json:"description,omitempty"`
	Properties map[string]PropertyDef `json:"properties,omitempty"`
	Archived   *bool                  `json:"archived,omitempty"`
}

// AppendBlockChildrenRequest represents a request to append block children
type AppendBlockChildrenRequest struct {
	Children []Block `json:"children"`
}

// CreateCommentRequest represents a create comment request
type CreateCommentRequest struct {
	Parent       Parent     `json:"parent"`
	RichText     []RichText `json:"rich_text"`
	DiscussionID string     `json:"discussion_id,omitempty"`
}
