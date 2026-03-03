package main

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestToolMiddlewareRecordsMetrics(t *testing.T) {
	toolCallsTotal.Reset()
	toolCallDuration.Reset()

	mw := observabilityMiddleware()
	handler := mw(func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{}, nil
	})

	req := mcp.CallToolRequest{}
	req.Params.Name = "search"

	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count := testutil.ToFloat64(toolCallsTotal.WithLabelValues("search", "success"))
	if count != 1 {
		t.Errorf("expected counter=1, got %f", count)
	}

	histCount := testutil.CollectAndCount(toolCallDuration)
	if histCount == 0 {
		t.Error("expected histogram metrics, got 0")
	}
}

func TestToolMiddlewareRecordsToolError(t *testing.T) {
	toolCallsTotal.Reset()

	mw := observabilityMiddleware()
	handler := mw(func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{IsError: true}, nil
	})

	req := mcp.CallToolRequest{}
	req.Params.Name = "create_page"

	_, _ = handler(context.Background(), req)

	count := testutil.ToFloat64(toolCallsTotal.WithLabelValues("create_page", "tool_error"))
	if count != 1 {
		t.Errorf("expected tool_error counter=1, got %f", count)
	}
}

func TestToolMiddlewareRecordsError(t *testing.T) {
	toolCallsTotal.Reset()

	mw := observabilityMiddleware()
	handler := mw(func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, context.DeadlineExceeded
	})

	req := mcp.CallToolRequest{}
	req.Params.Name = "get_page"

	_, _ = handler(context.Background(), req)

	count := testutil.ToFloat64(toolCallsTotal.WithLabelValues("get_page", "error"))
	if count != 1 {
		t.Errorf("expected error counter=1, got %f", count)
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"DEBUG", "DEBUG"},
		{"debug", "DEBUG"},
		{"WARN", "WARN"},
		{"ERROR", "ERROR"},
		{"INFO", "INFO"},
		{"", "INFO"},
		{"unknown", "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseLogLevel(tt.input)
			if got.String() != tt.want {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMetricsServerStartsAndStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		startMetricsServer(ctx, ":0")
		close(done)
	}()

	// Give the server a moment to start.
	time.Sleep(50 * time.Millisecond)

	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("metrics server did not shut down in time")
	}
}
