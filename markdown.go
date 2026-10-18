package logseq

import (
	"io"

	"github.com/andreoliwa/logseq-go/v2/content"
	"github.com/andreoliwa/logseq-go/v2/internal/markdown"
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

// WithStrictOutput selects upstream formatting and disables preservation.
// Deprecated: upstream formatting is the default.
func WithStrictOutput() MarkdownOption { return markdown.WithStrictOutput() }

// AsMarkdownWithOptions serializes with explicit formatting choices.
func AsMarkdownWithOptions(node content.Node, opts ...MarkdownOption) (string, error) {
	return markdown.AsString(node, opts...)
}

// WithPreservedMarkdown retains source formatting for one serialization call.
func WithPreservedMarkdown() MarkdownOption { return markdown.WithPreservedOutput() }

// WriteMarkdownWithOptions serializes with explicit formatting choices,
// propagating errors from the destination writer.
func WriteMarkdownWithOptions(node content.Node, w io.Writer, opts ...MarkdownOption) error {
	return markdown.NewWriter(w, opts...).Write(node)
}
