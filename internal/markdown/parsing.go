package markdown

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/andreoliwa/logseq-go/v2/content"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var markdownParser parser.Parser

var neverMatch = regexp.MustCompile(`$.^`)

func init() {
	markdownParser = parser.NewParser(
		parser.WithBlockParsers(
			util.Prioritized(parser.NewThematicBreakParser(), 200),
			util.Prioritized(parser.NewListParser(), 300),
			util.Prioritized(&sourceListItemParser{parser.NewListItemParser()}, 400),
			util.Prioritized(parser.NewCodeBlockParser(), 500),
			util.Prioritized(parser.NewATXHeadingParser(), 600),
			util.Prioritized(parser.NewFencedCodeBlockParser(), 700),
			util.Prioritized(parser.NewBlockquoteParser(), 800),
			util.Prioritized(&mathBlockParser{}, 896),
			util.Prioritized(&taskDateParser{}, 897),
			util.Prioritized(&logbookParser{}, 898),
			util.Prioritized(&beginEndParser{}, 899),
			util.Prioritized(parser.NewHTMLBlockParser(), 900),
			util.Prioritized(&footnoteDefinitionParser{}, 999),
			util.Prioritized(parser.NewParagraphParser(), 1000),
		),
		parser.WithInlineParsers(
			util.Prioritized(parser.NewCodeSpanParser(), 100),
			util.Prioritized(&footnoteRefParser{}, 194),
			util.Prioritized(&mathParser{}, 195),
			util.Prioritized(&strictPriorityParser{}, 196),
			util.Prioritized(&macroParser{}, 197),
			util.Prioritized(&blockRefParser{}, 198),
			util.Prioritized(&pageLinkParser{}, 199),
			util.Prioritized(&tagParser{}, 999),
			util.Prioritized(parser.NewLinkParser(), 200),
			util.Prioritized(parser.NewAutoLinkParser(), 300),
			util.Prioritized(parser.NewRawHTMLParser(), 400),
			util.Prioritized(parser.NewEmphasisParser(), 500),
			util.Prioritized(extension.NewStrikethroughParser(), 501),
			util.Prioritized(&highlightParser{}, 502),
			util.Prioritized(extension.NewLinkifyParser(
				extension.WithLinkifyEmailRegexp(neverMatch),
				extension.WithLinkifyWWWRegexp(neverMatch),
			), 998),
		),
		parser.WithParagraphTransformers(
			util.Prioritized(parser.LinkReferenceParagraphTransformer, 100),
			util.Prioritized(newTableParagraphTransformer(extension.NewTableParagraphTransformer()), 200),
		),
		parser.WithASTTransformers(
			util.Prioritized(extension.NewTableASTTransformer(), 0),
			util.Prioritized(defaultPropertiesASTTransformer, 200),
		),
	)
}

// ParseOption changes how Markdown is read for the parts of the syntax that a
// graph configures the shape of.
type ParseOption func(*parseOptions)

// parseOptionsKey is where the options of the current parse are kept, so that
// the parts of parsing that Goldmark drives can reach them.
var parseOptionsKey = parser.NewContextKey()

// builtInPropertiesSeparatedByCommas are the properties whose value Logseq
// always reads as a list of pages separated by commas, no matter what the
// graph configures.
var builtInPropertiesSeparatedByCommas = []string{"alias", "aliases", "tags"}

// parseOptions are the settings that parsing takes from the graph. The zero
// value is not usable; use defaultParseOptions for the compatible defaults.
type parseOptions struct {
	// propertiesSeparatedByCommas are the properties whose value is a list of
	// pages separated by commas.
	propertiesSeparatedByCommas map[string]struct{}

	// ignoredPageReferences are the properties whose value never points at a
	// page, even where it is written as a link.
	ignoredPageReferences map[string]struct{}
	typedLogbooks         bool
}

// defaultParseOptions use typed logbooks and Logseq built-in reference lists.
func defaultParseOptions() parseOptions {
	options := parseOptions{
		propertiesSeparatedByCommas: make(map[string]struct{}),
		ignoredPageReferences:       make(map[string]struct{}),
		typedLogbooks:               true,
	}

	for _, name := range builtInPropertiesSeparatedByCommas {
		options.propertiesSeparatedByCommas[name] = struct{}{}
	}
	return options
}

// WithTypedLogbooks opts into structured CLOCK and state-change nodes.
func WithTypedLogbooks() ParseOption { return func(o *parseOptions) { o.typedLogbooks = true } }

// WithRawLogbooks keeps every logbook entry as raw text rather than a typed node.
func WithRawLogbooks() ParseOption { return func(o *parseOptions) { o.typedLogbooks = false } }

// WithStrictPriorities is retained for callers; positional syntax is always enabled.
// Deprecated: priorities are position-sensitive by default.
func WithStrictPriorities() ParseOption { return func(o *parseOptions) {} }

// WithBuiltInPropertyLists enables the default implicit page references in Logseq lists.
func WithBuiltInPropertyLists() ParseOption {
	return func(o *parseOptions) {
		for _, name := range builtInPropertiesSeparatedByCommas {
			o.propertiesSeparatedByCommas[name] = struct{}{}
		}
	}
}

// WithPropertiesSeparatedByCommas adds properties whose value is a list of
// pages separated by commas, which is `:property/separated-by-commas`. The
// built-in lists require WithBuiltInPropertyLists.
func WithPropertiesSeparatedByCommas(names ...string) ParseOption {
	return func(o *parseOptions) {
		for _, name := range names {
			o.propertiesSeparatedByCommas[normalizePropertyName(name)] = struct{}{}
		}
	}
}

// WithIgnoredPageReferences adds properties whose value does not point at a
// page, which is `:ignored-page-references-keywords`.
func WithIgnoredPageReferences(names ...string) ParseOption {
	return func(o *parseOptions) {
		for _, name := range names {
			o.ignoredPageReferences[normalizePropertyName(name)] = struct{}{}
		}
	}
}

// isSeparatedByCommas checks if the value of the property with the given name
// is a list of pages separated by commas.
func (o *parseOptions) isSeparatedByCommas(name string) bool {
	_, ok := o.propertiesSeparatedByCommas[normalizePropertyName(name)]
	return ok
}

// ignoresPageReferences checks if the value of the property with the given
// name points at pages or is only text.
func (o *parseOptions) ignoresPageReferences(name string) bool {
	_, ok := o.ignoredPageReferences[normalizePropertyName(name)]
	return ok
}

// normalizePropertyName brings a property name into the form it is matched in,
// as Logseq does not distinguish between property names that only differ in
// case.
func normalizePropertyName(name string) string {
	return strings.ToLower(name)
}

func Parse(src []byte, opts ...ParseOption) (*content.Block, error) {
	options := defaultParseOptions()
	for _, opt := range opts {
		opt(&options)
	}

	context := parser.NewContext()
	context.Set(parseOptionsKey, &options)

	doc := markdownParser.Parse(text.NewReader(src), parser.WithContext(context))
	node, err := convert(src, doc)
	if err != nil {
		return nil, fmt.Errorf("Could not parse Markdown: %w", err)
	}

	if b, ok := node.(*content.Block); ok {
		return b, nil
	}

	return nil, errors.New("Could not parse Markdown")
}

func ParseString(src string, opts ...ParseOption) (*content.Block, error) {
	// Pre-process the input to handle escaped tildes in strikethrough
	preprocessed := preprocessEscapedTildes(src)
	block, err := Parse([]byte(preprocessed), opts...)
	if err != nil {
		return nil, err
	}

	// Post-process to restore escaped tildes
	postprocessEscapedTildes(block)
	return block, nil
}

// Placeholder for escaped tildes during parsing
const escapedTildePlaceholder = "\uE000" // Private Use Area character

func preprocessEscapedTildes(input string) string {
	// Replace \~ with a placeholder to prevent Goldmark from misinterpreting it
	return strings.ReplaceAll(input, "\\~", escapedTildePlaceholder)
}

func postprocessEscapedTildes(node content.Node) {
	// Recursively process all text nodes to restore escaped tildes
	if textNode, ok := node.(*content.Text); ok {
		// Replace the placeholder with just the tilde (unescaped)
		// This is correct for parsing - escaped tildes should become literal tildes
		textNode.Value = strings.ReplaceAll(textNode.Value, escapedTildePlaceholder, "~")
	}

	if hasChildren, ok := node.(content.HasChildren); ok {
		for child := hasChildren.FirstChild(); child != nil; child = child.NextSibling() {
			postprocessEscapedTildes(child)
		}
	}
}

// convert converts from the Goldmark AST into our AST.
func convert(src []byte, in ast.Node) (content.Node, error) {
	switch node := in.(type) {
	case *ast.Document:
		return convertToBlock(src, node)
	case *ast.Heading:
		return convertHeading(src, node)
	case *ast.Paragraph:
		return convertParagraph(src, node)
	case *ast.TextBlock:
		return convertTextBlock(src, node)
	case *ast.Text:
		return convertText(src, node)
	case *ast.Emphasis:
		return convertEmphasis(src, node)
	case *east.Strikethrough:
		return convertStrikethrough(src, node)
	case *highlight:
		return convertHighlight(src, node)
	case *footnoteRef:
		return content.NewFootnoteRef(node.Label), nil
	case *footnoteDefinition:
		return convertFootnoteDefinition(src, node)
	case *math:
		return content.NewMath(node.Value).WithDisplayed(node.Displayed), nil
	case *mathBlock:
		return convertMathBlock(src, node)
	case *ast.CodeSpan:
		return convertCodeSpan(src, node)
	case *ast.Link:
		return convertLink(src, node)
	case *ast.AutoLink:
		return content.NewAutoLink(string(node.URL(src))), nil
	case *tag:
		return content.NewHashtag(node.Page), nil
	case *pageLink:
		return content.NewPageLink(node.Page), nil
	case *pageRefText:
		return content.NewPageRefText(unescapeString(node.Segment.Value(src))).CaptureSource(string(node.Segment.Value(src))), nil
	case *blockRef:
		return content.NewBlockRef(node.ID), nil
	case *strictPriority:
		value := content.NewTaskPriority(node.Priority)
		value.OriginalText = node.OriginalText
		return value, nil
	case *macro:
		return convertMacro(src, node)
	case *ast.FencedCodeBlock:
		return convertFencedCodeBlock(src, node)
	case *ast.CodeBlock:
		return convertCodeBlock(src, node)
	case *ast.Blockquote:
		return convertBlockquote(src, node)
	case *ast.List:
		return convertList(src, node)
	case *east.Table:
		return convertTable(src, node)
	case *ast.RawHTML:
		return convertRawHTML(src, node)
	case *ast.HTMLBlock:
		return convertHTMLBlock(src, node)
	case *ast.Image:
		return convertImage(src, node)
	case *ast.ThematicBreak:
		return content.NewThematicBreak(), nil
	case *beginEnd:
		return convertBeginEnd(src, node)
	case *properties:
		return convertProperties(src, node)
	case *logbook:
		return convertLogbook(src, node)
	case *taskDate:
		return convertTaskDate(node)
	}

	return nil, fmt.Errorf("Could not convert node: %T", in)
}

func unescapeString(src []byte) string {
	if len(src) == 0 {
		return ""
	}

	var result []byte
	for i := 0; i < len(src); i++ {
		if src[i] == '\\' && i+1 < len(src) {
			next := src[i+1]
			// Check if the next character is escapable according to CommonMark
			if isEscapableCharacter(next) {
				result = append(result, next)
				i++ // Skip the next character since we've processed it
			} else {
				result = append(result, src[i])
			}
		} else {
			result = append(result, src[i])
		}
	}
	return string(result)
}

// isEscapableCharacter returns true if the character can be escaped with a backslash
// according to the CommonMark specification
func isEscapableCharacter(c byte) bool {
	// CommonMark escapable characters: !"#$%&'()*+,-./:;<=>?@[\]^_`{|}~
	switch c {
	case '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
		':', ';', '<', '=', '>', '?', '@', '[', '\\', ']', '^', '_', '`', '{', '|', '}', '~':
		return true
	default:
		return false
	}
}

// convertChildren converts all children of a node and adds them to the target.
func convertChildren(src []byte, node ast.Node, target content.HasChildren) error {
	var previousChild ast.Node
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		node, err := convert(src, child)
		if err != nil {
			return err
		}

		// If the previous child was a text node and the current child is a text
		// node can merge them in certain cases. This makes the output easier
		// to work with.
		if pText, pOk := previousChild.(*ast.Text); pOk {
			if text, ok := child.(*ast.Text); ok && canMergeTextNodes(pText, text) {
				previousNode := target.LastChild().(*content.Text)
				previousNode.Value += unescapeString(text.Segment.Value(src))
				previousChild = child

				previousNode.HardLineBreak = text.HardLineBreak()
				previousNode.SoftLineBreak = text.SoftLineBreak()
				continue
			}

		}

		target.AddChild(node)
		previousChild = child
	}

	return nil
}

func canMergeTextNodes(a *ast.Text, b *ast.Text) bool {
	if a.Segment.Stop != b.Segment.Start {
		return false
	}

	return true
}

// convertToBlock implements the Logseq outlining, where lists using `-` are
// converted into sub-blocks.
func convertToBlock(src []byte, node ast.Node) (*content.Block, error) {
	block := content.NewBlock()

	hasParsedBlock := false
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		// Check if the child is a list of blocks.
		if list, ok := child.(*ast.List); ok && list.Marker == '-' {
			// This is a list using the block marker, so children are parsed
			// as blocks.
			for listItem := list.FirstChild(); listItem != nil; listItem = listItem.NextSibling() {
				// Check if this list item contains literal extra dashes that should be preserved
				if shouldPreserveLiteralDashes(src, listItem) {
					// Convert this list item to a paragraph with literal dashes instead of nested blocks
					paragraph, err := convertListItemToLiteralDashes(src, listItem)
					if err != nil {
						return nil, err
					}
					listItemBlock := content.NewBlock(paragraph)
					block.AddChild(listItemBlock)
				} else {
					listItemBlock, err := convertToBlock(src, listItem)
					if err != nil {
						return nil, err
					}

					// If this is the first list item AND there's content before the list,
					// AND the list item has blank lines before it, set PreviousLineType on the first child.
					// Used when parsing metadata like aliases and filters.
					if listItem == list.FirstChild() && block.FirstChild() != nil && listItem.HasBlankPreviousLines() && listItemBlock.FirstChild() != nil {
						if pla, ok := listItemBlock.FirstChild().(content.PreviousLineAware); ok {
							pla.SetPreviousLineType(content.PreviousLineTypeBlank)
						}
					}

					block.AddChild(listItemBlock)
				}
			}

			hasParsedBlock = true
		} else {
			node, err := convert(src, child)
			if err != nil {
				return nil, err
			}

			// When converting a block the first paragraph might contain a task marker
			if p, ok := node.(*content.Paragraph); ok && block.FirstChild() == nil {
				convertTaskMarker(p, block)
			}

			if !hasParsedBlock {
				// No list of blocks encountered yet, treat nodes as content
				// to main block.
				block.AddChild(node)

				if p, ok := node.(*content.Paragraph); ok {
					// Logseq handles a "-" without a trailing space as block
					// but Goldmark will parse it as a text item in a paragraph
					var previousNode content.Node
					var previousNodeBreak bool
					for _, paragraphNode := range p.Children() {
						if text, ok := paragraphNode.(*content.Text); ok {
							if text.Value == "-" && previousNodeBreak {
								text.RemoveSelf()

								newBlock := content.NewBlock()
								block.AddChild(newBlock)

								if previousText, ok := previousNode.(*content.Text); ok {
									previousText.SoftLineBreak = false
									previousText.HardLineBreak = false
								}
							}

							previousNodeBreak = text.HardLineBreak || text.SoftLineBreak
						} else {
							previousNodeBreak = false
						}

						previousNode = paragraphNode
					}
				}
			} else {
				// We have parsed blocks so this is trailing content, in Logseq
				// this seems to be added to the last block.
				// However, if this is a thematic break that should be treated as a list item,
				// create a new block for it instead.
				if _, ok := node.(*content.ThematicBreak); ok {
					// Create a new block with a paragraph containing "---"
					newBlock := content.NewBlock(content.NewParagraph(content.NewText("---")))
					block.AddChild(newBlock)
				} else if lastBlock, ok := block.LastChild().(*content.Block); ok {
					// Find the node before the first block in sub-block.
					var lastNode content.Node
					for subNode := lastBlock.FirstChild(); subNode != nil; subNode = subNode.NextSibling() {
						if _, ok := subNode.(*content.Block); ok {
							break
						}
						lastNode = subNode
					}

					if lastNode == nil {
						// Likely no content, just add as the first child.
						lastBlock.PrependChild(node)
					} else {
						// Insert after the last node.
						lastBlock.InsertChildAfter(node, lastNode)
					}
				} else {
					return nil, errors.New("Last node is not a block")
				}
			}
		}
	}

	return block, nil
}

func convertTaskMarker(paragraph *content.Paragraph, block *content.Block) {
	textNode, ok := paragraph.FirstChild().(*content.Text)
	if !ok {
		return
	}

	potentialMarkerIdx := strings.Index(textNode.Value, " ")
	var potentialMarker string
	if potentialMarkerIdx < 0 {
		potentialMarker = textNode.Value
	} else {
		potentialMarker = textNode.Value[:potentialMarkerIdx]
	}

	taskStatus := content.ParseTaskStatus(potentialMarker)
	if taskStatus == content.TaskStatusNone {
		return
	}

	if potentialMarkerIdx < 0 {
		// The marker is the entire text, so there is nothing left of it.
		textNode.Value = ""
	} else {
		textNode.Value = textNode.Value[potentialMarkerIdx+1:]
	}

	if textNode.Value == "" {
		textNode.RemoveSelf()
	}

	taskMarker := content.NewTaskMarker(taskStatus)
	// Store parent references for use in TransitionTo transitions
	taskMarker.SetParentReferences(paragraph, block)
	paragraph.PrependChild(taskMarker)
}

// taskStatusFor maps a task marker to its status, returning TaskStatusNone if
// the word is not a marker.
func taskStatusFor(marker string) content.TaskStatus {
	switch marker {
	case "TODO":
		return content.TaskStatusTodo
	case "DONE":
		return content.TaskStatusDone
	case "DOING":
		return content.TaskStatusDoing
	case "LATER":
		return content.TaskStatusLater
	case "NOW":
		return content.TaskStatusNow
	case "CANCELLED":
		return content.TaskStatusCancelled
	case "CANCELED":
		return content.TaskStatusCanceled
	case "IN-PROGRESS":
		return content.TaskStatusInProgress
	case "WAIT":
		return content.TaskStatusWait
	case "WAITING":
		return content.TaskStatusWaiting
	}

	return content.TaskStatusNone
}

func convertParagraph(src []byte, node *ast.Paragraph) (*content.Paragraph, error) {
	paragraph := content.NewParagraph()
	err := convertChildren(src, node, paragraph)
	if err != nil {
		return nil, err
	}
	updatePreviousLine(node, paragraph)
	return paragraph, nil
}

func convertTextBlock(src []byte, node ast.Node) (*content.Paragraph, error) {
	paragraph := content.NewParagraph()
	err := convertChildren(src, node, paragraph)
	if err != nil {
		return nil, err
	}
	return paragraph, nil
}

func convertText(src []byte, node *ast.Text) (*content.Text, error) {
	value := unescapeString(node.Segment.Value(src))
	text := content.NewText(value)
	if isPropertyValue(node) {
		text.CaptureSource(string(node.Segment.Value(src)))
	}
	if node.SoftLineBreak() {
		text.SoftLineBreak = true
	} else if node.HardLineBreak() {
		text.HardLineBreak = true
	}
	return text, nil
}

func convertHeading(src []byte, node *ast.Heading) (*content.Heading, error) {
	heading := content.NewHeading(node.Level)
	err := convertChildren(src, node, heading)
	if err != nil {
		return nil, err
	}
	return heading, nil
}

func convertLink(src []byte, node *ast.Link) (*content.Link, error) {
	link := content.NewLink(unescapeString(node.Destination))
	err := convertChildren(src, node, link)
	if err != nil {
		return nil, err
	}
	link.Title = unescapeString(node.Title)
	link.OriginalURLAngleBrackets = propertyLinkUsesAngleBrackets(src, node)
	return link, nil
}

func convertMacro(src []byte, node *macro) (content.Node, error) {
	switch node.Name {
	case "query":
		// A query without a query string is not a query, keep it as a macro so
		// it is written back the way it was found.
		if len(node.Arguments) == 0 {
			break
		}

		return content.NewQuery(node.Arguments[0]), nil
	case "embed":
		// Either a [[page]] or a ((block))
		if len(node.Arguments) == 0 {
			break
		}

		arg := node.Arguments[0]
		if strings.HasPrefix(arg, "((") && strings.HasSuffix(arg, "))") {
			return content.NewBlockEmbed(arg[2 : len(arg)-2]), nil
		} else if strings.HasPrefix(arg, "[[") && strings.HasSuffix(arg, "]]") {
			return content.NewPageEmbed(arg[2 : len(arg)-2]), nil
		}
	case "cloze":
		// {{cloze answer \\ cue}} or {{cloze answer}}
		if len(node.Arguments) == 0 {
			break
		}

		arg := strings.Join(node.Arguments, ", ")
		cueIdx := strings.LastIndex(arg, "\\")
		if cueIdx < 0 {
			return content.NewCloze(
				strings.TrimSpace(arg),
			), nil
		} else {
			return content.NewClozeWithCue(
				strings.TrimSpace(arg[:cueIdx]),
				strings.TrimSpace(arg[cueIdx+1:]),
			), nil
		}
	}

	return content.NewMacro(node.Name, node.Arguments...), nil
}

func convertEmphasis(src []byte, node *ast.Emphasis) (content.HasChildren, error) {
	var result content.HasChildren
	if node.Level == 1 {
		emphasisChar := detectEmphasisCharacter(src, node)
		if isPropertyValue(node) {
			emphasisChar = detectPropertyEmphasisCharacter(src, node)
		}
		result = content.NewEmphasisWithCharacter(emphasisChar)
	} else if node.Level == 2 {
		strong := content.NewStrong()
		if isPropertyValue(node) {
			strong.Character = detectPropertyEmphasisCharacter(src, node)
		}
		result = strong
	} else {
		return nil, fmt.Errorf("Unsupported emphasis level: %d", node.Level)
	}

	err := convertChildren(src, node, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// detectEmphasisCharacter attempts to determine the original emphasis character
// used in the source text by examining the text around the emphasis node.
// This is necessary because Goldmark normalizes the emphasis characters to stars.
func detectEmphasisCharacter(src []byte, node *ast.Emphasis) rune {
	// Try to find the first text child to get position information
	firstChild := node.FirstChild()
	if firstChild == nil {
		return '*' // Default fallback
	}

	// If the first child is a text node, we can use its segment to find the emphasis character
	if textNode, ok := firstChild.(*ast.Text); ok {
		segment := textNode.Segment
		if segment.Start > 0 {
			// Look at the character just before the text content
			charBefore := src[segment.Start-1]
			if charBefore == '_' {
				return '_'
			}
		}
	}

	// Default to star if we can't determine the character
	return '*'
}

func convertStrikethrough(src []byte, node *east.Strikethrough) (*content.Strikethrough, error) {
	strikethrough := content.NewStrikethrough()
	err := convertChildren(src, node, strikethrough)
	if err != nil {
		return nil, err
	}
	return strikethrough, nil
}

func convertHighlight(src []byte, node *highlight) (*content.Highlight, error) {
	highlight := content.NewHighlight()
	err := convertChildren(src, node, highlight)
	if err != nil {
		return nil, err
	}
	return highlight, nil
}

func convertCodeSpan(src []byte, node *ast.CodeSpan) (*content.CodeSpan, error) {
	code := content.NewCodeSpan(string(node.Text(src)))
	if first, ok := node.FirstChild().(*ast.Text); ok && isPropertyValue(node) {
		start := first.Segment.Start
		// Goldmark may trim one padding space from each side of the value.
		if start > 0 && src[start-1] == ' ' {
			start--
		}
		end := start
		for start > 0 && src[start-1] == '`' {
			start--
		}
		code.OriginalMarker = string(src[start:end])
	}
	return code, nil
}

func convertFencedCodeBlock(src []byte, node *ast.FencedCodeBlock) (*content.CodeBlock, error) {
	// FencedCodeBlock contains raw data so we need to combine all the lines
	// into a single string.
	codeBuf := bytes.Buffer{}
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		_, _ = codeBuf.Write(line.Value(src))
	}

	code := content.NewCodeBlock(codeBuf.String())
	code.Language = string(node.Language(src))

	updatePreviousLine(node, code)

	return code, nil
}

func convertCodeBlock(src []byte, node *ast.CodeBlock) (*content.CodeBlock, error) {
	// CodeBlock contains raw data so we need to combine all the lines
	// into a single string.
	codeBuf := bytes.Buffer{}
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		_, _ = codeBuf.Write(line.Value(src))
	}

	code := content.NewCodeBlock(codeBuf.String())

	updatePreviousLine(node, code)
	return code, nil
}

func convertFootnoteDefinition(src []byte, node *footnoteDefinition) (*content.FootnoteDefinition, error) {
	definition := content.NewFootnoteDefinition(unescapeString([]byte(node.Label)))
	err := convertChildren(src, node, definition)
	if err != nil {
		return nil, err
	}

	updatePreviousLine(node, definition)

	return definition, nil
}

func convertMathBlock(src []byte, node *mathBlock) (*content.MathBlock, error) {
	// The lines of the formula are raw data, so they are put back together
	// into the LaTeX between the lines that mark the block out.
	valueBuf := bytes.Buffer{}
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		_, _ = valueBuf.Write(line.Value(src))
	}

	value := content.NewMathBlock(valueBuf.String())

	updatePreviousLine(node, value)

	return value, nil
}

func convertBlockquote(src []byte, node *ast.Blockquote) (*content.Blockquote, error) {
	blockquote := content.NewBlockquote()

	// Extract complete line format information
	lineFormats := extractBlockquoteLineFormats(src, node)
	blockquote.WithOriginalLineFormats(lineFormats)

	err := convertChildren(src, node, blockquote)
	if err != nil {
		return nil, err
	}

	updatePreviousLine(node, blockquote)

	return blockquote, nil
}

// extractBlockquoteLineFormats extracts the complete original formatting for each line
// of the blockquote, including whether each line has a ">" marker and the exact prefix
func extractBlockquoteLineFormats(src []byte, node *ast.Blockquote) []content.BlockquoteLineFormat {
	formats := make([]content.BlockquoteLineFormat, 0)

	// Find the range of the blockquote by looking at all text segments
	minStart := len(src)
	maxEnd := 0

	// Walk through all the children to find text segments
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if paragraph, ok := child.(*ast.Paragraph); ok {
			for textChild := paragraph.FirstChild(); textChild != nil; textChild = textChild.NextSibling() {
				if textNode, ok := textChild.(*ast.Text); ok {
					segment := textNode.Segment
					if segment.Start < minStart {
						minStart = segment.Start
					}
					if segment.Stop > maxEnd {
						maxEnd = segment.Stop
					}
				}
			}
		}
	}

	if minStart >= len(src) {
		return formats
	}

	// Find the actual start of the first line (go back to include blockquote markers)
	lineStart := minStart
	for lineStart > 0 && src[lineStart-1] != '\n' {
		lineStart--
	}

	// Find all lines within the blockquote range
	currentPos := lineStart
	for currentPos < maxEnd {
		// Find the end of this line
		lineEnd := maxEnd
		for i := currentPos; i < len(src); i++ {
			if src[i] == '\n' {
				lineEnd = i
				break
			}
		}

		// Extract the line content
		lineContent := string(src[currentPos:lineEnd])

		// Check if this line has a ">" marker and extract the prefix
		format := content.BlockquoteLineFormat{
			HasMarker: false,
			Prefix:    "",
		}

		// Look for ">" marker anywhere in the line
		contentStart := currentPos
		foundMarker := false
		markerPos := currentPos
		for i := currentPos; i < lineEnd; i++ {
			if src[i] == '>' {
				format.HasMarker = true
				foundMarker = true
				markerPos = i
				// Find where the actual content starts (after > and any spacing)
				contentStart = i + 1
				for contentStart < lineEnd && (src[contentStart] == ' ' || src[contentStart] == '\t') {
					contentStart++
				}

				// Determine the prefix based on context
				// If this looks like a list item (starts with "- " or similar),
				// only include the part from the ">" marker onwards
				lineContent := string(src[currentPos:lineEnd])
				if strings.HasPrefix(strings.TrimLeft(lineContent, " \t"), "- >") ||
					strings.HasPrefix(strings.TrimLeft(lineContent, " \t"), "* >") ||
					strings.HasPrefix(strings.TrimLeft(lineContent, " \t"), "+ >") {
					// This is a list item with blockquote, only include from ">" onwards
					format.Prefix = string(src[markerPos:contentStart])
				} else {
					// This is not a list item, include full indentation
					format.Prefix = string(src[currentPos:contentStart])
				}
				break
			}
		}

		if !foundMarker {
			// No ">" marker found, extract the indentation as prefix
			// Find the first non-whitespace character
			for i := currentPos; i < lineEnd; i++ {
				if src[i] != ' ' && src[i] != '\t' {
					format.Prefix = string(src[currentPos:i])
					break
				}
			}
		}

		// If the line is all whitespace, use the whole line as prefix
		if contentStart == lineEnd {
			format.Prefix = lineContent
		}

		formats = append(formats, format)

		// Move to the next line
		if lineEnd < len(src) && src[lineEnd] == '\n' {
			currentPos = lineEnd + 1
		} else {
			break
		}
	}

	return formats
}

// convertList converts an ast.List into either a list or a block.
func convertList(src []byte, node *ast.List) (*content.List, error) {
	list := content.NewListFromMarker(node.Marker)

	// For ordered lists, we need to check if the numbering is sequential
	// If it's sequential starting from 1, we don't store original numbers
	// If it's non-sequential or starts from a number other than 1, we store original numbers
	if list.Type == content.ListTypeOrdered {
		// First pass: extract all original numbers
		var originalNumbers []int
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			if listItem, ok := child.(*ast.ListItem); ok {
				originalNumber := extractOriginalListItemNumber(src, listItem)
				originalNumbers = append(originalNumbers, originalNumber)
			}
		}

		// Check if the numbering is sequential starting from 1
		isSequentialFromOne := true
		for i, num := range originalNumbers {
			if num != i+1 {
				isSequentialFromOne = false
				break
			}
		}

		// Convert list items, storing original numbers only if needed
		i := 0
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			item, err := convertListItem(src, child)
			if err != nil {
				return nil, err
			}

			// Store original number only if the list is not sequential from 1
			if !isSequentialFromOne && i < len(originalNumbers) && originalNumbers[i] > 0 {
				item.WithOriginalNumber(originalNumbers[i])
			}

			list.AddChild(item)
			i++
		}
	} else {
		// For unordered lists, just convert normally
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			item, err := convertListItem(src, child)
			if err != nil {
				return nil, err
			}
			list.AddChild(item)
		}
	}

	updatePreviousLine(node, list)

	return list, nil
}

func convertListItem(src []byte, node ast.Node) (*content.ListItem, error) {
	item := content.NewListItem()
	err := convertChildren(src, node, item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func convertTable(src []byte, node *east.Table) (*content.Table, error) {
	table := content.NewTable()
	for _, alignment := range node.Alignments {
		table.Alignments = append(table.Alignments, tableAlignmentFor(alignment))
	}

	// The header of the table is its own type in Goldmark, but both it and the
	// rows below it are a row of cells.
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		row, err := convertTableRow(src, child)
		if err != nil {
			return nil, err
		}
		table.AddChild(row)
	}

	updatePreviousLine(node, table)

	return table, nil
}

func convertTableRow(src []byte, node ast.Node) (*content.TableRow, error) {
	row := content.NewTableRow()
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*east.TableCell); !ok {
			return nil, fmt.Errorf("Could not convert node: %T", child)
		}

		cell := content.NewTableCell()
		err := convertChildren(src, child, cell)
		if err != nil {
			return nil, err
		}
		row.AddChild(cell)
	}
	return row, nil
}

func tableAlignmentFor(alignment east.Alignment) content.TableAlignment {
	switch alignment {
	case east.AlignLeft:
		return content.TableAlignmentLeft
	case east.AlignRight:
		return content.TableAlignmentRight
	case east.AlignCenter:
		return content.TableAlignmentCenter
	}

	return content.TableAlignmentNone
}

func convertRawHTML(src []byte, node *ast.RawHTML) (*content.RawHTML, error) {
	codeBuf := bytes.Buffer{}
	for i := 0; i < node.Segments.Len(); i++ {
		segment := node.Segments.At(i)
		_, _ = codeBuf.Write(segment.Value(src))
	}

	raw := content.NewRawHTML(codeBuf.String())
	return raw, nil
}

func convertHTMLBlock(src []byte, node *ast.HTMLBlock) (*content.RawHTMLBlock, error) {
	codeBuf := bytes.Buffer{}
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		_, _ = codeBuf.Write(line.Value(src))
	}

	raw := content.NewRawHTMLBlock(codeBuf.String())
	return raw, nil
}

func convertImage(src []byte, node *ast.Image) (*content.Image, error) {
	image := content.NewImage(string(node.Destination))
	image.Title = unescapeString(node.Title)

	err := convertChildren(src, node, image)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func convertProperties(src []byte, node *properties) (*content.Properties, error) {
	properties := content.NewProperties()
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		p, ok := child.(*property)
		if !ok {
			return nil, errors.New("Invalid child in properties")
		}

		// All values use regular Property/reference nodes. Separators remain text,
		// so preservation reflects edits without replaying an old whole value.
		prop := content.NewProperty(p.Name).WithPageRefsIgnored(p.PageRefsIgnored)
		if err := convertChildren(src, p, prop); err != nil {
			return nil, err
		}
		properties.AddChild(prop)
	}

	updatePreviousLineKeepingBlank(node, properties)

	// Check if there's a blank line after the properties
	if nextSibling := node.NextSibling(); nextSibling != nil {
		if nextSibling.HasBlankPreviousLines() {
			properties.SetHasBlankLineAfter(true)
		}
	}

	return properties, nil
}

func convertBeginEnd(src []byte, node *beginEnd) (content.Node, error) {
	codeBuf := bytes.Buffer{}
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		_, _ = codeBuf.Write(line.Value(src))
	}

	switch node.Variant {
	case "QUERY":
		return content.NewQueryCommand(codeBuf.String()), nil
	default:
		return content.NewAdvancedCommand(node.Variant, codeBuf.String()), nil
	}
}

func convertLogbook(src []byte, node *logbook) (content.Node, error) {
	logbook := content.NewLogbook()
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		value := strings.TrimSuffix(string(line.Value(src)), "\n")
		if node.Typed {
			logbook.AddChild(parseLogbookEntry(value))
		} else {
			logbook.AddChild(content.NewLogbookEntryRaw(value))
		}
	}

	updatePreviousLine(node, logbook)

	return logbook, nil
}

// shouldPreserveLiteralDashes checks if a list item contains only nested lists
// that should be flattened back to literal extra dashes.
func shouldPreserveLiteralDashes(src []byte, listItem ast.Node) bool {
	// Check if this list item contains only a single nested list
	if listItem.ChildCount() != 1 {
		return false
	}

	firstChild := listItem.FirstChild()
	if nestedList, ok := firstChild.(*ast.List); ok && nestedList.Marker == '-' {
		// This list item contains only a nested list with dash markers
		// This indicates it should be flattened to literal dashes
		return true
	}

	return false
}

// convertListItemToLiteralDashes converts a list item with nested lists to a paragraph
// that preserves the literal dash characters.
func convertListItemToLiteralDashes(src []byte, listItem ast.Node) (*content.Paragraph, error) {
	// Count the nesting depth and get the final text content
	dashCount := 0
	currentNode := listItem.FirstChild()

	// Traverse nested lists to count dashes
	for {
		if nestedList, ok := currentNode.(*ast.List); ok && nestedList.Marker == '-' {
			dashCount++
			if nestedList.ChildCount() == 1 {
				currentNode = nestedList.FirstChild().FirstChild()
			} else {
				break
			}
		} else {
			break
		}
	}

	// Get the text content from the deepest level
	var textContent string
	if textBlock, ok := currentNode.(*ast.TextBlock); ok {
		textContent = string(textBlock.Text(src))
	} else {
		// Fallback: try to get text from any text node
		textContent = "unknown content"
	}

	// Build the literal dash string: "- " repeated dashCount times + text
	literalDashes := ""
	for i := 0; i < dashCount; i++ {
		literalDashes += "- "
	}
	literalDashes += textContent

	return content.NewParagraph(content.NewText(literalDashes)), nil
}

// extractOriginalListItemNumber extracts the original number from the source text
// for an ordered list item by looking backwards from the first text content to find
// the number prefix (e.g., "14." or "19.")
func extractOriginalListItemNumber(src []byte, listItem *ast.ListItem) int {
	// Find the first text node in the list item to get position information
	var firstTextNode *ast.Text
	ast.Walk(listItem, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if textNode, ok := node.(*ast.Text); ok && firstTextNode == nil {
				firstTextNode = textNode
				return ast.WalkStop, nil
			}
		}
		return ast.WalkContinue, nil
	})

	if firstTextNode == nil {
		return 0
	}

	// Get the segment of the first text node
	segment := firstTextNode.Segment
	if segment.Start == 0 {
		return 0
	}

	// Look backwards from the text to find the list item marker
	// We need to find the beginning of the line and then parse the number
	lineStart := segment.Start
	for lineStart > 0 && src[lineStart-1] != '\n' {
		lineStart--
	}

	// Extract the line content up to the text
	lineContent := string(src[lineStart:segment.Start])

	// Parse the number from patterns like "14. " or "  19. " or "- > 23. "
	// Use regex to find the number before the dot
	re := regexp.MustCompile(`(\d+)\.`)
	matches := re.FindStringSubmatch(lineContent)
	if len(matches) >= 2 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			return num
		}
	}

	return 0
}

func convertTaskDate(node *taskDate) (content.Node, error) {
	var date *content.TaskDate
	if node.HasTime {
		date = content.NewTaskDateWithTime(node.DateType, node.Date)
	} else {
		date = content.NewTaskDate(node.DateType, node.Date)
	}
	date.Repeater = node.Repeater

	updatePreviousLineKeepingBlank(node, date)

	return date, nil
}

// updatePreviousLine updates the PreviousLineType of the target based on the
// node. This takes the information Goldmark parsed out and transfers it to
// our nodes.
func updatePreviousLine(node ast.Node, target content.PreviousLineAware) {
	if node.HasBlankPreviousLines() {
		target.SetPreviousLineType(content.PreviousLineTypeAutomatic)
	} else {
		// Parsing didn't indicate blank lines before the node, but we might
		// be the first node on this level in which case we set the type to
		// automatic
		if node.PreviousSibling() == nil {
			target.SetPreviousLineType(content.PreviousLineTypeAutomatic)
		} else {
			target.SetPreviousLineType(content.PreviousLineTypeNonBlank)
		}
	}
}

// updatePreviousLineKeepingBlank is updatePreviousLine for nodes that combine
// with the previous line by default. For those a blank line before the node
// has to be recorded explicitly, as the automatic behavior would drop it.
func updatePreviousLineKeepingBlank(node ast.Node, target content.PreviousLineAware) {
	if node.HasBlankPreviousLines() {
		target.SetPreviousLineType(content.PreviousLineTypeBlank)
	} else if node.PreviousSibling() == nil {
		// Parsing didn't indicate blank lines before the node, but we might
		// be the first node on this level in which case we set the type to
		// automatic
		target.SetPreviousLineType(content.PreviousLineTypeAutomatic)
	} else {
		target.SetPreviousLineType(content.PreviousLineTypeNonBlank)
	}
}

// isPropertyValue limits source-style additions to the migrated property representation.
func isPropertyValue(node ast.Node) bool {
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		if _, ok := parent.(*property); ok {
			return true
		}
	}
	return false
}

// detectPropertyEmphasisCharacter accounts for nested emphasis before the first text.
func detectPropertyEmphasisCharacter(src []byte, node *ast.Emphasis) rune {
	child := node.FirstChild()
	prefix := 0
	for {
		nested, ok := child.(*ast.Emphasis)
		if !ok {
			break
		}
		prefix += nested.Level
		child = nested.FirstChild()
	}
	if text, ok := child.(*ast.Text); ok && text.Segment.Start > prefix {
		if src[text.Segment.Start-prefix-1] == '_' {
			return '_'
		}
	}
	return '*'
}

// propertyLinkUsesAngleBrackets reads the wrapper after the label's source range.
func propertyLinkUsesAngleBrackets(src []byte, node *ast.Link) bool {
	if !isPropertyValue(node) {
		return false
	}
	stop := 0
	_ = ast.Walk(node, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if text, ok := child.(*ast.Text); entering && ok && text.Segment.Stop > stop {
			stop = text.Segment.Stop
		}
		return ast.WalkContinue, nil
	})
	if stop == 0 {
		return false
	}
	tail := string(src[stop:])
	if end := strings.IndexByte(tail, '\n'); end >= 0 {
		tail = tail[:end]
	}
	_, destination, ok := strings.Cut(tail, "](")
	if !ok {
		return false
	}
	destination = strings.TrimLeft(destination, " \t")
	return strings.HasPrefix(destination, "<"+string(node.Destination)+">")
}
