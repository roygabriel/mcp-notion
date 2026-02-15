package notion

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

func TestNormalizeID_WithDashes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "already formatted UUID",
			input: "12345678-1234-5678-9012-123456789012",
			want:  "12345678-1234-5678-9012-123456789012",
		},
		{
			name:  "UUID without dashes",
			input: "12345678123456789012123456789012",
			want:  "12345678-1234-5678-9012-123456789012",
		},
		{
			name:    "too short",
			input:   "12345",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   "12345678123456789012123456789012extra",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid hex characters",
			input:   "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NormalizeID(%q) expected error, got %q", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("NormalizeID(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("NormalizeID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	rl := NewRateLimiter(3)

	// First 3 calls should be fast
	start := time.Now()
	rl.Wait()
	rl.Wait()
	rl.Wait()
	elapsed := time.Since(start)

	if elapsed > 100*time.Millisecond {
		t.Errorf("first 3 calls should be fast, took %v", elapsed)
	}

	// 4th call should block
	start = time.Now()
	rl.Wait()
	elapsed = time.Since(start)

	if elapsed < 500*time.Millisecond {
		t.Errorf("4th call should block for ~1s, took %v", elapsed)
	}
}

func TestNotionError_Error(t *testing.T) {
	err := &NotionError{
		Object:  "error",
		Status:  404,
		Code:    "object_not_found",
		Message: "Could not find page",
	}

	if err.Error() != "Could not find page" {
		t.Errorf("Error() = %q, want %q", err.Error(), "Could not find page")
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

const validUUID = "12345678-1234-5678-9012-123456789012"

func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c := &Client{
		client:         resty.New(),
		apiVersion:     "2022-06-28",
		rateLimiter:    NewRateLimiter(1000),
		circuitBreaker: NewCircuitBreaker(),
	}
	c.client.SetBaseURL(server.URL)
	// Disable retries so error tests are fast
	c.client.SetRetryCount(0)
	return c
}

// notionErrorHandler returns an HTTP handler that responds with a NotionError JSON body.
func notionErrorHandler(status int, code, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(NotionError{
			Object:  "error",
			Status:  status,
			Code:    code,
			Message: message,
		})
	}
}

// ---------------------------------------------------------------------------
// newHTTPTransport
// ---------------------------------------------------------------------------

func TestNewHTTPTransport(t *testing.T) {
	tr := newHTTPTransport()
	if tr == nil {
		t.Fatal("newHTTPTransport returned nil")
	}
	if tr.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig is nil")
	}
	if tr.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %d, want %d", tr.TLSClientConfig.MinVersion, tls.VersionTLS12)
	}
	if tr.MaxIdleConns != 100 {
		t.Errorf("MaxIdleConns = %d, want 100", tr.MaxIdleConns)
	}
	if tr.MaxIdleConnsPerHost != 10 {
		t.Errorf("MaxIdleConnsPerHost = %d, want 10", tr.MaxIdleConnsPerHost)
	}
	if tr.TLSHandshakeTimeout != 5*time.Second {
		t.Errorf("TLSHandshakeTimeout = %v, want 5s", tr.TLSHandshakeTimeout)
	}
	if tr.IdleConnTimeout != 90*time.Second {
		t.Errorf("IdleConnTimeout = %v, want 90s", tr.IdleConnTimeout)
	}
}

// ---------------------------------------------------------------------------
// handleError
// ---------------------------------------------------------------------------

func TestHandleError_NotionErrorJSON(t *testing.T) {
	server := httptest.NewServer(notionErrorHandler(400, "validation_error", "Invalid request"))
	defer server.Close()

	c := &Client{client: resty.New(), rateLimiter: NewRateLimiter(1000), circuitBreaker: NewCircuitBreaker()}
	c.client.SetBaseURL(server.URL)
	resp, err := c.client.R().Get("/test")
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}

	got := c.handleError(resp)
	var ne *NotionError
	if !errors.As(got, &ne) {
		t.Fatalf("expected *NotionError, got %T: %v", got, got)
	}
	if ne.Code != "validation_error" {
		t.Errorf("Code = %q, want %q", ne.Code, "validation_error")
	}
}

func TestHandleError_ObjectNotFound(t *testing.T) {
	server := httptest.NewServer(notionErrorHandler(404, "object_not_found", "Not found"))
	defer server.Close()

	c := &Client{client: resty.New(), rateLimiter: NewRateLimiter(1000), circuitBreaker: NewCircuitBreaker()}
	c.client.SetBaseURL(server.URL)
	resp, _ := c.client.R().Get("/test")

	got := c.handleError(resp)
	if !strings.Contains(got.Error(), "notion.so/my-integrations") {
		t.Errorf("object_not_found should mention integration sharing, got: %v", got)
	}
}

func TestHandleError_NonJSONBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		_, _ = fmt.Fprint(w, "Bad Gateway")
	}))
	defer server.Close()

	c := &Client{client: resty.New(), rateLimiter: NewRateLimiter(1000), circuitBreaker: NewCircuitBreaker()}
	c.client.SetBaseURL(server.URL)
	resp, _ := c.client.R().Get("/test")

	got := c.handleError(resp)
	if !strings.Contains(got.Error(), "502") {
		t.Errorf("expected status code in fallback error, got: %v", got)
	}
}

// ---------------------------------------------------------------------------
// checkCircuitBreaker
// ---------------------------------------------------------------------------

func TestCheckCircuitBreaker_Open(t *testing.T) {
	c := &Client{
		client:         resty.New(),
		rateLimiter:    NewRateLimiter(1000),
		circuitBreaker: NewCircuitBreaker(),
	}
	// Trip the breaker by recording enough failures
	for i := 0; i < 5; i++ {
		c.circuitBreaker.RecordFailure()
	}

	err := c.checkCircuitBreaker()
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCheckCircuitBreaker_Closed(t *testing.T) {
	c := &Client{
		client:         resty.New(),
		rateLimiter:    NewRateLimiter(1000),
		circuitBreaker: NewCircuitBreaker(),
	}
	if err := c.checkCircuitBreaker(); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// isCircuitBreakerFailure
// ---------------------------------------------------------------------------

func TestIsCircuitBreakerFailure(t *testing.T) {
	t.Run("transport error returns true", func(t *testing.T) {
		if !isCircuitBreakerFailure(nil, errors.New("connection refused")) {
			t.Error("expected true for transport error")
		}
	})

	t.Run("5xx returns true", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
		defer server.Close()
		resp, _ := resty.New().R().Get(server.URL)
		if !isCircuitBreakerFailure(resp, nil) {
			t.Error("expected true for 500")
		}
	})

	t.Run("4xx returns false", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
		}))
		defer server.Close()
		resp, _ := resty.New().R().Get(server.URL)
		if isCircuitBreakerFailure(resp, nil) {
			t.Error("expected false for 404")
		}
	})

	t.Run("2xx returns false", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		}))
		defer server.Close()
		resp, _ := resty.New().R().Get(server.URL)
		if isCircuitBreakerFailure(resp, nil) {
			t.Error("expected false for 200")
		}
	})
}

// ---------------------------------------------------------------------------
// recordResult
// ---------------------------------------------------------------------------

func TestRecordResult(t *testing.T) {
	t.Run("success resets breaker", func(t *testing.T) {
		c := &Client{circuitBreaker: NewCircuitBreaker()}
		c.circuitBreaker.RecordFailure() // one failure
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		}))
		defer server.Close()
		resp, _ := resty.New().R().Get(server.URL)
		c.recordResult(resp, nil)
		if c.circuitBreaker.State() != CircuitClosed {
			t.Error("expected circuit closed after success")
		}
	})

	t.Run("5xx records failure", func(t *testing.T) {
		c := &Client{circuitBreaker: NewCircuitBreaker()}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(503)
		}))
		defer server.Close()
		resp, _ := resty.New().R().Get(server.URL)
		// Record 5 failures to open breaker
		for i := 0; i < 5; i++ {
			c.recordResult(resp, nil)
		}
		if c.circuitBreaker.State() != CircuitOpen {
			t.Errorf("expected circuit open after 5 failures, got %d", c.circuitBreaker.State())
		}
	})
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

func TestSearch_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SearchResponse{
			Object:  "list",
			Results: []SearchResult{{Object: "page", ID: validUUID}},
		})
	})
	c := newTestClient(t, mux)
	resp, err := c.Search(context.Background(), &SearchRequest{Query: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Object != "list" {
		t.Errorf("Object = %q, want %q", resp.Object, "list")
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
}

func TestSearch_PageSizeClamping(t *testing.T) {
	tests := []struct {
		name     string
		pageSize int
		want     int
	}{
		{"zero defaults to 100", 0, 100},
		{"over 100 clamped to 100", 200, 100},
		{"normal preserved", 50, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPageSize int
			mux := http.NewServeMux()
			mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
				var req SearchRequest
				_ = json.NewDecoder(r.Body).Decode(&req)
				gotPageSize = req.PageSize
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(SearchResponse{Object: "list"})
			})
			c := newTestClient(t, mux)
			_, _ = c.Search(context.Background(), &SearchRequest{PageSize: tt.pageSize})
			if gotPageSize != tt.want {
				t.Errorf("PageSize = %d, want %d", gotPageSize, tt.want)
			}
		})
	}
}

func TestSearch_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/search", notionErrorHandler(400, "validation_error", "bad request"))
	c := newTestClient(t, mux)
	_, err := c.Search(context.Background(), &SearchRequest{Query: "test"})
	if err == nil {
		t.Fatal("expected error")
	}
	var ne *NotionError
	if !errors.As(err, &ne) {
		t.Fatalf("expected *NotionError, got %T: %v", err, err)
	}
}

func TestSearch_CircuitOpen(t *testing.T) {
	c := &Client{
		client:         resty.New(),
		rateLimiter:    NewRateLimiter(1000),
		circuitBreaker: NewCircuitBreaker(),
	}
	for i := 0; i < 5; i++ {
		c.circuitBreaker.RecordFailure()
	}
	_, err := c.Search(context.Background(), &SearchRequest{Query: "test"})
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// GET-by-ID methods (table-driven): GetDatabase, GetPage, GetBlock, GetUser
// ---------------------------------------------------------------------------

func TestGetByID_Success(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		call     func(c *Client, ctx context.Context) (any, error)
		wantObj  string
	}{
		{
			name: "GetDatabase",
			path: "/databases/" + validUUID,
			call: func(c *Client, ctx context.Context) (any, error) {
				return c.GetDatabase(ctx, validUUID)
			},
			wantObj: "database",
		},
		{
			name: "GetPage",
			path: "/pages/" + validUUID,
			call: func(c *Client, ctx context.Context) (any, error) {
				return c.GetPage(ctx, validUUID)
			},
			wantObj: "page",
		},
		{
			name: "GetBlock",
			path: "/blocks/" + validUUID,
			call: func(c *Client, ctx context.Context) (any, error) {
				return c.GetBlock(ctx, validUUID)
			},
			wantObj: "block",
		},
		{
			name: "GetUser",
			path: "/users/" + validUUID,
			call: func(c *Client, ctx context.Context) (any, error) {
				return c.GetUser(ctx, validUUID)
			},
			wantObj: "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc(tt.path, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprintf(w, `{"object":%q,"id":%q}`, tt.wantObj, validUUID)
			})
			c := newTestClient(t, mux)
			result, err := tt.call(c, context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("result is nil")
			}
		})
	}
}

func TestGetByID_InvalidID(t *testing.T) {
	tests := []struct {
		name string
		call func(c *Client, ctx context.Context) (any, error)
	}{
		{"GetDatabase", func(c *Client, ctx context.Context) (any, error) { return c.GetDatabase(ctx, "bad") }},
		{"GetPage", func(c *Client, ctx context.Context) (any, error) { return c.GetPage(ctx, "bad") }},
		{"GetBlock", func(c *Client, ctx context.Context) (any, error) { return c.GetBlock(ctx, "bad") }},
		{"GetUser", func(c *Client, ctx context.Context) (any, error) { return c.GetUser(ctx, "bad") }},
	}
	c := newTestClient(t, http.NewServeMux())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.call(c, context.Background())
			if err == nil {
				t.Fatal("expected error for invalid ID")
			}
			if !strings.Contains(err.Error(), "invalid") {
				t.Errorf("expected 'invalid' in error, got: %v", err)
			}
		})
	}
}

func TestGetByID_APIError(t *testing.T) {
	tests := []struct {
		name string
		path string
		call func(c *Client, ctx context.Context) (any, error)
	}{
		{
			name: "GetDatabase",
			path: "/databases/" + validUUID,
			call: func(c *Client, ctx context.Context) (any, error) { return c.GetDatabase(ctx, validUUID) },
		},
		{
			name: "GetPage",
			path: "/pages/" + validUUID,
			call: func(c *Client, ctx context.Context) (any, error) { return c.GetPage(ctx, validUUID) },
		},
		{
			name: "GetBlock",
			path: "/blocks/" + validUUID,
			call: func(c *Client, ctx context.Context) (any, error) { return c.GetBlock(ctx, validUUID) },
		},
		{
			name: "GetUser",
			path: "/users/" + validUUID,
			call: func(c *Client, ctx context.Context) (any, error) { return c.GetUser(ctx, validUUID) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc(tt.path, notionErrorHandler(404, "object_not_found", "Not found"))
			c := newTestClient(t, mux)
			_, err := tt.call(c, context.Background())
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// QueryDatabase
// ---------------------------------------------------------------------------

func TestQueryDatabase_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/databases/"+validUUID+"/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(QueryDatabaseResponse{
			Object:  "list",
			Results: []Page{{Object: "page", ID: validUUID}},
		})
	})
	c := newTestClient(t, mux)
	resp, err := c.QueryDatabase(context.Background(), validUUID, &QueryDatabaseRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Object != "list" || len(resp.Results) != 1 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestQueryDatabase_InvalidID(t *testing.T) {
	c := newTestClient(t, http.NewServeMux())
	_, err := c.QueryDatabase(context.Background(), "bad", &QueryDatabaseRequest{})
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid ID error, got: %v", err)
	}
}

func TestQueryDatabase_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/databases/"+validUUID+"/query", notionErrorHandler(404, "object_not_found", "Not found"))
	c := newTestClient(t, mux)
	_, err := c.QueryDatabase(context.Background(), validUUID, &QueryDatabaseRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// CreateDatabase
// ---------------------------------------------------------------------------

func TestCreateDatabase_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/databases", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Database{Object: "database", ID: validUUID})
	})
	c := newTestClient(t, mux)
	resp, err := c.CreateDatabase(context.Background(), &CreateDatabaseRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Object != "database" {
		t.Errorf("Object = %q, want database", resp.Object)
	}
}

func TestCreateDatabase_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/databases", notionErrorHandler(400, "validation_error", "bad"))
	c := newTestClient(t, mux)
	_, err := c.CreateDatabase(context.Background(), &CreateDatabaseRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// UpdateDatabase
// ---------------------------------------------------------------------------

func TestUpdateDatabase_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/databases/"+validUUID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Database{Object: "database", ID: validUUID})
	})
	c := newTestClient(t, mux)
	resp, err := c.UpdateDatabase(context.Background(), validUUID, &UpdateDatabaseRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Object != "database" {
		t.Errorf("Object = %q, want database", resp.Object)
	}
}

func TestUpdateDatabase_InvalidID(t *testing.T) {
	c := newTestClient(t, http.NewServeMux())
	_, err := c.UpdateDatabase(context.Background(), "bad", &UpdateDatabaseRequest{})
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid ID error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CreatePage
// ---------------------------------------------------------------------------

func TestCreatePage_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/pages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Page{Object: "page", ID: validUUID})
	})
	c := newTestClient(t, mux)
	resp, err := c.CreatePage(context.Background(), &CreatePageRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Object != "page" {
		t.Errorf("Object = %q, want page", resp.Object)
	}
}

func TestCreatePage_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/pages", notionErrorHandler(400, "validation_error", "bad"))
	c := newTestClient(t, mux)
	_, err := c.CreatePage(context.Background(), &CreatePageRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// UpdatePage
// ---------------------------------------------------------------------------

func TestUpdatePage_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/pages/"+validUUID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Page{Object: "page", ID: validUUID})
	})
	c := newTestClient(t, mux)
	resp, err := c.UpdatePage(context.Background(), validUUID, &UpdatePageRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Object != "page" {
		t.Errorf("Object = %q, want page", resp.Object)
	}
}

func TestUpdatePage_InvalidID(t *testing.T) {
	c := newTestClient(t, http.NewServeMux())
	_, err := c.UpdatePage(context.Background(), "bad", &UpdatePageRequest{})
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid ID error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// MovePage
// ---------------------------------------------------------------------------

func TestMovePage_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/pages/"+validUUID+"/move", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Page{Object: "page", ID: validUUID})
	})
	c := newTestClient(t, mux)
	resp, err := c.MovePage(context.Background(), validUUID, &MovePageRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Object != "page" {
		t.Errorf("Object = %q, want page", resp.Object)
	}
}

func TestMovePage_InvalidID(t *testing.T) {
	c := newTestClient(t, http.NewServeMux())
	_, err := c.MovePage(context.Background(), "bad", &MovePageRequest{})
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid ID error, got: %v", err)
	}
}

func TestMovePage_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/pages/"+validUUID+"/move", notionErrorHandler(404, "object_not_found", "Not found"))
	c := newTestClient(t, mux)
	_, err := c.MovePage(context.Background(), validUUID, &MovePageRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// GetBlockChildren
// ---------------------------------------------------------------------------

func TestGetBlockChildren_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/blocks/"+validUUID+"/children", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// Verify query params
		if r.URL.Query().Get("page_size") != "10" {
			t.Errorf("page_size = %q, want 10", r.URL.Query().Get("page_size"))
		}
		if r.URL.Query().Get("start_cursor") != "abc" {
			t.Errorf("start_cursor = %q, want abc", r.URL.Query().Get("start_cursor"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(BlockListResponse{
			Object:  "list",
			Results: []Block{{Object: "block", ID: validUUID, Type: "paragraph"}},
		})
	})
	c := newTestClient(t, mux)
	resp, err := c.GetBlockChildren(context.Background(), validUUID, 10, "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Object != "list" || len(resp.Results) != 1 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetBlockChildren_InvalidID(t *testing.T) {
	c := newTestClient(t, http.NewServeMux())
	_, err := c.GetBlockChildren(context.Background(), "bad", 10, "")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid ID error, got: %v", err)
	}
}

func TestGetBlockChildren_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/blocks/"+validUUID+"/children", notionErrorHandler(404, "object_not_found", "Not found"))
	c := newTestClient(t, mux)
	_, err := c.GetBlockChildren(context.Background(), validUUID, 0, "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// AppendBlockChildren
// ---------------------------------------------------------------------------

func TestAppendBlockChildren_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/blocks/"+validUUID+"/children", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(BlockListResponse{Object: "list", Results: []Block{{Object: "block", ID: validUUID}}})
	})
	c := newTestClient(t, mux)
	resp, err := c.AppendBlockChildren(context.Background(), validUUID, &AppendBlockChildrenRequest{
		Children: []Block{{Type: "paragraph"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Object != "list" {
		t.Errorf("Object = %q, want list", resp.Object)
	}
}

func TestAppendBlockChildren_InvalidID(t *testing.T) {
	c := newTestClient(t, http.NewServeMux())
	_, err := c.AppendBlockChildren(context.Background(), "bad", &AppendBlockChildrenRequest{})
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid ID error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// UpdateBlock
// ---------------------------------------------------------------------------

func TestUpdateBlock_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/blocks/"+validUUID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Block{Object: "block", ID: validUUID, Type: "paragraph"})
	})
	c := newTestClient(t, mux)
	resp, err := c.UpdateBlock(context.Background(), validUUID, &Block{Type: "paragraph"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Object != "block" {
		t.Errorf("Object = %q, want block", resp.Object)
	}
}

func TestUpdateBlock_InvalidID(t *testing.T) {
	c := newTestClient(t, http.NewServeMux())
	_, err := c.UpdateBlock(context.Background(), "bad", &Block{})
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid ID error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// DeleteBlock
// ---------------------------------------------------------------------------

func TestDeleteBlock_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/blocks/"+validUUID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Block{Object: "block", ID: validUUID, Archived: true})
	})
	c := newTestClient(t, mux)
	resp, err := c.DeleteBlock(context.Background(), validUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Archived {
		t.Error("expected Archived=true")
	}
}

func TestDeleteBlock_InvalidID(t *testing.T) {
	c := newTestClient(t, http.NewServeMux())
	_, err := c.DeleteBlock(context.Background(), "bad")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid ID error, got: %v", err)
	}
}

func TestDeleteBlock_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/blocks/"+validUUID, notionErrorHandler(404, "object_not_found", "Not found"))
	c := newTestClient(t, mux)
	_, err := c.DeleteBlock(context.Background(), validUUID)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// GetComments
// ---------------------------------------------------------------------------

func TestGetComments_WithBlockID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("block_id") != validUUID {
			t.Errorf("block_id = %q, want %q", r.URL.Query().Get("block_id"), validUUID)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CommentListResponse{
			Object:  "list",
			Results: []Comment{{Object: "comment", ID: validUUID}},
		})
	})
	c := newTestClient(t, mux)
	resp, err := c.GetComments(context.Background(), validUUID, "", 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(resp.Results))
	}
}

func TestGetComments_WithPageID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page_id") != validUUID {
			t.Errorf("page_id = %q, want %q", r.URL.Query().Get("page_id"), validUUID)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CommentListResponse{Object: "list"})
	})
	c := newTestClient(t, mux)
	_, err := c.GetComments(context.Background(), "", validUUID, 50, "cursor123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetComments_NeitherID(t *testing.T) {
	c := newTestClient(t, http.NewServeMux())
	_, err := c.GetComments(context.Background(), "", "", 0, "")
	if err == nil || !strings.Contains(err.Error(), "either") {
		t.Errorf("expected 'either block_id or page_id' error, got: %v", err)
	}
}

func TestGetComments_InvalidBlockID(t *testing.T) {
	c := newTestClient(t, http.NewServeMux())
	_, err := c.GetComments(context.Background(), "bad", "", 0, "")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid ID error, got: %v", err)
	}
}

func TestGetComments_InvalidPageID(t *testing.T) {
	c := newTestClient(t, http.NewServeMux())
	_, err := c.GetComments(context.Background(), "", "bad", 0, "")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid ID error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CreateComment
// ---------------------------------------------------------------------------

func TestCreateComment_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Comment{Object: "comment", ID: validUUID})
	})
	c := newTestClient(t, mux)
	resp, err := c.CreateComment(context.Background(), &CreateCommentRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Object != "comment" {
		t.Errorf("Object = %q, want comment", resp.Object)
	}
}

func TestCreateComment_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/comments", notionErrorHandler(400, "validation_error", "bad"))
	c := newTestClient(t, mux)
	_, err := c.CreateComment(context.Background(), &CreateCommentRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// ListUsers
// ---------------------------------------------------------------------------

func TestListUsers_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UserListResponse{
			Object:  "list",
			Results: []User{{Object: "user", ID: validUUID, Name: "Alice"}},
		})
	})
	c := newTestClient(t, mux)
	resp, err := c.ListUsers(context.Background(), 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Results) != 1 || resp.Results[0].Name != "Alice" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestListUsers_WithCursor(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("start_cursor") != "cur123" {
			t.Errorf("start_cursor = %q, want cur123", r.URL.Query().Get("start_cursor"))
		}
		if r.URL.Query().Get("page_size") != "25" {
			t.Errorf("page_size = %q, want 25", r.URL.Query().Get("page_size"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(UserListResponse{Object: "list"})
	})
	c := newTestClient(t, mux)
	_, err := c.ListUsers(context.Background(), 25, "cur123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListUsers_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users", notionErrorHandler(403, "restricted_resource", "forbidden"))
	c := newTestClient(t, mux)
	_, err := c.ListUsers(context.Background(), 0, "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// GetBotUser
// ---------------------------------------------------------------------------

func TestGetBotUser_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/me", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(User{Object: "user", ID: validUUID, Type: "bot", Name: "TestBot"})
	})
	c := newTestClient(t, mux)
	resp, err := c.GetBotUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Type != "bot" || resp.Name != "TestBot" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetBotUser_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/me", notionErrorHandler(401, "unauthorized", "bad token"))
	c := newTestClient(t, mux)
	_, err := c.GetBotUser(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
