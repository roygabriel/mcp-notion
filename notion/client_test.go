package notion

import (
	"testing"
	"time"
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
