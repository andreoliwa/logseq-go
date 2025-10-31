package markdown

import (
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var propertyRegex = regexp.MustCompile(`^([a-zA-Z0-9_-]+)::`)

var propertiesKind = ast.NewNodeKind("Properties")

type properties struct {
	ast.BaseBlock
}

func (*properties) Kind() ast.NodeKind {
	return propertiesKind
}

func (n *properties) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

var propertyKind = ast.NewNodeKind("Property")

type property struct {
	ast.BaseBlock
	Name string
}

func (*property) Kind() ast.NodeKind {
	return propertyKind
}

func (n *property) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

type propertiesASTTransformer struct {
}

var defaultPropertiesASTTransformer = &propertiesASTTransformer{}

// Transform paragraphs by looking for properties. A property in Logseq can
// interrupt a paragraph and looks like `key:: value`.
//
// We go through text nodes and handle those that are on a new line and contain
// a property name and then adopt all of the following text nodes until we find
// a line break.
//
// If a property is found just after another property, it is considered to be
// part of the same properties block.
func (t *propertiesASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	// Collect all Paragraph and TextBlock nodes first, then transform them
	// We can't modify the tree while walking it because it breaks the walker
	var nodesToTransform []ast.Node

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && (n.Kind() == ast.KindParagraph || n.Kind() == ast.KindTextBlock) {
			nodesToTransform = append(nodesToTransform, n)
		}
		return ast.WalkContinue, nil
	})

	// Now transform all collected nodes
	for _, n := range nodesToTransform {
		t.transformTextBlockOrParagraph(n, reader)
	}
}

func (t *propertiesASTTransformer) transformTextBlockOrParagraph(node ast.Node, reader text.Reader) {
	wasPreviousLinebreak := true
	var currentProperties *properties
	var currentProperty *property

	var next ast.Node
	for child := node.FirstChild(); child != nil; child = next {
		next = child.NextSibling()

		if currentProperty != nil {
			// Currently reading the value of a property
			textNode, isText := child.(*ast.Text)
			if isText {
				if !textNode.Segment.IsEmpty() {
					currentProperty.AppendChild(currentProperty, child)
				} else {
					textNode.Parent().RemoveChild(textNode.Parent(), child)
				}
			} else {
				currentProperty.AppendChild(currentProperty, child)
			}

			if isText && (textNode.HardLineBreak() || textNode.SoftLineBreak()) {
				// End of the property value due to a line break
				currentProperty = nil

				textNode.SetHardLineBreak(false)
				textNode.SetSoftLineBreak(false)

				wasPreviousLinebreak = true
			}
		} else {
			textNode, isText := child.(*ast.Text)
			if !isText {
				node = maybeSplitParagraph(node, currentProperties, child)

				currentProperties = nil
				wasPreviousLinebreak = false
				continue
			}

			if wasPreviousLinebreak {
				// Potentially a new property
				potentialName := string(reader.Value(textNode.Segment))

				// In Goldmark the space after :: will either be part of the
				// current text node or the next one.
				matches := propertyRegex.FindStringSubmatchIndex(potentialName)
				if matches == nil {
					// Not a property
					node = maybeSplitParagraph(node, currentProperties, child)
					currentProperties = nil
					wasPreviousLinebreak = textNode.HardLineBreak() || textNode.SoftLineBreak()
					continue
				}

				// Check if there is a space after the ::
				var propertyValueStart int
				if !strings.HasPrefix(potentialName[matches[3]+2:], " ") {
					// There isn't a space after :: in the current text node
					nextTextNode, _ := next.(*ast.Text)
					if startsWithSpace(nextTextNode, reader) {
						// The text node has a space, update it to remove the space
						nextTextNode.Segment = nextTextNode.Segment.WithStart(nextTextNode.Segment.Start + 1)
					} else {
						// The space is missing, not parsing as property
						node = maybeSplitParagraph(node, currentProperties, child)
						currentProperties = nil
						wasPreviousLinebreak = textNode.HardLineBreak() || textNode.SoftLineBreak()
						continue
					}
					propertyValueStart = -1 // Value is in next text node
				} else {
					// The space is in the current text node, so the value starts after "key:: "
					propertyValueStart = matches[3] + 3 // +2 for "::" and +1 for the space
				}

				if currentProperties == nil {
					// This is a new block of properties that splits the paragraph
					currentProperties = &properties{}
					node.Parent().InsertAfter(node.Parent(), node, currentProperties)

					// If the properties is the first child of the paragraph we
					// set explicit blank line information if the paragraph
					// is not the first child at its level
					if child.PreviousSibling() == nil && node.PreviousSibling() != nil {
						currentProperties.SetBlankPreviousLines(node.HasBlankPreviousLines())
					}

					// The paragraph does not have any blank lines before it
					node.SetBlankPreviousLines(false)
				}

				currentProperty = &property{}
				currentProperty.Name = potentialName[matches[2]:matches[3]]

				currentProperties.AppendChild(currentProperties, currentProperty)

				// The previous node no longer has a line break
				if previousTextNode, ok := textNode.PreviousSibling().(*ast.Text); ok {
					previousTextNode.SetHardLineBreak(false)
					previousTextNode.SetSoftLineBreak(false)
				}

				// Handle the property value
				if propertyValueStart >= 0 {
					// The property value is in the same text node
					propertyValue := potentialName[propertyValueStart:]
					if len(propertyValue) > 0 {
						// Create a new text node for the property value
						valueTextNode := ast.NewTextSegment(textNode.Segment.WithStart(textNode.Segment.Start + propertyValueStart))
						valueTextNode.SetHardLineBreak(textNode.HardLineBreak())
						valueTextNode.SetSoftLineBreak(textNode.SoftLineBreak())
						currentProperty.AppendChild(currentProperty, valueTextNode)
					}
				}

				// Remove the text node with the parameter name
				node.RemoveChild(node, child)
			} else {
				wasPreviousLinebreak = textNode.HardLineBreak() || textNode.SoftLineBreak()
			}
		}
	}

	if node.FirstChild() == nil {
		// The paragraph is now empty
		node.Parent().RemoveChild(node.Parent(), node)
	}
}

func startsWithSpace(node *ast.Text, reader text.Reader) bool {
	if node == nil {
		return false
	}

	value := string(reader.Value(node.Segment))
	return strings.HasPrefix(value, " ")
}

func maybeSplitParagraph(node ast.Node, divider *properties, firstChildOfNewParagraph ast.Node) ast.Node {
	if divider == nil {
		return node
	}

	newParagraph := &ast.Paragraph{}
	node.Parent().InsertAfter(node.Parent(), divider, newParagraph)

	for child := firstChildOfNewParagraph; child != nil; {
		next := child.NextSibling()
		newParagraph.AppendChild(newParagraph, child)
		child = next
	}

	if node.FirstChild() == nil {
		// The paragraph is now empty, remove it
		node.Parent().RemoveChild(node.Parent(), node)
	}

	return newParagraph
}
