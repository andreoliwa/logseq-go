package markdown

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"

	"github.com/andreoliwa/logseq-go/content"
)

var urlRegexp = regexp.MustCompile(`^(?:http|https|ftp)://[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-z]+(?::\d+)?(?:[/#?][-a-zA-Z0-9@:%_+.~#$!?&/=\(\);,'">\^{}\[\]` + "`" + `]*)?`)

type EscapeFunc func(rune, rune) bool

func EscapeNone(rune, rune) bool {
	return false
}

func EscapePotentialMarkdown(prev rune, r rune) bool {
	// Don't escape brackets, underscores, or stars - let Logseq handle them
	// This allows Logseq to recognize unescaped Markdown characters

	// Don't escape ~~ if the first ~ is already escaped
	if prev == '~' && r == '~' {
		return true
	}

	return false
}

func EscapeLinkText(prev rune, r rune) bool {
	// Don't escape brackets in link text since they're already protected by the link syntax
	if r == '*' || r == '_' {
		return true
	}

	// Don't escape ~~ if the first ~ is already escaped
	if prev == '~' && r == '~' {
		return true
	}

	return false
}

func EscapeLinkTitle(prev rune, r rune) bool {
	return r == '"' || r == '\'' || r == ')'
}

func EscapeWikiLink(prev rune, r rune) bool {
	return r == ']'
}

func EscapeBlockRef(prev rune, r rune) bool {
	return r == ')'
}

func EscapeMacroQuotedArgument(prev rune, r rune) bool {
	return r == '"'
}

func EscapeString(str string, f EscapeFunc) string {
	out := strings.Builder{}
	runes := []rune(str)

	for i, r := range runes {
		var prev rune
		if i > 0 {
			prev = runes[i-1]
		}

		// Check if this character is already escaped
		alreadyEscaped := false
		if i > 0 && runes[i-1] == '\\' {
			// Count consecutive backslashes before this character
			backslashes := 0
			for j := i - 1; j >= 0 && runes[j] == '\\'; j-- {
				backslashes++
			}
			// If odd number of backslashes, the character is escaped
			alreadyEscaped = backslashes%2 == 1
		}

		if !alreadyEscaped && f(prev, r) {
			out.WriteRune('\\')
		}

		out.WriteRune(r)
	}

	return out.String()
}

// Output is used to write Markdown to an output buffer. It will help keep
// track of list indentation and when to add newlines.
type Output struct {
	out              *writer
	insideLink       bool // Track if we're currently writing inside a link
	blockIndentLevel int  // Track the current Logseq block indentation level
}

// NewWriter creates a new Markdown writer.
func NewWriter(out io.Writer) *Output {
	return &Output{
		out: newWriter(out),
	}
}

func AsString(n content.Node) (string, error) {
	out := strings.Builder{}
	w := NewWriter(&out)
	if err := w.Write(n); err != nil {
		return "", err
	}

	return out.String(), nil
}

func Write(n content.Node, out io.Writer) error {
	w := NewWriter(out)
	return w.Write(n)
}

func (w *Output) Write(n content.Node) error {
	switch node := n.(type) {
	case *content.RawHTML:
		return w.writeRaw(node.HTML)
	case *content.Text:
		return w.writeText(node)
	case *content.Emphasis:
		return w.writeEmphasis(node)
	case *content.Strong:
		return w.writeStrong(node)
	case *content.Strikethrough:
		return w.writeStrikethrough(node)
	case *content.CodeSpan:
		return w.writeCodeSpan(node)
	case *content.Link:
		return w.writeLink(node)
	case *content.AutoLink:
		return w.writeAutoLink(node)
	case *content.PageLink:
		return w.writePageLink(node)
	case *content.Hashtag:
		return w.writeHashtag(node)
	case *content.Priority:
		return w.writePriority(node)
	case *content.BlockRef:
		return w.writeBlockRef(node)
	case *content.Image:
		return w.writeImage(node)
	case *content.Macro:
		return w.writeMacro(node, node.Name, node.Arguments)
	case *content.Query:
		return w.writeMacro(node, "query", []string{node.Query})
	case *content.PageEmbed:
		return w.writeMacro(node, "embed", []string{"[[" + EscapeString(node.To, EscapeWikiLink) + "]]"})
	case *content.BlockEmbed:
		return w.writeMacro(node, "embed", []string{"((" + EscapeString(node.ID, EscapeBlockRef) + "))"})
	case *content.Cloze:
		return w.writeCloze(node)
	case *content.Heading:
		return w.writeHeading(node)
	case *content.RawHTMLBlock:
		return w.writeRawHTMLBlock(node)
	case *content.Paragraph:
		return w.writeParagraph(node)
	case *content.List:
		return w.writeList(node)
	case *content.Blockquote:
		return w.writeBlockquote(node)
	case *content.CodeBlock:
		return w.writeCodeBlock(node)
	case *content.ThematicBreak:
		return w.writeThematicBreak(node)
	case *content.Block:
		return w.writeBlock(node)
	case *content.Properties:
		return w.writeProperties(node)
	case *content.AdvancedCommand:
		return w.writeAdvancedCommand(node)
	case *content.QueryCommand:
		return w.writeBeginEnd(node, "QUERY", node.Query)
	case *content.TaskMarker:
		return w.writeTaskMarker(node)
	case *content.Logbook:
		return w.writeLogbook(node)
	default:
		return fmt.Errorf("unsupported node: %T", node)
	}
}

func (w *Output) writeRaw(s string) error {
	return w.out.WriteString(s)
}

// writeRawPreservingWhitespace writes content with current indentation applied to each line,
// but preserves the internal whitespace of each line exactly as it is.
// This is used for content like code blocks where we want to preserve exact whitespace within lines.
func (w *Output) writeRawPreservingWhitespace(s string) error {
	if len(s) == 0 {
		return nil
	}

	// Mark that we've written content at this level
	w.out.didWrite[len(w.out.didWrite)-1] = true

	// Split the content into lines and apply indentation to each line
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		// For the first line, check if we need to write indentation
		if i == 0 && w.out.lastWasLineBreak {
			_, err := w.out.output.Write(w.out.OnlyIndent())
			if err != nil {
				return err
			}
			_, err = w.out.output.Write(w.out.TrailingWhitespace())
			if err != nil {
				return err
			}
		}

		// Write the line content exactly as it is (preserving internal whitespace)
		_, err := w.out.output.Write([]byte(line))
		if err != nil {
			return err
		}

		// Add newline if this is not the last line
		if i < len(lines)-1 {
			_, err := w.out.output.Write([]byte("\n"))
			if err != nil {
				return err
			}

			// For subsequent lines, always write the block indentation
			// This ensures that even empty lines get the proper indentation
			_, err = w.out.output.Write(w.out.OnlyIndent())
			if err != nil {
				return err
			}
			_, err = w.out.output.Write(w.out.TrailingWhitespace())
			if err != nil {
				return err
			}
		}
	}

	// Update the lastWasLineBreak state based on whether the content ends with a newline
	w.out.lastWasLineBreak = len(s) > 0 && s[len(s)-1] == '\n'

	return nil
}

func (w *Output) write(s string, escapeFunc EscapeFunc) error {
	escaped := EscapeString(s, escapeFunc)
	return w.writeRaw(escaped)
}

func (w *Output) startBlock(node content.BlockNode, marker string) error {
	return w.startBlockWithAutomaticBehavior(node, marker, true)
}

// writeNewlinePrefix handles writing the appropriate newline prefix based on the previous line type
func (w *Output) writeNewlinePrefix(node content.BlockNode, doubleNewLineForAutomatic bool) error {
	if !w.out.HasWrittenAtCurrentIndent() {
		return nil
	}

	var prefix string
	if pl, ok := node.(content.PreviousLineAware); ok {
		switch pl.PreviousLineType() {
		case content.PreviousLineTypeBlank:
			prefix = "\n\n"
		case content.PreviousLineTypeNonBlank:
			prefix = "\n"
		case content.PreviousLineTypeAutomatic:
			if doubleNewLineForAutomatic {
				prefix = "\n\n"
			} else {
				prefix = "\n"
			}
		default:
			return fmt.Errorf("unknown previous line type: %d", pl.PreviousLineType())
		}
	} else {
		if doubleNewLineForAutomatic {
			prefix = "\n\n"
		} else {
			prefix = "\n"
		}
	}

	// Handle Logseq indentation for empty lines
	if prefix == "\n\n" && w.blockIndentLevel > 0 {
		// For the double newline, we need to write:
		// 1. First newline (end of previous content)
		// 2. Indented empty line
		// 3. Newline to end the empty line
		logseqIndent := strings.Repeat("\t", w.blockIndentLevel) + "  "

		// Write: newline + indented empty line + newline + indentation for next content
		toWrite := "\n" + logseqIndent + "\n" + logseqIndent
		_, err := w.out.output.Write([]byte(toWrite))
		if err != nil {
			return err
		}

		// Update the writer state to indicate we've written content but not a line break
		// This is important so the next content doesn't get extra indentation
		w.out.lastWasLineBreak = false
		w.out.didWrite[len(w.out.didWrite)-1] = true

		return nil
	}

	return w.out.WriteString(prefix)
}

func (w *Output) startBlockWithAutomaticBehavior(node content.BlockNode, marker string, doubleNewLineForAutomatic bool) error {
	err := w.writeNewlinePrefix(node, doubleNewLineForAutomatic)
	if err != nil {
		return err
	}

	w.out.PushIndentation(marker)
	return nil
}

func (w *Output) endBlock() {
	w.out.PopIndentation()
}

func (w *Output) writeChildren(node content.HasChildren) error {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		err := w.Write(child)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Output) writeText(node *content.Text) error {
	// Use different escaping based on context
	escapeFunc := EscapePotentialMarkdown
	if w.insideLink {
		escapeFunc = EscapeLinkText
	}

	err := w.write(node.Value, escapeFunc)
	if err != nil {
		return err
	}

	if node.SoftLineBreak {
		err := w.writeRaw("\n")
		if err != nil {
			return err
		}
	} else if node.HardLineBreak {
		err := w.writeRaw("\\\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Output) writeEmphasis(node *content.Emphasis) error {
	if _, ok := node.PreviousSibling().(*content.Emphasis); ok {
		// Writing two emphasis nodes next to each other is not valid Markdown,
		// so we add a space between them as a compromise.
		err := w.writeRaw(" ")
		if err != nil {
			return err
		}
	}

	// Use the stored character instead of hardcoded star
	emphasisChar := string(node.Character)
	err := w.writeRaw(emphasisChar)
	if err != nil {
		return err
	}

	err = w.writeChildren(node)
	if err != nil {
		return err
	}

	err = w.writeRaw(emphasisChar)
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writeStrong(node *content.Strong) error {
	if _, ok := node.PreviousSibling().(*content.Strong); ok {
		// Writing two strong nodes next to each other is not valid Markdown,
		// so we add a space between them as a compromise.
		err := w.writeRaw(" ")
		if err != nil {
			return err
		}
	}

	err := w.writeRaw("**")
	if err != nil {
		return err
	}

	err = w.writeChildren(node)
	if err != nil {
		return err
	}

	err = w.writeRaw("**")
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writeStrikethrough(node *content.Strikethrough) error {
	err := w.writeRaw("~~")
	if err != nil {
		return err
	}

	err = w.writeChildren(node)
	if err != nil {
		return err
	}

	err = w.writeRaw("~~")
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writeCodeSpan(node *content.CodeSpan) error {
	// First find the longest sequence of backticks in the value so can use
	// the correct marker.
	longestSequence := 0
	for i := 0; i < len(node.Value); i++ {
		if node.Value[i] != '`' {
			continue
		}

		if longestSequence == 0 {
			longestSequence = 1
		} else if node.Value[i-1] == '`' {
			longestSequence++
		}
	}
	marker := strings.Repeat("`", longestSequence+1)

	err := w.writeRaw(marker)
	if err != nil {
		return err
	}

	err = w.writeRaw(node.Value)
	if err != nil {
		return err
	}

	err = w.writeRaw(marker)
	if err != nil {
		return err
	}
	return nil
}

func (w *Output) writeLink(node *content.Link) error {
	err := w.writeRaw("[")
	if err != nil {
		return err
	}

	// Set context flag to indicate we're inside a link
	oldInsideLink := w.insideLink
	w.insideLink = true

	err = w.writeChildren(node)

	// Restore previous context
	w.insideLink = oldInsideLink

	if err != nil {
		return err
	}

	err = w.writeRaw("](")
	if err != nil {
		return err
	}

	err = w.write(node.URL, EscapeNone)
	if err != nil {
		return err
	}

	if node.Title != "" {
		err = w.writeRaw(" '")
		if err != nil {
			return err
		}

		err = w.write(node.Title, EscapeLinkTitle)
		if err != nil {
			return err
		}

		err = w.writeRaw("'")
		if err != nil {
			return err
		}
	}

	err = w.writeRaw(")")
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writeAutoLink(node *content.AutoLink) error {
	if urlRegexp.Match([]byte(node.URL)) {
		// No need for brackets, Logseq will automatically linkify the URL.
		return w.writeRaw(node.URL)
	}

	err := w.writeRaw("<")
	if err != nil {
		return err
	}

	err = w.writeRaw(node.URL)
	if err != nil {
		return err
	}

	err = w.writeRaw(">")
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writePageLink(node *content.PageLink) error {
	err := w.writeRaw("[[")
	if err != nil {
		return err
	}

	err = w.write(node.To, EscapeWikiLink)
	if err != nil {
		return err
	}

	err = w.writeRaw("]]")
	if err != nil {
		return err
	}

	return nil
}

// writeHashtag writes *content.PageLink as `#to` or `#[[to]]`. The extended
// syntax is used if the target contains whitespace.
func (w *Output) writeHashtag(node *content.Hashtag) error {
	err := w.writeRaw("#")
	if err != nil {
		return err
	}

	writeExtended := false
	for _, r := range node.To {
		if unicode.IsSpace(r) {
			writeExtended = true
			break
		}
	}

	if writeExtended {
		err = w.writeRaw("[[")
		if err != nil {
			return err
		}
	}

	err = w.write(node.To, EscapePotentialMarkdown)
	if err != nil {
		return err
	}

	if writeExtended {
		err = w.writeRaw("]]")
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Output) writePriority(node *content.Priority) error {
	// If we have original text (for both valid and invalid priorities), use it to preserve case
	if node.OriginalText != "" {
		return w.writeRaw(node.OriginalText)
	}

	// Fallback: For valid priorities without original text, construct the output
	err := w.writeRaw("[#")
	if err != nil {
		return err
	}

	var letter string
	switch node.Priority {
	case content.PriorityHigh:
		letter = "A"
	case content.PriorityMedium:
		letter = "B"
	case content.PriorityLow:
		letter = "C"
	default:
		return fmt.Errorf("unknown priority value: %v", node.Priority)
	}

	err = w.writeRaw(letter)
	if err != nil {
		return err
	}

	err = w.writeRaw("] ")
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writeBlockRef(node *content.BlockRef) error {
	err := w.writeRaw("((")
	if err != nil {
		return err
	}

	err = w.write(node.ID, EscapeWikiLink)
	if err != nil {
		return err
	}

	err = w.writeRaw("))")
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writeImage(node *content.Image) error {
	err := w.writeRaw("![")
	if err != nil {
		return err
	}

	err = w.writeChildren(node)
	if err != nil {
		return err
	}

	err = w.writeRaw("](")
	if err != nil {
		return err
	}

	err = w.write(node.URL, EscapeNone)
	if err != nil {
		return err
	}

	if node.Title != "" {
		err = w.writeRaw(" '")
		if err != nil {
			return err
		}

		err = w.write(node.Title, EscapeLinkTitle)
		if err != nil {
			return err
		}

		err = w.writeRaw("'")
		if err != nil {
			return err
		}
	}

	err = w.writeRaw(")")
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writeCloze(node *content.Cloze) error {
	cue := strings.TrimSpace(node.Cue)
	answer := strings.TrimSpace(node.Answer)
	if cue != "" {
		return w.writeMacro(node, "cloze", []string{answer + " \\ " + cue})
	} else {
		return w.writeMacro(node, "cloze", []string{answer})
	}
}

func (w *Output) writeMacro(node content.Node, name string, arguments []string) error {
	err := w.writeRaw("{{")
	if err != nil {
		return err
	}

	// Validate the macro name, it can not contain whitespace.
	for _, r := range name {
		if unicode.IsSpace(r) {
			return fmt.Errorf("macro name can not contain whitespace")
		}
	}

	err = w.writeRaw(name)
	if err != nil {
		return err
	}

	if arguments != nil {
		for i, arg := range arguments {
			if i == 0 {
				err = w.writeRaw(" ")
				if err != nil {
					return err
				}
			} else {
				err = w.writeRaw(", ")
				if err != nil {
					return err
				}
			}

			// Check if the argument contains a comma, if so we need to quote
			// the argument.
			quoted := false
			for _, r := range arg {
				if r == ',' {
					quoted = true
					break
				}
			}

			if quoted {
				err = w.writeRaw("\"")
				if err != nil {
					return err
				}

				err = w.write(arg, EscapeMacroQuotedArgument)
				if err != nil {
					return err
				}

				err = w.writeRaw("\"")
				if err != nil {
					return err
				}
			} else {
				err = w.write(arg, EscapeNone)
				if err != nil {
					return err
				}
			}
		}
	}

	err = w.writeRaw("}}")
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writeRawHTMLBlock(node *content.RawHTMLBlock) error {
	err := w.startBlock(node, "")
	if err != nil {
		return err
	}

	err = w.writeRaw(node.HTML)
	if err != nil {
		return err
	}

	w.endBlock()
	return nil
}

func (w *Output) writeHeading(node *content.Heading) error {
	// Handle newlines before the heading
	// Since headings don't implement PreviousLineAware, this will use default behavior (\n\n)
	err := w.writeNewlinePrefix(node, true)
	if err != nil {
		return err
	}

	// Write the heading marker directly (not as indentation)
	err = w.writeRaw(strings.Repeat("#", node.Level) + " ")
	if err != nil {
		return err
	}

	// Write the heading content inline
	err = w.writeChildren(node)
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writeParagraph(node *content.Paragraph) error {
	doubleNewLine := true
	if _, previousProperties := node.PreviousSibling().(*content.Properties); previousProperties {
		doubleNewLine = false
	}

	err := w.startBlockWithAutomaticBehavior(node, "", doubleNewLine)
	if err != nil {
		return err
	}

	err = w.writeChildren(node)
	if err != nil {
		return err
	}

	w.endBlock()
	return nil
}

func (w *Output) writeList(node *content.List) error {
	err := w.startBlock(node, "")
	if err != nil {
		return err
	}

	i := 0
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		listItem, ok := child.(*content.ListItem)
		if !ok {
			return fmt.Errorf("unsupported list child: %T", child)
		}

		var marker string
		if node.Type == content.ListTypeOrdered {
			var number int
			if listItem.OriginalNumber > 0 {
				// Use the preserved original number
				number = listItem.OriginalNumber
			} else {
				// Fallback to sequential numbering starting from 1
				number = i + 1
			}
			marker = fmt.Sprintf("%d", number) + string(node.Marker)
		} else {
			marker = string(node.Marker)
		}

		err := w.out.WriteString(marker + " ")
		if err != nil {
			return err
		}

		w.out.PushIndentation(strings.Repeat(" ", len(marker)+1))

		err = w.writeChildren(child.(content.HasChildren))
		if err != nil {
			return err
		}

		w.out.PopIndentation()

		if child.NextSibling() != nil {
			err = w.writeRaw("\n")
			if err != nil {
				return err
			}
		}

		i++
	}

	w.endBlock()
	return nil
}

func (w *Output) writeBlockquote(node *content.Blockquote) error {
	// Check if we have original line format information
	if len(node.OriginalLineFormats) > 0 {
		return w.writeBlockquoteWithOriginalFormats(node)
	}

	// Fallback to old behavior if no line format information is available
	// Use default marker
	marker := "> "

	// Check if we're inside a nested structure (like a list item)
	isNested := w.out.IndentationLevel() > 0

	if isNested {
		// Special handling for blockquotes inside list items:
		// - First line gets the ">" marker written directly
		// - Subsequent lines get no additional indentation (rely on existing list indentation)

		// Handle newlines before the blockquote
		err := w.writeNewlinePrefix(node, true)
		if err != nil {
			return err
		}

		// Push empty indentation for continuation lines
		w.out.PushIndentation("")

		// Write the marker for the first line
		if !w.out.lastWasLineBreak {
			// We're in the middle of a line (like after "* "), write marker directly
			markerBytes := []byte(marker)
			_, err := w.out.output.Write(markerBytes)
			if err != nil {
				return err
			}
		} else {
			// We're at the beginning of a line, write marker normally
			err := w.writeRaw(marker)
			if err != nil {
				return err
			}
		}

		err = w.writeChildren(node)
		if err != nil {
			return err
		}

		w.endBlock()
		return nil
	} else {
		// Standard blockquote handling: use marker as indentation for all lines
		err := w.startBlock(node, marker)
		if err != nil {
			return err
		}

		if !w.out.lastWasLineBreak {
			// This is a hack to make sure that the indicator is written in lists
			// if the blockquote is the first item in a list item.
			markerBytes := []byte(marker)
			_, err = w.out.output.Write(markerBytes)
			if err != nil {
				return err
			}
		}

		err = w.writeChildren(node)
		if err != nil {
			return err
		}

		w.endBlock()
		return nil
	}
}

func (w *Output) writeBlockquoteWithOriginalFormats(node *content.Blockquote) error {
	// Handle newlines before the blockquote
	err := w.writeNewlinePrefix(node, true)
	if err != nil {
		return err
	}

	// Push empty indentation since we'll handle all formatting manually
	w.out.PushIndentation("")

	// Manually write the content with original line formats
	err = w.writeBlockquoteContentWithFormats(node, node.OriginalLineFormats)
	if err != nil {
		return err
	}

	w.endBlock()
	return nil
}

func (w *Output) writeBlockquoteContentWithFormats(node *content.Blockquote, lineFormats []content.BlockquoteLineFormat) error {
	lineIndex := 0
	isFirstLine := true

	// Walk through all children
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if paragraph, ok := child.(*content.Paragraph); ok {
			// Process all inline nodes in the paragraph together to maintain line structure
			err := w.writeBlockquoteParagraphWithFormats(paragraph, lineFormats, &lineIndex, &isFirstLine)
			if err != nil {
				return err
			}
		} else {
			// For non-paragraph children, use normal writing
			err := w.Write(child)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Output) writeBlockquoteParagraphWithFormats(paragraph *content.Paragraph, lineFormats []content.BlockquoteLineFormat, lineIndex *int, isFirstLine *bool) error {
	// We need to process all inline nodes while tracking line boundaries
	// The key insight is that line breaks only occur within Text nodes, not between different node types

	for inlineNode := paragraph.FirstChild(); inlineNode != nil; inlineNode = inlineNode.NextSibling() {
		if text, ok := inlineNode.(*content.Text); ok {
			// Handle text nodes with potential line breaks
			err := w.writeBlockquoteTextWithFormats(text, lineFormats, lineIndex, isFirstLine)
			if err != nil {
				return err
			}
		} else {
			// Handle non-text inline nodes (PageLink, etc.)
			// These should be written inline without affecting line structure
			// We need to use a special method that doesn't handle line breaks
			err := w.writeInlineNodeInBlockquote(inlineNode)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Output) writeInlineNodeInBlockquote(node content.Node) error {
	// Write inline nodes without triggering line break handling
	// This is similar to w.Write() but bypasses text line break processing
	switch n := node.(type) {
	case *content.PageLink:
		return w.writePageLink(n)
	case *content.Link:
		return w.writeLinkInBlockquote(n)
	case *content.Emphasis:
		return w.writeEmphasisInBlockquote(n)
	case *content.Strong:
		return w.writeStrongInBlockquote(n)
	case *content.CodeSpan:
		return w.writeCodeSpan(n)
	case *content.Hashtag:
		return w.writeHashtag(n)
	case *content.BlockRef:
		return w.writeBlockRef(n)
	default:
		// For other node types, fall back to normal writing but this might cause issues
		return w.Write(node)
	}
}

func (w *Output) writeLinkInBlockquote(node *content.Link) error {
	w.insideLink = true
	defer func() { w.insideLink = false }()

	err := w.writeRaw("[")
	if err != nil {
		return err
	}

	// Write link text without line break handling
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if text, ok := child.(*content.Text); ok {
			// Write text content without line break processing
			err := w.write(text.Value, EscapeLinkText)
			if err != nil {
				return err
			}
		} else {
			err := w.writeInlineNodeInBlockquote(child)
			if err != nil {
				return err
			}
		}
	}

	err = w.writeRaw("](")
	if err != nil {
		return err
	}

	err = w.write(node.URL, EscapeNone)
	if err != nil {
		return err
	}

	if node.Title != "" {
		err = w.writeRaw(" \"")
		if err != nil {
			return err
		}

		err = w.write(node.Title, EscapeLinkTitle)
		if err != nil {
			return err
		}

		err = w.writeRaw("\"")
		if err != nil {
			return err
		}
	}

	err = w.writeRaw(")")
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writeEmphasisInBlockquote(node *content.Emphasis) error {
	emphasisChar := string(node.Character)
	err := w.writeRaw(emphasisChar)
	if err != nil {
		return err
	}

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if text, ok := child.(*content.Text); ok {
			err := w.write(text.Value, EscapePotentialMarkdown)
			if err != nil {
				return err
			}
		} else {
			err := w.writeInlineNodeInBlockquote(child)
			if err != nil {
				return err
			}
		}
	}

	err = w.writeRaw(emphasisChar)
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writeStrongInBlockquote(node *content.Strong) error {
	err := w.writeRaw("**")
	if err != nil {
		return err
	}

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if text, ok := child.(*content.Text); ok {
			err := w.write(text.Value, EscapePotentialMarkdown)
			if err != nil {
				return err
			}
		} else {
			err := w.writeInlineNodeInBlockquote(child)
			if err != nil {
				return err
			}
		}
	}

	err = w.writeRaw("**")
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writeBlockquoteTextWithFormats(text *content.Text, lineFormats []content.BlockquoteLineFormat, lineIndex *int, isFirstLine *bool) error {
	// Split the text by lines
	lines := strings.Split(text.Value, "\n")

	for i, line := range lines {
		if i > 0 {
			// This is a continuation line within this text node (actual line break in the text)
			if *lineIndex < len(lineFormats) {
				format := lineFormats[*lineIndex]
				// Write newline + prefix directly to bypass indentation system
				_, err := w.out.output.Write([]byte("\n" + format.Prefix))
				if err != nil {
					return err
				}
				w.out.lastWasLineBreak = false // We've written content after the newline
			} else {
				_, err := w.out.output.Write([]byte("\n"))
				if err != nil {
					return err
				}
				w.out.lastWasLineBreak = true
			}
			*lineIndex++
		} else if *isFirstLine {
			// This is the very first text content in the blockquote
			if *lineIndex < len(lineFormats) {
				format := lineFormats[*lineIndex]
				if !w.out.lastWasLineBreak {
					// We're in the middle of a line (like after "* ")
					_, err := w.out.output.Write([]byte(format.Prefix))
					if err != nil {
						return err
					}
				} else {
					// We're at the beginning of a line
					err := w.writeRaw(format.Prefix)
					if err != nil {
						return err
					}
				}
			}
			*lineIndex++
			*isFirstLine = false
		}
		// For i == 0 && !*isFirstLine: this is the first line of a subsequent text node
		// We don't need to write any prefix, just continue on the same line

		// Write the line content
		if line != "" {
			err := w.write(line, EscapePotentialMarkdown)
			if err != nil {
				return err
			}
		}
	}

	// Handle line breaks after writing the text content
	if text.SoftLineBreak {
		// Soft line break - write the next line format
		if *lineIndex < len(lineFormats) {
			format := lineFormats[*lineIndex]
			// Write newline + prefix directly to bypass indentation system
			_, err := w.out.output.Write([]byte("\n" + format.Prefix))
			if err != nil {
				return err
			}
			w.out.lastWasLineBreak = false // We've written content after the newline
		} else {
			_, err := w.out.output.Write([]byte("\n"))
			if err != nil {
				return err
			}
			w.out.lastWasLineBreak = true
		}
		*lineIndex++
	} else if text.HardLineBreak {
		err := w.writeRaw("\\")
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Output) writeCodeBlock(node *content.CodeBlock) error {
	err := w.startBlock(node, "")
	if err != nil {
		return err
	}

	err = w.writeRaw("```")
	if err != nil {
		return err
	}

	if node.Language != "" {
		err = w.writeRaw(node.Language)
		if err != nil {
			return err
		}
	}

	err = w.writeRaw("\n")
	if err != nil {
		return err
	}

	// Use writeRawPreservingWhitespace for the code content to preserve tabs and spaces
	err = w.writeRawPreservingWhitespace(node.Code)
	if err != nil {
		return err
	}

	// If the code does not end with a blank line, we add a newline
	if !strings.HasSuffix(node.Code, "\n") {
		err = w.writeRaw("\n")
		if err != nil {
			return err
		}
	}

	// Reset the lastWasLineBreak state to false so the closing ``` doesn't get extra indentation
	// The closing ``` should be at the same level as the opening ```
	w.out.lastWasLineBreak = false

	err = w.writeRaw("```")
	if err != nil {
		return err
	}

	w.endBlock()
	return nil
}

func (w *Output) writeThematicBreak(node *content.ThematicBreak) error {
	err := w.startBlock(node, "")
	if err != nil {
		return err
	}

	err = w.writeRaw("---")
	if err != nil {
		return err
	}

	w.endBlock()
	return nil
}

func (w *Output) writeBlock(node *content.Block) error {
	err := w.startBlock(node, "")
	if err != nil {
		return err
	}

	// Track block indentation level
	hasParentBlock := false
	if _, ok := node.Parent().(*content.Block); ok {
		hasParentBlock = true
	}

	// Track whether we incremented the level for this block
	incrementedLevel := false

	// Only increment for blocks that have parent blocks that are also blocks
	// The first level of blocks (children of root) should be level 0
	if hasParentBlock {
		// Check if the parent's parent is also a block
		grandParent := node.Parent().Parent()
		if _, ok := grandParent.(*content.Block); ok {
			w.blockIndentLevel++
			incrementedLevel = true
		}
	}

	// Write the content first
	for _, child := range node.Content() {
		err := w.Write(child)
		if err != nil {
			return err
		}
	}

	w.endBlock()

	previousIndent := ""
	if hasParentBlock && w.out.IndentationLevel() > 0 {
		// As Logseq uses tabs for indentation of blocks we pop the current
		// indentation which is the two spaces to align content with "- " of
		// the list item. This allows the indentation to be only tabs for
		// blocks
		previousIndent = w.out.PopIndentation()
	}

	// Output the sub blocks
	blocks := node.Blocks()
	if len(blocks) > 0 {
		if w.out.HasWrittenAtCurrentIndent() {
			err := w.out.WriteString("\n")
			if err != nil {
				return err
			}
		}

		if hasParentBlock {
			w.out.PushIndentation("\t")
		} else {
			w.out.PushIndentation("")
		}

		i := 0
		for _, child := range blocks {
			i++
			err := w.out.WriteString("- ")
			if err != nil {
				return err
			}

			w.out.PushIndentation("  ")

			err = w.Write(child)
			if err != nil {
				return err
			}

			if child.NextSibling() != nil {
				err = w.writeRaw("\n")
				if err != nil {
					return err
				}
			}

			w.out.PopIndentation()
		}

		w.out.PopIndentation()
	}

	if hasParentBlock {
		// Push the previous indentation back on the stack
		w.out.PushIndentation(previousIndent)
		// Only decrement if we actually incremented for this block
		if incrementedLevel {
			w.blockIndentLevel--
		}
	}

	return nil
}

func (w *Output) writeProperties(node *content.Properties) error {
	err := w.startBlockWithAutomaticBehavior(node, "", false)
	if err != nil {
		return err
	}

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*content.Property); !ok {
			return fmt.Errorf("unsupported properties child: %T", child)
		}

		property := child.(*content.Property)
		err := w.writeRaw(property.Name)
		if err != nil {
			return err
		}

		err = w.writeRaw(":: ")
		if err != nil {
			return err
		}

		err = w.writeChildren(property)
		if err != nil {
			return err
		}

		if child.NextSibling() != nil {
			err = w.writeRaw("\n")
			if err != nil {
				return err
			}
		}
	}

	w.endBlock()
	return nil
}

func (w *Output) writeAdvancedCommand(node *content.AdvancedCommand) error {
	return w.writeBeginEnd(node, node.Type, node.Value)
}

func (w *Output) writeBeginEnd(node content.BlockNode, variant string, value string) error {
	err := w.startBlock(node, "")
	if err != nil {
		return err
	}

	err = w.writeRaw("#+BEGIN_" + variant + "\n")
	if err != nil {
		return err
	}

	err = w.writeRaw(value)
	if err != nil {
		return err
	}

	if !w.out.lastWasLineBreak {
		err = w.writeRaw("\n")
		if err != nil {
			return err
		}
	}

	err = w.writeRaw("#+END_" + variant)
	if err != nil {
		return err
	}

	w.endBlock()
	return nil
}

func (w *Output) writeTaskMarker(node *content.TaskMarker) error {
	var err error
	switch node.Status {
	case content.TaskStatusNone:
		return nil
	case content.TaskStatusTodo:
		err = w.writeRaw("TODO")
	case content.TaskStatusDoing:
		err = w.writeRaw("DOING")
	case content.TaskStatusDone:
		err = w.writeRaw("DONE")
	case content.TaskStatusLater:
		err = w.writeRaw("LATER")
	case content.TaskStatusNow:
		err = w.writeRaw("NOW")
	case content.TaskStatusCancelled:
		err = w.writeRaw("CANCELLED")
	case content.TaskStatusCanceled:
		err = w.writeRaw("CANCELED")
	case content.TaskStatusInProgress:
		err = w.writeRaw("IN-PROGRESS")
	case content.TaskStatusWait:
		err = w.writeRaw("WAIT")
	case content.TaskStatusWaiting:
		err = w.writeRaw("WAITING")
	default:
		return fmt.Errorf("unsupported task status: %d", node.Status)
	}

	if err != nil {
		return err
	}

	if node.NextSibling() != nil {
		err = w.writeRaw(" ")
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Output) writeLogbook(node *content.Logbook) error {
	err := w.startBlock(node, "")
	if err != nil {
		return err
	}

	err = w.writeRaw(":LOGBOOK:\n")
	if err != nil {
		return err
	}

	for _, entry := range node.Children() {
		switch e := entry.(type) {
		case *content.LogbookEntryRaw:
			err = w.writeRaw(e.Value)
			if err != nil {
				return err
			}

			if !strings.HasSuffix(e.Value, "\n") {
				err = w.writeRaw("\n")
				if err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unsupported logbook entry: %T", entry)
		}
	}

	err = w.writeRaw(":END:")
	if err != nil {
		return err
	}

	w.endBlock()
	return nil
}
