package content

import "strings"

type PriorityValue int

const (
	PriorityNone PriorityValue = iota
	PriorityHigh
	PriorityMedium
	PriorityLow
)

type Priority struct {
	baseNode

	Priority     PriorityValue
	OriginalText string // Preserves original text for invalid priorities
}

func NewPriority(priority PriorityValue) *Priority {
	return &Priority{
		Priority: priority,
	}
}

// ParsePriorityFromLetter parses a priority from a letter (A, B, C or a, b, c).
func ParsePriorityFromLetter(letter string) PriorityValue {
	switch strings.ToUpper(letter) {
	case "A":
		return PriorityHigh
	case "B":
		return PriorityMedium
	case "C":
		return PriorityLow
	default:
		return PriorityNone
	}
}

// NewPriorityFromString creates a new priority from a letter string.
func NewPriorityFromString(letter string) *Priority {
	return &Priority{
		Priority: ParsePriorityFromLetter(letter),
	}
}

// WithPriority sets the priority value.
func (p *Priority) WithPriority(priority PriorityValue) *Priority {
	p.Priority = priority
	return p
}

func (p *Priority) isInline() {}
