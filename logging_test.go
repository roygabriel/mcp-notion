package main

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestRedactingHandler_Message(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	handler := newRedactingHandler(inner, "my-secret-token")
	logger := slog.New(handler)

	logger.Info("connecting with token my-secret-token to server")

	output := buf.String()
	if strings.Contains(output, "my-secret-token") {
		t.Errorf("expected secret to be redacted from message, got: %s", output)
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output, got: %s", output)
	}
}

func TestRedactingHandler_StringAttr(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	handler := newRedactingHandler(inner, "secret123")
	logger := slog.New(handler)

	logger.Info("request", "token", "bearer secret123")

	output := buf.String()
	if strings.Contains(output, "secret123") {
		t.Errorf("expected secret to be redacted from attribute value, got: %s", output)
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output, got: %s", output)
	}
}

func TestRedactingHandler_GroupAttr(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	handler := newRedactingHandler(inner, "group-secret")
	logger := slog.New(handler)

	logger.LogAttrs(context.Background(), slog.LevelInfo, "grouped",
		slog.Group("config",
			slog.String("api_key", "group-secret"),
			slog.String("host", "example.com"),
		),
	)

	output := buf.String()
	if strings.Contains(output, "group-secret") {
		t.Errorf("expected secret to be redacted from group attribute, got: %s", output)
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output, got: %s", output)
	}
}

func TestRedactingHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	handler := newRedactingHandler(inner, "preattached-secret")
	logger := slog.New(handler)

	loggerWithAttrs := logger.With("api_key", "preattached-secret")
	loggerWithAttrs.Info("doing work")

	output := buf.String()
	if strings.Contains(output, "preattached-secret") {
		t.Errorf("expected secret to be redacted from WithAttrs attribute, got: %s", output)
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output, got: %s", output)
	}
}

func TestRedactingHandler_EmptySecret(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	handler := newRedactingHandler(inner, "", "real-secret", "")
	logger := slog.New(handler)

	// Should not panic and should still redact the non-empty secret.
	logger.Info("message with real-secret inside")

	output := buf.String()
	if strings.Contains(output, "real-secret") {
		t.Errorf("expected non-empty secret to be redacted, got: %s", output)
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output, got: %s", output)
	}
}

func TestRedactingHandler_NoSecrets(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	handler := newRedactingHandler(inner)
	logger := slog.New(handler)

	logger.Info("hello world", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "hello world") {
		t.Errorf("expected message to pass through unchanged, got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("expected attribute to pass through unchanged, got: %s", output)
	}
}

func TestRedactingHandler_MultipleSecrets(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	handler := newRedactingHandler(inner, "secret-aaa", "secret-bbb", "secret-ccc")
	logger := slog.New(handler)

	logger.Info("using secret-aaa and secret-bbb", "token", "secret-ccc")

	output := buf.String()
	for _, secret := range []string{"secret-aaa", "secret-bbb", "secret-ccc"} {
		if strings.Contains(output, secret) {
			t.Errorf("expected %q to be redacted, got: %s", secret, output)
		}
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output, got: %s", output)
	}
}

func TestRedactingHandler_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	handler := newRedactingHandler(inner, "group-token-xyz")
	logger := slog.New(handler)

	groupedLogger := logger.WithGroup("request")
	groupedLogger.Info("api call", "token", "group-token-xyz")

	output := buf.String()
	if strings.Contains(output, "group-token-xyz") {
		t.Errorf("expected secret to be redacted in WithGroup context, got: %s", output)
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output, got: %s", output)
	}
}

func TestRedactingHandler_Enabled(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})
	handler := newRedactingHandler(inner, "secret")

	if handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expected Info to be disabled when handler level is Warn")
	}
	if !handler.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("expected Warn to be enabled when handler level is Warn")
	}
	if !handler.Enabled(context.Background(), slog.LevelError) {
		t.Error("expected Error to be enabled when handler level is Warn")
	}
}
