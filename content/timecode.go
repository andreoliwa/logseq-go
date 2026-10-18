package content

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// TimeOfDay is a validated 24-hour clock value. Its zero value is midnight.
type TimeOfDay struct{ hour, minute int }

// NewTimeOfDay validates an hour and minute without guessing invalid values.
func NewTimeOfDay(hour, minute int) (TimeOfDay, bool) {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return TimeOfDay{}, false
	}
	return TimeOfDay{hour: hour, minute: minute}, true
}

// Hour returns the hour from 0 through 23.
func (t TimeOfDay) Hour() int { return t.hour }

// Minute returns the minute from 0 through 59.
func (t TimeOfDay) Minute() int { return t.minute }

// Minutes returns minutes since midnight for comparisons.
func (t TimeOfDay) Minutes() int { return t.hour*60 + t.minute }

// String returns the canonical plain hh:mm representation.
func (t TimeOfDay) String() string { return fmt.Sprintf("%02d:%02d", t.hour, t.minute) }

var leadingTimePattern = regexp.MustCompile(`^([0-9]{1,2})[:hH.,;]([0-9]{2})`)

// LeadingTime discovers a first-line time after optional task and priority
// markers. It reads text through bold/italic wrappers only and never mutates
// the block, resolves references, or searches child blocks or continuation lines.
func LeadingTime(block *Block) (TimeOfDay, bool) {
	_, _, value, ok := locateLeadingTime(block)
	return value, ok
}

func timeParagraph(block *Block) *Paragraph {
	for _, node := range block.Children() {
		if properties, ok := node.(*Properties); ok {
			// Empty containers may be added by Properties() reads. Nonempty
			// properties occupy the first line, so following text is not leading.
			if properties.FirstChild() != nil {
				return nil
			}
			continue
		}
		paragraph, _ := node.(*Paragraph)
		return paragraph
	}
	return nil
}

func timeStart(paragraph *Paragraph) Node {
	node := paragraph.FirstChild()
	if _, ok := node.(*TaskMarker); ok {
		node = node.NextSibling()
	}
	// Prefix nodes carry their own separator when rendered. Some parsed text
	// nodes also retain whitespace between prefixes.
	skipSpace := func() {
		for node != nil {
			text, ok := node.(*Text)
			if !ok || text.HardLineBreak || text.SoftLineBreak || strings.ContainsAny(text.Value, "\r\n") || strings.TrimSpace(text.Value) != "" {
				break
			}
			node = node.NextSibling()
		}
	}
	skipSpace()
	if _, ok := node.(*TaskPriority); ok {
		node = node.NextSibling()
		skipSpace()
	}
	return node
}

// timeText exposes only plain and emphasis text. A sentinel prevents a time
// immediately before code, a link, or another unsupported node from matching
// as though that node were end-of-line.
func timeText(node Node) string {
	switch n := node.(type) {
	case *Text:
		result := n.Value
		if n.HardLineBreak || n.SoftLineBreak {
			result += "\n"
		}
		return result
	case *Emphasis:
		return timeChildrenText(n)
	case *Strong:
		return timeChildrenText(n)
	default:
		return "\x00"
	}
}

func timeChildrenText(parent HasChildren) string {
	var result strings.Builder
	for _, node := range parent.Children() {
		result.WriteString(timeText(node))
	}
	return result.String()
}

func locateLeadingTime(block *Block) (*Paragraph, Node, TimeOfDay, bool) {
	paragraph := timeParagraph(block)
	if paragraph == nil {
		return nil, nil, TimeOfDay{}, false
	}
	start := timeStart(paragraph)
	var text strings.Builder
	for node := start; node != nil; node = node.NextSibling() {
		text.WriteString(timeText(node))
	}
	firstLine, _, _ := strings.Cut(text.String(), "\n")
	match := leadingTimePattern.FindStringSubmatch(firstLine)
	if match == nil {
		return nil, nil, TimeOfDay{}, false
	}
	if len(firstLine) > len(match[0]) {
		next, _ := utf8.DecodeRuneInString(firstLine[len(match[0]):])
		if !unicode.IsSpace(next) {
			return nil, nil, TimeOfDay{}, false
		}
	}
	hour, _ := strconv.Atoi(match[1])
	minute, _ := strconv.Atoi(match[2])
	value, ok := NewTimeOfDay(hour, minute)
	return paragraph, start, value, ok
}

// NormalizeLeadingTime explicitly converts a recognized first-line time to
// plain hh:mm, preserving description formatting, metadata and continuation
// lines. It reports false for canonical, invalid, or absent times.
func NormalizeLeadingTime(block *Block) bool {
	paragraph, start, value, ok := locateLeadingTime(block)
	if !ok {
		return false
	}
	var visible strings.Builder
	for node := start; node != nil; node = node.NextSibling() {
		visible.WriteString(timeText(node))
	}
	source := visible.String()
	match := leadingTimePattern.FindStringSubmatch(source)
	tokenLength := len(match[1]) + 1 + len(match[2])
	if text, plain := start.(*Text); plain && strings.HasPrefix(text.Value, value.String()) {
		return false
	}
	// Lift the separator out of any shared wrappers too: otherwise the
	// description would start with whitespace inside emphasis and cease to parse.
	end := tokenLength
	for _, character := range source[tokenLength:] {
		if !unicode.IsSpace(character) {
			break
		}
		end += utf8.RuneLen(character)
		if character == '\n' {
			break
		}
	}
	remaining := end
	separator := source[tokenLength:end]
	replacement := NewText(value.String() + strings.TrimSuffix(separator, "\n"))
	if strings.HasSuffix(separator, "\n") {
		replacement.WithSoftLineBreak()
	}
	for node := start; node != nil && remaining > 0; {
		next := node.NextSibling()
		consumeTimePrefix(paragraph, node, &remaining, replacement)
		node = next
	}
	// The original start might have been removed. Insert at the original prefix
	// boundary, not before the task or priority marker.
	anchor := timeStart(paragraph)
	if anchor == nil {
		paragraph.AddChild(replacement)
	} else {
		paragraph.InsertChildBefore(replacement, anchor)
	}
	return true
}

// consumeTimePrefix trims only the matched visible prefix, keeping existing
// suffix nodes and wrapper types. Break flags belong to the surviving suffix.
func consumeTimePrefix(parent HasChildren, node Node, remaining *int, replacement *Text) {
	if text, ok := node.(*Text); ok {
		count := min(*remaining, len(text.Value))
		text.Value = text.Value[count:]
		*remaining -= count
		if text.Value == "" && *remaining > 0 && (text.HardLineBreak || text.SoftLineBreak) {
			if text.HardLineBreak {
				replacement.WithHardLineBreak()
			}
			*remaining--
			text.WithNoLineBreak()
		}
		if text.Value == "" && !text.HardLineBreak && !text.SoftLineBreak {
			parent.RemoveChild(node)
		}
		return
	}
	wrapper, ok := node.(HasChildren)
	if !ok {
		return
	}
	for _, child := range wrapper.Children() {
		if *remaining == 0 {
			break
		}
		consumeTimePrefix(wrapper, child, remaining, replacement)
	}
	if wrapper.FirstChild() == nil {
		parent.RemoveChild(node)
	}
}
