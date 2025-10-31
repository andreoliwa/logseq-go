package content

// Properties is a collection of Property nodes.
type Properties struct {
	baseNodeWithChildren
	previousLineAwareImpl
	// hasBlankLineAfter indicates whether there was a blank line after the properties in the original input.
	// This is used during output to determine if a blank line should be added.
	hasBlankLineAfter bool
}

// NewProperties creates a new Properties node with the given Property or Alias children.
func NewProperties(children ...Node) *Properties {
	p := &Properties{}
	p.self = p
	p.childValidator = allowOnlyPropertiesOrAlias
	for _, child := range children {
		p.AddChild(child)
	}
	return p
}

func (p *Properties) WithPreviousLineType(t PreviousLineType) *Properties {
	p.previousLineType = t
	return p
}

// HasBlankLineAfter returns whether there was a blank line after the properties in the original input
func (p *Properties) HasBlankLineAfter() bool {
	return p.hasBlankLineAfter
}

// SetHasBlankLineAfter sets whether there was a blank line after the properties
func (p *Properties) SetHasBlankLineAfter(hasBlankLineAfter bool) {
	p.hasBlankLineAfter = hasBlankLineAfter
}

// GetAsNode gets a Property node by name. Will return nil if no Property with the given name exists.
func (p *Properties) GetAsNode(key string) *Property {
	for _, child := range p.Children() {
		if property, ok := child.(*Property); ok && property.Name == key {
			return property
		}
	}

	return nil
}

// Get gets the value of a Property by name. Will return an empty slice if no Property with the given name exists.
func (p *Properties) Get(key string) NodeList {
	property := p.GetAsNode(key)
	if property == nil {
		return NodeList{}
	}

	return property.Children()
}

// Set a Property node by name. If a Property with the given name already exists, it will be replaced.
func (p *Properties) Set(key string, nodes ...Node) {
	property := p.GetAsNode(key)
	if property == nil {
		property = NewProperty(key)
		p.AddChild(property)
	}

	property.SetChildren(nodes...)
}

// Remove a Property node by name. If a Property with the given name does not exist this does nothing.
func (p *Properties) Remove(key string) {
	property := p.GetAsNode(key)
	if property != nil {
		p.RemoveChild(property)
	}
}

func (p *Properties) debug(p2 *debugPrinter) {
	p2.StartType("Properties")
	debugPreviousLineAware(p2, p)
	p2.Children(p)
	p2.EndType()
}

func (p *Properties) isBlock() {}

var _ Node = (*Properties)(nil)
var _ BlockNode = (*Properties)(nil)

// Property is a node that represents a property, which is a key that can have multiple values.
type Property struct {
	baseNodeWithChildren

	// Name is the name of the property.
	Name string
}

// NewProperty creates a new Property node with the given name and values.
func NewProperty(name string, children ...Node) *Property {
	property := &Property{Name: name}
	property.self = property
	property.childValidator = allowOnlyInlineNodes
	property.AddChildren(children...)
	return property
}

// WithName sets the name of the property.
func (p *Property) WithName(name string) *Property {
	p.Name = name
	return p
}

func (p *Property) debug(p2 *debugPrinter) {
	p2.StartType("Property")
	p2.Field("Name", p.Name)
	p2.Children(p)
	p2.EndType()
}

func allowOnlyPropertiesOrAlias(node Node) bool {
	_, okProperty := node.(*Property)
	_, okAlias := node.(*Alias)
	return okProperty || okAlias
}

// PropertyNameAlias is the name of the alias property.
const PropertyNameAlias = "alias"

// Alias is a property that represents an alias, which is an alternative name for a page.
type Alias struct {
	baseNodeWithChildren

	// OriginalValue stores the original text with spacing for output preservation.
	// This is used to maintain the exact formatting when writing back to markdown.
	OriginalValue string
}

// NewAlias creates a new Alias node with the given values.
func NewAlias(children ...Node) *Alias {
	alias := &Alias{}
	alias.self = alias
	alias.childValidator = allowOnlyInlineNodes
	alias.AddChildren(children...)
	return alias
}

func (a *Alias) debug(p2 *debugPrinter) {
	p2.StartType("Alias")
	p2.Children(a)
	p2.EndType()
}
