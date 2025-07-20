package markdown

import (
	"unicode"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var tagKind = ast.NewNodeKind("Hashtag")

type tag struct {
	ast.BaseInline
	Page string
}

func (*tag) Kind() ast.NodeKind {
	return tagKind
}

func (n *tag) Dump(src []byte, level int) {
}

// tagParser parses Logseq style tags in Goldmark.
type tagParser struct {
}

func (t *tagParser) Trigger() []byte {
	return []byte{'#'}
}

func (t *tagParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, seg := block.PeekLine()

	if len(line) == 0 || line[0] != '#' {
		return nil
	}

	// Check if this looks like it's part of a Markdown link
	// We need to detect hashtags that are inside link text, but be careful not to
	// interfere with regular hashtags that just happen to be on the same line as a link
	if len(line) > 1 {
		// First, find where the hashtag content ends (space or special char)
		hashtagEnd := 1
		for i := 1; i < len(line); i++ {
			r := rune(line[i])
			if unicode.IsSpace(r) || r == ']' || r == ')' {
				hashtagEnd = i
				break
			}
		}

		// Pattern 1: #something](url) - hashtag immediately followed by ](
		if hashtagEnd < len(line) && line[hashtagEnd] == ']' && hashtagEnd+1 < len(line) && line[hashtagEnd+1] == '(' {
			// Found pattern #something](url) - we're definitely in link text
			return nil
		}

		// Pattern 2: #something) - hashtag in URL followed by )
		if hashtagEnd < len(line) && line[hashtagEnd] == ')' {
			// Check if there's a reasonable URL-like pattern before the #
			beforeHash := line[1:hashtagEnd]
			if len(beforeHash) > 0 && (containsURLPattern(beforeHash) || containsDomainPattern(beforeHash)) {
				return nil
			}
		}

		// Pattern 3: #something followed by other content and then ](url)
		// This handles cases like "#first #second #third](url)" where we need to detect
		// that we're inside link text even when there's content after the hashtag
		if hashtagEnd < len(line) {
			// Look for ](url) pattern after the hashtag, but only if there's no
			// intervening content that would suggest we're not in a link
			for i := hashtagEnd; i < len(line)-1; i++ {
				if line[i] == ']' && line[i+1] == '(' {
					// Found ](url) pattern
					// Check if there's a reasonable URL after the (
					urlStart := i + 2
					if urlStart < len(line) {
						// Look for the closing ) to find the URL
						for j := urlStart; j < len(line); j++ {
							if line[j] == ')' {
								url := line[urlStart:j]
								// If it looks like a URL, and we haven't seen any content that
								// would suggest we're not in a link, skip this hashtag
								if len(url) > 0 && (containsURLPattern(url) || containsDomainPattern(url) || len(url) > 3) {
									// Additional check: make sure there's no standalone text before the ]
									// that would suggest this hashtag is not part of the link
									textBetween := line[hashtagEnd:i]
									// If the text between hashtag and ] contains only spaces and other hashtags,
									// then we're likely inside link text
									if isLinkTextContent(textBetween) {
										return nil
									}
								}
								break
							}
						}
					}
					break
				}
			}
		}
	}

	line = line[1:]

	end := 0
	var value []byte

	// Check for extended tag syntax which wraps tags in [[...]], this is how
	// Logseq supports tags with spaces.
	if len(line) > 1 && line[0] == '[' && line[1] == '[' {
		// This is an extended tag so scan until the closing ]].
		for i := 1; i < len(line)-1; i++ {
			if line[i] == ']' && line[i+1] == ']' {
				end = i
				break
			}
		}

		// Didn't find the closing ]], so this isn't a tag.
		if end == 0 {
			return nil
		}

		// The value of the tag is the text between the [[ and ]].
		value = line[2:end]
		end += 2
	} else {
		// TODO: Does Logseq support Unicode tags?
		// Scan until a Unicode space character is found.
		for i, r := range line {
			if unicode.IsSpace(rune(r)) {
				end = i
				break
			}
		}

		if end == 0 {
			// No space found, assume the tag is until end of line.
			end = len(line)
		}

		value = line[:end]
	}

	seg = seg.WithStop(seg.Start + end + 1)

	n := tag{
		Page: string(value),
	}
	block.Advance(seg.Len())
	return &n
}

var _ parser.InlineParser = (*tagParser)(nil)

// containsURLPattern checks if the byte slice contains URL-like patterns
func containsURLPattern(b []byte) bool {
	s := string(b)
	return len(s) > 4 && ((len(s) > 7 && (s[:7] == "http://" || s[:8] == "https://")) ||
		(len(s) > 6 && s[:6] == "ftp://"))
}

// containsDomainPattern checks if the byte slice looks like a domain or path
func containsDomainPattern(b []byte) bool {
	s := string(b)
	// Look for domain-like patterns: contains dots and alphanumeric characters
	hasDot := false
	hasAlphaNum := false
	for _, r := range s {
		if r == '.' {
			hasDot = true
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			hasAlphaNum = true
		}
	}
	return hasDot && hasAlphaNum && len(s) > 3
}

// isLinkTextContent checks if the byte slice contains only content that would
// typically be found inside link text (spaces, hashtags, alphanumeric chars)
func isLinkTextContent(b []byte) bool {
	for _, char := range b {
		r := rune(char)
		// Allow spaces, hashtags, and typical text characters
		if !unicode.IsSpace(r) && r != '#' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			// If we find other special characters, it might not be link text
			return false
		}
	}
	return true
}
