package notion

import (
	"strings"
	"testing"
)

func TestMarkdownToBlocks(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTypes []string
		wantCount int
	}{
		{
			name:      "heading 1",
			input:     "# Hello World",
			wantTypes: []string{"heading_1"},
			wantCount: 1,
		},
		{
			name:      "heading 2",
			input:     "## Section",
			wantTypes: []string{"heading_2"},
			wantCount: 1,
		},
		{
			name:      "heading 3",
			input:     "### Subsection",
			wantTypes: []string{"heading_3"},
			wantCount: 1,
		},
		{
			name:      "bulleted list with dash",
			input:     "- Item 1\n- Item 2",
			wantTypes: []string{"bulleted_list_item", "bulleted_list_item"},
			wantCount: 2,
		},
		{
			name:      "bulleted list with asterisk",
			input:     "* Item 1",
			wantTypes: []string{"bulleted_list_item"},
			wantCount: 1,
		},
		{
			name:      "numbered list",
			input:     "1. First\n2. Second",
			wantTypes: []string{"numbered_list_item", "numbered_list_item"},
			wantCount: 2,
		},
		{
			name:      "unchecked todo",
			input:     "- [ ] Task",
			wantTypes: []string{"to_do"},
			wantCount: 1,
		},
		{
			name:      "checked todo",
			input:     "- [x] Done",
			wantTypes: []string{"to_do"},
			wantCount: 1,
		},
		{
			name:      "quote",
			input:     "> Some quote",
			wantTypes: []string{"quote"},
			wantCount: 1,
		},
		{
			name:      "divider dashes",
			input:     "---",
			wantTypes: []string{"divider"},
			wantCount: 1,
		},
		{
			name:      "divider asterisks",
			input:     "***",
			wantTypes: []string{"divider"},
			wantCount: 1,
		},
		{
			name:      "code block",
			input:     "```go\nfmt.Println(\"hi\")\n```",
			wantTypes: []string{"code"},
			wantCount: 1,
		},
		{
			name:      "paragraph",
			input:     "Just some text",
			wantTypes: []string{"paragraph"},
			wantCount: 1,
		},
		{
			name:      "skip empty lines",
			input:     "Line 1\n\nLine 2",
			wantTypes: []string{"paragraph", "paragraph"},
			wantCount: 2,
		},
		{
			name:      "mixed content",
			input:     "# Title\n\nSome text\n\n- Item\n\n> Quote",
			wantTypes: []string{"heading_1", "paragraph", "bulleted_list_item", "quote"},
			wantCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := MarkdownToBlocks(tt.input)

			if len(blocks) != tt.wantCount {
				t.Errorf("MarkdownToBlocks() returned %d blocks, want %d", len(blocks), tt.wantCount)
				for i, b := range blocks {
					t.Logf("  block[%d]: type=%s", i, b.Type)
				}
				return
			}

			for i, wantType := range tt.wantTypes {
				if blocks[i].Type != wantType {
					t.Errorf("block[%d].Type = %q, want %q", i, blocks[i].Type, wantType)
				}
			}
		})
	}
}

func TestMarkdownToBlocks_CodeBlockLanguage(t *testing.T) {
	blocks := MarkdownToBlocks("```python\nprint('hello')\n```")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Code == nil {
		t.Fatal("expected code block")
	}
	if blocks[0].Code.Language != "python" {
		t.Errorf("language = %q, want %q", blocks[0].Code.Language, "python")
	}
}

func TestMarkdownToBlocks_UnclosedCodeBlock(t *testing.T) {
	blocks := MarkdownToBlocks("```\nsome code without closing")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block for unclosed code block, got %d", len(blocks))
	}
	if blocks[0].Type != "code" {
		t.Errorf("expected code block, got %q", blocks[0].Type)
	}
}

func TestExtractPlainText(t *testing.T) {
	tests := []struct {
		name  string
		input []RichText
		want  string
	}{
		{
			name:  "empty",
			input: []RichText{},
			want:  "",
		},
		{
			name: "plain text from PlainText field",
			input: []RichText{
				{PlainText: "Hello"},
				{PlainText: " World"},
			},
			want: "Hello World",
		},
		{
			name: "plain text from Text.Content",
			input: []RichText{
				{Text: &TextContent{Content: "Hello"}},
			},
			want: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPlainText(tt.input)
			if got != tt.want {
				t.Errorf("ExtractPlainText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRichTextToMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input []RichText
		want  string
	}{
		{
			name: "bold text",
			input: []RichText{
				{
					Text:        &TextContent{Content: "bold"},
					Annotations: &Annotations{Bold: true},
				},
			},
			want: "**bold**",
		},
		{
			name: "italic text",
			input: []RichText{
				{
					Text:        &TextContent{Content: "italic"},
					Annotations: &Annotations{Italic: true},
				},
			},
			want: "*italic*",
		},
		{
			name: "code text",
			input: []RichText{
				{
					Text:        &TextContent{Content: "code"},
					Annotations: &Annotations{Code: true},
				},
			},
			want: "`code`",
		},
		{
			name: "strikethrough text",
			input: []RichText{
				{
					Text:        &TextContent{Content: "deleted"},
					Annotations: &Annotations{Strikethrough: true},
				},
			},
			want: "~~deleted~~",
		},
		{
			name: "link text",
			input: []RichText{
				{
					Text: &TextContent{
						Content: "click",
						Link:    &LinkObject{URL: "https://example.com"},
					},
				},
			},
			want: "[click](https://example.com)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RichTextToMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("RichTextToMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBlockToMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		block Block
		want  string
	}{
		{
			name: "paragraph",
			block: Block{
				Type: "paragraph",
				Paragraph: &RichTextBlock{
					RichText: NewRichText("hello"),
				},
			},
			want: "hello\n",
		},
		{
			name: "heading_1",
			block: Block{
				Type: "heading_1",
				Heading1: &RichTextBlock{
					RichText: NewRichText("Title"),
				},
			},
			want: "# Title\n",
		},
		{
			name:  "divider",
			block: Block{Type: "divider"},
			want:  "---\n",
		},
		{
			name: "todo unchecked",
			block: Block{
				Type: "to_do",
				ToDo: &ToDoBlock{
					RichText: NewRichText("task"),
					Checked:  false,
				},
			},
			want: "- [ ] task\n",
		},
		{
			name: "todo checked",
			block: Block{
				Type: "to_do",
				ToDo: &ToDoBlock{
					RichText: NewRichText("done"),
					Checked:  true,
				},
			},
			want: "- [x] done\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BlockToMarkdown(&tt.block, 0)
			if got != tt.want {
				t.Errorf("BlockToMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBlocksToMarkdown(t *testing.T) {
	blocks := []Block{
		NewHeadingBlock(1, "Title"),
		NewParagraphBlock("Content"),
	}

	result := BlocksToMarkdown(blocks, 0)

	if !strings.Contains(result, "# Title") {
		t.Error("expected markdown to contain '# Title'")
	}
	if !strings.Contains(result, "Content") {
		t.Error("expected markdown to contain 'Content'")
	}
}

func TestNewRichText(t *testing.T) {
	rt := NewRichText("test")
	if len(rt) != 1 {
		t.Fatalf("expected 1 element, got %d", len(rt))
	}
	if rt[0].Type != "text" {
		t.Errorf("Type = %q, want %q", rt[0].Type, "text")
	}
	if rt[0].Text.Content != "test" {
		t.Errorf("Content = %q, want %q", rt[0].Text.Content, "test")
	}
}

func TestNewHeadingBlock_InvalidLevel(t *testing.T) {
	block := NewHeadingBlock(99, "test")
	if block.Type != "heading_2" {
		t.Errorf("invalid heading level should default to heading_2, got %q", block.Type)
	}
}

func TestNewCodeBlock_DefaultLanguage(t *testing.T) {
	block := NewCodeBlock("code", "")
	if block.Code.Language != "plain text" {
		t.Errorf("empty language should default to 'plain text', got %q", block.Code.Language)
	}
}
