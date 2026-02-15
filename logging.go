package main

import (
	"context"
	"log/slog"
	"strings"
)

// redactingHandler wraps an slog.Handler to redact secret values from log output.
type redactingHandler struct {
	inner   slog.Handler
	secrets []string
}

// newRedactingHandler creates a new redactingHandler that replaces occurrences
// of the given secrets with "[REDACTED]" in log messages and string attributes.
// Empty strings in the secrets list are filtered out.
func newRedactingHandler(inner slog.Handler, secrets ...string) *redactingHandler {
	filtered := make([]string, 0, len(secrets))
	for _, s := range secrets {
		if s != "" {
			filtered = append(filtered, s)
		}
	}
	return &redactingHandler{inner: inner, secrets: filtered}
}

func (h *redactingHandler) redact(s string) string {
	for _, secret := range h.secrets {
		s = strings.ReplaceAll(s, secret, "[REDACTED]")
	}
	return s
}

func (h *redactingHandler) redactAttr(a slog.Attr) slog.Attr {
	switch a.Value.Kind() {
	case slog.KindString:
		a.Value = slog.StringValue(h.redact(a.Value.String()))
	case slog.KindGroup:
		attrs := a.Value.Group()
		redacted := make([]slog.Attr, len(attrs))
		for i, ga := range attrs {
			redacted[i] = h.redactAttr(ga)
		}
		a.Value = slog.GroupValue(redacted...)
	}
	return a
}

// Enabled reports whether the handler handles records at the given level.
func (h *redactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle redacts secrets from the record message and attributes before
// passing to the inner handler.
func (h *redactingHandler) Handle(ctx context.Context, r slog.Record) error {
	r.Message = h.redact(r.Message)
	var redactedAttrs []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		redactedAttrs = append(redactedAttrs, h.redactAttr(a))
		return true
	})
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	newRecord.AddAttrs(redactedAttrs...)
	return h.inner.Handle(ctx, newRecord)
}

// WithAttrs returns a new handler with the given attributes redacted.
func (h *redactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	redacted := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		redacted[i] = h.redactAttr(a)
	}
	return &redactingHandler{
		inner:   h.inner.WithAttrs(redacted),
		secrets: h.secrets,
	}
}

// WithGroup returns a new handler with the given group name.
func (h *redactingHandler) WithGroup(name string) slog.Handler {
	return &redactingHandler{
		inner:   h.inner.WithGroup(name),
		secrets: h.secrets,
	}
}
