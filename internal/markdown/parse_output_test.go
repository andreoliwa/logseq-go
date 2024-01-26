package markdown_test

import (
	"github.com/aholstenson/logseq-go/internal/markdown"
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

		FullyEqual("Italic text", "*Basic* content")
		FullyEqual("Italic text with newline", "*Basic\ncontent*")
		FullyEqual("Italic text with hard newline", "*Basic\\\ncontent*")
		Varies("Italic text with hard newline via two spaces", "*Basic  \ncontent*", "*Basic\\\ncontent*")

		FullyEqual("Strikethrough text", "~~Basic~~ content")
		FullyEqual("Strikethrough text with newline", "~~Basic\ncontent~~")
		FullyEqual("Strikethrough text with hard newline", "~~Basic\\\ncontent~~")
		Varies("Strikethrough text with hard newline via two spaces", "~~Basic  \ncontent~~", "~~Basic\\\ncontent~~")

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
})