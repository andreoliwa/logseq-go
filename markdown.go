package logseq

import (
	"io"

	"github.com/andreoliwa/logseq-go/content"
	"github.com/andreoliwa/logseq-go/internal/markdown"
)

// AsMarkdown serializes a content node back to a Markdown string.
func AsMarkdown(node content.Node) (string, error) {
	return markdown.AsString(node)
}

// WriteMarkdown serializes a content node as Markdown, writing to the given writer.
func WriteMarkdown(node content.Node, w io.Writer) error {
	return markdown.Write(node, w)
}

// MarkdownOption configures the additive serialization entry point.
type MarkdownOption = markdown.Option

// WithStrictOutput opts into stricter escaping of literal Markdown.
func WithStrictOutput() MarkdownOption { return markdown.WithStrictOutput() }

// AsMarkdownWithOptions serializes with explicit formatting choices.
func AsMarkdownWithOptions(node content.Node, opts ...MarkdownOption) (string, error) {
	return markdown.AsString(node, opts...)
}
