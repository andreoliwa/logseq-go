package logseq

import (
	"github.com/andreoliwa/logseq-go/content"
	"github.com/andreoliwa/logseq-go/internal/markdown"
)

// ParseBlock parses markdown into a block.
func ParseBlock(text string) (*content.Block, error) {
	return markdown.ParseString(text)
}

// ParseNodes parses markdown into a list of nodes.
func ParseNodes(text string) (content.NodeList, error) {
	block, err := ParseBlock(text)
	if err != nil {
		return nil, err
	}

	return block.Children(), nil
}

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
