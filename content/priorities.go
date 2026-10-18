package content

import "strings"

type Priority int

const (
	PriorityNone Priority = iota
	PriorityA
	PriorityB
	PriorityC
)

// Readable aliases for the upstream letter priorities.
const (
	PriorityHigh   = PriorityA
	PriorityMedium = PriorityB
	PriorityLow    = PriorityC
)

type TaskPriority struct {
	baseNode

	Priority     Priority
	OriginalText string // Preserves original text for invalid priorities
}

func NewPriority(priority Priority) *TaskPriority {
	return &TaskPriority{
		Priority: priority,
	}
}

// ParsePriorityFromLetter parses a priority from a letter (A, B, C or a, b, c).
func ParsePriorityFromLetter(letter string) Priority {
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
func NewPriorityFromString(letter string) *TaskPriority {
	return &TaskPriority{
		Priority: ParsePriorityFromLetter(letter),
	}
}

// WithPriority sets the priority value.
func (p *TaskPriority) WithPriority(priority Priority) *TaskPriority {
	p.Priority = priority
	return p
}

func (p *TaskPriority) isInline() {}
