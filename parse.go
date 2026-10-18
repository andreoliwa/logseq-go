package logseq

import (
	"github.com/andreoliwa/logseq-go/v2/content"
	"github.com/andreoliwa/logseq-go/v2/internal/markdown"
)

// ParseBlock parses markdown into a block, read the way Logseq reads a graph
// that configures nothing. Use [Graph.ParseBlock] for markdown that belongs to
// a graph, so that the settings of that graph are used instead.
func ParseBlock(text string) (*content.Block, error) {
	return markdown.ParseString(text)
}

// ParseBlock parses markdown into a block as the settings of the graph say it
// should be read, which is the same way opening a page reads it.
func (g *Graph) ParseBlock(text string) (*content.Block, error) {
	return markdown.ParseString(text, g.markdownParseOptions()...)
}

// ParseNodes parses markdown into a list of nodes, read the way Logseq reads a
// graph that configures nothing. Use [Graph.ParseNodes] for markdown that
// belongs to a graph, so that the settings of that graph are used instead.
func ParseNodes(text string) (content.NodeList, error) {
	block, err := ParseBlock(text)
	if err != nil {
		return nil, err
	}

	return block.Children(), nil
}

// ParseNodes parses markdown into a list of nodes as the settings of the graph
// say it should be read.
func (g *Graph) ParseNodes(text string) (content.NodeList, error) {
	block, err := g.ParseBlock(text)
	if err != nil {
		return nil, err
	}

	return block.Children(), nil
}

// AsString converts a node into Markdown, written the way Logseq writes a
// graph that configures nothing. Use [Graph.AsString] for a node that belongs
// to a graph, so that the settings of that graph are used instead.
func AsString(node content.Node) (string, error) {
	return markdown.AsString(node)
}

// BlockSourceRange identifies a root outline block's half-open source byte range.
// The range includes metadata, continuation lines and descendants.
type BlockSourceRange = markdown.BlockSourceRange

// ParseSource parses Markdown and returns raw source ranges for root dash
// outlines. outlined is false for unbulleted input or a leading prose prefix;
// ranges still identify root outlines after a prefix. Leading whitespace and
// page properties do not disqualify an outline. Bytes before the first range
// belong to the page. Parsing does not normalize or serialize the source.
// On range errors, the parsed root and outlined flag remain available; callers
// needing single-block semantics can ignore ranges when outlined is false.
func ParseSource(source string) (*content.Block, []BlockSourceRange, bool, error) {
	return markdown.ParseSource(source)
}

// AsString converts a node into Markdown as the settings of the graph say it
// should be written, which is the same way saving a page writes it.
func (g *Graph) AsString(node content.Node) (string, error) {
	return markdown.AsString(node, g.markdownOptions()...)
}

// markdownOptions are the options for writing Markdown that match the settings
// of the graph.
func (g *Graph) markdownOptions() []markdown.Option {
	opts := []markdown.Option{markdown.WithLogbookSeconds(g.config.Logbook.WithSecondSupport)}
	if g.options.strictMarkdownOutput {
		opts = append(opts, markdown.WithStrictOutput())
	}
	return opts
}

// markdownParseOptions are the options for reading Markdown that match the
// settings of the graph.
func (g *Graph) markdownParseOptions() []markdown.ParseOption {
	opts := append([]markdown.ParseOption(nil), g.options.parseOptions...)
	return append(opts,
		markdown.WithPropertiesSeparatedByCommas(g.config.PropertiesSeparatedByCommas...),
		markdown.WithIgnoredPageReferences(g.config.IgnoredPageReferencesKeywords...),
	)
}

// ParseOption configures the additive parsing entry point.
type ParseOption = markdown.ParseOption

// WithTypedLogbooks opts into structured CLOCK and state-change nodes.
func WithTypedLogbooks() ParseOption { return markdown.WithTypedLogbooks() }

// WithStrictPriorities opts into position-sensitive Logseq priority syntax.
func WithStrictPriorities() ParseOption { return markdown.WithStrictPriorities() }

// WithBuiltInPropertyLists opts into implicit page references in built-in lists.
func WithBuiltInPropertyLists() ParseOption { return markdown.WithBuiltInPropertyLists() }

// ParseBlockWithOptions parses with explicit syntax or representation choices.
func ParseBlockWithOptions(text string, opts ...ParseOption) (*content.Block, error) {
	return markdown.ParseString(text, opts...)
}
