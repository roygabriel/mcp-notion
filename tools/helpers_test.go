package tools

import "testing"

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"exact match", "Hello", "Hello", true},
		{"case insensitive", "Hello World", "hello", true},
		{"not found", "Hello", "xyz", false},
		{"empty substr", "Hello", "", true},
		{"empty string", "", "hello", false},
		{"substr longer than s", "Hi", "Hello", false},
		{"unicode", "Notion API", "notion", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsIgnoreCase(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}
