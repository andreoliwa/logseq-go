package markdown_test

import (
	"github.com/andreoliwa/logseq-go/internal/markdown"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func parseAndOutput(input string) string {
	block, err := markdown.ParseString(input)
	Expect(err).ToNot(HaveOccurred())
	v, err := markdown.AsString(block)
	Expect(err).ToNot(HaveOccurred())
	return v
}

func FullyEqual(name string, input string) {
	It(name, func() {
		v := parseAndOutput(input)
		Expect(v).To(Equal(input))
	})
}

func Varies(name string, input string, output string) {
	It(name, func() {
		v := parseAndOutput(input)
		Expect(v).To(Equal(output))
	})
}

var _ = Describe("Parsing then outputting", func() {
	Describe("Paragraphs", func() {
		FullyEqual("Paragraph", "Basic content")
		FullyEqual("Paragraph with soft newline", "Basic\ncontent")
		FullyEqual("Paragraph with hard newline via backslash", "Basic\\\ncontent")
		Varies("Paragraph with hard newline via two spaces", "Basic  \ncontent", "Basic\\\ncontent")

		FullyEqual("Multiple paragraphs", "Basic content\n\nMore content")
	})

	Describe("Inline formatting", func() {
		FullyEqual("Bold text", "**Basic** content")
		FullyEqual("Bold text with newline", "**Basic\ncontent**")
		FullyEqual("Bold text with hard newline", "**Basic\\\ncontent**")
		Varies("Bold text with hard newline via two spaces", "**Basic  \ncontent**", "**Basic\\\ncontent**")

		FullyEqual("Italic text with stars", "*Basic* content")
		FullyEqual("Italic text with underscore", "_Basic_ content")
		FullyEqual("Italic text with newline", "*Basic\ncontent*")
		FullyEqual("Italic text with hard newline", "*Basic\\\ncontent*")
		Varies("Italic text with hard newline via two spaces", "*Basic  \ncontent*", "*Basic\\\ncontent*")

		// Edge cases for emphasis character preservation
		FullyEqual("Multiple underscore emphasis", "_first_ and _second_ underscore")
		FullyEqual("Multiple star emphasis", "*first* and *second* star")
		FullyEqual("Mixed underscore and star emphasis", "_underscore_ and *star* mixed")
		FullyEqual("Bold and underscore italic", "**bold** and _italic_")
		FullyEqual("Bold and star italic", "**bold** and *italic*")
		FullyEqual("Underscore emphasis at start", "_start_ of text")
		FullyEqual("Underscore emphasis at end", "text at _end_")
		FullyEqual("Star emphasis at start", "*start* of text")
		FullyEqual("Star emphasis at end", "text at *end*")

		FullyEqual("Strikethrough text", "~~Basic~~ content")
		FullyEqual("Strikethrough text with newline", "~~Basic\ncontent~~")
		FullyEqual("Strikethrough text with hard newline", "~~Basic\\\ncontent~~")
		Varies("Strikethrough text with hard newline via two spaces", "~~Basic  \ncontent~~", "~~Basic\\\ncontent~~")
		FullyEqual("Strikethrough text containing escaped ~~", "~~Bas~\\~ic~~ content")

		// Code text maintains spaces and newlines
		FullyEqual("Code text", "`Basic` content")
		FullyEqual("Code text maintains newline", "`Basic\ncontent`")
		FullyEqual("Code text maintains spaces before 'hard' newline", "`Basic  \ncontent`")
	})

	Describe("Heading", func() {
		FullyEqual("Heading 1", "# Heading")
		FullyEqual("Heading 2", "## Heading")
		FullyEqual("Heading 3", "### Heading")
		FullyEqual("Heading 4", "#### Heading")
		FullyEqual("Heading 5", "##### Heading")
		FullyEqual("Heading 6", "###### Heading")
	})

	Describe("Heading combined with paragraph", func() {
		FullyEqual("Heading 1", "# Heading\n\nParagraph")
		FullyEqual("Heading 2", "## Heading\n\nParagraph")
		FullyEqual("Heading 3", "### Heading\n\nParagraph")
		FullyEqual("Heading 4", "#### Heading\n\nParagraph")
		FullyEqual("Heading 5", "##### Heading\n\nParagraph")
		FullyEqual("Heading 6", "###### Heading\n\nParagraph")
	})

	Describe("Code blocks", func() {
		FullyEqual("Code block", "```go\nfunc main() {\n\tfmt.Println(\"Hello world\")\n}\n```")
		FullyEqual("Code block with newline", "```go\nfunc main() {\n\tfmt.Println(\"Hello world\")\n}\n```\n\nParagraph")

		FullyEqual("Code block after paragraph", "Paragraph\n\n```go\nfunc main() {\n\tfmt.Println(\"Hello world\")\n}\n```")
		FullyEqual("Code block interrupting paragraph", "Paragraph\n```go\nfunc main() {\n\tfmt.Println(\"Hello world\")\n}\n```")
	})

	Describe("Block quotes", func() {
		FullyEqual("Block quote", "> This is a blockquote")
		FullyEqual("Block quote with multiple spaces", ">    This is a blockquote with spaces")
		FullyEqual("Block quote with page links", "> This is a blockquote with [[page link]] and text after")
		FullyEqual("Block quote with hashtag", "> This is a blockquote with #hashtag and text after")
		FullyEqual("Block quote with bold text", "> This is a blockquote with **bold text** and text after")
		FullyEqual("Block quote with italic text", "> This is a blockquote with *italic text* and text after")
		FullyEqual("Block quote with strikethrough text", "> This is a blockquote with ~~strikethrough text~~ and text after")
		FullyEqual("Block quote with code text", "> This is a blockquote with `code text` and text after")
		FullyEqual("Block quote with link", "> This is a blockquote with [link](https://example.com) and text after")
	})

	Describe("Macros", func() {
		FullyEqual("Macro with no arguments", "{{poem}}")
		FullyEqual("Macro with one argument", "{{poem red}}")
		FullyEqual("Macro with two arguments", "{{poem red, blue}}")
		Varies("Macro with two arguments, no space", "{{poem red,blue}}", "{{poem red, blue}}")
		FullyEqual("Macro with one argument and spaces", "{{poem red blue}}")
		FullyEqual("Macro with quoted argument with comma", "{{poem \"red, blue\"}}")

		Describe("Invalid", func() {
			FullyEqual("Macro without end", "{{poem red blue")
		})
	})

	Describe("Properties", func() {
		FullyEqual("Single property", "key:: value")
		FullyEqual("Multiple properties", "key:: value\nkey2:: value2")
		FullyEqual("Properties followed by trailing paragraph", "key:: value\nParagraph")
		FullyEqual("Paragraphs interrupted by properties", "Paragraph\nkey:: value")
		FullyEqual("Paragraphs interrupted by properties followed by more paragraph", "Paragraph\nkey:: value\nParagraph")
		FullyEqual("Paragraph followed by properties", "Paragraph\n\nkey:: value")
	})

	Describe("Links", func() {
		Describe("Markdown links with hashtags", func() {
			FullyEqual("URL with hashtag anchor", "- Any link with the hashtag symbol like [http://example.com/path/to/page#anchor](http://example.com/path/to/page#anchor) should be preserved")
			FullyEqual("Link text with hashtag", "- A link to a Slack channel like [#random](https://some-company.slack.com/archives/D01234NC6DEF) should not have the closing bracket escaped")
			FullyEqual("Regular hashtag should still work", "- This is a regular #hashtag that should be preserved")
			FullyEqual("Multiple links with hashtags", "- Check [#general](https://slack.com/general) and [https://example.com#section](https://example.com#section)")
			FullyEqual("Link with hashtag in both text and URL", "- Link [#anchor](https://example.com#anchor) has hashtag in both parts")
			FullyEqual("Complex URL with multiple hashtags", "- Complex [link](https://example.com/path#section1#section2) with multiple hashtags")
			FullyEqual("Hashtag at start of link text", "- Start with [#hashtag text](https://example.com) in link")
			FullyEqual("Hashtag at end of link text", "- End with [text #hashtag](https://example.com) in link")
			FullyEqual("Multiple hashtags in link text", "- Multiple [#first #second #third](https://example.com) hashtags in text")
			FullyEqual("URL with query parameters and hashtag", "- URL with [query params](https://example.com/path?param=value#anchor) and hashtag")
			FullyEqual("Mixed regular and link hashtags", "- Regular #tag and [#link-tag](https://example.com) mixed together")
			FullyEqual("Nested brackets with hashtags", "- Nested [text with [nested] and #hashtag](https://example.com) content")
		})

		Describe("Edge cases", func() {
			FullyEqual("Hashtag immediately after link", "- Link [text](url)#hashtag should work")
			FullyEqual("Hashtag immediately before link", "- Text #hashtag[link](url) should work")
			FullyEqual("Multiple links on same line", "- Multiple [#first](url1) and [#second](url2) links")
			Varies("Link with escaped brackets in text", "- Link [text with \\[escaped\\] brackets #tag](url) should work", "- Link [text with [escaped] brackets #tag](url) should work")
			FullyEqual("Link with special characters and hashtag", "- Special [chars & symbols #tag word_with_underscores](https://example.com/path?a=1&b=2#anchor) in link")
		})

		Describe("URLs with parentheses", func() {
			FullyEqual("URL with parentheses should not be escaped", "[93 (Thelema) | Symbolism Wiki | Fandom](https://symbolism.fandom.com/wiki/93_(Thelema))")
			FullyEqual("URL with multiple parentheses", "[Wikipedia article](https://en.wikipedia.org/wiki/Function_(mathematics))")
			FullyEqual("URL with nested parentheses", "[Complex URL](https://example.com/path/(nested(content)))")
		})
	})

	Describe("Tasks", func() {
		Describe("Markers", func() {
			FullyEqual("TODO Task", "TODO Task")
			FullyEqual("DOING Task", "DOING Task")
			FullyEqual("DONE Task", "DONE Task")
			FullyEqual("LATER Task", "LATER Task")
			FullyEqual("NOW Task", "NOW Task")
			FullyEqual("CANCELLED Task", "CANCELLED Task")
			FullyEqual("CANCELED Task", "CANCELED Task")
			FullyEqual("IN-PROGRESS Task", "IN-PROGRESS Task")
			FullyEqual("WAIT Task", "WAIT Task")
			FullyEqual("WAITING Task", "WAITING Task")

			FullyEqual("Mixed case Todo", "Todo mundo")
			FullyEqual("Lowercase later", "later tonight")
			FullyEqual("Mixed case Now", "Now or never")

			Varies("Task with leading space", " TODO Task", "TODO Task")
		})
	})

	Describe("Priorities", func() {
		FullyEqual("Uppercase A", "[#A] First")
		FullyEqual("Uppercase B", "[#B] Second")
		FullyEqual("Uppercase C", "[#C] Third")

		FullyEqual("Lowercase a", "[#a] First")
		FullyEqual("Lowercase b", "[#b] Second")
		FullyEqual("Lowercase c", "[#c] Third")

		FullyEqual("After task marker", "DOING [#A] Task")

		FullyEqual("Only the first priority is parsed", "[#A] Only one priority per line [#B] [#C]")

		FullyEqual("Invalid letter", "[#D] Invalid")
		FullyEqual("Invalid letter without space", "[#D]Invalid")

		// Spaces
		FullyEqual("Preserve spaces", "[#A]     Preserve")
		FullyEqual("Valid letter but no space", "[#A]Valid")
		FullyEqual("Space in the middle", "[# A] Middle")
		FullyEqual("Space before", "[ #B] Before")
		FullyEqual("Space after", "[#B ] After")
		FullyEqual("Space before and after", "[ #C ] Before and after")
	})
})
