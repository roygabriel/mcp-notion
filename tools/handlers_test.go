package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rgabriel/mcp-notion/notion"
)

// mockNotionClient implements notion.NotionClient for testing.
type mockNotionClient struct {
	// Configurable return values
	searchResp       *notion.SearchResponse
	getPageResp      *notion.Page
	createPageResp   *notion.Page
	updatePageResp   *notion.Page
	deleteBlockResp  *notion.Block
	blockChildrenRes *notion.BlockListResponse
	err              error
}

func (m *mockNotionClient) Search(_ context.Context, _ *notion.SearchRequest) (*notion.SearchResponse, error) {
	return m.searchResp, m.err
}

func (m *mockNotionClient) GetDatabase(_ context.Context, _ string) (*notion.Database, error) {
	return nil, m.err
}

func (m *mockNotionClient) QueryDatabase(_ context.Context, _ string, _ *notion.QueryDatabaseRequest) (*notion.QueryDatabaseResponse, error) {
	return nil, m.err
}

func (m *mockNotionClient) CreateDatabase(_ context.Context, _ *notion.CreateDatabaseRequest) (*notion.Database, error) {
	return nil, m.err
}

func (m *mockNotionClient) UpdateDatabase(_ context.Context, _ string, _ *notion.UpdateDatabaseRequest) (*notion.Database, error) {
	return nil, m.err
}

func (m *mockNotionClient) GetPage(_ context.Context, _ string) (*notion.Page, error) {
	return m.getPageResp, m.err
}

func (m *mockNotionClient) CreatePage(_ context.Context, _ *notion.CreatePageRequest) (*notion.Page, error) {
	return m.createPageResp, m.err
}

func (m *mockNotionClient) UpdatePage(_ context.Context, _ string, _ *notion.UpdatePageRequest) (*notion.Page, error) {
	return m.updatePageResp, m.err
}

func (m *mockNotionClient) MovePage(_ context.Context, _ string, _ *notion.MovePageRequest) (*notion.Page, error) {
	return nil, m.err
}

func (m *mockNotionClient) GetBlock(_ context.Context, _ string) (*notion.Block, error) {
	return nil, m.err
}

func (m *mockNotionClient) GetBlockChildren(_ context.Context, _ string, _ int, _ string) (*notion.BlockListResponse, error) {
	return m.blockChildrenRes, m.err
}

func (m *mockNotionClient) AppendBlockChildren(_ context.Context, _ string, _ *notion.AppendBlockChildrenRequest) (*notion.BlockListResponse, error) {
	return m.blockChildrenRes, m.err
}

func (m *mockNotionClient) UpdateBlock(_ context.Context, _ string, _ *notion.Block) (*notion.Block, error) {
	return nil, m.err
}

func (m *mockNotionClient) DeleteBlock(_ context.Context, _ string) (*notion.Block, error) {
	return m.deleteBlockResp, m.err
}

func (m *mockNotionClient) GetComments(_ context.Context, _, _ string, _ int, _ string) (*notion.CommentListResponse, error) {
	return nil, m.err
}

func (m *mockNotionClient) CreateComment(_ context.Context, _ *notion.CreateCommentRequest) (*notion.Comment, error) {
	return nil, m.err
}

func (m *mockNotionClient) ListUsers(_ context.Context, _ int, _ string) (*notion.UserListResponse, error) {
	return nil, m.err
}

func (m *mockNotionClient) GetUser(_ context.Context, _ string) (*notion.User, error) {
	return nil, m.err
}

func (m *mockNotionClient) GetBotUser(_ context.Context) (*notion.User, error) {
	return nil, m.err
}

// helper to build a CallToolRequest with given arguments
func makeRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func TestGetPageHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		getPageResp: &notion.Page{
			ID:  "page-123",
			URL: "https://notion.so/page-123",
			Properties: map[string]any{
				"title": map[string]any{"type": "title"},
			},
		},
	}

	handler := GetPageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Parse the JSON response
	textContent := result.Content[0].(mcp.TextContent)
	var response map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response["id"] != "page-123" {
		t.Errorf("expected id 'page-123', got %v", response["id"])
	}
}

func TestGetPageHandler_MissingPageID(t *testing.T) {
	client := &mockNotionClient{}
	handler := GetPageHandler(client)

	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing page_id")
	}
}

func TestGetPageHandler_APIError(t *testing.T) {
	client := &mockNotionClient{
		err: fmt.Errorf("API connection failed"),
	}
	handler := GetPageHandler(client)

	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for API failure")
	}
	textContent := result.Content[0].(mcp.TextContent)
	if textContent.Text == "" {
		t.Error("expected non-empty error message")
	}
}

func TestSearchHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		searchResp: &notion.SearchResponse{
			Results: []notion.SearchResult{
				{Object: "page", ID: "page-1"},
				{Object: "database", ID: "db-1"},
			},
			HasMore: false,
		},
	}

	handler := SearchHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"query": "test",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error")
	}

	textContent := result.Content[0].(mcp.TextContent)
	var response map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response["count"] != float64(2) {
		t.Errorf("expected count 2, got %v", response["count"])
	}
}

func TestSearchHandler_APIError(t *testing.T) {
	client := &mockNotionClient{
		err: fmt.Errorf("search failed"),
	}
	handler := SearchHandler(client)

	result, err := handler(context.Background(), makeRequest(map[string]any{
		"query": "test",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
}

func TestDeletePageHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		updatePageResp: &notion.Page{
			ID:       "page-123",
			Archived: true,
		},
	}

	handler := DeletePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error")
	}

	textContent := result.Content[0].(mcp.TextContent)
	var response map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response["success"] != true {
		t.Errorf("expected success true, got %v", response["success"])
	}
	if response["id"] != "page-123" {
		t.Errorf("expected id 'page-123', got %v", response["id"])
	}
}

func TestDeletePageHandler_MissingPageID(t *testing.T) {
	client := &mockNotionClient{}
	handler := DeletePageHandler(client)

	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing page_id")
	}
}

func TestCreateSimplePageHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		createPageResp: &notion.Page{
			ID:  "new-page-123",
			URL: "https://notion.so/new-page-123",
		},
	}

	handler := CreateSimplePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"title":          "Test Page",
		"parent_page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"content":        "# Hello\nSome content here",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error")
	}

	textContent := result.Content[0].(mcp.TextContent)
	var response map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response["id"] != "new-page-123" {
		t.Errorf("expected id 'new-page-123', got %v", response["id"])
	}
	if response["success"] != true {
		t.Errorf("expected success true, got %v", response["success"])
	}
}

func TestCreateSimplePageHandler_MissingTitle(t *testing.T) {
	client := &mockNotionClient{}
	handler := CreateSimplePageHandler(client)

	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing title")
	}
}

func TestCreateSimplePageHandler_APIError(t *testing.T) {
	client := &mockNotionClient{
		err: fmt.Errorf("creation failed"),
	}
	handler := CreateSimplePageHandler(client)

	result, err := handler(context.Background(), makeRequest(map[string]any{
		"title":          "Test",
		"parent_page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for API failure")
	}
}

func TestDeleteBlockHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		deleteBlockResp: &notion.Block{
			ID: "block-456",
		},
	}

	handler := DeleteBlockHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"block_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error")
	}

	textContent := result.Content[0].(mcp.TextContent)
	var response map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if response["id"] != "block-456" {
		t.Errorf("expected id 'block-456', got %v", response["id"])
	}
}

func TestDeleteBlockHandler_MissingBlockID(t *testing.T) {
	client := &mockNotionClient{}
	handler := DeleteBlockHandler(client)

	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing block_id")
	}
}
