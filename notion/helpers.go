package notion

import (
	"regexp"
	"strings"
	"time"
)

var numberedListRegexp = regexp.MustCompile(`^\d+\.\s`)

// NewRichText creates a simple rich text array from a string
func NewRichText(content string) []RichText {
	return []RichText{
		{
			Type: "text",
			Text: &TextContent{
				Content: content,
			},
		},
	}
}

// NewRichTextBold creates a bold rich text array
func NewRichTextBold(content string) []RichText {
	return []RichText{
		{
			Type: "text",
			Text: &TextContent{
				Content: content,
			},
			Annotations: &Annotations{
				Bold: true,
			},
		},
	}
}

// NewRichTextItalic creates an italic rich text array
func NewRichTextItalic(content string) []RichText {
	return []RichText{
		{
			Type: "text",
			Text: &TextContent{
				Content: content,
			},
			Annotations: &Annotations{
				Italic: true,
			},
		},
	}
}

// NewRichTextCode creates a code-formatted rich text array
func NewRichTextCode(content string) []RichText {
	return []RichText{
		{
			Type: "text",
			Text: &TextContent{
				Content: content,
			},
			Annotations: &Annotations{
				Code: true,
			},
		},
	}
}

// NewRichTextLink creates a link rich text array
func NewRichTextLink(content, url string) []RichText {
	return []RichText{
		{
			Type: "text",
			Text: &TextContent{
				Content: content,
				Link: &LinkObject{
					URL: url,
				},
			},
		},
	}
}

// NewParagraphBlock creates a paragraph block
func NewParagraphBlock(text string) Block {
	return Block{
		Object: "block",
		Type:   "paragraph",
		Paragraph: &RichTextBlock{
			RichText: NewRichText(text),
		},
	}
}

// NewHeadingBlock creates a heading block (level 1, 2, or 3)
func NewHeadingBlock(level int, text string) Block {
	richText := NewRichText(text)
	block := Block{
		Object: "block",
	}

	switch level {
	case 1:
		block.Type = "heading_1"
		block.Heading1 = &RichTextBlock{RichText: richText}
	case 2:
		block.Type = "heading_2"
		block.Heading2 = &RichTextBlock{RichText: richText}
	case 3:
		block.Type = "heading_3"
		block.Heading3 = &RichTextBlock{RichText: richText}
	default:
		// Default to heading 2
		block.Type = "heading_2"
		block.Heading2 = &RichTextBlock{RichText: richText}
	}

	return block
}

// NewBulletedListItemBlock creates a bulleted list item block
func NewBulletedListItemBlock(text string) Block {
	return Block{
		Object: "block",
		Type:   "bulleted_list_item",
		BulletedListItem: &RichTextBlock{
			RichText: NewRichText(text),
		},
	}
}

// NewNumberedListItemBlock creates a numbered list item block
func NewNumberedListItemBlock(text string) Block {
	return Block{
		Object: "block",
		Type:   "numbered_list_item",
		NumberedListItem: &RichTextBlock{
			RichText: NewRichText(text),
		},
	}
}

// NewToDoBlock creates a to-do block
func NewToDoBlock(text string, checked bool) Block {
	return Block{
		Object: "block",
		Type:   "to_do",
		ToDo: &ToDoBlock{
			RichText: NewRichText(text),
			Checked:  checked,
		},
	}
}

// NewToggleBlock creates a toggle block
func NewToggleBlock(text string) Block {
	return Block{
		Object: "block",
		Type:   "toggle",
		Toggle: &RichTextBlock{
			RichText: NewRichText(text),
		},
	}
}

// NewCodeBlock creates a code block
func NewCodeBlock(text, language string) Block {
	if language == "" {
		language = "plain text"
	}
	return Block{
		Object: "block",
		Type:   "code",
		Code: &CodeBlock{
			RichText: NewRichText(text),
			Language: language,
		},
	}
}

// NewQuoteBlock creates a quote block
func NewQuoteBlock(text string) Block {
	return Block{
		Object: "block",
		Type:   "quote",
		Quote: &RichTextBlock{
			RichText: NewRichText(text),
		},
	}
}

// NewCalloutBlock creates a callout block
func NewCalloutBlock(text, emoji string) Block {
	block := Block{
		Object: "block",
		Type:   "callout",
		Callout: &CalloutBlock{
			RichText: NewRichText(text),
		},
	}
	if emoji != "" {
		block.Callout.Icon = &Icon{
			Type:  "emoji",
			Emoji: emoji,
		}
	}
	return block
}

// NewDividerBlock creates a divider block
func NewDividerBlock() Block {
	return Block{
		Object:  "block",
		Type:    "divider",
		Divider: map[string]any{},
	}
}

// NewBookmarkBlock creates a bookmark block
func NewBookmarkBlock(url string) Block {
	return Block{
		Object: "block",
		Type:   "bookmark",
		Bookmark: &BookmarkBlock{
			URL: url,
		},
	}
}

// NewTitleProperty creates a title property value
func NewTitleProperty(text string) map[string]any {
	return map[string]any{
		"title": []map[string]any{
			{
				"text": map[string]string{
					"content": text,
				},
			},
		},
	}
}

// NewRichTextProperty creates a rich text property value
func NewRichTextProperty(text string) map[string]any {
	return map[string]any{
		"rich_text": []map[string]any{
			{
				"text": map[string]string{
					"content": text,
				},
			},
		},
	}
}

// NewNumberProperty creates a number property value
func NewNumberProperty(value float64) map[string]any {
	return map[string]any{
		"number": value,
	}
}

// NewSelectProperty creates a select property value
func NewSelectProperty(value string) map[string]any {
	return map[string]any{
		"select": map[string]string{
			"name": value,
		},
	}
}

// NewMultiSelectProperty creates a multi-select property value
func NewMultiSelectProperty(values []string) map[string]any {
	options := make([]map[string]string, len(values))
	for i, v := range values {
		options[i] = map[string]string{"name": v}
	}
	return map[string]any{
		"multi_select": options,
	}
}

// NewDateProperty creates a date property value
func NewDateProperty(start time.Time, end *time.Time) map[string]any {
	dateObj := map[string]any{
		"start": start.Format(time.RFC3339),
	}
	if end != nil {
		dateObj["end"] = end.Format(time.RFC3339)
	}
	return map[string]any{
		"date": dateObj,
	}
}

// NewCheckboxProperty creates a checkbox property value
func NewCheckboxProperty(checked bool) map[string]any {
	return map[string]any{
		"checkbox": checked,
	}
}

// NewURLProperty creates a URL property value
func NewURLProperty(url string) map[string]any {
	return map[string]any{
		"url": url,
	}
}

// NewEmailProperty creates an email property value
func NewEmailProperty(email string) map[string]any {
	return map[string]any{
		"email": email,
	}
}

// NewPhoneNumberProperty creates a phone number property value
func NewPhoneNumberProperty(phone string) map[string]any {
	return map[string]any{
		"phone_number": phone,
	}
}

// MarkdownToBlocks converts simple markdown-like text to Notion blocks
func MarkdownToBlocks(markdown string) []Block {
	blocks := make([]Block, 0)
	lines := strings.Split(markdown, "\n")

	var inCodeBlock bool
	var codeContent []string
	var codeLanguage string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle code blocks
		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				// Start code block
				inCodeBlock = true
				codeLanguage = strings.TrimPrefix(trimmed, "```")
				if codeLanguage == "" {
					codeLanguage = "plain text"
				}
				codeContent = make([]string, 0)
			} else {
				// End code block
				inCodeBlock = false
				blocks = append(blocks, NewCodeBlock(strings.Join(codeContent, "\n"), codeLanguage))
				codeContent = nil
			}
			continue
		}

		if inCodeBlock {
			codeContent = append(codeContent, line)
			continue
		}

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Heading 1
		if strings.HasPrefix(trimmed, "# ") {
			text := strings.TrimPrefix(trimmed, "# ")
			blocks = append(blocks, NewHeadingBlock(1, text))
			continue
		}

		// Heading 2
		if strings.HasPrefix(trimmed, "## ") {
			text := strings.TrimPrefix(trimmed, "## ")
			blocks = append(blocks, NewHeadingBlock(2, text))
			continue
		}

		// Heading 3
		if strings.HasPrefix(trimmed, "### ") {
			text := strings.TrimPrefix(trimmed, "### ")
			blocks = append(blocks, NewHeadingBlock(3, text))
			continue
		}

		// To-do items (must be checked before bulleted list since "- [ ]" starts with "- ")
		if strings.HasPrefix(trimmed, "- [ ] ") {
			text := strings.TrimPrefix(trimmed, "- [ ] ")
			blocks = append(blocks, NewToDoBlock(text, false))
			continue
		}
		if strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "- [X] ") {
			text := strings.TrimPrefix(strings.TrimPrefix(trimmed, "- [x] "), "- [X] ")
			blocks = append(blocks, NewToDoBlock(text, true))
			continue
		}

		// Bulleted list
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			text := strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
			blocks = append(blocks, NewBulletedListItemBlock(text))
			continue
		}

		// Numbered list (simple detection)
		if numberedListRegexp.MatchString(trimmed) {
			text := numberedListRegexp.ReplaceAllString(trimmed, "")
			blocks = append(blocks, NewNumberedListItemBlock(text))
			continue
		}

		// Quote
		if strings.HasPrefix(trimmed, "> ") {
			text := strings.TrimPrefix(trimmed, "> ")
			blocks = append(blocks, NewQuoteBlock(text))
			continue
		}

		// Divider
		if trimmed == "---" || trimmed == "***" {
			blocks = append(blocks, NewDividerBlock())
			continue
		}

		// Default to paragraph
		blocks = append(blocks, NewParagraphBlock(line))
	}

	// Handle unclosed code block
	if inCodeBlock && len(codeContent) > 0 {
		blocks = append(blocks, NewCodeBlock(strings.Join(codeContent, "\n"), codeLanguage))
	}

	return blocks
}

// ExtractPlainText extracts plain text from rich text array
func ExtractPlainText(richText []RichText) string {
	var parts []string
	for _, rt := range richText {
		if rt.PlainText != "" {
			parts = append(parts, rt.PlainText)
		} else if rt.Text != nil {
			parts = append(parts, rt.Text.Content)
		}
	}
	return strings.Join(parts, "")
}

// RichTextToMarkdown converts rich text array to markdown with formatting
func RichTextToMarkdown(richText []RichText) string {
	var result strings.Builder

	for _, rt := range richText {
		text := ""
		if rt.Text != nil {
			text = rt.Text.Content
		} else if rt.PlainText != "" {
			text = rt.PlainText
		}

		if text == "" {
			continue
		}

		// Apply formatting annotations
		if rt.Annotations != nil {
			if rt.Annotations.Code {
				text = "`" + text + "`"
			} else {
				if rt.Annotations.Bold {
					text = "**" + text + "**"
				}
				if rt.Annotations.Italic {
					text = "*" + text + "*"
				}
				if rt.Annotations.Strikethrough {
					text = "~~" + text + "~~"
				}
			}
		}

		// Apply link if present
		if rt.Href != "" || (rt.Text != nil && rt.Text.Link != nil && rt.Text.Link.URL != "") {
			url := rt.Href
			if url == "" && rt.Text != nil && rt.Text.Link != nil {
				url = rt.Text.Link.URL
			}
			text = "[" + text + "](" + url + ")"
		}

		result.WriteString(text)
	}

	return result.String()
}

// BlockToMarkdown converts a Notion block to markdown
func BlockToMarkdown(block *Block, indent int) string {
	indentStr := strings.Repeat("  ", indent)

	switch block.Type {
	case "paragraph":
		if block.Paragraph != nil && len(block.Paragraph.RichText) > 0 {
			return indentStr + RichTextToMarkdown(block.Paragraph.RichText) + "\n"
		}
		return indentStr + "\n"

	case "heading_1":
		if block.Heading1 != nil && len(block.Heading1.RichText) > 0 {
			return indentStr + "# " + RichTextToMarkdown(block.Heading1.RichText) + "\n"
		}

	case "heading_2":
		if block.Heading2 != nil && len(block.Heading2.RichText) > 0 {
			return indentStr + "## " + RichTextToMarkdown(block.Heading2.RichText) + "\n"
		}

	case "heading_3":
		if block.Heading3 != nil && len(block.Heading3.RichText) > 0 {
			return indentStr + "### " + RichTextToMarkdown(block.Heading3.RichText) + "\n"
		}

	case "bulleted_list_item":
		if block.BulletedListItem != nil && len(block.BulletedListItem.RichText) > 0 {
			return indentStr + "- " + RichTextToMarkdown(block.BulletedListItem.RichText) + "\n"
		}

	case "numbered_list_item":
		if block.NumberedListItem != nil && len(block.NumberedListItem.RichText) > 0 {
			return indentStr + "1. " + RichTextToMarkdown(block.NumberedListItem.RichText) + "\n"
		}

	case "to_do":
		if block.ToDo != nil && len(block.ToDo.RichText) > 0 {
			checkbox := "[ ]"
			if block.ToDo.Checked {
				checkbox = "[x]"
			}
			return indentStr + "- " + checkbox + " " + RichTextToMarkdown(block.ToDo.RichText) + "\n"
		}

	case "toggle":
		if block.Toggle != nil && len(block.Toggle.RichText) > 0 {
			return indentStr + "- " + RichTextToMarkdown(block.Toggle.RichText) + "\n"
		}

	case "code":
		if block.Code != nil && len(block.Code.RichText) > 0 {
			language := block.Code.Language
			if language == "" {
				language = "text"
			}
			code := RichTextToMarkdown(block.Code.RichText)
			return indentStr + "```" + language + "\n" + code + "\n" + indentStr + "```\n"
		}

	case "quote":
		if block.Quote != nil && len(block.Quote.RichText) > 0 {
			return indentStr + "> " + RichTextToMarkdown(block.Quote.RichText) + "\n"
		}

	case "callout":
		if block.Callout != nil && len(block.Callout.RichText) > 0 {
			icon := "💡"
			if block.Callout.Icon != nil && block.Callout.Icon.Emoji != "" {
				icon = block.Callout.Icon.Emoji
			}
			return indentStr + "> " + icon + " " + RichTextToMarkdown(block.Callout.RichText) + "\n"
		}

	case "divider":
		return indentStr + "---\n"

	case "table_of_contents":
		return indentStr + "_Table of Contents_\n"

	case "bookmark":
		if block.Bookmark != nil {
			caption := ""
			if len(block.Bookmark.Caption) > 0 {
				caption = RichTextToMarkdown(block.Bookmark.Caption)
			}
			if caption == "" {
				caption = block.Bookmark.URL
			}
			return indentStr + "[" + caption + "](" + block.Bookmark.URL + ")\n"
		}

	case "child_page":
		if block.ChildPage != nil {
			return indentStr + "[📄 " + block.ChildPage.Title + "](notion page)\n"
		}

	case "child_database":
		if block.ChildDatabase != nil {
			return indentStr + "[🗄️ " + block.ChildDatabase.Title + "](notion database)\n"
		}
	}

	return ""
}

// BlocksToMarkdown converts an array of blocks to markdown
func BlocksToMarkdown(blocks []Block, indent int) string {
	var result strings.Builder
	for _, block := range blocks {
		result.WriteString(BlockToMarkdown(&block, indent))
	}
	return result.String()
}
