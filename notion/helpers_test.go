package notion

import (
	"strings"
	"testing"
	"time"
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

// --- New Rich Text Constructor Tests ---

func TestNewRichTextBold(t *testing.T) {
	rt := NewRichTextBold("strong")
	if len(rt) != 1 {
		t.Fatalf("expected 1 element, got %d", len(rt))
	}
	if rt[0].Type != "text" {
		t.Errorf("Type = %q, want %q", rt[0].Type, "text")
	}
	if rt[0].Text == nil || rt[0].Text.Content != "strong" {
		t.Errorf("Content = %v, want %q", rt[0].Text, "strong")
	}
	if rt[0].Annotations == nil || !rt[0].Annotations.Bold {
		t.Error("expected Bold annotation to be true")
	}
}

func TestNewRichTextItalic(t *testing.T) {
	rt := NewRichTextItalic("emphasis")
	if len(rt) != 1 {
		t.Fatalf("expected 1 element, got %d", len(rt))
	}
	if rt[0].Type != "text" {
		t.Errorf("Type = %q, want %q", rt[0].Type, "text")
	}
	if rt[0].Text == nil || rt[0].Text.Content != "emphasis" {
		t.Errorf("Content = %v, want %q", rt[0].Text, "emphasis")
	}
	if rt[0].Annotations == nil || !rt[0].Annotations.Italic {
		t.Error("expected Italic annotation to be true")
	}
}

func TestNewRichTextCode(t *testing.T) {
	rt := NewRichTextCode("snippet")
	if len(rt) != 1 {
		t.Fatalf("expected 1 element, got %d", len(rt))
	}
	if rt[0].Type != "text" {
		t.Errorf("Type = %q, want %q", rt[0].Type, "text")
	}
	if rt[0].Text == nil || rt[0].Text.Content != "snippet" {
		t.Errorf("Content = %v, want %q", rt[0].Text, "snippet")
	}
	if rt[0].Annotations == nil || !rt[0].Annotations.Code {
		t.Error("expected Code annotation to be true")
	}
}

func TestNewRichTextLink(t *testing.T) {
	rt := NewRichTextLink("click here", "https://example.com")
	if len(rt) != 1 {
		t.Fatalf("expected 1 element, got %d", len(rt))
	}
	if rt[0].Type != "text" {
		t.Errorf("Type = %q, want %q", rt[0].Type, "text")
	}
	if rt[0].Text == nil {
		t.Fatal("expected Text to be non-nil")
	}
	if rt[0].Text.Content != "click here" {
		t.Errorf("Content = %q, want %q", rt[0].Text.Content, "click here")
	}
	if rt[0].Text.Link == nil {
		t.Fatal("expected Link to be non-nil")
	}
	if rt[0].Text.Link.URL != "https://example.com" {
		t.Errorf("Link.URL = %q, want %q", rt[0].Text.Link.URL, "https://example.com")
	}
}

// --- New Block Constructor Tests ---

func TestNewToggleBlock(t *testing.T) {
	block := NewToggleBlock("toggle content")
	if block.Object != "block" {
		t.Errorf("Object = %q, want %q", block.Object, "block")
	}
	if block.Type != "toggle" {
		t.Errorf("Type = %q, want %q", block.Type, "toggle")
	}
	if block.Toggle == nil {
		t.Fatal("expected Toggle to be non-nil")
	}
	text := ExtractPlainText(block.Toggle.RichText)
	if text != "toggle content" {
		t.Errorf("Toggle text = %q, want %q", text, "toggle content")
	}
}

func TestNewCalloutBlock(t *testing.T) {
	t.Run("with emoji", func(t *testing.T) {
		block := NewCalloutBlock("important note", "!")
		if block.Object != "block" {
			t.Errorf("Object = %q, want %q", block.Object, "block")
		}
		if block.Type != "callout" {
			t.Errorf("Type = %q, want %q", block.Type, "callout")
		}
		if block.Callout == nil {
			t.Fatal("expected Callout to be non-nil")
		}
		text := ExtractPlainText(block.Callout.RichText)
		if text != "important note" {
			t.Errorf("Callout text = %q, want %q", text, "important note")
		}
		if block.Callout.Icon == nil {
			t.Fatal("expected Icon to be non-nil")
		}
		if block.Callout.Icon.Type != "emoji" {
			t.Errorf("Icon.Type = %q, want %q", block.Callout.Icon.Type, "emoji")
		}
		if block.Callout.Icon.Emoji != "!" {
			t.Errorf("Icon.Emoji = %q, want %q", block.Callout.Icon.Emoji, "!")
		}
	})

	t.Run("without emoji", func(t *testing.T) {
		block := NewCalloutBlock("plain callout", "")
		if block.Type != "callout" {
			t.Errorf("Type = %q, want %q", block.Type, "callout")
		}
		if block.Callout == nil {
			t.Fatal("expected Callout to be non-nil")
		}
		text := ExtractPlainText(block.Callout.RichText)
		if text != "plain callout" {
			t.Errorf("Callout text = %q, want %q", text, "plain callout")
		}
		if block.Callout.Icon != nil {
			t.Error("expected Icon to be nil when no emoji provided")
		}
	})
}

func TestNewBookmarkBlock(t *testing.T) {
	block := NewBookmarkBlock("https://notion.so")
	if block.Object != "block" {
		t.Errorf("Object = %q, want %q", block.Object, "block")
	}
	if block.Type != "bookmark" {
		t.Errorf("Type = %q, want %q", block.Type, "bookmark")
	}
	if block.Bookmark == nil {
		t.Fatal("expected Bookmark to be non-nil")
	}
	if block.Bookmark.URL != "https://notion.so" {
		t.Errorf("Bookmark.URL = %q, want %q", block.Bookmark.URL, "https://notion.so")
	}
}

// --- Property Constructor Tests ---

func TestNewTitleProperty(t *testing.T) {
	prop := NewTitleProperty("My Page Title")
	titleSlice, ok := prop["title"].([]map[string]any)
	if !ok {
		t.Fatalf("expected title to be []map[string]any, got %T", prop["title"])
	}
	if len(titleSlice) != 1 {
		t.Fatalf("expected 1 title element, got %d", len(titleSlice))
	}
	textMap, ok := titleSlice[0]["text"].(map[string]string)
	if !ok {
		t.Fatalf("expected text to be map[string]string, got %T", titleSlice[0]["text"])
	}
	if textMap["content"] != "My Page Title" {
		t.Errorf("content = %q, want %q", textMap["content"], "My Page Title")
	}
}

func TestNewRichTextProperty(t *testing.T) {
	prop := NewRichTextProperty("Some text content")
	rtSlice, ok := prop["rich_text"].([]map[string]any)
	if !ok {
		t.Fatalf("expected rich_text to be []map[string]any, got %T", prop["rich_text"])
	}
	if len(rtSlice) != 1 {
		t.Fatalf("expected 1 rich_text element, got %d", len(rtSlice))
	}
	textMap, ok := rtSlice[0]["text"].(map[string]string)
	if !ok {
		t.Fatalf("expected text to be map[string]string, got %T", rtSlice[0]["text"])
	}
	if textMap["content"] != "Some text content" {
		t.Errorf("content = %q, want %q", textMap["content"], "Some text content")
	}
}

func TestNewNumberProperty(t *testing.T) {
	tests := []struct {
		name  string
		value float64
	}{
		{"positive integer", 42},
		{"negative", -3.14},
		{"zero", 0},
		{"decimal", 99.99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := NewNumberProperty(tt.value)
			got, ok := prop["number"].(float64)
			if !ok {
				t.Fatalf("expected number to be float64, got %T", prop["number"])
			}
			if got != tt.value {
				t.Errorf("number = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestNewSelectProperty(t *testing.T) {
	prop := NewSelectProperty("Option A")
	selectMap, ok := prop["select"].(map[string]string)
	if !ok {
		t.Fatalf("expected select to be map[string]string, got %T", prop["select"])
	}
	if selectMap["name"] != "Option A" {
		t.Errorf("name = %q, want %q", selectMap["name"], "Option A")
	}
}

func TestNewMultiSelectProperty(t *testing.T) {
	t.Run("multiple values", func(t *testing.T) {
		prop := NewMultiSelectProperty([]string{"Tag1", "Tag2", "Tag3"})
		options, ok := prop["multi_select"].([]map[string]string)
		if !ok {
			t.Fatalf("expected multi_select to be []map[string]string, got %T", prop["multi_select"])
		}
		if len(options) != 3 {
			t.Fatalf("expected 3 options, got %d", len(options))
		}
		want := []string{"Tag1", "Tag2", "Tag3"}
		for i, opt := range options {
			if opt["name"] != want[i] {
				t.Errorf("option[%d] = %q, want %q", i, opt["name"], want[i])
			}
		}
	})

	t.Run("empty values", func(t *testing.T) {
		prop := NewMultiSelectProperty([]string{})
		options, ok := prop["multi_select"].([]map[string]string)
		if !ok {
			t.Fatalf("expected multi_select to be []map[string]string, got %T", prop["multi_select"])
		}
		if len(options) != 0 {
			t.Errorf("expected 0 options, got %d", len(options))
		}
	})
}

func TestNewDateProperty(t *testing.T) {
	startTime := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)

	t.Run("without end date", func(t *testing.T) {
		prop := NewDateProperty(startTime, nil)
		dateObj, ok := prop["date"].(map[string]any)
		if !ok {
			t.Fatalf("expected date to be map[string]any, got %T", prop["date"])
		}
		startStr, ok := dateObj["start"].(string)
		if !ok {
			t.Fatalf("expected start to be string, got %T", dateObj["start"])
		}
		wantStart := startTime.Format(time.RFC3339)
		if startStr != wantStart {
			t.Errorf("start = %q, want %q", startStr, wantStart)
		}
		if _, exists := dateObj["end"]; exists {
			t.Error("expected end key to not exist when end is nil")
		}
	})

	t.Run("with end date", func(t *testing.T) {
		endTime := time.Date(2025, 6, 20, 18, 0, 0, 0, time.UTC)
		prop := NewDateProperty(startTime, &endTime)
		dateObj, ok := prop["date"].(map[string]any)
		if !ok {
			t.Fatalf("expected date to be map[string]any, got %T", prop["date"])
		}
		startStr, ok := dateObj["start"].(string)
		if !ok {
			t.Fatalf("expected start to be string, got %T", dateObj["start"])
		}
		wantStart := startTime.Format(time.RFC3339)
		if startStr != wantStart {
			t.Errorf("start = %q, want %q", startStr, wantStart)
		}
		endStr, ok := dateObj["end"].(string)
		if !ok {
			t.Fatalf("expected end to be string, got %T", dateObj["end"])
		}
		wantEnd := endTime.Format(time.RFC3339)
		if endStr != wantEnd {
			t.Errorf("end = %q, want %q", endStr, wantEnd)
		}
	})
}

func TestNewCheckboxProperty(t *testing.T) {
	tests := []struct {
		name    string
		checked bool
	}{
		{"checked", true},
		{"unchecked", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := NewCheckboxProperty(tt.checked)
			got, ok := prop["checkbox"].(bool)
			if !ok {
				t.Fatalf("expected checkbox to be bool, got %T", prop["checkbox"])
			}
			if got != tt.checked {
				t.Errorf("checkbox = %v, want %v", got, tt.checked)
			}
		})
	}
}

func TestNewURLProperty(t *testing.T) {
	prop := NewURLProperty("https://example.com/page")
	got, ok := prop["url"].(string)
	if !ok {
		t.Fatalf("expected url to be string, got %T", prop["url"])
	}
	if got != "https://example.com/page" {
		t.Errorf("url = %q, want %q", got, "https://example.com/page")
	}
}

func TestNewEmailProperty(t *testing.T) {
	prop := NewEmailProperty("user@example.com")
	got, ok := prop["email"].(string)
	if !ok {
		t.Fatalf("expected email to be string, got %T", prop["email"])
	}
	if got != "user@example.com" {
		t.Errorf("email = %q, want %q", got, "user@example.com")
	}
}

func TestNewPhoneNumberProperty(t *testing.T) {
	prop := NewPhoneNumberProperty("+1-555-0123")
	got, ok := prop["phone_number"].(string)
	if !ok {
		t.Fatalf("expected phone_number to be string, got %T", prop["phone_number"])
	}
	if got != "+1-555-0123" {
		t.Errorf("phone_number = %q, want %q", got, "+1-555-0123")
	}
}

// --- Additional BlockToMarkdown Tests ---

func TestBlockToMarkdown_AdditionalCases(t *testing.T) {
	tests := []struct {
		name  string
		block Block
		want  string
	}{
		{
			name: "heading_2",
			block: Block{
				Type: "heading_2",
				Heading2: &RichTextBlock{
					RichText: NewRichText("Subtitle"),
				},
			},
			want: "## Subtitle\n",
		},
		{
			name: "heading_3",
			block: Block{
				Type: "heading_3",
				Heading3: &RichTextBlock{
					RichText: NewRichText("Sub-subtitle"),
				},
			},
			want: "### Sub-subtitle\n",
		},
		{
			name: "bulleted_list_item",
			block: Block{
				Type: "bulleted_list_item",
				BulletedListItem: &RichTextBlock{
					RichText: NewRichText("bullet point"),
				},
			},
			want: "- bullet point\n",
		},
		{
			name: "numbered_list_item",
			block: Block{
				Type: "numbered_list_item",
				NumberedListItem: &RichTextBlock{
					RichText: NewRichText("first item"),
				},
			},
			want: "1. first item\n",
		},
		{
			name: "toggle",
			block: Block{
				Type: "toggle",
				Toggle: &RichTextBlock{
					RichText: NewRichText("toggle summary"),
				},
			},
			want: "- toggle summary\n",
		},
		{
			name: "code block with language",
			block: Block{
				Type: "code",
				Code: &CodeBlock{
					RichText: NewRichText("fmt.Println(\"hello\")"),
					Language: "go",
				},
			},
			want: "```go\nfmt.Println(\"hello\")\n```\n",
		},
		{
			name: "code block with empty language",
			block: Block{
				Type: "code",
				Code: &CodeBlock{
					RichText: NewRichText("some code"),
					Language: "",
				},
			},
			want: "```text\nsome code\n```\n",
		},
		{
			name: "quote",
			block: Block{
				Type: "quote",
				Quote: &RichTextBlock{
					RichText: NewRichText("wise words"),
				},
			},
			want: "> wise words\n",
		},
		{
			name: "callout with custom emoji",
			block: Block{
				Type: "callout",
				Callout: &CalloutBlock{
					RichText: NewRichText("warning message"),
					Icon: &Icon{
						Type:  "emoji",
						Emoji: "!",
					},
				},
			},
			want: "> ! warning message\n",
		},
		{
			name: "callout without emoji uses default",
			block: Block{
				Type: "callout",
				Callout: &CalloutBlock{
					RichText: NewRichText("info message"),
				},
			},
			want: "> \xf0\x9f\x92\xa1 info message\n",
		},
		{
			name: "callout with nil icon uses default",
			block: Block{
				Type: "callout",
				Callout: &CalloutBlock{
					RichText: NewRichText("tip"),
					Icon:     nil,
				},
			},
			want: "> \xf0\x9f\x92\xa1 tip\n",
		},
		{
			name: "callout with icon but empty emoji uses default",
			block: Block{
				Type: "callout",
				Callout: &CalloutBlock{
					RichText: NewRichText("note"),
					Icon:     &Icon{Type: "emoji", Emoji: ""},
				},
			},
			want: "> \xf0\x9f\x92\xa1 note\n",
		},
		{
			name:  "table_of_contents",
			block: Block{Type: "table_of_contents"},
			want:  "_Table of Contents_\n",
		},
		{
			name: "bookmark without caption",
			block: Block{
				Type: "bookmark",
				Bookmark: &BookmarkBlock{
					URL: "https://example.com",
				},
			},
			want: "[https://example.com](https://example.com)\n",
		},
		{
			name: "bookmark with caption",
			block: Block{
				Type: "bookmark",
				Bookmark: &BookmarkBlock{
					URL:     "https://example.com",
					Caption: NewRichText("Example Site"),
				},
			},
			want: "[Example Site](https://example.com)\n",
		},
		{
			name: "child_page",
			block: Block{
				Type: "child_page",
				ChildPage: &ChildPageBlock{
					Title: "Sub Page",
				},
			},
			want: "[\xf0\x9f\x93\x84 Sub Page](notion page)\n",
		},
		{
			name: "child_database",
			block: Block{
				Type: "child_database",
				ChildDatabase: &ChildDatabaseBlock{
					Title: "My Database",
				},
			},
			want: "[\xf0\x9f\x97\x84\xef\xb8\x8f My Database](notion database)\n",
		},
		{
			name: "empty paragraph (nil Paragraph)",
			block: Block{
				Type:      "paragraph",
				Paragraph: nil,
			},
			want: "\n",
		},
		{
			name: "paragraph with empty RichText slice",
			block: Block{
				Type: "paragraph",
				Paragraph: &RichTextBlock{
					RichText: []RichText{},
				},
			},
			want: "\n",
		},
		{
			name: "unknown block type",
			block: Block{
				Type: "unsupported_block_type",
			},
			want: "",
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

func TestBlockToMarkdown_WithIndent(t *testing.T) {
	block := Block{
		Type: "paragraph",
		Paragraph: &RichTextBlock{
			RichText: NewRichText("indented"),
		},
	}
	got := BlockToMarkdown(&block, 2)
	want := "    indented\n"
	if got != want {
		t.Errorf("BlockToMarkdown() with indent=2 = %q, want %q", got, want)
	}
}

// --- Additional RichTextToMarkdown Tests ---

func TestRichTextToMarkdown_AdditionalCases(t *testing.T) {
	tests := []struct {
		name  string
		input []RichText
		want  string
	}{
		{
			name: "bold and italic combined",
			input: []RichText{
				{
					Text:        &TextContent{Content: "both"},
					Annotations: &Annotations{Bold: true, Italic: true},
				},
			},
			want: "***both***",
		},
		{
			name: "text with Href field (not Text.Link)",
			input: []RichText{
				{
					Text: &TextContent{Content: "linked via href"},
					Href: "https://href-example.com",
				},
			},
			want: "[linked via href](https://href-example.com)",
		},
		{
			name: "empty text should be skipped",
			input: []RichText{
				{
					Text: &TextContent{Content: ""},
				},
				{
					Text: &TextContent{Content: "visible"},
				},
			},
			want: "visible",
		},
		{
			name: "text with only PlainText field (no Text)",
			input: []RichText{
				{
					PlainText: "fallback text",
				},
			},
			want: "fallback text",
		},
		{
			name: "nil annotations means no formatting",
			input: []RichText{
				{
					Text:        &TextContent{Content: "plain"},
					Annotations: nil,
				},
			},
			want: "plain",
		},
		{
			name:  "empty input",
			input: []RichText{},
			want:  "",
		},
		{
			name: "bold with link via Text.Link",
			input: []RichText{
				{
					Text: &TextContent{
						Content: "bold link",
						Link:    &LinkObject{URL: "https://bold.example.com"},
					},
					Annotations: &Annotations{Bold: true},
				},
			},
			want: "[**bold link**](https://bold.example.com)",
		},
		{
			name: "Href takes priority when both Href and Text.Link present",
			input: []RichText{
				{
					Text: &TextContent{
						Content: "link text",
						Link:    &LinkObject{URL: "https://text-link.com"},
					},
					Href: "https://href-link.com",
				},
			},
			want: "[link text](https://href-link.com)",
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

// --- NewHeadingBlock Level Tests ---

func TestNewHeadingBlock_Levels(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		wantType string
	}{
		{"level 1", 1, "heading_1"},
		{"level 2", 2, "heading_2"},
		{"level 3", 3, "heading_3"},
		{"level 0 defaults to heading_2", 0, "heading_2"},
		{"level 4 defaults to heading_2", 4, "heading_2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := NewHeadingBlock(tt.level, "test heading")
			if block.Type != tt.wantType {
				t.Errorf("NewHeadingBlock(%d).Type = %q, want %q", tt.level, block.Type, tt.wantType)
			}
			if block.Object != "block" {
				t.Errorf("Object = %q, want %q", block.Object, "block")
			}
			// Verify the correct heading field is populated
			switch tt.wantType {
			case "heading_1":
				if block.Heading1 == nil {
					t.Error("expected Heading1 to be non-nil")
				}
			case "heading_2":
				if block.Heading2 == nil {
					t.Error("expected Heading2 to be non-nil")
				}
			case "heading_3":
				if block.Heading3 == nil {
					t.Error("expected Heading3 to be non-nil")
				}
			}
		})
	}
}

// --- NewCodeBlock with Language Test ---

func TestNewCodeBlock_WithLanguage(t *testing.T) {
	block := NewCodeBlock("print('hello')", "python")
	if block.Type != "code" {
		t.Errorf("Type = %q, want %q", block.Type, "code")
	}
	if block.Code == nil {
		t.Fatal("expected Code to be non-nil")
	}
	if block.Code.Language != "python" {
		t.Errorf("Language = %q, want %q", block.Code.Language, "python")
	}
	text := ExtractPlainText(block.Code.RichText)
	if text != "print('hello')" {
		t.Errorf("Code text = %q, want %q", text, "print('hello')")
	}
}

// --- BlocksToMarkdown Additional Tests ---

func TestBlocksToMarkdown_MixedContent(t *testing.T) {
	blocks := []Block{
		NewHeadingBlock(1, "Title"),
		NewParagraphBlock("Intro paragraph"),
		NewBulletedListItemBlock("Item 1"),
		NewBulletedListItemBlock("Item 2"),
		NewDividerBlock(),
		NewCodeBlock("x = 1", "python"),
		NewQuoteBlock("Famous quote"),
	}

	result := BlocksToMarkdown(blocks, 0)

	expected := []string{
		"# Title",
		"Intro paragraph",
		"- Item 1",
		"- Item 2",
		"---",
		"```python",
		"x = 1",
		"```",
		"> Famous quote",
	}
	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("BlocksToMarkdown result missing %q\ngot: %s", exp, result)
		}
	}
}

func TestBlocksToMarkdown_EmptyBlocks(t *testing.T) {
	result := BlocksToMarkdown([]Block{}, 0)
	if result != "" {
		t.Errorf("expected empty string for empty blocks, got %q", result)
	}
}
