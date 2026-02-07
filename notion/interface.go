package notion

import "context"

// NotionClient defines the interface for interacting with the Notion API.
// This enables dependency injection and testing with mock implementations.
type NotionClient interface {
	Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
	GetDatabase(ctx context.Context, databaseID string) (*Database, error)
	QueryDatabase(ctx context.Context, databaseID string, req *QueryDatabaseRequest) (*QueryDatabaseResponse, error)
	CreateDatabase(ctx context.Context, req *CreateDatabaseRequest) (*Database, error)
	UpdateDatabase(ctx context.Context, databaseID string, req *UpdateDatabaseRequest) (*Database, error)
	GetPage(ctx context.Context, pageID string) (*Page, error)
	CreatePage(ctx context.Context, req *CreatePageRequest) (*Page, error)
	UpdatePage(ctx context.Context, pageID string, req *UpdatePageRequest) (*Page, error)
	MovePage(ctx context.Context, pageID string, req *MovePageRequest) (*Page, error)
	GetBlock(ctx context.Context, blockID string) (*Block, error)
	GetBlockChildren(ctx context.Context, blockID string, pageSize int, startCursor string) (*BlockListResponse, error)
	AppendBlockChildren(ctx context.Context, blockID string, req *AppendBlockChildrenRequest) (*BlockListResponse, error)
	UpdateBlock(ctx context.Context, blockID string, block *Block) (*Block, error)
	DeleteBlock(ctx context.Context, blockID string) (*Block, error)
	GetComments(ctx context.Context, blockID, pageID string, pageSize int, startCursor string) (*CommentListResponse, error)
	CreateComment(ctx context.Context, req *CreateCommentRequest) (*Comment, error)
	ListUsers(ctx context.Context, pageSize int, startCursor string) (*UserListResponse, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	GetBotUser(ctx context.Context) (*User, error)
}
