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
	// Per-method response and error fields
	searchResp          *notion.SearchResponse
	searchErr           error
	getDatabaseResp     *notion.Database
	getDatabaseErr      error
	queryDatabaseResp   *notion.QueryDatabaseResponse
	queryDatabaseErr    error
	createDatabaseResp  *notion.Database
	createDatabaseErr   error
	updateDatabaseResp  *notion.Database
	updateDatabaseErr   error
	getPageResp         *notion.Page
	getPageErr          error
	createPageResp      *notion.Page
	createPageErr       error
	updatePageResp      *notion.Page
	updatePageErr       error
	movePageResp        *notion.Page
	movePageErr         error
	getBlockResp        *notion.Block
	getBlockErr         error
	blockChildrenResp   *notion.BlockListResponse
	blockChildrenErr    error
	appendBlockResp     *notion.BlockListResponse
	appendBlockErr      error
	updateBlockResp     *notion.Block
	updateBlockErr      error
	deleteBlockResp     *notion.Block
	deleteBlockErr      error
	getCommentsResp     *notion.CommentListResponse
	getCommentsErr      error
	createCommentResp   *notion.Comment
	createCommentErr    error
	listUsersResp       *notion.UserListResponse
	listUsersErr        error
	getUserResp         *notion.User
	getUserErr          error
	getBotUserResp      *notion.User
	getBotUserErr       error

	// Sequential responses for batch operations
	createPageResponses []*notion.Page
	createPageErrors    []error
	createPageCallCount int

	updatePageResponses []*notion.Page
	updatePageErrors    []error
	updatePageCallCount int

	// Fallback generic error (backward compat)
	err error
}

func (m *mockNotionClient) Search(_ context.Context, _ *notion.SearchRequest) (*notion.SearchResponse, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.searchResp, nil
}

func (m *mockNotionClient) GetDatabase(_ context.Context, _ string) (*notion.Database, error) {
	if m.getDatabaseErr != nil {
		return nil, m.getDatabaseErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.getDatabaseResp, nil
}

func (m *mockNotionClient) QueryDatabase(_ context.Context, _ string, _ *notion.QueryDatabaseRequest) (*notion.QueryDatabaseResponse, error) {
	if m.queryDatabaseErr != nil {
		return nil, m.queryDatabaseErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.queryDatabaseResp, nil
}

func (m *mockNotionClient) CreateDatabase(_ context.Context, _ *notion.CreateDatabaseRequest) (*notion.Database, error) {
	if m.createDatabaseErr != nil {
		return nil, m.createDatabaseErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.createDatabaseResp, nil
}

func (m *mockNotionClient) UpdateDatabase(_ context.Context, _ string, _ *notion.UpdateDatabaseRequest) (*notion.Database, error) {
	if m.updateDatabaseErr != nil {
		return nil, m.updateDatabaseErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.updateDatabaseResp, nil
}

func (m *mockNotionClient) GetPage(_ context.Context, _ string) (*notion.Page, error) {
	if m.getPageErr != nil {
		return nil, m.getPageErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.getPageResp, nil
}

func (m *mockNotionClient) CreatePage(_ context.Context, _ *notion.CreatePageRequest) (*notion.Page, error) {
	// Support sequential responses for batch ops
	if len(m.createPageResponses) > 0 || len(m.createPageErrors) > 0 {
		idx := m.createPageCallCount
		m.createPageCallCount++
		var respErr error
		if idx < len(m.createPageErrors) {
			respErr = m.createPageErrors[idx]
		}
		if respErr != nil {
			return nil, respErr
		}
		if idx < len(m.createPageResponses) {
			return m.createPageResponses[idx], nil
		}
		return nil, fmt.Errorf("no more mock responses")
	}
	if m.createPageErr != nil {
		return nil, m.createPageErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.createPageResp, nil
}

func (m *mockNotionClient) UpdatePage(_ context.Context, _ string, _ *notion.UpdatePageRequest) (*notion.Page, error) {
	// Support sequential responses for batch ops
	if len(m.updatePageResponses) > 0 || len(m.updatePageErrors) > 0 {
		idx := m.updatePageCallCount
		m.updatePageCallCount++
		var respErr error
		if idx < len(m.updatePageErrors) {
			respErr = m.updatePageErrors[idx]
		}
		if respErr != nil {
			return nil, respErr
		}
		if idx < len(m.updatePageResponses) {
			return m.updatePageResponses[idx], nil
		}
		return nil, fmt.Errorf("no more mock responses")
	}
	if m.updatePageErr != nil {
		return nil, m.updatePageErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.updatePageResp, nil
}

func (m *mockNotionClient) MovePage(_ context.Context, _ string, _ *notion.MovePageRequest) (*notion.Page, error) {
	if m.movePageErr != nil {
		return nil, m.movePageErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.movePageResp, nil
}

func (m *mockNotionClient) GetBlock(_ context.Context, _ string) (*notion.Block, error) {
	if m.getBlockErr != nil {
		return nil, m.getBlockErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.getBlockResp, nil
}

func (m *mockNotionClient) GetBlockChildren(_ context.Context, _ string, _ int, _ string) (*notion.BlockListResponse, error) {
	if m.blockChildrenErr != nil {
		return nil, m.blockChildrenErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.blockChildrenResp, nil
}

func (m *mockNotionClient) AppendBlockChildren(_ context.Context, _ string, _ *notion.AppendBlockChildrenRequest) (*notion.BlockListResponse, error) {
	if m.appendBlockErr != nil {
		return nil, m.appendBlockErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.appendBlockResp, nil
}

func (m *mockNotionClient) UpdateBlock(_ context.Context, _ string, _ *notion.Block) (*notion.Block, error) {
	if m.updateBlockErr != nil {
		return nil, m.updateBlockErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.updateBlockResp, nil
}

func (m *mockNotionClient) DeleteBlock(_ context.Context, _ string) (*notion.Block, error) {
	if m.deleteBlockErr != nil {
		return nil, m.deleteBlockErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.deleteBlockResp, nil
}

func (m *mockNotionClient) GetComments(_ context.Context, _, _ string, _ int, _ string) (*notion.CommentListResponse, error) {
	if m.getCommentsErr != nil {
		return nil, m.getCommentsErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.getCommentsResp, nil
}

func (m *mockNotionClient) CreateComment(_ context.Context, _ *notion.CreateCommentRequest) (*notion.Comment, error) {
	if m.createCommentErr != nil {
		return nil, m.createCommentErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.createCommentResp, nil
}

func (m *mockNotionClient) ListUsers(_ context.Context, _ int, _ string) (*notion.UserListResponse, error) {
	if m.listUsersErr != nil {
		return nil, m.listUsersErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.listUsersResp, nil
}

func (m *mockNotionClient) GetUser(_ context.Context, _ string) (*notion.User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.getUserResp, nil
}

func (m *mockNotionClient) GetBotUser(_ context.Context) (*notion.User, error) {
	if m.getBotUserErr != nil {
		return nil, m.getBotUserErr
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.getBotUserResp, nil
}

// helper to build a CallToolRequest with given arguments
func makeRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func parseResponse(t *testing.T, result *mcp.CallToolResult) map[string]any {
	t.Helper()
	textContent := result.Content[0].(mcp.TextContent)
	var response map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}
	return response
}

// ==================== Search Tests ====================

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

	response := parseResponse(t, result)
	if response["count"] != float64(2) {
		t.Errorf("expected count 2, got %v", response["count"])
	}
}

func TestSearchHandler_WithFilter(t *testing.T) {
	client := &mockNotionClient{
		searchResp: &notion.SearchResponse{
			Results: []notion.SearchResult{{Object: "page", ID: "p1"}},
		},
	}

	handler := SearchHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"query":  "test",
		"filter": map[string]any{"property": "object", "value": "page"},
		"sort":   map[string]any{"direction": "descending", "timestamp": "last_edited_time"},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestSearchHandler_APIError(t *testing.T) {
	client := &mockNotionClient{
		searchErr: fmt.Errorf("search failed"),
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

func TestListDatabasesHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		searchResp: &notion.SearchResponse{
			Results: []notion.SearchResult{
				{
					Object: "database",
					ID:     "db-1",
					Title:  []notion.RichText{{PlainText: "My DB"}},
				},
			},
		},
	}

	handler := ListDatabasesHandler(client)
	result, err := handler(context.Background(), makeRequest(nil))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["count"] != float64(1) {
		t.Errorf("expected count 1, got %v", response["count"])
	}
}

func TestListDatabasesHandler_Error(t *testing.T) {
	client := &mockNotionClient{searchErr: fmt.Errorf("fail")}
	handler := ListDatabasesHandler(client)
	result, err := handler(context.Background(), makeRequest(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
}

func TestGetDatabaseHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		getDatabaseResp: &notion.Database{
			ID:    "db-123",
			URL:   "https://notion.so/db-123",
			Title: []notion.RichText{{PlainText: "Test DB"}},
			Properties: map[string]notion.PropertyDef{
				"Name": {Type: "title"},
			},
		},
	}

	handler := GetDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["id"] != "db-123" {
		t.Errorf("expected id 'db-123', got %v", response["id"])
	}
}

func TestGetDatabaseHandler_MissingID(t *testing.T) {
	client := &mockNotionClient{}
	handler := GetDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing database_id")
	}
}

func TestGetDatabaseHandler_Error(t *testing.T) {
	client := &mockNotionClient{getDatabaseErr: fmt.Errorf("not found")}
	handler := GetDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
}

// ==================== Page Tests ====================

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

	response := parseResponse(t, result)
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
	client := &mockNotionClient{getPageErr: fmt.Errorf("API connection failed")}
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
}

func TestGetPageContentHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		blockChildrenResp: &notion.BlockListResponse{
			Results: []notion.Block{
				{ID: "b1", Type: "paragraph"},
			},
		},
	}

	handler := GetPageContentHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["count"] != float64(1) {
		t.Errorf("expected count 1, got %v", response["count"])
	}
}

func TestGetPageContentHandler_MissingPageID(t *testing.T) {
	client := &mockNotionClient{}
	handler := GetPageContentHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing page_id")
	}
}

func TestGetPageContentHandler_Error(t *testing.T) {
	client := &mockNotionClient{blockChildrenErr: fmt.Errorf("fail")}
	handler := GetPageContentHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestCreatePageHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		createPageResp: &notion.Page{
			ID:  "new-page",
			URL: "https://notion.so/new-page",
		},
	}

	handler := CreatePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent":     map[string]any{"page_id": "parent-123"},
		"properties": map[string]any{"title": map[string]any{"title": []any{}}},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["success"] != true {
		t.Error("expected success true")
	}
}

func TestCreatePageHandler_MissingParent(t *testing.T) {
	client := &mockNotionClient{}
	handler := CreatePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"properties": map[string]any{"title": "test"},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing parent")
	}
}

func TestCreatePageHandler_MissingProperties(t *testing.T) {
	client := &mockNotionClient{}
	handler := CreatePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent": map[string]any{"page_id": "p1"},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing properties")
	}
}

func TestCreatePageHandler_InvalidParent(t *testing.T) {
	client := &mockNotionClient{}
	handler := CreatePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent":     map[string]any{"workspace": true},
		"properties": map[string]any{"title": "test"},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for invalid parent")
	}
}

func TestCreatePageHandler_WithDatabaseParent(t *testing.T) {
	client := &mockNotionClient{
		createPageResp: &notion.Page{ID: "new", URL: "https://notion.so/new"},
	}

	handler := CreatePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent":     map[string]any{"database_id": "db-123"},
		"properties": map[string]any{"Name": map[string]any{"title": []any{}}},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestCreatePageHandler_Error(t *testing.T) {
	client := &mockNotionClient{createPageErr: fmt.Errorf("fail")}
	handler := CreatePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent":     map[string]any{"page_id": "p1"},
		"properties": map[string]any{"title": "test"},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestUpdatePageHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		updatePageResp: &notion.Page{
			ID:  "page-123",
			URL: "https://notion.so/page-123",
		},
	}

	handler := UpdatePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id":    "aabbccdd-1122-3344-5566-778899aabbcc",
		"properties": map[string]any{"Status": map[string]any{"select": map[string]any{"name": "Done"}}},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestUpdatePageHandler_MissingPageID(t *testing.T) {
	client := &mockNotionClient{}
	handler := UpdatePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing page_id")
	}
}

func TestUpdatePageHandler_WithArchived(t *testing.T) {
	client := &mockNotionClient{
		updatePageResp: &notion.Page{ID: "p1", URL: "https://notion.so/p1"},
	}

	handler := UpdatePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id":  "aabbccdd-1122-3344-5566-778899aabbcc",
		"archived": true,
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestUpdatePageHandler_Error(t *testing.T) {
	client := &mockNotionClient{updatePageErr: fmt.Errorf("fail")}
	handler := UpdatePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
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

	response := parseResponse(t, result)
	if response["success"] != true {
		t.Errorf("expected success true, got %v", response["success"])
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

func TestDeletePageHandler_Error(t *testing.T) {
	client := &mockNotionClient{updatePageErr: fmt.Errorf("fail")}
	handler := DeletePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestMovePageHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		movePageResp: &notion.Page{
			ID:  "page-123",
			URL: "https://notion.so/page-123",
			Parent: &notion.Parent{
				Type:   "page_id",
				PageID: "new-parent",
			},
		},
	}

	handler := MovePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"parent":  map[string]any{"type": "page_id", "page_id": "new-parent"},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["success"] != true {
		t.Error("expected success true")
	}
}

func TestMovePageHandler_MissingPageID(t *testing.T) {
	client := &mockNotionClient{}
	handler := MovePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent": map[string]any{"type": "page_id", "page_id": "p1"},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing page_id")
	}
}

func TestMovePageHandler_MissingParent(t *testing.T) {
	client := &mockNotionClient{}
	handler := MovePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing parent")
	}
}

func TestMovePageHandler_MissingParentType(t *testing.T) {
	client := &mockNotionClient{}
	handler := MovePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"parent":  map[string]any{"page_id": "p1"},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing parent type")
	}
}

func TestMovePageHandler_InvalidParentType(t *testing.T) {
	client := &mockNotionClient{}
	handler := MovePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"parent":  map[string]any{"type": "workspace"},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for invalid parent type")
	}
}

func TestMovePageHandler_Error(t *testing.T) {
	client := &mockNotionClient{movePageErr: fmt.Errorf("fail")}
	handler := MovePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"parent":  map[string]any{"type": "page_id", "page_id": "p1"},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

// ==================== Database Tests ====================

func TestQueryDatabaseHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		queryDatabaseResp: &notion.QueryDatabaseResponse{
			Results: []notion.Page{{ID: "p1"}, {ID: "p2"}},
		},
	}

	handler := QueryDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["count"] != float64(2) {
		t.Errorf("expected count 2, got %v", response["count"])
	}
}

func TestQueryDatabaseHandler_MissingID(t *testing.T) {
	client := &mockNotionClient{}
	handler := QueryDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing database_id")
	}
}

func TestQueryDatabaseHandler_WithFilterAndSorts(t *testing.T) {
	client := &mockNotionClient{
		queryDatabaseResp: &notion.QueryDatabaseResponse{
			Results: []notion.Page{{ID: "p1"}},
		},
	}

	handler := QueryDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"filter":      map[string]any{"property": "Status", "select": map[string]any{"equals": "Done"}},
		"sorts":       []any{map[string]any{"property": "Name", "direction": "ascending"}},
		"page_size":   float64(50),
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestQueryDatabaseHandler_Error(t *testing.T) {
	client := &mockNotionClient{queryDatabaseErr: fmt.Errorf("fail")}
	handler := QueryDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestCreateDatabaseHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		createDatabaseResp: &notion.Database{
			ID:  "db-new",
			URL: "https://notion.so/db-new",
		},
	}

	handler := CreateDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent": map[string]any{"page_id": "parent-page"},
		"title":  []any{map[string]any{"type": "text", "text": map[string]any{"content": "New DB"}}},
		"properties": map[string]any{
			"Name": map[string]any{"type": "title", "title": map[string]any{}},
		},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["success"] != true {
		t.Error("expected success true")
	}
}

func TestCreateDatabaseHandler_MissingParent(t *testing.T) {
	client := &mockNotionClient{}
	handler := CreateDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"title":      []any{map[string]any{"type": "text"}},
		"properties": map[string]any{"Name": map[string]any{"type": "title"}},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing parent")
	}
}

func TestCreateDatabaseHandler_MissingTitle(t *testing.T) {
	client := &mockNotionClient{}
	handler := CreateDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent":     map[string]any{"page_id": "p1"},
		"properties": map[string]any{"Name": map[string]any{"type": "title"}},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing title")
	}
}

func TestCreateDatabaseHandler_MissingProperties(t *testing.T) {
	client := &mockNotionClient{}
	handler := CreateDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent": map[string]any{"page_id": "p1"},
		"title":  []any{map[string]any{"type": "text"}},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing properties")
	}
}

func TestCreateDatabaseHandler_Error(t *testing.T) {
	client := &mockNotionClient{createDatabaseErr: fmt.Errorf("fail")}
	handler := CreateDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent":     map[string]any{"page_id": "p1"},
		"title":      []any{map[string]any{"type": "text", "text": map[string]any{"content": "DB"}}},
		"properties": map[string]any{"Name": map[string]any{"type": "title", "title": map[string]any{}}},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestUpdateDatabaseHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		updateDatabaseResp: &notion.Database{
			ID:  "db-123",
			URL: "https://notion.so/db-123",
		},
	}

	handler := UpdateDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"title":       []any{map[string]any{"type": "text", "text": map[string]any{"content": "Updated"}}},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestUpdateDatabaseHandler_MissingID(t *testing.T) {
	client := &mockNotionClient{}
	handler := UpdateDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing database_id")
	}
}

func TestUpdateDatabaseHandler_Error(t *testing.T) {
	client := &mockNotionClient{updateDatabaseErr: fmt.Errorf("fail")}
	handler := UpdateDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

// ==================== Block Tests ====================

func TestGetBlockHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		getBlockResp: &notion.Block{ID: "block-1", Type: "paragraph"},
	}

	handler := GetBlockHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"block_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestGetBlockHandler_MissingID(t *testing.T) {
	client := &mockNotionClient{}
	handler := GetBlockHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing block_id")
	}
}

func TestGetBlockHandler_Error(t *testing.T) {
	client := &mockNotionClient{getBlockErr: fmt.Errorf("fail")}
	handler := GetBlockHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"block_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestGetBlockChildrenHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		blockChildrenResp: &notion.BlockListResponse{
			Results: []notion.Block{{ID: "b1"}, {ID: "b2"}},
		},
	}

	handler := GetBlockChildrenHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"block_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["count"] != float64(2) {
		t.Errorf("expected count 2, got %v", response["count"])
	}
}

func TestGetBlockChildrenHandler_MissingID(t *testing.T) {
	client := &mockNotionClient{}
	handler := GetBlockChildrenHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestGetBlockChildrenHandler_Error(t *testing.T) {
	client := &mockNotionClient{blockChildrenErr: fmt.Errorf("fail")}
	handler := GetBlockChildrenHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"block_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestAppendBlocksHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		appendBlockResp: &notion.BlockListResponse{
			Results: []notion.Block{{ID: "new-b1"}},
		},
	}

	handler := AppendBlocksHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"block_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"children": []any{map[string]any{"type": "paragraph", "paragraph": map[string]any{"rich_text": []any{}}}},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestAppendBlocksHandler_MissingBlockID(t *testing.T) {
	client := &mockNotionClient{}
	handler := AppendBlocksHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"children": []any{map[string]any{"type": "paragraph"}},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing block_id")
	}
}

func TestAppendBlocksHandler_MissingChildren(t *testing.T) {
	client := &mockNotionClient{}
	handler := AppendBlocksHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"block_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing children")
	}
}

func TestAppendBlocksHandler_Error(t *testing.T) {
	client := &mockNotionClient{appendBlockErr: fmt.Errorf("fail")}
	handler := AppendBlocksHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"block_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"children": []any{map[string]any{"type": "paragraph", "paragraph": map[string]any{"rich_text": []any{}}}},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestUpdateBlockHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		updateBlockResp: &notion.Block{ID: "b1", Type: "paragraph"},
	}

	handler := UpdateBlockHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"block_id":  "aabbccdd-1122-3344-5566-778899aabbcc",
		"paragraph": map[string]any{"rich_text": []any{}},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestUpdateBlockHandler_MissingBlockID(t *testing.T) {
	client := &mockNotionClient{}
	handler := UpdateBlockHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing block_id")
	}
}

func TestUpdateBlockHandler_NoUpdateFields(t *testing.T) {
	client := &mockNotionClient{}
	handler := UpdateBlockHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"block_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for no update fields")
	}
}

func TestUpdateBlockHandler_Error(t *testing.T) {
	client := &mockNotionClient{updateBlockErr: fmt.Errorf("fail")}
	handler := UpdateBlockHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"block_id":  "aabbccdd-1122-3344-5566-778899aabbcc",
		"paragraph": map[string]any{},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestDeleteBlockHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		deleteBlockResp: &notion.Block{ID: "block-456"},
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

	response := parseResponse(t, result)
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

func TestDeleteBlockHandler_Error(t *testing.T) {
	client := &mockNotionClient{deleteBlockErr: fmt.Errorf("fail")}
	handler := DeleteBlockHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"block_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

// ==================== Comment Tests ====================

func TestGetCommentsHandler_WithBlockID(t *testing.T) {
	client := &mockNotionClient{
		getCommentsResp: &notion.CommentListResponse{
			Results: []notion.Comment{{ID: "c1"}},
		},
	}

	handler := GetCommentsHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"block_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["count"] != float64(1) {
		t.Errorf("expected count 1, got %v", response["count"])
	}
}

func TestGetCommentsHandler_WithPageID(t *testing.T) {
	client := &mockNotionClient{
		getCommentsResp: &notion.CommentListResponse{
			Results: []notion.Comment{{ID: "c1"}, {ID: "c2"}},
		},
	}

	handler := GetCommentsHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestGetCommentsHandler_MissingBothIDs(t *testing.T) {
	client := &mockNotionClient{}
	handler := GetCommentsHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing both IDs")
	}
}

func TestGetCommentsHandler_Error(t *testing.T) {
	client := &mockNotionClient{getCommentsErr: fmt.Errorf("fail")}
	handler := GetCommentsHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestCreateCommentHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		createCommentResp: &notion.Comment{
			ID:           "comment-1",
			DiscussionID: "disc-1",
		},
	}

	handler := CreateCommentHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent":    map[string]any{"page_id": "page-1"},
		"rich_text": []any{map[string]any{"type": "text", "text": map[string]any{"content": "Hello"}}},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["success"] != true {
		t.Error("expected success true")
	}
}

func TestCreateCommentHandler_MissingParent(t *testing.T) {
	client := &mockNotionClient{}
	handler := CreateCommentHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"rich_text": []any{map[string]any{"type": "text"}},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing parent")
	}
}

func TestCreateCommentHandler_MissingRichText(t *testing.T) {
	client := &mockNotionClient{}
	handler := CreateCommentHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent": map[string]any{"page_id": "p1"},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing rich_text")
	}
}

func TestCreateCommentHandler_InvalidParent(t *testing.T) {
	client := &mockNotionClient{}
	handler := CreateCommentHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent":    map[string]any{"database_id": "db1"},
		"rich_text": []any{map[string]any{"type": "text"}},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for invalid parent (no page_id)")
	}
}

func TestCreateCommentHandler_WithDiscussionID(t *testing.T) {
	client := &mockNotionClient{
		createCommentResp: &notion.Comment{ID: "c1", DiscussionID: "d1"},
	}

	handler := CreateCommentHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent":        map[string]any{"page_id": "p1"},
		"rich_text":     []any{map[string]any{"type": "text", "text": map[string]any{"content": "reply"}}},
		"discussion_id": "d1",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestCreateCommentHandler_Error(t *testing.T) {
	client := &mockNotionClient{createCommentErr: fmt.Errorf("fail")}
	handler := CreateCommentHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent":    map[string]any{"page_id": "p1"},
		"rich_text": []any{map[string]any{"type": "text", "text": map[string]any{"content": "Hello"}}},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

// ==================== User Tests ====================

func TestListUsersHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		listUsersResp: &notion.UserListResponse{
			Results: []notion.User{
				{ID: "u1", Name: "Alice"},
				{ID: "u2", Name: "Bob"},
			},
		},
	}

	handler := ListUsersHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["count"] != float64(2) {
		t.Errorf("expected count 2, got %v", response["count"])
	}
}

func TestListUsersHandler_WithPagination(t *testing.T) {
	client := &mockNotionClient{
		listUsersResp: &notion.UserListResponse{
			Results: []notion.User{{ID: "u1"}},
		},
	}

	handler := ListUsersHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_size":    float64(10),
		"start_cursor": "cursor-123",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestListUsersHandler_Error(t *testing.T) {
	client := &mockNotionClient{listUsersErr: fmt.Errorf("fail")}
	handler := ListUsersHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestGetUserHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		getUserResp: &notion.User{
			ID:   "user-123",
			Name: "Alice",
			Type: "person",
		},
	}

	handler := GetUserHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"user_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestGetUserHandler_MissingID(t *testing.T) {
	client := &mockNotionClient{}
	handler := GetUserHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing user_id")
	}
}

func TestGetUserHandler_Error(t *testing.T) {
	client := &mockNotionClient{getUserErr: fmt.Errorf("fail")}
	handler := GetUserHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"user_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestGetBotUserHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		getBotUserResp: &notion.User{
			ID:   "bot-123",
			Name: "My Bot",
			Type: "bot",
			Bot: &notion.Bot{
				WorkspaceName: "Test Workspace",
			},
		},
	}

	handler := GetBotUserHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["name"] != "My Bot" {
		t.Errorf("expected name 'My Bot', got %v", response["name"])
	}
}

func TestGetBotUserHandler_Error(t *testing.T) {
	client := &mockNotionClient{getBotUserErr: fmt.Errorf("fail")}
	handler := GetBotUserHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

// ==================== Helper Tests ====================

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

	response := parseResponse(t, result)
	if response["id"] != "new-page-123" {
		t.Errorf("expected id 'new-page-123', got %v", response["id"])
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

func TestCreateSimplePageHandler_WithDatabaseParent(t *testing.T) {
	client := &mockNotionClient{
		createPageResp: &notion.Page{ID: "p1", URL: "https://notion.so/p1"},
	}

	handler := CreateSimplePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"title":              "Test",
		"parent_database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestCreateSimplePageHandler_WorkspaceDefault(t *testing.T) {
	client := &mockNotionClient{
		createPageResp: &notion.Page{ID: "p1", URL: "https://notion.so/p1"},
	}

	handler := CreateSimplePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"title": "Workspace Page",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestCreateSimplePageHandler_APIError(t *testing.T) {
	client := &mockNotionClient{createPageErr: fmt.Errorf("creation failed")}
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

func TestAppendToPageHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		appendBlockResp: &notion.BlockListResponse{
			Results: []notion.Block{{ID: "b1"}, {ID: "b2"}},
		},
	}

	handler := AppendToPageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"content": "# New Section\nSome text",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["success"] != true {
		t.Error("expected success true")
	}
}

func TestAppendToPageHandler_MissingPageID(t *testing.T) {
	client := &mockNotionClient{}
	handler := AppendToPageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"content": "text",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing page_id")
	}
}

func TestAppendToPageHandler_MissingContent(t *testing.T) {
	client := &mockNotionClient{}
	handler := AppendToPageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing content")
	}
}

func TestAppendToPageHandler_Error(t *testing.T) {
	client := &mockNotionClient{appendBlockErr: fmt.Errorf("fail")}
	handler := AppendToPageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"content": "Some text",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestSearchInDatabaseHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		queryDatabaseResp: &notion.QueryDatabaseResponse{
			Results: []notion.Page{
				{
					ID: "p1",
					Properties: map[string]any{
						"Name": map[string]any{
							"title": []any{
								map[string]any{"plain_text": "Test Page"},
							},
						},
					},
				},
			},
		},
	}

	handler := SearchInDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"query":       "Test",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestSearchInDatabaseHandler_MissingDatabaseID(t *testing.T) {
	client := &mockNotionClient{}
	handler := SearchInDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"query": "test",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing database_id")
	}
}

func TestSearchInDatabaseHandler_MissingQuery(t *testing.T) {
	client := &mockNotionClient{}
	handler := SearchInDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing query")
	}
}

func TestSearchInDatabaseHandler_WithPropertyFilters(t *testing.T) {
	client := &mockNotionClient{
		queryDatabaseResp: &notion.QueryDatabaseResponse{
			Results: []notion.Page{{ID: "p1", Properties: map[string]any{}}},
		},
	}

	handler := SearchInDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id":      "aabbccdd-1122-3344-5566-778899aabbcc",
		"query":            "test",
		"property_filters": map[string]any{"Status": "Done", "Active": true, "Priority": float64(1)},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestSearchInDatabaseHandler_Error(t *testing.T) {
	client := &mockNotionClient{queryDatabaseErr: fmt.Errorf("fail")}
	handler := SearchInDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"query":       "test",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

// ==================== Batch Tests ====================

func TestBatchCreatePagesHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		createPageResponses: []*notion.Page{
			{ID: "p1", URL: "https://notion.so/p1"},
			{ID: "p2", URL: "https://notion.so/p2"},
		},
		createPageErrors: []error{nil, nil},
	}

	handler := BatchCreatePagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pages": []any{
			map[string]any{
				"parent":     map[string]any{"page_id": "parent-1"},
				"properties": map[string]any{"title": map[string]any{}},
			},
			map[string]any{
				"parent":     map[string]any{"page_id": "parent-1"},
				"properties": map[string]any{"title": map[string]any{}},
			},
		},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	summary := response["summary"].(map[string]any)
	if summary["succeeded"] != float64(2) {
		t.Errorf("expected 2 succeeded, got %v", summary["succeeded"])
	}
}

func TestBatchCreatePagesHandler_MissingPages(t *testing.T) {
	client := &mockNotionClient{}
	handler := BatchCreatePagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing pages")
	}
}

func TestBatchCreatePagesHandler_PartialFailureStops(t *testing.T) {
	client := &mockNotionClient{
		createPageResponses: []*notion.Page{{ID: "p1", URL: "u1"}},
		createPageErrors:    []error{nil, fmt.Errorf("fail")},
	}

	handler := BatchCreatePagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pages": []any{
			map[string]any{"parent": map[string]any{"page_id": "p1"}, "properties": map[string]any{"t": "v"}},
			map[string]any{"parent": map[string]any{"page_id": "p1"}, "properties": map[string]any{"t": "v"}},
			map[string]any{"parent": map[string]any{"page_id": "p1"}, "properties": map[string]any{"t": "v"}},
		},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	response := parseResponse(t, result)
	summary := response["summary"].(map[string]any)
	if summary["succeeded"] != float64(1) {
		t.Errorf("expected 1 succeeded, got %v", summary["succeeded"])
	}
	if summary["failed"] != float64(1) {
		t.Errorf("expected 1 failed, got %v", summary["failed"])
	}
}

func TestBatchCreatePagesHandler_ContinueOnError(t *testing.T) {
	client := &mockNotionClient{
		createPageResponses: []*notion.Page{{ID: "p1", URL: "u1"}, nil, {ID: "p3", URL: "u3"}},
		createPageErrors:    []error{nil, fmt.Errorf("fail"), nil},
	}

	handler := BatchCreatePagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pages": []any{
			map[string]any{"parent": map[string]any{"page_id": "p1"}, "properties": map[string]any{"t": "v"}},
			map[string]any{"parent": map[string]any{"page_id": "p1"}, "properties": map[string]any{"t": "v"}},
			map[string]any{"parent": map[string]any{"page_id": "p1"}, "properties": map[string]any{"t": "v"}},
		},
		"continue_on_error": true,
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	response := parseResponse(t, result)
	summary := response["summary"].(map[string]any)
	if summary["succeeded"] != float64(2) {
		t.Errorf("expected 2 succeeded, got %v", summary["succeeded"])
	}
	if summary["failed"] != float64(1) {
		t.Errorf("expected 1 failed, got %v", summary["failed"])
	}
}

func TestBatchCreatePagesHandler_InvalidPageFormat(t *testing.T) {
	client := &mockNotionClient{}
	handler := BatchCreatePagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"pages":             []any{"not-a-map"},
		"continue_on_error": true,
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	response := parseResponse(t, result)
	summary := response["summary"].(map[string]any)
	if summary["failed"] != float64(1) {
		t.Errorf("expected 1 failed, got %v", summary["failed"])
	}
}

func TestBatchUpdatePagesHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		updatePageResponses: []*notion.Page{
			{ID: "p1", URL: "u1"},
			{ID: "p2", URL: "u2"},
		},
		updatePageErrors: []error{nil, nil},
	}

	handler := BatchUpdatePagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_ids":   []any{"aabbccdd-1122-3344-5566-778899aabbcc", "aabbccdd-1122-3344-5566-778899aabbdd"},
		"properties": map[string]any{"Status": map[string]any{"select": map[string]any{"name": "Done"}}},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestBatchUpdatePagesHandler_MissingPageIDs(t *testing.T) {
	client := &mockNotionClient{}
	handler := BatchUpdatePagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing page_ids")
	}
}

func TestBatchUpdatePagesHandler_EmptyStringIDs(t *testing.T) {
	client := &mockNotionClient{}
	handler := BatchUpdatePagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_ids": []any{"", ""},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for empty page IDs")
	}
}

func TestBatchDeletePagesHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		updatePageResponses: []*notion.Page{{ID: "p1"}, {ID: "p2"}},
		updatePageErrors:    []error{nil, nil},
	}

	handler := BatchDeletePagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_ids": []any{"aabbccdd-1122-3344-5566-778899aabbcc", "aabbccdd-1122-3344-5566-778899aabbdd"},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestBatchDeletePagesHandler_MissingPageIDs(t *testing.T) {
	client := &mockNotionClient{}
	handler := BatchDeletePagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing page_ids")
	}
}

func TestBatchDeletePagesHandler_ContinueOnError(t *testing.T) {
	client := &mockNotionClient{
		updatePageResponses: []*notion.Page{{ID: "p1"}, nil, {ID: "p3"}},
		updatePageErrors:    []error{nil, fmt.Errorf("fail"), nil},
	}

	handler := BatchDeletePagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_ids":          []any{"p1", "p2", "p3"},
		"continue_on_error": true,
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	response := parseResponse(t, result)
	summary := response["summary"].(map[string]any)
	if summary["succeeded"] != float64(2) {
		t.Errorf("expected 2 succeeded, got %v", summary["succeeded"])
	}
}

// ==================== Template Tests ====================

func TestCreatePageFromTemplateHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		createPageResp: &notion.Page{
			ID:  "template-page",
			URL: "https://notion.so/template-page",
		},
	}

	handler := CreatePageFromTemplateHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"template_name":  "meeting_notes",
		"title":          "Team Sync",
		"parent_page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["template"] != "meeting_notes" {
		t.Errorf("expected template 'meeting_notes', got %v", response["template"])
	}
}

func TestCreatePageFromTemplateHandler_MissingTemplateName(t *testing.T) {
	client := &mockNotionClient{}
	handler := CreatePageFromTemplateHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"title": "Test",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing template_name")
	}
}

func TestCreatePageFromTemplateHandler_MissingTitle(t *testing.T) {
	client := &mockNotionClient{}
	handler := CreatePageFromTemplateHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"template_name": "meeting_notes",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing title")
	}
}

func TestCreatePageFromTemplateHandler_InvalidTemplate(t *testing.T) {
	client := &mockNotionClient{}
	handler := CreatePageFromTemplateHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"template_name": "nonexistent",
		"title":         "Test",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for invalid template")
	}
}

func TestCreatePageFromTemplateHandler_WithVariables(t *testing.T) {
	client := &mockNotionClient{
		createPageResp: &notion.Page{ID: "p1", URL: "https://notion.so/p1"},
	}

	handler := CreatePageFromTemplateHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"template_name":  "project_brief",
		"title":          "{project_name} Brief",
		"parent_page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
		"variables":      map[string]any{"project_name": "Alpha"},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestCreatePageFromTemplateHandler_WithDatabaseParent(t *testing.T) {
	client := &mockNotionClient{
		createPageResp: &notion.Page{ID: "p1", URL: "https://notion.so/p1"},
	}

	handler := CreatePageFromTemplateHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"template_name":      "daily_log",
		"title":              "Log {date}",
		"parent_database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestCreatePageFromTemplateHandler_AllTemplates(t *testing.T) {
	templates := []string{"meeting_notes", "daily_log", "project_brief", "sprint_planning", "retrospective"}
	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			client := &mockNotionClient{
				createPageResp: &notion.Page{ID: "p1", URL: "https://notion.so/p1"},
			}
			handler := CreatePageFromTemplateHandler(client)
			result, err := handler(context.Background(), makeRequest(map[string]any{
				"template_name":  tmpl,
				"title":          "Test",
				"parent_page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
			}))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.IsError {
				t.Fatalf("expected success for template %s", tmpl)
			}
		})
	}
}

func TestCreatePageFromTemplateHandler_Error(t *testing.T) {
	client := &mockNotionClient{createPageErr: fmt.Errorf("fail")}
	handler := CreatePageFromTemplateHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"template_name":  "meeting_notes",
		"title":          "Test",
		"parent_page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestListTemplatesHandler_Success(t *testing.T) {
	client := &mockNotionClient{}

	handler := ListTemplatesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["count"] != float64(5) {
		t.Errorf("expected 5 templates, got %v", response["count"])
	}
}

// ==================== Export Tests ====================

func TestExportPageAsMarkdownHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		getPageResp: &notion.Page{
			ID:  "page-1",
			URL: "https://notion.so/page-1",
			Properties: map[string]any{
				"title": map[string]any{
					"title": []any{
						map[string]any{"plain_text": "Test Page"},
					},
				},
			},
		},
		blockChildrenResp: &notion.BlockListResponse{
			Results: []notion.Block{
				{
					Type: "paragraph",
					Paragraph: &notion.RichTextBlock{
						RichText: notion.NewRichText("Hello world"),
					},
				},
			},
		},
	}

	handler := ExportPageAsMarkdownHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	md, ok := response["markdown"].(string)
	if !ok {
		t.Fatal("expected markdown field")
	}
	if md == "" {
		t.Error("expected non-empty markdown")
	}
}

func TestExportPageAsMarkdownHandler_MissingPageID(t *testing.T) {
	client := &mockNotionClient{}
	handler := ExportPageAsMarkdownHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing page_id")
	}
}

func TestExportPageAsMarkdownHandler_WithFrontmatter(t *testing.T) {
	client := &mockNotionClient{
		getPageResp: &notion.Page{
			ID:         "page-1",
			URL:        "https://notion.so/page-1",
			Properties: map[string]any{"title": map[string]any{"title": []any{map[string]any{"plain_text": "Test"}}}},
		},
		blockChildrenResp: &notion.BlockListResponse{Results: []notion.Block{}},
	}

	handler := ExportPageAsMarkdownHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id":     "aabbccdd-1122-3344-5566-778899aabbcc",
		"frontmatter": true,
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestExportPageAsMarkdownHandler_GetPageError(t *testing.T) {
	client := &mockNotionClient{getPageErr: fmt.Errorf("fail")}
	handler := ExportPageAsMarkdownHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestExportPageAsMarkdownHandler_GetBlocksError(t *testing.T) {
	client := &mockNotionClient{
		getPageResp:      &notion.Page{ID: "p1", Properties: map[string]any{}},
		blockChildrenErr: fmt.Errorf("fail"),
	}
	handler := ExportPageAsMarkdownHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestExportDatabaseAsCSVHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		getDatabaseResp: &notion.Database{
			ID: "db-1",
			Properties: map[string]notion.PropertyDef{
				"Name":   {Type: "title"},
				"Status": {Type: "select"},
			},
		},
		queryDatabaseResp: &notion.QueryDatabaseResponse{
			Results: []notion.Page{
				{
					ID: "p1",
					Properties: map[string]any{
						"Name":   map[string]any{"title": []any{map[string]any{"plain_text": "Page 1"}}},
						"Status": map[string]any{"select": map[string]any{"name": "Done"}},
					},
				},
			},
		},
	}

	handler := ExportDatabaseAsCSVHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	response := parseResponse(t, result)
	if response["row_count"] != float64(1) {
		t.Errorf("expected row_count 1, got %v", response["row_count"])
	}
}

func TestExportDatabaseAsCSVHandler_MissingID(t *testing.T) {
	client := &mockNotionClient{}
	handler := ExportDatabaseAsCSVHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing database_id")
	}
}

func TestExportDatabaseAsCSVHandler_GetDatabaseError(t *testing.T) {
	client := &mockNotionClient{getDatabaseErr: fmt.Errorf("fail")}
	handler := ExportDatabaseAsCSVHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestExportDatabaseAsCSVHandler_QueryError(t *testing.T) {
	client := &mockNotionClient{
		getDatabaseResp:  &notion.Database{ID: "db-1", Properties: map[string]notion.PropertyDef{}},
		queryDatabaseErr: fmt.Errorf("fail"),
	}
	handler := ExportDatabaseAsCSVHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

// ==================== Smart Query Tests ====================

func TestGetRecentlyEditedHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		searchResp: &notion.SearchResponse{
			Results: []notion.SearchResult{
				{Object: "page", ID: "p1"},
			},
		},
	}

	handler := GetRecentlyEditedHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestGetRecentlyEditedHandler_WithParams(t *testing.T) {
	client := &mockNotionClient{
		searchResp: &notion.SearchResponse{Results: []notion.SearchResult{}},
	}

	handler := GetRecentlyEditedHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"days":        float64(3),
		"limit":       float64(10),
		"database_id": "db-1",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestGetRecentlyEditedHandler_Error(t *testing.T) {
	client := &mockNotionClient{searchErr: fmt.Errorf("fail")}
	handler := GetRecentlyEditedHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

func TestGetRelatedPagesHandler_Success(t *testing.T) {
	client := &mockNotionClient{
		getPageResp: &notion.Page{
			ID: "page-1",
			Properties: map[string]any{
				"Related": map[string]any{
					"relation": []any{
						map[string]any{"id": "rel-1"},
					},
				},
			},
		},
	}

	handler := GetRelatedPagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestGetRelatedPagesHandler_MissingPageID(t *testing.T) {
	client := &mockNotionClient{}
	handler := GetRelatedPagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing page_id")
	}
}

func TestGetRelatedPagesHandler_WithPropertyFilter(t *testing.T) {
	client := &mockNotionClient{
		getPageResp: &notion.Page{
			ID: "p1",
			Properties: map[string]any{
				"Tasks": map[string]any{
					"relation": []any{map[string]any{"id": "t1"}},
				},
				"Docs": map[string]any{
					"relation": []any{map[string]any{"id": "d1"}},
				},
			},
		},
	}

	handler := GetRelatedPagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id":           "aabbccdd-1122-3344-5566-778899aabbcc",
		"relation_property": "Tasks",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestGetRelatedPagesHandler_Error(t *testing.T) {
	client := &mockNotionClient{getPageErr: fmt.Errorf("fail")}
	handler := GetRelatedPagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "aabbccdd-1122-3344-5566-778899aabbcc",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error")
	}
}

// ==================== Export Helper Tests ====================

func TestExtractPageTitle(t *testing.T) {
	tests := []struct {
		name  string
		props map[string]any
		want  string
	}{
		{
			name:  "title property",
			props: map[string]any{"title": map[string]any{"title": []any{map[string]any{"plain_text": "Hello"}}}},
			want:  "Hello",
		},
		{
			name:  "Name property",
			props: map[string]any{"Name": map[string]any{"title": []any{map[string]any{"plain_text": "World"}}}},
			want:  "World",
		},
		{
			name:  "no title",
			props: map[string]any{},
			want:  "Untitled",
		},
		{
			name:  "empty title array",
			props: map[string]any{"title": map[string]any{"title": []any{}}},
			want:  "Untitled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPageTitle(tt.props)
			if got != tt.want {
				t.Errorf("extractPageTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractPropertyValue(t *testing.T) {
	tests := []struct {
		name string
		prop any
		want string
	}{
		{"nil", nil, ""},
		{"not a map", "string", ""},
		{
			"title",
			map[string]any{"title": []any{map[string]any{"plain_text": "Hello"}}},
			"Hello",
		},
		{
			"rich_text",
			map[string]any{"rich_text": []any{map[string]any{"plain_text": "World"}}},
			"World",
		},
		{
			"number",
			map[string]any{"number": float64(42)},
			"42",
		},
		{
			"select",
			map[string]any{"select": map[string]any{"name": "Done"}},
			"Done",
		},
		{
			"multi_select",
			map[string]any{"multi_select": []any{map[string]any{"name": "A"}, map[string]any{"name": "B"}}},
			"A; B",
		},
		{
			"date start only",
			map[string]any{"date": map[string]any{"start": "2024-01-01"}},
			"2024-01-01",
		},
		{
			"date range",
			map[string]any{"date": map[string]any{"start": "2024-01-01", "end": "2024-01-31"}},
			"2024-01-01 → 2024-01-31",
		},
		{
			"checkbox true",
			map[string]any{"checkbox": true},
			"true",
		},
		{
			"checkbox false",
			map[string]any{"checkbox": false},
			"false",
		},
		{
			"url",
			map[string]any{"url": "https://example.com"},
			"https://example.com",
		},
		{
			"email",
			map[string]any{"email": "test@example.com"},
			"test@example.com",
		},
		{
			"phone_number",
			map[string]any{"phone_number": "+1234567890"},
			"+1234567890",
		},
		{
			"people",
			map[string]any{"people": []any{map[string]any{"name": "Alice"}, map[string]any{"name": "Bob"}}},
			"Alice; Bob",
		},
		{
			"files",
			map[string]any{"files": []any{map[string]any{"name": "doc.pdf"}}},
			"doc.pdf",
		},
		{
			"created_time",
			map[string]any{"created_time": "2024-01-01T00:00:00Z"},
			"2024-01-01T00:00:00Z",
		},
		{
			"last_edited_time",
			map[string]any{"last_edited_time": "2024-02-01T00:00:00Z"},
			"2024-02-01T00:00:00Z",
		},
		{
			"unknown type",
			map[string]any{"unknown_field": "value"},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPropertyValue(tt.prop)
			if got != tt.want {
				t.Errorf("extractPropertyValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractRichTextArray(t *testing.T) {
	tests := []struct {
		name  string
		input []any
		want  string
	}{
		{"empty", []any{}, ""},
		{"single", []any{map[string]any{"plain_text": "Hello"}}, "Hello"},
		{"multiple", []any{map[string]any{"plain_text": "Hello "}, map[string]any{"plain_text": "World"}}, "Hello World"},
		{"non-map item", []any{"not-a-map"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractRichTextArray(tt.input)
			if got != tt.want {
				t.Errorf("extractRichTextArray() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ==================== Template Helper Tests ====================

func TestGetTemplates(t *testing.T) {
	templates := GetTemplates()
	expectedNames := []string{"meeting_notes", "daily_log", "project_brief", "sprint_planning", "retrospective"}

	for _, name := range expectedNames {
		if _, ok := templates[name]; !ok {
			t.Errorf("expected template %q to exist", name)
		}
	}

	if len(templates) != len(expectedNames) {
		t.Errorf("expected %d templates, got %d", len(expectedNames), len(templates))
	}
}

func TestReplaceVariables(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		variables map[string]string
		contains  []string
	}{
		{
			name:      "custom variable",
			text:      "Project: {project_name}",
			variables: map[string]string{"project_name": "Alpha"},
			contains:  []string{"Project: Alpha"},
		},
		{
			name:      "no variables",
			text:      "No vars here",
			variables: map[string]string{},
			contains:  []string{"No vars here"},
		},
		{
			name:      "date variable present",
			text:      "Log {date}",
			variables: map[string]string{},
			contains:  []string{"Log "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceVariables(tt.text, tt.variables)
			for _, c := range tt.contains {
				if !containsIgnoreCase(got, c) {
					t.Errorf("replaceVariables() = %q, expected to contain %q", got, c)
				}
			}
		})
	}
}

func TestGetBlockStructurePreview(t *testing.T) {
	tests := []struct {
		name  string
		block notion.Block
		want  string
	}{
		{
			name:  "heading_2",
			block: notion.NewHeadingBlock(2, "Section"),
			want:  "## Section",
		},
		{
			name:  "paragraph",
			block: notion.NewParagraphBlock("text"),
			want:  "Paragraph: text",
		},
		{
			name:  "bullet",
			block: notion.NewBulletedListItemBlock("item"),
			want:  "• item",
		},
		{
			name:  "todo unchecked",
			block: notion.NewToDoBlock("task", false),
			want:  "☐ task",
		},
		{
			name:  "todo checked",
			block: notion.NewToDoBlock("done", true),
			want:  "☑ done",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getBlockStructurePreview(tt.block)
			if got != tt.want {
				t.Errorf("getBlockStructurePreview() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTemplateNames(t *testing.T) {
	names := getTemplateNames()
	if len(names) != 5 {
		t.Errorf("expected 5 template names, got %d", len(names))
	}
}

// ==================== Smart Query Helper Tests ====================

func TestFindPropertyName(t *testing.T) {
	props := map[string]notion.PropertyDef{
		"Status":      {Type: "select"},
		"Assigned To": {Type: "people"},
		"Due Date":    {Type: "date"},
	}

	tests := []struct {
		name          string
		possibleNames []string
		want          string
	}{
		{"found first", []string{"Status", "State"}, "Status"},
		{"found second", []string{"State", "Status"}, "Status"},
		{"not found", []string{"Priority", "Level"}, ""},
		{"assignee", []string{"Assigned To", "Assignee"}, "Assigned To"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findPropertyName(props, tt.possibleNames)
			if got != tt.want {
				t.Errorf("findPropertyName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ==================== buildCreatePageRequest Tests ====================

func TestBuildCreatePageRequest_Success(t *testing.T) {
	pageMap := map[string]any{
		"parent":     map[string]any{"page_id": "p1"},
		"properties": map[string]any{"title": map[string]any{}},
	}

	req, err := buildCreatePageRequest(pageMap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Parent.PageID != "p1" {
		t.Errorf("expected parent page_id 'p1', got %q", req.Parent.PageID)
	}
}

func TestBuildCreatePageRequest_DatabaseParent(t *testing.T) {
	pageMap := map[string]any{
		"parent":     map[string]any{"database_id": "db1"},
		"properties": map[string]any{"Name": map[string]any{}},
	}

	req, err := buildCreatePageRequest(pageMap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Parent.DatabaseID != "db1" {
		t.Errorf("expected parent database_id 'db1', got %q", req.Parent.DatabaseID)
	}
}

func TestBuildCreatePageRequest_MissingParent(t *testing.T) {
	pageMap := map[string]any{
		"properties": map[string]any{"title": map[string]any{}},
	}

	_, err := buildCreatePageRequest(pageMap)
	if err == nil {
		t.Fatal("expected error for missing parent")
	}
}

func TestBuildCreatePageRequest_MissingProperties(t *testing.T) {
	pageMap := map[string]any{
		"parent": map[string]any{"page_id": "p1"},
	}

	_, err := buildCreatePageRequest(pageMap)
	if err == nil {
		t.Fatal("expected error for missing properties")
	}
}

func TestBuildCreatePageRequest_InvalidParent(t *testing.T) {
	pageMap := map[string]any{
		"parent":     map[string]any{"workspace": true},
		"properties": map[string]any{"title": map[string]any{}},
	}

	_, err := buildCreatePageRequest(pageMap)
	if err == nil {
		t.Fatal("expected error for invalid parent")
	}
}

// ==================== GetMyTasks Tests ====================

func TestGetMyTasksHandler_SuccessWithDatabaseID(t *testing.T) {
	client := &mockNotionClient{
		getBotUserResp: &notion.User{
			Object: "user",
			ID:     "bot-user-123",
			Type:   "bot",
			Name:   "Test Bot",
		},
		getDatabaseResp: &notion.Database{
			Properties: map[string]notion.PropertyDef{
				"Name":        {Type: "title"},
				"Status":      {Type: "select"},
				"Assigned To": {Type: "people"},
				"Due Date":    {Type: "date"},
			},
		},
		queryDatabaseResp: &notion.QueryDatabaseResponse{
			Results: []notion.Page{
				{
					ID:  "page-1",
					URL: "https://notion.so/page-1",
					Properties: map[string]any{
						"Name": map[string]any{
							"type": "title",
							"title": []any{
								map[string]any{"plain_text": "Task 1"},
							},
						},
						"Status": map[string]any{
							"type": "select",
							"select": map[string]any{
								"name": "In Progress",
							},
						},
					},
				},
				{
					ID:  "page-2",
					URL: "https://notion.so/page-2",
					Properties: map[string]any{
						"Name": map[string]any{
							"type": "title",
							"title": []any{
								map[string]any{"plain_text": "Task 2"},
							},
						},
						"Status": map[string]any{
							"type": "select",
							"select": map[string]any{
								"name": "To Do",
							},
						},
					},
				},
			},
			HasMore: false,
		},
	}

	handler := GetMyTasksHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "db-456",
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success, got error")
	}

	response := parseResponse(t, result)
	if response["count"] != float64(2) {
		t.Errorf("expected count 2, got %v", response["count"])
	}
	if response["database_id"] != "db-456" {
		t.Errorf("expected database_id db-456, got %v", response["database_id"])
	}
	tasks, ok := response["tasks"].([]any)
	if !ok {
		t.Fatal("expected tasks array in response")
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	task1 := tasks[0].(map[string]any)
	if task1["id"] != "page-1" {
		t.Errorf("expected task id page-1, got %v", task1["id"])
	}
	if task1["title"] != "Task 1" {
		t.Errorf("expected task title 'Task 1', got %v", task1["title"])
	}
	if task1["status"] != "In Progress" {
		t.Errorf("expected task status 'In Progress', got %v", task1["status"])
	}
}

func TestGetMyTasksHandler_SuccessAutoFindDatabase(t *testing.T) {
	client := &mockNotionClient{
		getBotUserResp: &notion.User{
			Object: "user",
			ID:     "bot-user-123",
			Type:   "bot",
			Name:   "Test Bot",
		},
		searchResp: &notion.SearchResponse{
			Results: []notion.SearchResult{
				{
					Object: "database",
					ID:     "db-auto-123",
					Title:  notion.NewRichText("My Tasks"),
				},
			},
		},
		getDatabaseResp: &notion.Database{
			Properties: map[string]notion.PropertyDef{
				"Name":        {Type: "title"},
				"Status":      {Type: "select"},
				"Assigned To": {Type: "people"},
				"Due Date":    {Type: "date"},
			},
		},
		queryDatabaseResp: &notion.QueryDatabaseResponse{
			Results: []notion.Page{
				{
					ID:  "page-1",
					URL: "https://notion.so/page-1",
					Properties: map[string]any{
						"Name": map[string]any{
							"type": "title",
							"title": []any{
								map[string]any{"plain_text": "Auto-found Task"},
							},
						},
					},
				},
			},
			HasMore: false,
		},
	}

	handler := GetMyTasksHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success, got error")
	}

	response := parseResponse(t, result)
	if response["database_id"] != "db-auto-123" {
		t.Errorf("expected auto-found database_id db-auto-123, got %v", response["database_id"])
	}
	if response["count"] != float64(1) {
		t.Errorf("expected count 1, got %v", response["count"])
	}
	tasks := response["tasks"].([]any)
	task1 := tasks[0].(map[string]any)
	if task1["title"] != "Auto-found Task" {
		t.Errorf("expected title 'Auto-found Task', got %v", task1["title"])
	}
}

func TestGetMyTasksHandler_GetBotUserError(t *testing.T) {
	client := &mockNotionClient{
		getBotUserErr: fmt.Errorf("bot user auth failed"),
	}

	handler := GetMyTasksHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "db-456",
	}))

	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result")
	}
}

func TestGetMyTasksHandler_NoDatabaseFound(t *testing.T) {
	client := &mockNotionClient{
		getBotUserResp: &notion.User{
			Object: "user",
			ID:     "bot-user-123",
			Type:   "bot",
			Name:   "Test Bot",
		},
		searchResp: &notion.SearchResponse{
			Results: []notion.SearchResult{
				{
					Object: "database",
					ID:     "db-unrelated",
					Title:  notion.NewRichText("Project Tracker"),
				},
			},
		},
	}

	handler := GetMyTasksHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{}))

	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result when no task database found")
	}
}

func TestGetMyTasksHandler_GetDatabaseError(t *testing.T) {
	client := &mockNotionClient{
		getBotUserResp: &notion.User{
			Object: "user",
			ID:     "bot-user-123",
			Type:   "bot",
			Name:   "Test Bot",
		},
		getDatabaseErr: fmt.Errorf("database not found"),
	}

	handler := GetMyTasksHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "db-456",
	}))

	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result when GetDatabase fails")
	}
}

func TestGetMyTasksHandler_QueryDatabaseError(t *testing.T) {
	client := &mockNotionClient{
		getBotUserResp: &notion.User{
			Object: "user",
			ID:     "bot-user-123",
			Type:   "bot",
			Name:   "Test Bot",
		},
		getDatabaseResp: &notion.Database{
			Properties: map[string]notion.PropertyDef{
				"Name":        {Type: "title"},
				"Status":      {Type: "select"},
				"Assigned To": {Type: "people"},
				"Due Date":    {Type: "date"},
			},
		},
		queryDatabaseErr: fmt.Errorf("query failed: rate limited"),
	}

	handler := GetMyTasksHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "db-456",
	}))

	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result when QueryDatabase fails")
	}
}

// ==================== exportBlockChildren Tests ====================

func TestExportBlockChildren_Success(t *testing.T) {
	client := &mockNotionClient{
		blockChildrenResp: &notion.BlockListResponse{
			Results: []notion.Block{
				{
					ID:   "b1",
					Type: "paragraph",
					Paragraph: &notion.RichTextBlock{
						RichText: []notion.RichText{{PlainText: "Hello world"}},
					},
				},
				{
					ID:   "b2",
					Type: "heading_1",
					Heading1: &notion.RichTextBlock{
						RichText: []notion.RichText{{PlainText: "Title"}},
					},
				},
			},
			HasMore: false,
		},
	}

	result, err := exportBlockChildren(context.Background(), client, "page-123", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty markdown output")
	}
}

func TestExportBlockChildren_WithPagination(t *testing.T) {
	// The mock always returns the same response, but we test the pagination code path
	// by setting HasMore=true on the first call
	callCount := 0
	client := &paginatingBlockClient{
		mockNotionClient: mockNotionClient{},
		blockChildrenFn: func(ctx context.Context, blockID string, pageSize int, cursor string) (*notion.BlockListResponse, error) {
			callCount++
			if callCount == 1 {
				return &notion.BlockListResponse{
					Results: []notion.Block{
						{ID: "b1", Type: "paragraph", Paragraph: &notion.RichTextBlock{
							RichText: []notion.RichText{{PlainText: "First page"}},
						}},
					},
					HasMore:    true,
					NextCursor: "cursor-abc",
				}, nil
			}
			return &notion.BlockListResponse{
				Results: []notion.Block{
					{ID: "b2", Type: "paragraph", Paragraph: &notion.RichTextBlock{
						RichText: []notion.RichText{{PlainText: "Second page"}},
					}},
				},
				HasMore: false,
			}, nil
		},
	}

	result, err := exportBlockChildren(context.Background(), client, "page-123", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestExportBlockChildren_WithChildren(t *testing.T) {
	callCount := 0
	client := &paginatingBlockClient{
		mockNotionClient: mockNotionClient{},
		blockChildrenFn: func(ctx context.Context, blockID string, pageSize int, cursor string) (*notion.BlockListResponse, error) {
			callCount++
			if blockID == "page-123" {
				return &notion.BlockListResponse{
					Results: []notion.Block{
						{
							ID:          "b1",
							Type:        "toggle",
							HasChildren: true,
							Toggle: &notion.RichTextBlock{
								RichText: []notion.RichText{{PlainText: "Toggle"}},
							},
						},
					},
					HasMore: false,
				}, nil
			}
			// Child blocks
			return &notion.BlockListResponse{
				Results: []notion.Block{
					{ID: "child-1", Type: "paragraph", Paragraph: &notion.RichTextBlock{
						RichText: []notion.RichText{{PlainText: "Child content"}},
					}},
				},
				HasMore: false,
			}, nil
		},
	}

	result, err := exportBlockChildren(context.Background(), client, "page-123", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty result with children")
	}
}

func TestExportBlockChildren_Error(t *testing.T) {
	client := &mockNotionClient{
		blockChildrenErr: fmt.Errorf("api error"),
	}

	_, err := exportBlockChildren(context.Background(), client, "page-123", 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

// paginatingBlockClient wraps mockNotionClient but overrides GetBlockChildren
// to support per-call behavior for pagination tests.
type paginatingBlockClient struct {
	mockNotionClient
	blockChildrenFn func(ctx context.Context, blockID string, pageSize int, cursor string) (*notion.BlockListResponse, error)
}

func (p *paginatingBlockClient) GetBlockChildren(ctx context.Context, blockID string, pageSize int, cursor string) (*notion.BlockListResponse, error) {
	if p.blockChildrenFn != nil {
		return p.blockChildrenFn(ctx, blockID, pageSize, cursor)
	}
	return p.mockNotionClient.GetBlockChildren(ctx, blockID, pageSize, cursor)
}

// ==================== Additional branch coverage tests ====================

func TestCreatePageHandler_WithChildren(t *testing.T) {
	client := &mockNotionClient{
		createPageResp: &notion.Page{ID: "new-page", URL: "https://notion.so/new-page"},
	}

	handler := CreatePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent":     map[string]any{"page_id": "parent-123"},
		"properties": map[string]any{"title": map[string]any{"title": []map[string]any{{"text": map[string]any{"content": "Test"}}}}},
		"children": []any{
			map[string]any{"type": "paragraph", "paragraph": map[string]any{"rich_text": []map[string]any{{"text": map[string]any{"content": "Hello"}}}}},
		},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestCreatePageHandler_WithIconAndCover(t *testing.T) {
	client := &mockNotionClient{
		createPageResp: &notion.Page{ID: "new-page", URL: "https://notion.so/new-page"},
	}

	handler := CreatePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent":     map[string]any{"page_id": "parent-123"},
		"properties": map[string]any{"title": "Test"},
		"icon":       map[string]any{"type": "emoji", "emoji": "📝"},
		"cover":      map[string]any{"type": "external", "external": map[string]any{"url": "https://example.com/img.png"}},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestUpdatePageHandler_WithIconAndCover(t *testing.T) {
	client := &mockNotionClient{
		updatePageResp: &notion.Page{ID: "page-123", URL: "https://notion.so/page-123"},
	}

	handler := UpdatePageHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_id": "page-123",
		"icon":    map[string]any{"type": "emoji", "emoji": "🔥"},
		"cover":   map[string]any{"type": "external", "external": map[string]any{"url": "https://example.com/cover.png"}},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestUpdateDatabaseHandler_WithTitleAndDescription(t *testing.T) {
	client := &mockNotionClient{
		updateDatabaseResp: &notion.Database{ID: "db-123", URL: "https://notion.so/db-123"},
	}

	handler := UpdateDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"database_id": "db-123",
		"title":       []any{map[string]any{"text": map[string]any{"content": "New Title"}}},
		"description": []any{map[string]any{"text": map[string]any{"content": "New Description"}}},
		"properties":  map[string]any{"Status": map[string]any{"select": map[string]any{}}},
		"archived":    true,
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestCreateDatabaseHandler_WithIconAndCover(t *testing.T) {
	client := &mockNotionClient{
		createDatabaseResp: &notion.Database{ID: "db-new", URL: "https://notion.so/db-new"},
	}

	handler := CreateDatabaseHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"parent":     map[string]any{"page_id": "parent-123"},
		"title":      []any{map[string]any{"text": map[string]any{"content": "My DB"}}},
		"properties": map[string]any{"Name": map[string]any{"title": map[string]any{}}},
		"icon":       map[string]any{"type": "emoji", "emoji": "📊"},
		"cover":      map[string]any{"type": "external", "external": map[string]any{"url": "https://example.com/cover.png"}},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestBatchUpdatePagesHandler_WithProperties(t *testing.T) {
	client := &mockNotionClient{
		updatePageResponses: []*notion.Page{
			{ID: "p1", URL: "https://notion.so/p1"},
			{ID: "p2", URL: "https://notion.so/p2"},
		},
		updatePageErrors: []error{nil, nil},
	}

	handler := BatchUpdatePagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_ids":   []any{"p1", "p2"},
		"properties": map[string]any{"Status": map[string]any{"select": map[string]any{"name": "Done"}}},
		"archived":   true,
		"icon":       map[string]any{"type": "emoji", "emoji": "✅"},
		"cover":      map[string]any{"type": "external", "external": map[string]any{"url": "https://example.com/cover.png"}},
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	resp := parseResponse(t, result)
	summary := resp["summary"].(map[string]any)
	if int(summary["succeeded"].(float64)) != 2 {
		t.Errorf("expected 2 succeeded, got %v", summary["succeeded"])
	}
}

func TestBatchUpdatePagesHandler_ContinueOnError(t *testing.T) {
	client := &mockNotionClient{
		updatePageResponses: []*notion.Page{
			nil,
			{ID: "p2", URL: "https://notion.so/p2"},
		},
		updatePageErrors: []error{
			fmt.Errorf("update failed"),
			nil,
		},
	}

	handler := BatchUpdatePagesHandler(client)
	result, err := handler(context.Background(), makeRequest(map[string]any{
		"page_ids":          []any{"p1", "p2"},
		"continue_on_error": true,
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success result (partial)")
	}
	resp := parseResponse(t, result)
	summary := resp["summary"].(map[string]any)
	if int(summary["succeeded"].(float64)) != 1 {
		t.Errorf("expected 1 succeeded, got %v", summary["succeeded"])
	}
	if int(summary["failed"].(float64)) != 1 {
		t.Errorf("expected 1 failed, got %v", summary["failed"])
	}
}
