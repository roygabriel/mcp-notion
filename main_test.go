package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestConcurrencyMiddleware_AllowsUpToMax(t *testing.T) {
	const maxConcurrent = 3
	middleware := concurrencyMiddleware(maxConcurrent)
	handler := middleware(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Simulate work
		time.Sleep(50 * time.Millisecond)
		return mcp.NewToolResultText("ok"), nil
	})

	var wg sync.WaitGroup
	results := make(chan error, maxConcurrent)

	for i := 0; i < maxConcurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := handler(context.Background(), mcp.CallToolRequest{})
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	for err := range results {
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	}
}

func TestConcurrencyMiddleware_BlocksBeyondMax(t *testing.T) {
	const maxConcurrent = 2
	middleware := concurrencyMiddleware(maxConcurrent)

	// This channel keeps goroutines blocked inside the handler until we release them.
	block := make(chan struct{})

	handler := middleware(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		<-block
		return mcp.NewToolResultText("ok"), nil
	})

	var wg sync.WaitGroup

	// Launch maxConcurrent goroutines that will occupy all semaphore slots.
	for i := 0; i < maxConcurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler(context.Background(), mcp.CallToolRequest{})
		}()
	}

	// Give goroutines time to acquire semaphore slots.
	time.Sleep(50 * time.Millisecond)

	// Launch one extra goroutine that should be blocked.
	extraDone := make(chan struct{})
	go func() {
		handler(context.Background(), mcp.CallToolRequest{})
		close(extraDone)
	}()

	// Verify the extra goroutine is still blocked.
	select {
	case <-extraDone:
		t.Fatal("expected extra goroutine to be blocked, but it completed")
	case <-time.After(100 * time.Millisecond):
		// Good: extra goroutine is blocked as expected.
	}

	// Release all blocked goroutines.
	close(block)

	// Wait for the extra goroutine to complete.
	select {
	case <-extraDone:
		// Good: extra goroutine completed after release.
	case <-time.After(2 * time.Second):
		t.Fatal("extra goroutine did not complete after release")
	}

	wg.Wait()
}

func TestConcurrencyMiddleware_ContextCancellation(t *testing.T) {
	const maxConcurrent = 1
	middleware := concurrencyMiddleware(maxConcurrent)

	// Block the single slot permanently.
	block := make(chan struct{})
	handler := middleware(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		<-block
		return mcp.NewToolResultText("ok"), nil
	})

	// Occupy the single slot.
	go func() {
		handler(context.Background(), mcp.CallToolRequest{})
	}()
	time.Sleep(50 * time.Millisecond)

	// Create a cancelled context and try to call the handler.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	result, err := handler(ctx, mcp.CallToolRequest{})
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got: %v", result)
	}

	close(block)
}

func TestObservabilityMiddleware_LogsToolCall(t *testing.T) {
	middleware := observabilityMiddleware()
	called := false
	handler := middleware(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		called = true
		return mcp.NewToolResultText("success"), nil
	})

	req := mcp.CallToolRequest{}
	req.Params.Name = "search"

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected next handler to be called")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestObservabilityMiddleware_AuditDestructive(t *testing.T) {
	middleware := observabilityMiddleware()
	called := false
	handler := middleware(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		called = true
		return mcp.NewToolResultText("deleted"), nil
	})

	req := mcp.CallToolRequest{}
	req.Params.Name = "delete_page"
	req.Params.Arguments = map[string]any{"page_id": "abc123"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected next handler to be called for destructive tool")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestDestructiveToolsMap(t *testing.T) {
	expectedTools := []string{
		"delete_page",
		"delete_block",
		"update_page",
		"update_block",
		"move_page",
		"create_page",
		"create_database",
		"update_database",
		"append_blocks",
		"create_comment",
		"batch_create_pages",
		"batch_update_pages",
		"batch_delete_pages",
		"create_simple_page",
		"append_to_page",
		"create_page_from_template",
	}

	for _, tool := range expectedTools {
		if !destructiveTools[tool] {
			t.Errorf("expected %q to be in destructiveTools map", tool)
		}
	}

	// Verify a read-only tool is NOT in the map.
	readOnlyTools := []string{"search", "get_page", "list_users", "get_block"}
	for _, tool := range readOnlyTools {
		if destructiveTools[tool] {
			t.Errorf("expected %q to NOT be in destructiveTools map", tool)
		}
	}
}

// Verify concurrencyMiddleware conforms to the ToolHandlerMiddleware type.
var _ server.ToolHandlerMiddleware = concurrencyMiddleware(1)
