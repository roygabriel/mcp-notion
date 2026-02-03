package notion

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/rgabriel/mcp-notion/config"
)

const (
	baseURL         = "https://api.notion.com/v1"
	maxRetries      = 5
	initialBackoff  = 1 * time.Second
	maxBackoff      = 32 * time.Second
	maxResultsLimit = 1000
)

// Client wraps the Resty HTTP client with Notion-specific functionality
type Client struct {
	client      *resty.Client
	apiVersion  string
	rateLimiter *RateLimiter
}

// RateLimiter tracks request timestamps to enforce rate limits
type RateLimiter struct {
	mu         sync.Mutex
	timestamps []time.Time
	maxPerSec  int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxPerSec int) *RateLimiter {
	return &RateLimiter{
		timestamps: make([]time.Time, 0),
		maxPerSec:  maxPerSec,
	}
}

// Wait blocks until a request can be made without exceeding rate limits
func (r *RateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	oneSecondAgo := now.Add(-1 * time.Second)

	// Remove timestamps older than 1 second
	validTimestamps := make([]time.Time, 0)
	for _, ts := range r.timestamps {
		if ts.After(oneSecondAgo) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	r.timestamps = validTimestamps

	// If we're at the limit, wait until we can make another request
	if len(r.timestamps) >= r.maxPerSec {
		oldestValid := r.timestamps[0]
		waitDuration := time.Until(oldestValid.Add(1 * time.Second))
		if waitDuration > 0 {
			time.Sleep(waitDuration)
		}
		// Clean up again after waiting
		now = time.Now()
		oneSecondAgo = now.Add(-1 * time.Second)
		validTimestamps = make([]time.Time, 0)
		for _, ts := range r.timestamps {
			if ts.After(oneSecondAgo) {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		r.timestamps = validTimestamps
	}

	// Record this request
	r.timestamps = append(r.timestamps, now)
}

// NewClient creates a new Notion API client
func NewClient(cfg *config.Config) *Client {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetTimeout(time.Duration(cfg.NotionTimeout) * time.Second)
	client.SetHeader("Authorization", "Bearer "+cfg.NotionAPIToken)
	client.SetHeader("Notion-Version", cfg.NotionAPIVersion)
	client.SetHeader("Content-Type", "application/json")

	// Add retry mechanism for rate limiting
	client.SetRetryCount(maxRetries)
	client.SetRetryWaitTime(initialBackoff)
	client.SetRetryMaxWaitTime(maxBackoff)
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		// Retry on 429 (rate limit) or 5xx errors
		return r.StatusCode() == 429 || r.StatusCode() >= 500
	})

	return &Client{
		client:      client,
		apiVersion:  cfg.NotionAPIVersion,
		rateLimiter: NewRateLimiter(3), // Notion's limit is ~3 req/sec
	}
}

// NormalizeID converts a UUID string to Notion's format (with dashes)
func NormalizeID(id string) (string, error) {
	// Remove any existing dashes
	cleaned := strings.ReplaceAll(id, "-", "")
	
	// Validate length
	if len(cleaned) != 32 {
		return "", fmt.Errorf("invalid UUID length: expected 32 characters, got %d", len(cleaned))
	}

	// Parse and format with dashes
	parsed, err := uuid.Parse(cleaned[:8] + "-" + cleaned[8:12] + "-" + cleaned[12:16] + "-" + cleaned[16:20] + "-" + cleaned[20:])
	if err != nil {
		return "", fmt.Errorf("invalid UUID format: %w", err)
	}

	return parsed.String(), nil
}

// handleError processes Notion API error responses
func (c *Client) handleError(resp *resty.Response) error {
	var notionErr NotionError
	if err := json.Unmarshal(resp.Body(), &notionErr); err == nil && notionErr.Message != "" {
		// Add context for common errors
		if notionErr.Code == "object_not_found" {
			return fmt.Errorf("%s (make sure the page/database is shared with your integration at notion.so/my-integrations)", notionErr.Message)
		}
		return &notionErr
	}
	return fmt.Errorf("API error: %s (status: %d)", resp.Status(), resp.StatusCode())
}

// Search searches across all pages and databases
func (c *Client) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	c.rateLimiter.Wait()

	if req.PageSize == 0 {
		req.PageSize = 100
	} else if req.PageSize > 100 {
		req.PageSize = 100
	}

	var result SearchResponse
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/search")

	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// GetDatabase retrieves a database by ID
func (c *Client) GetDatabase(ctx context.Context, databaseID string) (*Database, error) {
	c.rateLimiter.Wait()

	normalizedID, err := NormalizeID(databaseID)
	if err != nil {
		return nil, fmt.Errorf("invalid database ID: %w", err)
	}

	var result Database
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/databases/" + normalizedID)

	if err != nil {
		return nil, fmt.Errorf("get database request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// QueryDatabase queries a database with filters and sorts
func (c *Client) QueryDatabase(ctx context.Context, databaseID string, req *QueryDatabaseRequest) (*QueryDatabaseResponse, error) {
	c.rateLimiter.Wait()

	normalizedID, err := NormalizeID(databaseID)
	if err != nil {
		return nil, fmt.Errorf("invalid database ID: %w", err)
	}

	if req.PageSize == 0 {
		req.PageSize = 100
	} else if req.PageSize > 100 {
		req.PageSize = 100
	}

	var result QueryDatabaseResponse
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/databases/" + normalizedID + "/query")

	if err != nil {
		return nil, fmt.Errorf("query database request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// CreateDatabase creates a new database
func (c *Client) CreateDatabase(ctx context.Context, req *CreateDatabaseRequest) (*Database, error) {
	c.rateLimiter.Wait()

	var result Database
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/databases")

	if err != nil {
		return nil, fmt.Errorf("create database request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// UpdateDatabase updates a database
func (c *Client) UpdateDatabase(ctx context.Context, databaseID string, req *UpdateDatabaseRequest) (*Database, error) {
	c.rateLimiter.Wait()

	normalizedID, err := NormalizeID(databaseID)
	if err != nil {
		return nil, fmt.Errorf("invalid database ID: %w", err)
	}

	var result Database
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Patch("/databases/" + normalizedID)

	if err != nil {
		return nil, fmt.Errorf("update database request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// GetPage retrieves a page by ID
func (c *Client) GetPage(ctx context.Context, pageID string) (*Page, error) {
	c.rateLimiter.Wait()

	normalizedID, err := NormalizeID(pageID)
	if err != nil {
		return nil, fmt.Errorf("invalid page ID: %w", err)
	}

	var result Page
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/pages/" + normalizedID)

	if err != nil {
		return nil, fmt.Errorf("get page request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// CreatePage creates a new page
func (c *Client) CreatePage(ctx context.Context, req *CreatePageRequest) (*Page, error) {
	c.rateLimiter.Wait()

	var result Page
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/pages")

	if err != nil {
		return nil, fmt.Errorf("create page request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// UpdatePage updates a page
func (c *Client) UpdatePage(ctx context.Context, pageID string, req *UpdatePageRequest) (*Page, error) {
	c.rateLimiter.Wait()

	normalizedID, err := NormalizeID(pageID)
	if err != nil {
		return nil, fmt.Errorf("invalid page ID: %w", err)
	}

	var result Page
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Patch("/pages/" + normalizedID)

	if err != nil {
		return nil, fmt.Errorf("update page request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// GetBlock retrieves a block by ID
func (c *Client) GetBlock(ctx context.Context, blockID string) (*Block, error) {
	c.rateLimiter.Wait()

	normalizedID, err := NormalizeID(blockID)
	if err != nil {
		return nil, fmt.Errorf("invalid block ID: %w", err)
	}

	var result Block
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/blocks/" + normalizedID)

	if err != nil {
		return nil, fmt.Errorf("get block request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// GetBlockChildren retrieves children of a block
func (c *Client) GetBlockChildren(ctx context.Context, blockID string, pageSize int, startCursor string) (*BlockListResponse, error) {
	c.rateLimiter.Wait()

	normalizedID, err := NormalizeID(blockID)
	if err != nil {
		return nil, fmt.Errorf("invalid block ID: %w", err)
	}

	if pageSize == 0 {
		pageSize = 100
	} else if pageSize > 100 {
		pageSize = 100
	}

	req := c.client.R().SetContext(ctx).SetResult(&BlockListResponse{})
	req.SetQueryParam("page_size", fmt.Sprintf("%d", pageSize))
	if startCursor != "" {
		req.SetQueryParam("start_cursor", startCursor)
	}

	var result BlockListResponse
	resp, err := req.
		SetResult(&result).
		Get("/blocks/" + normalizedID + "/children")

	if err != nil {
		return nil, fmt.Errorf("get block children request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// AppendBlockChildren appends children to a block
func (c *Client) AppendBlockChildren(ctx context.Context, blockID string, req *AppendBlockChildrenRequest) (*BlockListResponse, error) {
	c.rateLimiter.Wait()

	normalizedID, err := NormalizeID(blockID)
	if err != nil {
		return nil, fmt.Errorf("invalid block ID: %w", err)
	}

	var result BlockListResponse
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Patch("/blocks/" + normalizedID + "/children")

	if err != nil {
		return nil, fmt.Errorf("append block children request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// UpdateBlock updates a block
func (c *Client) UpdateBlock(ctx context.Context, blockID string, block *Block) (*Block, error) {
	c.rateLimiter.Wait()

	normalizedID, err := NormalizeID(blockID)
	if err != nil {
		return nil, fmt.Errorf("invalid block ID: %w", err)
	}

	var result Block
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(block).
		SetResult(&result).
		Patch("/blocks/" + normalizedID)

	if err != nil {
		return nil, fmt.Errorf("update block request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// DeleteBlock deletes a block
func (c *Client) DeleteBlock(ctx context.Context, blockID string) (*Block, error) {
	c.rateLimiter.Wait()

	normalizedID, err := NormalizeID(blockID)
	if err != nil {
		return nil, fmt.Errorf("invalid block ID: %w", err)
	}

	var result Block
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Delete("/blocks/" + normalizedID)

	if err != nil {
		return nil, fmt.Errorf("delete block request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// GetComments retrieves comments
func (c *Client) GetComments(ctx context.Context, blockID, pageID string, pageSize int, startCursor string) (*CommentListResponse, error) {
	c.rateLimiter.Wait()

	if blockID == "" && pageID == "" {
		return nil, fmt.Errorf("either block_id or page_id must be provided")
	}

	if pageSize == 0 {
		pageSize = 100
	} else if pageSize > 100 {
		pageSize = 100
	}

	req := c.client.R().SetContext(ctx)
	if blockID != "" {
		normalizedID, err := NormalizeID(blockID)
		if err != nil {
			return nil, fmt.Errorf("invalid block ID: %w", err)
		}
		req.SetQueryParam("block_id", normalizedID)
	} else {
		normalizedID, err := NormalizeID(pageID)
		if err != nil {
			return nil, fmt.Errorf("invalid page ID: %w", err)
		}
		req.SetQueryParam("page_id", normalizedID)
	}

	req.SetQueryParam("page_size", fmt.Sprintf("%d", pageSize))
	if startCursor != "" {
		req.SetQueryParam("start_cursor", startCursor)
	}

	var result CommentListResponse
	resp, err := req.
		SetResult(&result).
		Get("/comments")

	if err != nil {
		return nil, fmt.Errorf("get comments request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// CreateComment creates a comment
func (c *Client) CreateComment(ctx context.Context, req *CreateCommentRequest) (*Comment, error) {
	c.rateLimiter.Wait()

	var result Comment
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/comments")

	if err != nil {
		return nil, fmt.Errorf("create comment request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// ListUsers lists all users
func (c *Client) ListUsers(ctx context.Context, pageSize int, startCursor string) (*UserListResponse, error) {
	c.rateLimiter.Wait()

	if pageSize == 0 {
		pageSize = 100
	} else if pageSize > 100 {
		pageSize = 100
	}

	req := c.client.R().SetContext(ctx)
	req.SetQueryParam("page_size", fmt.Sprintf("%d", pageSize))
	if startCursor != "" {
		req.SetQueryParam("start_cursor", startCursor)
	}

	var result UserListResponse
	resp, err := req.
		SetResult(&result).
		Get("/users")

	if err != nil {
		return nil, fmt.Errorf("list users request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// GetUser retrieves a user by ID
func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	c.rateLimiter.Wait()

	normalizedID, err := NormalizeID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var result User
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/users/" + normalizedID)

	if err != nil {
		return nil, fmt.Errorf("get user request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}

// GetBotUser retrieves the current bot user
func (c *Client) GetBotUser(ctx context.Context) (*User, error) {
	c.rateLimiter.Wait()

	var result User
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/users/me")

	if err != nil {
		return nil, fmt.Errorf("get bot user request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleError(resp)
	}

	return &result, nil
}
