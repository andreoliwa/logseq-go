package markdown

import (
	"fmt"

	"github.com/andreoliwa/logseq-go/v2/content"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// BlockSourceRange identifies a root block's half-open byte range in the source.
// It includes descendants and all supporting content up to the next root block.
type BlockSourceRange struct {
	Block *content.Block
	Start int
	End   int
}

const sourceStartAttribute = "logseq-source-start"

// sourceListItemParser retains the existing parser's decisions, including empty
// items and markers inside fences, while recording original byte positions.
type sourceListItemParser struct{ parser.BlockParser }

func (p *sourceListItemParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	_, position := reader.Position()
	start := position.Start
	source := reader.Source()
	for start > 0 && source[start-1] != '\n' {
		start--
	}
	node, state := p.BlockParser.Open(parent, reader, pc)
	if node != nil {
		node.SetAttributeString(sourceStartAttribute, start)
	}
	return node, state
}

// ParseSource returns the ordinary parsed content and source ranges for root
// dash outlines. outlined reports whether the input starts with an outline,
// allowing leading whitespace and page properties but not ordinary prose.
// Ranges are available even when a prose prefix makes outlined false. Source
// before the first range belongs to the page, not to its first root block.
// On range errors, root and outlined remain available, but ranges is nil.
func ParseSource(source string) (*content.Block, []BlockSourceRange, bool, error) {
	root, err := ParseString(source)
	if err != nil {
		return nil, nil, false, err
	}
	// Parse the original bytes for offsets. ParseString preprocesses escaped
	// tildes, which can change byte lengths; its AST remains the public result.
	doc := markdownParser.Parse(text.NewReader([]byte(source)))
	var starts []int
	outlined := true
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		if list, ok := child.(*ast.List); ok && list.Marker == '-' {
			for item := list.FirstChild(); item != nil; item = item.NextSibling() {
				start, ok := item.AttributeString(sourceStartAttribute)
				if !ok {
					return root, nil, outlined, fmt.Errorf("missing outline source position")
				}
				starts = append(starts, start.(int))
			}
		} else if len(starts) == 0 {
			if _, ok := child.(*properties); !ok {
				outlined = false
			}
		}
	}
	if len(starts) == 0 {
		return root, nil, false, nil
	}
	var blocks []*content.Block
	for child := root.FirstChild(); child != nil; child = child.NextSibling() {
		if block, ok := child.(*content.Block); ok {
			blocks = append(blocks, block)
		}
	}
	if len(blocks) != len(starts) {
		return root, nil, outlined, fmt.Errorf("outline source ranges do not match parsed blocks")
	}
	ranges := make([]BlockSourceRange, len(starts))
	for i, start := range starts {
		end := len(source)
		if i+1 < len(starts) {
			end = starts[i+1]
		}
		if start < 0 || start >= end || end > len(source) {
			return root, nil, outlined, fmt.Errorf("invalid outline source range [%d:%d]", start, end)
		}
		ranges[i] = BlockSourceRange{Block: blocks[i], Start: start, End: end}
	}
	return root, ranges, outlined, nil
}
