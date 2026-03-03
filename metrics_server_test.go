package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMetricsServerEndpoints(t *testing.T) {
	// Find a free port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ensure at least one metric label set exists so the counter appears in output.
	toolCallsTotal.WithLabelValues("_test", "success").Inc()

	go startMetricsServer(ctx, addr)
	time.Sleep(100 * time.Millisecond)

	client := &http.Client{Timeout: 2 * time.Second}

	t.Run("metrics endpoint", func(t *testing.T) {
		resp, err := client.Get(fmt.Sprintf("http://%s/metrics", addr))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		body := string(bodyBytes)
		if !strings.Contains(body, "mcp_tool_calls_total") {
			t.Error("expected mcp_tool_calls_total in metrics output")
		}
	})

	t.Run("healthz endpoint", func(t *testing.T) {
		resp, err := client.Get(fmt.Sprintf("http://%s/healthz", addr))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if result["status"] != "ok" {
			t.Errorf("expected status=ok, got %q", result["status"])
		}
	})
}
