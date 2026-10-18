package markdown

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/andreoliwa/logseq-go/v2/content"
)

var urlRegexp = regexp.MustCompile(`^(?:http|https|ftp)://[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-z]+(?::\d+)?(?:[/#?][-a-zA-Z0-9@:%_+.~#$!?&/=\(\);,'">\^{}\[\]` + "`" + `]*)?`)

// EscapeFunc escapes a string so that reading it back gives the same value.
type EscapeFunc func(string) string

func EscapeNone(str string) string {
	return str
}

// literalBracketsRegexp matches bracketed syntax that Logseq only recognizes
// when the brackets are left alone, which is the priority of a task such as
// `[#A]`. Escaping those would turn them into plain text.
var literalBracketsRegexp = regexp.MustCompile(`\[#[^\[\]\s]+\]`)

// EscapePotentialMarkdown escapes characters in text that would otherwise be
// read back as formatting.
func EscapePotentialMarkdown(str string) string {
	out := strings.Builder{}

	str = escapeMath(str)

	end := 0
	for _, match := range literalBracketsRegexp.FindAllStringIndex(str, -1) {
		if match[1] < len(str) && (str[match[1]] == '(' || str[match[1]] == '[') {
			// Followed by a link destination or a link label, so the brackets
			// have to stay escaped to not be read back as a link.
			continue
		}

		out.WriteString(escapeRunes(str[end:match[0]], escapePotentialMarkdownRune))
		out.WriteString(str[match[0]:match[1]])
		end = match[1]
	}

	out.WriteString(escapeRunes(str[end:], escapePotentialMarkdownRune))
	return out.String()
}

func EscapeStrictMarkdown(str string) string {
	out := strings.Builder{}

	str = escapeMath(str)

	end := 0
	for _, match := range literalBracketsRegexp.FindAllStringIndex(str, -1) {
		if match[1] < len(str) && (str[match[1]] == '(' || str[match[1]] == '[') {
			// Followed by a link destination or a link label, so the brackets
			// have to stay escaped to not be read back as a link.
			continue
		}

		out.WriteString(escapeRunes(str[end:match[0]], escapeStrictMarkdownRune))
		out.WriteString(str[match[0]:match[1]])
		end = match[1]
	}

	out.WriteString(escapeRunes(str[end:], escapeStrictMarkdownRune))
	return out.String()
}

func escapeStrictMarkdownRune(prev rune, r rune) bool {
	if r == '*' || r == '_' || r == '[' || r == ']' {
		return true
	}
	return escapePotentialMarkdownRune(prev, r)
}

// escapeMath escapes the dollar signs of the parts of a text that would be
// read back as a LaTeX formula, so that text such as `$x$` stays text. The
// dollar signs within a formula are escaped as well, as leaving one of them
// alone would let it pair up with something later in the text.
func escapeMath(str string) string {
	if !strings.Contains(str, "$") {
		return str
	}

	out := strings.Builder{}
	src := []byte(str)
	for i := 0; i < len(src); {
		if src[i] != '$' {
			out.WriteByte(src[i])
			i++
			continue
		}

		_, length, _, ok := mathAt(src[i:])
		if !ok {
			out.WriteByte('$')
			i++
			continue
		}

		out.WriteString(strings.ReplaceAll(string(src[i:i+length]), "$", `\$`))
		i += length
	}

	return out.String()
}

func escapePotentialMarkdownRune(prev rune, r rune) bool {
	// Preserve literal brackets, underscores and stars for Logseq.

	// Don't escape ~~ if the first ~ is already escaped
	if prev == '~' && r == '~' {
		return true
	}

	if prev == '^' && r == '^' {
		return true
	}

	return false
}

func EscapeLinkText(str string) string {
	return escapeRunes(str, func(prev rune, r rune) bool {
		// Don't escape brackets, underscores, or stars in link text since they're already protected by the link syntax
		// Underscores should remain unescaped regardless of context
		if r == '*' {
			return true
		}

		// Don't escape ~~ if the first ~ is already escaped
		if prev == '~' && r == '~' {
			return true
		}

		return false
	})
}

func EscapeLinkURL(str string) string {
	return escapeRunes(str, func(prev rune, r rune) bool {
		return r == '(' || r == ')'
	})
}

func EscapeLinkTitle(str string) string {
	return escapeRunes(str, func(prev rune, r rune) bool {
		return r == '"' || r == '\'' || r == ')'
	})
}

func EscapeWikiLink(str string) string {
	return escapeRunes(str, func(prev rune, r rune) bool {
		return r == ']'
	})
}

// EscapeFootnoteLabel escapes the label of a footnote, which is read up to the
// first closing bracket in it.
func EscapeFootnoteLabel(str string) string {
	return escapeRunes(str, func(prev rune, r rune) bool {
		return r == ']' || r == '['
	})
}

func EscapeBlockRef(str string) string {
	return escapeRunes(str, func(prev rune, r rune) bool {
		return r == ')'
	})
}

func EscapeMacroQuotedArgument(str string) string {
	return escapeRunes(str, func(prev rune, r rune) bool {
		return r == '"'
	})
}

// escapeRunes puts a backslash in front of every rune the given function picks
// out, passing it the rune before it for the cases that need that context.
func escapeRunes(str string, f func(prev rune, r rune) bool) string {
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

// Option changes how Markdown is written for the parts of the syntax that a
// graph configures the shape of.
type Option func(*outputOptions)

// outputOptions are the settings that the writer takes from the graph. The
// zero value is not usable, use defaultOutputOptions to get the defaults of
// Logseq.
type outputOptions struct {
	// logbookWithSeconds is whether the times in logbook entries are written
	// with seconds.
	logbookWithSeconds bool
	strict             bool
}

// defaultOutputOptions are what Logseq does for a graph that does not
// configure anything else.
func defaultOutputOptions() outputOptions {
	return outputOptions{
		logbookWithSeconds: true,
	}
}

// WithStrictOutput escapes literal Markdown and omits absent task priorities.
func WithStrictOutput() Option { return func(o *outputOptions) { o.strict = true } }

// WithLogbookSeconds sets whether the times in logbook entries are written
// with seconds, which is `:with-second-support?` of `:logbook/settings`.
func WithLogbookSeconds(withSeconds bool) Option {
	return func(o *outputOptions) {
		o.logbookWithSeconds = withSeconds
	}
}

// Output is used to write Markdown to an output buffer. It will help keep
// track of list indentation and when to add newlines.
type Output struct {
	out              *writer
	opts             outputOptions
	insideLink       bool // Track whether text is protected by link syntax.
	blockIndentLevel int  // Track Logseq block indentation.
}

// NewWriter creates a new Markdown writer.
func NewWriter(out io.Writer, opts ...Option) *Output {
	options := defaultOutputOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Output{
		out:  newWriter(out),
		opts: options,
	}
}

func AsString(n content.Node, opts ...Option) (string, error) {
	out := strings.Builder{}
	w := NewWriter(&out, opts...)
	if err := w.Write(n); err != nil {
		return "", err
	}

	return out.String(), nil
}

func Write(n content.Node, out io.Writer, opts ...Option) error {
	w := NewWriter(out, opts...)
	return w.Write(n)
}

func (w *Output) Write(n content.Node) error {
	switch node := n.(type) {
	case *content.RawHTML:
		return w.writeRaw(node.HTML)
	case *content.Text:
		return w.writeText(node)
	case *content.RawText:
		// Raw text is already Markdown, so it is written as it is.
		return w.writeRaw(node.Value)
	case *content.Emphasis:
		return w.writeEmphasis(node)
	case *content.Strong:
		return w.writeStrong(node)
	case *content.Strikethrough:
		return w.writeStrikethrough(node)
	case *content.Highlight:
		return w.writeHighlight(node)
	case *content.FootnoteRef:
		return w.writeFootnoteRef(node)
	case *content.FootnoteDefinition:
		return w.writeFootnoteDefinition(node)
	case *content.Math:
		return w.writeMath(node)
	case *content.MathBlock:
		return w.writeMathBlock(node)
	case *content.CodeSpan:
		return w.writeCodeSpan(node)
	case *content.Link:
		return w.writeLink(node)
	case *content.AutoLink:
		return w.writeAutoLink(node)
	case *content.PageLink:
		return w.writePageLink(node)
	case *content.PageRefText:
		// The title is the whole of the reference, written the same way text is.
		if w.opts.strict {
			return w.write(node.To, EscapeStrictMarkdown)
		}
		return w.write(node.To, EscapePotentialMarkdown)
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
		return w.writeMacro(node, "embed", []string{"[[" + EscapeWikiLink(node.To) + "]]"})
	case *content.BlockEmbed:
		return w.writeMacro(node, "embed", []string{"((" + EscapeBlockRef(node.ID) + "))"})
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
	case *content.Table:
		return w.writeTable(node)
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
	case *content.TaskDate:
		return w.writeTaskDate(node)
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
	return w.writeRaw(escapeFunc(s))
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
	if w.opts.strict {
		escapeFunc = EscapeStrictMarkdown
	}
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
	if w.opts.strict {
		emphasisChar = "*"
	}
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

func (w *Output) writeHighlight(node *content.Highlight) error {
	err := w.writeRaw("^^")
	if err != nil {
		return err
	}

	err = w.writeChildren(node)
	if err != nil {
		return err
	}

	err = w.writeRaw("^^")
	if err != nil {
		return err
	}

	return nil
}

func (w *Output) writeFootnoteRef(node *content.FootnoteRef) error {
	return w.writeRaw("[^" + EscapeFootnoteLabel(node.Label) + "]")
}

func (w *Output) writeFootnoteDefinition(node *content.FootnoteDefinition) error {
	err := w.startBlock(node, "")
	if err != nil {
		return err
	}

	err = w.writeRaw("[^" + EscapeFootnoteLabel(node.Label) + "]: ")
	if err != nil {
		return err
	}

	// The content of a footnote carries on over the lines that are indented
	// under the definition.
	w.out.PushIndentation("    ")

	err = w.writeChildren(node)
	if err != nil {
		return err
	}

	w.out.PopIndentation()

	w.endBlock()
	return nil
}

func (w *Output) writeMath(node *content.Math) error {
	marker := "$"
	if node.Displayed {
		marker = "$$"
	}

	err := w.writeRaw(marker)
	if err != nil {
		return err
	}

	err = w.writeRaw(node.Value)
	if err != nil {
		return err
	}

	return w.writeRaw(marker)
}

func (w *Output) writeMathBlock(node *content.MathBlock) error {
	err := w.startBlock(node, "")
	if err != nil {
		return err
	}

	err = w.writeRaw("$$\n")
	if err != nil {
		return err
	}

	err = w.writeRaw(node.Value)
	if err != nil {
		return err
	}

	// The line that ends the block needs a line of its own.
	if !strings.HasSuffix(node.Value, "\n") {
		err = w.writeRaw("\n")
		if err != nil {
			return err
		}
	}

	err = w.writeRaw("$$")
	if err != nil {
		return err
	}

	w.endBlock()
	return nil
}

func (w *Output) writeCodeSpan(node *content.CodeSpan) error {
	// The marker has to be longer than the longest run of backticks in the
	// value, or it would end the code span early.
	longestRun := 0
	currentRun := 0
	for i := 0; i < len(node.Value); i++ {
		if node.Value[i] == '`' {
			currentRun++
			if currentRun > longestRun {
				longestRun = currentRun
			}
		} else {
			currentRun = 0
		}
	}
	marker := strings.Repeat("`", longestRun+1)

	// A code span that starts or ends with a backtick needs to be padded with
	// spaces to keep the backtick apart from the marker. The same padding is
	// needed for a value surrounded by spaces, as a reader strips one space
	// from each end when both are present.
	value := node.Value
	if needsCodeSpanPadding(value) {
		value = " " + value + " "
	}

	err := w.writeRaw(marker)
	if err != nil {
		return err
	}

	err = w.writeRaw(value)
	if err != nil {
		return err
	}

	err = w.writeRaw(marker)
	if err != nil {
		return err
	}
	return nil
}

// needsCodeSpanPadding checks if a code span value has to be written with a
// space at each end to survive being read back.
func needsCodeSpanPadding(value string) bool {
	if strings.HasPrefix(value, "`") || strings.HasSuffix(value, "`") {
		return true
	}

	// A value of only spaces is kept as is by a reader, and padding it would
	// make it longer instead.
	if strings.Trim(value, " ") == "" {
		return false
	}

	return strings.HasPrefix(value, " ") && strings.HasSuffix(value, " ")
}

func (w *Output) writeLink(node *content.Link) error {
	err := w.writeRaw("[")
	if err != nil {
		return err
	}

	// Set context flag to indicate we're inside a link
	oldInsideLink := w.insideLink
	w.insideLink = !w.opts.strict

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

	escapeURL := EscapeNone
	if w.opts.strict {
		escapeURL = EscapeLinkURL
	}
	err = w.write(node.URL, escapeURL)
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

	escapeTarget := EscapePotentialMarkdown
	if w.opts.strict {
		escapeTarget = EscapeWikiLink
	}
	err = w.write(node.To, escapeTarget)
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
	if w.opts.strict {
		return w.writeTaskPriority(node)
	}
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

	escapeURL := EscapeNone
	if w.opts.strict {
		escapeURL = EscapeLinkURL
	}
	err = w.write(node.URL, escapeURL)
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
	if w.opts.strict {

		err := w.startBlock(node, strings.Repeat("#", node.Level)+" ")
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
			if !w.opts.strict && listItem.OriginalNumber > 0 {
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

// writeTable writes a table in the GitHub Flavored Markdown syntax, padding
// the cells so that the pipes of every row line up in a text editor.
func (w *Output) writeTable(node *content.Table) error {
	rows, err := w.tableRows(node)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		return fmt.Errorf("table has no rows")
	}

	widths := tableColumnWidths(rows)
	for len(widths) < len(node.Alignments) {
		// A column that the table sets the alignment of, but that no row has a
		// cell for, is still written out.
		widths = append(widths, 3)
	}

	err = w.startBlock(node, "")
	if err != nil {
		return err
	}

	// The header names the columns and is followed by the row that sets how
	// the text of each column lines up.
	err = w.writeRaw(tableRow(rows[0], widths))
	if err != nil {
		return err
	}

	err = w.writeRaw("\n" + tableDelimiterRow(node, widths))
	if err != nil {
		return err
	}

	for _, row := range rows[1:] {
		err = w.writeRaw("\n" + tableRow(row, widths))
		if err != nil {
			return err
		}
	}

	w.endBlock()
	return nil
}

// tableRows writes out the cells of every row, so that the width of each
// column is known before any of it is written.
func (w *Output) tableRows(node *content.Table) ([][]string, error) {
	rows := make([][]string, 0)
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		row, ok := child.(*content.TableRow)
		if !ok {
			return nil, fmt.Errorf("unsupported table child: %T", child)
		}

		cells := make([]string, 0)
		for cellNode := row.FirstChild(); cellNode != nil; cellNode = cellNode.NextSibling() {
			cell, ok := cellNode.(*content.TableCell)
			if !ok {
				return nil, fmt.Errorf("unsupported table row child: %T", cellNode)
			}

			value, err := w.tableCell(cell)
			if err != nil {
				return nil, err
			}

			cells = append(cells, value)
		}

		rows = append(rows, cells)
	}

	return rows, nil
}

// tableCell writes the content of a cell to the text that goes between the
// pipes of a row. A row is a single line, so line breaks become spaces, and
// the pipes in the content are escaped to keep them out of the shape of the
// table.
func (w *Output) tableCell(node *content.TableCell) (string, error) {
	out := strings.Builder{}
	cellWriter := &Output{
		out:  newWriter(&out),
		opts: w.opts,
	}

	err := cellWriter.writeChildren(node)
	if err != nil {
		return "", err
	}

	value := out.String()
	value = strings.ReplaceAll(value, "\\\n", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.ReplaceAll(value, "|", "\\|"), nil
}

// tableColumnWidths measures how wide each column has to be for the content of
// every cell in it to fit.
func tableColumnWidths(rows [][]string) []int {
	widths := make([]int, 0)
	for _, cells := range rows {
		for i, cell := range cells {
			// Three is the shortest a column can be, as that is what the row
			// that sets the alignment of a centered column needs.
			for i >= len(widths) {
				widths = append(widths, 3)
			}

			if width := utf8.RuneCountInString(cell); width > widths[i] {
				widths[i] = width
			}
		}
	}

	return widths
}

// tableRow writes one row of a table, padding every cell to the width of its
// column. Rows that are missing cells get empty ones, as a row that is shorter
// than the header is read back with the columns it does not have left empty.
func tableRow(cells []string, widths []int) string {
	row := strings.Builder{}
	for i, width := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}

		row.WriteString("| ")
		row.WriteString(cell)
		row.WriteString(strings.Repeat(" ", width-utf8.RuneCountInString(cell)))
		row.WriteString(" ")
	}

	row.WriteString("|")
	return row.String()
}

// tableDelimiterRow writes the row below the header, which sets how the text
// of each column lines up.
func tableDelimiterRow(node *content.Table, widths []int) string {
	row := strings.Builder{}
	for i, width := range widths {
		row.WriteString("| ")

		switch node.AlignmentOf(i) {
		case content.TableAlignmentLeft:
			row.WriteString(":" + strings.Repeat("-", width-1))
		case content.TableAlignmentRight:
			row.WriteString(strings.Repeat("-", width-1) + ":")
		case content.TableAlignmentCenter:
			row.WriteString(":" + strings.Repeat("-", width-2) + ":")
		default:
			row.WriteString(strings.Repeat("-", width))
		}

		row.WriteString(" ")
	}

	row.WriteString("|")
	return row.String()
}

func (w *Output) writeBlockquote(node *content.Blockquote) error {
	if w.opts.strict {

		err := w.startBlock(node, "> ")
		if err != nil {
			return err
		}

		if !w.out.lastWasLineBreak {
			// This is a hack to make sure that the indicator is written in lists
			// if the blockquote is the first item in a list item.
			_, err = w.out.output.Write([]byte{'>', ' '})
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

	if w.opts.strict {
		err = w.writeRaw(node.Code)
	} else {
		// Use writeRawPreservingWhitespace for the code content to preserve tabs and spaces
		err = w.writeRawPreservingWhitespace(node.Code)
	}
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

	if !w.opts.strict {
		// Reset the lastWasLineBreak state to false so the closing ``` doesn't get extra indentation
		// The closing ``` should be at the same level as the opening ```
		w.out.lastWasLineBreak = false
	}

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
	if _, ok := node.Parent().(*content.Block); ok && !writesWithoutBullet(node) {
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
			// Check if the first block's first child has PreviousLineTypeBlank
			// to determine if we need an extra newline
			// The first block of a page contains metadata like aliases and filters, and has a blank line after it.
			needsBlankLine := false
			if firstBlock := blocks[0]; firstBlock.FirstChild() != nil {
				if pla, ok := firstBlock.FirstChild().(content.PreviousLineAware); ok {
					if pla.PreviousLineType() == content.PreviousLineTypeBlank {
						needsBlankLine = true
					}
				}
			}

			var err error
			if needsBlankLine {
				err = w.out.WriteString("\n\n")
			} else {
				err = w.out.WriteString("\n")
			}
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

			if writesWithoutBullet(child) {

				// The pre-block is the content before the first bullet of the
				// page, so it is written as is with no marker to indent under.
				err := w.Write(child)
				if err != nil {
					return err
				}

				if child.NextSibling() != nil {
					if !w.opts.strict {
						if _, endsWithProperties := child.LastChild().(*content.Properties); endsWithProperties {
							if err := w.writeRaw("\n"); err != nil {
								return err
							}
						}
					}
					err = w.writeRaw("\n")
					if err != nil {
						return err
					}
				}

				continue
			}

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

// writesWithoutBullet checks if a block is a pre-block in a position where the
// bullet-less form parses back into the same structure, which is as the first
// block of a page.
func writesWithoutBullet(node *content.Block) bool {
	if !node.IsPreBlock() || node.PreviousSibling() != nil {
		return false
	}

	parent, ok := node.Parent().(*content.Block)
	return ok && parent.Parent() == nil
}

func (w *Output) writeProperties(node *content.Properties) error {
	err := w.startBlockWithAutomaticBehavior(node, "", false)
	if err != nil {
		return err
	}

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		// Handle both Property and Alias nodes
		if property, ok := child.(*content.Property); ok {
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
		} else if alias, ok := child.(*content.Alias); ok {
			err := w.writeRaw(content.PropertyNameAlias + ":: ")
			if err != nil {
				return err
			}

			// Use OriginalValue if available to preserve exact spacing
			if alias.OriginalValue != "" {
				err = w.writeRaw(alias.OriginalValue)
				if err != nil {
					return err
				}
			} else {
				// Fallback: write all alias values separated by commas
				isFirst := true
				for aliasChild := alias.FirstChild(); aliasChild != nil; aliasChild = aliasChild.NextSibling() {
					if !isFirst {
						err = w.writeRaw(",")
						if err != nil {
							return err
						}
					}
					isFirst = false

					err = w.Write(aliasChild)
					if err != nil {
						return err
					}
				}
			}
		} else {
			return fmt.Errorf("unsupported properties child: %T", child)
		}

		if child.NextSibling() != nil {
			err = w.writeRaw("\n")
			if err != nil {
				return err
			}
		}
	}

	// Properties at the top of a page should always have a blank line after them,
	// except when followed by a Paragraph, or when the blank line already exists.
	// Check if this Properties node is the first child of the root block.
	if nextSibling := node.NextSibling(); nextSibling != nil && node.PreviousSibling() == nil {
		// Don't add blank line if next sibling is a Paragraph
		if _, isParagraph := nextSibling.(*content.Paragraph); isParagraph {
			goto skipBlankLine
		}

		// Don't add blank line if it already exists in the original input
		if node.HasBlankLineAfter() {
			goto skipBlankLine
		}

		if parent := node.Parent(); parent != nil {
			if block, ok := parent.(*content.Block); ok && block.Parent() == nil {
				// This is a top-level properties block at the start of the page, add blank line
				err = w.writeRaw("\n")
				if err != nil {
					return err
				}
			}
		}
	}
skipBlankLine:

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
	if node.Status == content.TaskStatusNone {
		return nil
	}

	marker := node.Status.String()
	if marker == "" {
		return fmt.Errorf("unsupported task status: %d", node.Status)
	}

	err := w.writeRaw(marker)
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

func (w *Output) writeTaskPriority(node *content.TaskPriority) error {
	var err error
	switch node.Priority {
	case content.PriorityNone:
		return nil
	case content.PriorityA:
		err = w.writeRaw("[#A]")
	case content.PriorityB:
		err = w.writeRaw("[#B]")
	case content.PriorityC:
		err = w.writeRaw("[#C]")
	default:
		return fmt.Errorf("unsupported priority: %d", node.Priority)
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

func (w *Output) writeTaskDate(node *content.TaskDate) error {
	var keyword string
	switch node.Type {
	case content.TaskDateTypeScheduled:
		keyword = "SCHEDULED"
	case content.TaskDateTypeDeadline:
		keyword = "DEADLINE"
	default:
		return fmt.Errorf("unsupported task date type: %d", node.Type)
	}

	value := strings.Builder{}
	value.WriteString(keyword)
	value.WriteString(": <")
	value.WriteString(node.Date.Format("2006-01-02 Mon"))

	if node.HasTime {
		value.WriteString(node.Date.Format(" 15:04"))
	}

	if node.Repeater != nil {
		repeater := node.Repeater.String()
		if repeater == "" {
			return fmt.Errorf("unsupported repeater: %+v", *node.Repeater)
		}

		value.WriteString(" ")
		value.WriteString(repeater)
	}

	value.WriteString(">")

	// Task dates belong to the line above them, so they never get a blank line
	// of their own when the previous line type is automatic.
	err := w.startBlockWithAutomaticBehavior(node, "", false)
	if err != nil {
		return err
	}

	err = w.writeRaw(value.String())
	if err != nil {
		return err
	}

	w.endBlock()
	return nil
}

// logbookLayout is the timestamp format of the graph, which leaves the seconds
// out when the graph is set up without second support.
func (w *Output) logbookLayout() string {
	if w.opts.logbookWithSeconds {
		return logbookTimeLayout
	}

	return logbookTimeLayoutWithoutSeconds
}

// logbookClock formats a clock entry, deriving the duration from the two
// times so it always matches them.
func (w *Output) logbookClock(node *content.LogbookEntryClock) (string, error) {
	layout := w.logbookLayout()

	value := "CLOCK: [" + node.Start.Format(layout) + "]"
	if node.IsRunning() {
		return value, nil
	}

	if node.Duration() < 0 {
		return "", fmt.Errorf("logbook clock ends before it starts: %s", value)
	}

	value += "--[" + node.End.Format(layout) + "] =>  "

	if !w.opts.logbookWithSeconds {
		// The duration is measured between the times as they are written, so
		// that it adds up for whoever reads the entry.
		duration := node.End.Truncate(time.Minute).Sub(node.Start.Truncate(time.Minute))
		return value + fmt.Sprintf(
			"%02d:%02d",
			int(duration/time.Hour),
			int(duration/time.Minute)%60,
		), nil
	}

	duration := node.Duration()
	return value + fmt.Sprintf(
		"%02d:%02d:%02d",
		int(duration/time.Hour),
		int(duration/time.Minute)%60,
		int(duration/time.Second)%60,
	), nil
}

func (w *Output) logbookStateChange(node *content.LogbookEntryStateChange) (string, error) {
	to := node.To.String()
	if to == "" {
		return "", fmt.Errorf("unsupported task status: %d", node.To)
	}

	value := `* State "` + to + `"`

	if node.From != content.TaskStatusNone {
		from := node.From.String()
		if from == "" {
			return "", fmt.Errorf("unsupported task status: %d", node.From)
		}

		value += ` from "` + from + `"`
	}

	return value + " [" + node.Time.Format(w.logbookLayout()) + "]", nil
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
		var value string
		switch e := entry.(type) {
		case *content.LogbookEntryRaw:
			value = e.Value
		case *content.LogbookEntryClock:
			value, err = w.logbookClock(e)
		case *content.LogbookEntryStateChange:
			value, err = w.logbookStateChange(e)
		default:
			return fmt.Errorf("unsupported logbook entry: %T", entry)
		}

		if err != nil {
			return err
		}

		err = w.writeRaw(value)
		if err != nil {
			return err
		}

		if !strings.HasSuffix(value, "\n") {
			err = w.writeRaw("\n")
			if err != nil {
				return err
			}
		}
	}

	err = w.writeRaw(":END:")
	if err != nil {
		return err
	}

	w.endBlock()
	return nil
}
