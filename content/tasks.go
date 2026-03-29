package content

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// TaskStatus is the type of task.
type TaskStatus int

const (
	TaskStatusNone TaskStatus = iota
	// TaskStatusTodo is a TODO task.
	TaskStatusTodo
	// TaskStatusDoing is a DOING task.
	TaskStatusDoing
	// TaskStatusDone is a DONE task.
	TaskStatusDone
	// TaskStatusLater is a LATER task.
	TaskStatusLater
	// TaskStatusNow is a NOW task.
	TaskStatusNow
	// TaskStatusCancelled is a CANCELLED task.
	TaskStatusCancelled
	// TaskStatusCanceled is a CANCELED task.
	TaskStatusCanceled
	// TaskStatusInProgress is a IN-PROGRESS task.
	TaskStatusInProgress
	// TaskStatusWait is a WAIT task.
	TaskStatusWait
	// TaskStatusWaiting is a WAITING task.
	TaskStatusWaiting
)

// Task status string constants for use in comparisons and content parsing.
const (
	TaskStringTodo       = "TODO"
	TaskStringDoing      = "DOING"
	TaskStringDone       = "DONE"
	TaskStringLater      = "LATER"
	TaskStringNow        = "NOW"
	TaskStringCancelled  = "CANCELLED"
	TaskStringCanceled   = "CANCELED"
	TaskStringInProgress = "IN-PROGRESS"
	TaskStringWait       = "WAIT"
	TaskStringWaiting    = "WAITING"
)

// taskStatusString maps each TaskStatus to its uppercase string representation.
var taskStatusString = map[TaskStatus]string{
	TaskStatusNone:       "",
	TaskStatusTodo:       TaskStringTodo,
	TaskStatusDoing:      TaskStringDoing,
	TaskStatusDone:       TaskStringDone,
	TaskStatusLater:      TaskStringLater,
	TaskStatusNow:        TaskStringNow,
	TaskStatusCancelled:  TaskStringCancelled,
	TaskStatusCanceled:   TaskStringCanceled,
	TaskStatusInProgress: TaskStringInProgress,
	TaskStatusWait:       TaskStringWait,
	TaskStatusWaiting:    TaskStringWaiting,
}

// String returns the uppercase string representation of the task status (e.g., "DOING", "DONE", etc.).
// Returns an empty string for TaskStatusNone.
func (t TaskStatus) String() string {
	if s, ok := taskStatusString[t]; ok {
		return s
	}
	return ""
}

// TaskStatusStrings returns all valid task status strings (e.g., "DOING", "DONE", etc.).
// Excludes TaskStatusNone.
func TaskStatusStrings() []string {
	statuses := make([]string, 0, len(taskStatusString)-1)
	for status, str := range taskStatusString {
		if status != TaskStatusNone {
			statuses = append(statuses, str)
		}
	}
	return statuses
}

// TaskCategory represents a category of task statuses.
// Multiple task statuses can belong to the same category (e.g., DOING, NOW, IN-PROGRESS).
type TaskCategory int

const (
	TaskCategoryNone TaskCategory = iota
	TaskCategoryTodo
	TaskCategoryDoing
	TaskCategoryCancelled
	TaskCategoryWaiting
	TaskCategoryDone
)

// taskStatusToCategory maps each TaskStatus to its TaskCategory.
var taskStatusToCategory = map[TaskStatus]TaskCategory{
	TaskStatusNone:       TaskCategoryNone,
	TaskStatusTodo:       TaskCategoryTodo,
	TaskStatusLater:      TaskCategoryTodo,
	TaskStatusDoing:      TaskCategoryDoing,
	TaskStatusNow:        TaskCategoryDoing,
	TaskStatusInProgress: TaskCategoryDoing,
	TaskStatusCancelled:  TaskCategoryCancelled,
	TaskStatusCanceled:   TaskCategoryCancelled,
	TaskStatusWaiting:    TaskCategoryWaiting,
	TaskStatusWait:       TaskCategoryWaiting,
	TaskStatusDone:       TaskCategoryDone,
}

// Category returns the category of the task status.
func (t TaskStatus) Category() TaskCategory {
	if category, ok := taskStatusToCategory[t]; ok {
		return category
	}
	return TaskCategoryNone
}

type TaskMarker struct {
	baseNode

	Status TaskStatus

	// parentParagraph is a reference to the parent Paragraph node.
	// This is set during parsing and used for accessing the parent block.
	parentParagraph *Paragraph

	// parentBlock is a reference to the grandparent Block node.
	// This is set during parsing and used for transition side effects.
	parentBlock *Block

	// timeNow is a function that returns the current time.
	// This is used for dependency injection in tests.
	timeNow func() time.Time
}

func NewTaskMarker(t TaskStatus) *TaskMarker {
	return &TaskMarker{
		Status:  t,
		timeNow: time.Now,
	}
}

// SetParentReferences stores references to the parent Paragraph and grandparent Block.
// This is called during parsing to enable transition side effects in WithStatus.
func (t *TaskMarker) SetParentReferences(paragraph *Paragraph, block *Block) {
	t.parentParagraph = paragraph
	t.parentBlock = block
}

// ParentParagraph returns the parent Paragraph reference.
func (t *TaskMarker) ParentParagraph() *Paragraph {
	return t.parentParagraph
}

// ParentBlock returns the parent Block reference.
func (t *TaskMarker) ParentBlock() *Block {
	return t.parentBlock
}

// SetTimeNow sets the time provider function for testing.
func (t *TaskMarker) SetTimeNow(timeNow func() time.Time) {
	t.timeNow = timeNow
}

// ParseTaskStatus parses a task status from a TO DO/DOING/DONE/etc. string.
// Logseq only considers uppercase task statuses.
func ParseTaskStatus(status string) TaskStatus {
	switch status {
	case TaskStringTodo:
		return TaskStatusTodo
	case TaskStringDone:
		return TaskStatusDone
	case TaskStringDoing:
		return TaskStatusDoing
	case TaskStringLater:
		return TaskStatusLater
	case TaskStringNow:
		return TaskStatusNow
	case TaskStringCancelled:
		return TaskStatusCancelled
	case TaskStringCanceled:
		return TaskStatusCanceled
	case TaskStringInProgress:
		return TaskStatusInProgress
	case TaskStringWait:
		return TaskStatusWait
	case TaskStringWaiting:
		return TaskStatusWaiting
	default:
		return TaskStatusNone
	}
}

// NewTaskMarkerFromString creates a new task marker from a TO DO/DOING/DONE/etc. string.
func NewTaskMarkerFromString(t string) *TaskMarker {
	return &TaskMarker{
		Status: ParseTaskStatus(t),
	}
}

// WithStatus sets the status of the task marker and performs any necessary
// transition side effects (e.g., stopping running clocks, removing properties).
// Returns an error if the transition fails.
func (t *TaskMarker) WithStatus(status TaskStatus) (*TaskMarker, error) {
	oldStatus := t.Status

	// Only execute transitions if we have parent block and status is changing
	if t.parentBlock != nil && status != oldStatus {
		if status == TaskStatusTodo {
			if err := executeTransitionToTodo(oldStatus, t.parentBlock, t.timeNow()); err != nil {
				return nil, err
			}
		}
		// TODO: Future transitions to other statuses will be added here
	}

	t.Status = status
	return t, nil
}

func (t *TaskMarker) debug(p *debugPrinter) {
	p.StartType("TaskMarker")
	switch t.Status {
	case TaskStatusNone:
		p.Field("type", "none")
	case TaskStatusTodo:
		p.Field("type", "todo")
	case TaskStatusDoing:
		p.Field("type", "doing")
	case TaskStatusDone:
		p.Field("type", "done")
	case TaskStatusLater:
		p.Field("type", "later")
	case TaskStatusNow:
		p.Field("type", "now")
	case TaskStatusCancelled:
		p.Field("type", "cancelled")
	case TaskStatusInProgress:
		p.Field("type", "in-progress")
	case TaskStatusWait:
		p.Field("type", "wait")
	case TaskStatusWaiting:
		p.Field("type", "waiting")
	}
	p.EndType()
}

func (t *TaskMarker) isInline() {}

// Logbook represents a logbook of a task. Logseq will manage these
// automatically when a task changes state. They are used both for tracking if
// a task has been completed, for use with repeating tasks and for time tracking
// if the user has enabled that feature.
//
// These are commonly part of a `Block` with a task marker.
//
// A logbook node can only contain children of type `LogbookEntry`.
// TODO: feat: add support for Clock entries in a Logbook and maybe LogbookEntryRaw won't be needed
type Logbook struct {
	baseNodeWithChildren
	previousLineAwareImpl
}

func NewLogbook(entries ...LogbookEntry) *Logbook {
	l := &Logbook{}
	l.self = l
	l.childValidator = allowOnlyLogbookEntries
	for _, entry := range entries {
		l.AddChild(entry)
	}
	return l
}

func (l *Logbook) WithPreviousLineType(t PreviousLineType) *Logbook {
	l.previousLineType = t
	return l
}

func (l *Logbook) debug(p *debugPrinter) {
	p.StartType("TaskLogbook")
	debugPreviousLineAware(p, l)
	p.Children(l)
	p.EndType()
}

func (l *Logbook) isBlock() {}

var _ BlockNode = (*Logbook)(nil)

// LogbookEntry represents a single entry in a logbook.
type LogbookEntry interface {
	Node
	isLogbookEntry()
}

// LogbookEntryRaw represents a raw logbook entry, this is used for entries that
// are not supported by this library.
type LogbookEntryRaw struct {
	baseNode
	Value string
}

func NewLogbookEntryRaw(value string) *LogbookEntryRaw {
	return &LogbookEntryRaw{
		Value: value,
	}
}

// WithValue sets the value of the logbook entry.
func (t *LogbookEntryRaw) WithValue(value string) *LogbookEntryRaw {
	t.Value = value
	return t
}

func (t *LogbookEntryRaw) debug(p *debugPrinter) {
	p.StartType("LogbookEntryRaw")
	p.Field("value", t.Value)
	p.EndType()
}

func (t *LogbookEntryRaw) isLogbookEntry() {}

func allowOnlyLogbookEntries(n Node) bool {
	_, ok := n.(LogbookEntry)
	return ok
}

// ClockEntryData represents a parsed CLOCK entry.
// CLOCK entries remain stored as raw strings in LogbookEntryRaw, but this struct
// is used for parsing and formatting them.
type ClockEntryData struct {
	StartTime time.Time
	EndTime   *time.Time // nil if clock is running
	Duration  *time.Duration
}

// Logseq CLOCK format: "CLOCK: [2023-06-26 Mon 17:25:56]--[2023-06-26 Mon 17:25:58] =>  00:00:01"
// Running CLOCK format: "CLOCK: [2023-06-26 Mon 17:25:56]"
const logseqTimeFormat = "2006-01-02 Mon 15:04:05"

var clockEntryRegex = regexp.MustCompile(`^CLOCK: \[([^\]]+)\](?:--\[([^\]]+)\])?(?: => +(.+))?$`)

// ParseClockEntry parses a CLOCK entry string into a ClockEntryData struct.
// Returns an error if the format is invalid.
// Times are parsed in the local timezone since Logseq stores times without timezone information.
func ParseClockEntry(raw string) (*ClockEntryData, error) {
	matches := clockEntryRegex.FindStringSubmatch(raw)
	if matches == nil {
		return nil, fmt.Errorf("invalid CLOCK entry format: %s", raw)
	}

	startTime, err := time.ParseInLocation(logseqTimeFormat, matches[1], time.Local)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %w", err)
	}

	entry := &ClockEntryData{
		StartTime: startTime,
	}

	// Parse end time if present (matches[2])
	if matches[2] != "" {
		endTime, err := time.ParseInLocation(logseqTimeFormat, matches[2], time.Local)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end time: %w", err)
		}
		entry.EndTime = &endTime
	}

	// Parse duration if present (matches[3])
	if matches[3] != "" {
		duration, err := parseDuration(strings.TrimSpace(matches[3]))
		if err != nil {
			return nil, fmt.Errorf("failed to parse duration: %w", err)
		}
		entry.Duration = &duration
	}

	return entry, nil
}

// parseDuration parses a duration string in the format "HH:MM:SS".
func parseDuration(s string) (time.Duration, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	var hours, minutes, seconds int
	_, err := fmt.Sscanf(s, "%d:%d:%d", &hours, &minutes, &seconds)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second, nil
}

// CalculateClockDuration calculates the duration between start and end times.
// This is useful for calculating CLOCK entry durations.
func CalculateClockDuration(start, end time.Time) time.Duration {
	return end.Sub(start)
}

// FormatClockEntry formats a ClockEntryData struct back into a CLOCK entry string.
func FormatClockEntry(entry *ClockEntryData) string {
	result := fmt.Sprintf("CLOCK: [%s]", entry.StartTime.Format(logseqTimeFormat))

	if entry.EndTime != nil {
		result += fmt.Sprintf("--[%s]", entry.EndTime.Format(logseqTimeFormat))

		if entry.Duration != nil {
			totalSeconds := int(entry.Duration.Seconds())
			hours := totalSeconds / 3600
			minutes := (totalSeconds % 3600) / 60
			seconds := totalSeconds % 60
			result += fmt.Sprintf(" =>  %02d:%02d:%02d", hours, minutes, seconds)
		}
	}

	return result
}

// IsRunningClock checks if a CLOCK entry is running (has no end time).
func IsRunningClock(raw string) bool {
	entry, err := ParseClockEntry(raw)
	if err != nil {
		return false
	}
	return entry.EndTime == nil
}

// StopClock stops a running CLOCK entry by adding the end time and calculating duration.
// Returns the updated CLOCK entry string.
func StopClock(raw string, now time.Time) (string, error) {
	entry, err := ParseClockEntry(raw)
	if err != nil {
		return "", err
	}

	if entry.EndTime != nil {
		return "", fmt.Errorf("CLOCK entry is not running")
	}

	entry.EndTime = &now
	duration := CalculateClockDuration(entry.StartTime, now)
	entry.Duration = &duration

	return FormatClockEntry(entry), nil
}

// findLogbookInBlock finds the Logbook node in a Block's children.
// Returns nil if no Logbook is found.
func findLogbookInBlock(block *Block) *Logbook {
	for child := block.FirstChild(); child != nil; child = child.NextSibling() {
		if logbook, ok := child.(*Logbook); ok {
			return logbook
		}
	}
	return nil
}

// stopRunningClockInBlock finds a running CLOCK entry in the block's logbook
// and stops it by setting the end time and calculating the duration.
// Returns (true, nil) if a CLOCK entry was stopped.
// Returns (false, nil) if there was no running CLOCK or no logbook.
// Returns (false, error) if stopping the CLOCK failed.
func stopRunningClockInBlock(block *Block, now time.Time) (bool, error) {
	logbook := findLogbookInBlock(block)
	if logbook == nil {
		return false, nil // No logbook, nothing to do
	}

	// Iterate through logbook entries to find a running CLOCK
	// TODO: if there are multiple running clocks, this is an invalid state that cannot be fixed automatically
	for child := logbook.FirstChild(); child != nil; child = child.NextSibling() {
		entry, ok := child.(*LogbookEntryRaw)
		if !ok {
			continue
		}

		if IsRunningClock(entry.Value) {
			stoppedClock, err := StopClock(entry.Value, now)
			if err != nil {
				return false, fmt.Errorf("failed to stop running CLOCK: %w", err)
			}
			entry.Value = stoppedClock
			return true, nil // Successfully stopped the clock
		}
	}

	return false, nil // No running CLOCK found, nothing to do
}

// removePropertyFromBlock removes a property from the block's properties.
// Does nothing if the property doesn't exist.
func removePropertyFromBlock(block *Block, propertyName string) {
	properties := block.Properties()
	properties.Remove(propertyName)
}

// executeTransitionToTodo executes the side effects when transitioning to TODO status.
// This implements the state machine logic for transitions TO TaskStatusTodo.
func executeTransitionToTodo(fromStatus TaskStatus, block *Block, now time.Time) error {
	fromCategory := fromStatus.Category()

	switch fromCategory {
	case TaskCategoryNone, TaskCategoryTodo, TaskCategoryWaiting:
		// No action needed for these transitions
		return nil

	case TaskCategoryDoing:
		// Stop any running CLOCK entries
		_, err := stopRunningClockInBlock(block, now)
		return err

	case TaskCategoryDone:
		// TODO: feat: read name of completed property from "DONE task property" plugin
		//  ❯ rg completed ~/.logseq/settings/
		//  ~/.logseq/settings/logseq-plugin-confirmation-done-task.json
		//  17:  "customPropertyName": "completed",
		removePropertyFromBlock(block, "completed")
		return nil

	case TaskCategoryCancelled:
		// TODO: feat: read name of cancelled property from "DONE task property" plugin
		//  ❯ rg cancelled ~/.logseq/settings/
		//  ~/.logseq/settings/logseq-plugin-confirmation-done-task.json
		//  29:  "cancelledTaskPropertyName": "cancelled",
		//  30:  "cancelledTaskTime": false,
		//  31:  "removePropertyWithoutCANCELLEDtask": true,
		removePropertyFromBlock(block, "cancelled")
		return nil

	default:
		// Unknown category - no action
		return nil
	}
}
