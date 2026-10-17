package content

import (
	"fmt"
	"regexp"
	"strconv"
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

// TaskPriority is the fork's existing priority node under an additive name.
type TaskPriority = Priority

const (
	PriorityA = PriorityHigh
	PriorityB = PriorityMedium
	PriorityC = PriorityLow
)

func NewTaskPriority(priority PriorityValue) *TaskPriority { return NewPriority(priority) }

// TaskDateType is the type of a date that is attached to a task.
type TaskDateType int

const (
	// TaskDateTypeScheduled is a `SCHEDULED` date, which is when work on a
	// task is intended to start.
	TaskDateTypeScheduled TaskDateType = iota
	// TaskDateTypeDeadline is a `DEADLINE` date, which is when a task has to
	// be finished.
	TaskDateTypeDeadline
)

// RepeaterType controls how the next occurrence of a repeating date is picked
// when the task is completed.
type RepeaterType int

const (
	// RepeaterTypeCumulate, written as `+`, moves the date forward one
	// interval, even if that leaves it in the past.
	RepeaterTypeCumulate RepeaterType = iota
	// RepeaterTypeCatchUp, written as `++`, moves the date forward whole
	// intervals until it is in the future.
	RepeaterTypeCatchUp
	// RepeaterTypeRestart, written as `.+`, moves the date to one interval
	// after the day the task was completed.
	RepeaterTypeRestart
)

// RepeaterUnit is the unit of the interval of a Repeater.
type RepeaterUnit int

const (
	// RepeaterUnitHour repeats every N hours, written as `h`.
	RepeaterUnitHour RepeaterUnit = iota
	// RepeaterUnitDay repeats every N days, written as `d`.
	RepeaterUnitDay
	// RepeaterUnitWeek repeats every N weeks, written as `w`.
	RepeaterUnitWeek
	// RepeaterUnitMonth repeats every N months, written as `m`.
	RepeaterUnitMonth
	// RepeaterUnitYear repeats every N years, written as `y`.
	RepeaterUnitYear
)

// Repeater describes how a `SCHEDULED` or `DEADLINE` date repeats when the
// task it belongs to is completed.
type Repeater struct {
	// Type is how the next occurrence is picked.
	Type RepeaterType
	// Value is how many units to move the date forward by.
	Value int
	// Unit is the unit of Value.
	Unit RepeaterUnit
}

func NewRepeater(repeaterType RepeaterType, value int, unit RepeaterUnit) *Repeater {
	return &Repeater{
		Type:  repeaterType,
		Value: value,
		Unit:  unit,
	}
}

// String formats the repeater the way it is stored in Markdown, such as `.+1d`.
// Returns an empty string if the type or unit is not one of the known values.
func (r *Repeater) String() string {
	var prefix string
	switch r.Type {
	case RepeaterTypeCumulate:
		prefix = "+"
	case RepeaterTypeCatchUp:
		prefix = "++"
	case RepeaterTypeRestart:
		prefix = ".+"
	default:
		return ""
	}

	var unit string
	switch r.Unit {
	case RepeaterUnitHour:
		unit = "h"
	case RepeaterUnitDay:
		unit = "d"
	case RepeaterUnitWeek:
		unit = "w"
	case RepeaterUnitMonth:
		unit = "m"
	case RepeaterUnitYear:
		unit = "y"
	default:
		return ""
	}

	return prefix + strconv.Itoa(r.Value) + unit
}

// TaskDate is a `SCHEDULED` or `DEADLINE` date belonging to a task. Logseq
// writes these on their own lines directly after the content of the block that
// has the task marker:
//
//	TODO Water the plants
//	SCHEDULED: <2024-01-15 Mon 09:00 .+3d>
//
// The date can carry a time of day and a Repeater, both of which are optional.
type TaskDate struct {
	baseNode
	previousLineAwareImpl

	// Type is whether this is a scheduled date or a deadline.
	Type TaskDateType

	// Date is the day this date refers to. When HasTime is set the time of day
	// is part of the date, otherwise it is midnight and not written out.
	Date time.Time

	// HasTime indicates if the time of day of Date is meaningful.
	HasTime bool

	// Repeater describes how the date repeats, or nil if it does not repeat.
	Repeater *Repeater
}

// NewTaskDate creates a date without a time of day. Any time of day in the
// given time is dropped.
func NewTaskDate(dateType TaskDateType, date time.Time) *TaskDate {
	return &TaskDate{
		Type: dateType,
		Date: truncateToDay(date),
	}
}

// NewTaskDateWithTime creates a date that includes the time of day. Seconds
// and smaller units are dropped, as Logseq stores minute precision.
func NewTaskDateWithTime(dateType TaskDateType, date time.Time) *TaskDate {
	return &TaskDate{
		Type:    dateType,
		Date:    truncateToMinute(date),
		HasTime: true,
	}
}

// NewScheduled creates a `SCHEDULED` date without a time of day.
func NewScheduled(date time.Time) *TaskDate {
	return NewTaskDate(TaskDateTypeScheduled, date)
}

// NewScheduledWithTime creates a `SCHEDULED` date that includes the time of day.
func NewScheduledWithTime(date time.Time) *TaskDate {
	return NewTaskDateWithTime(TaskDateTypeScheduled, date)
}

// NewDeadline creates a `DEADLINE` date without a time of day.
func NewDeadline(date time.Time) *TaskDate {
	return NewTaskDate(TaskDateTypeDeadline, date)
}

// NewDeadlineWithTime creates a `DEADLINE` date that includes the time of day.
func NewDeadlineWithTime(date time.Time) *TaskDate {
	return NewTaskDateWithTime(TaskDateTypeDeadline, date)
}

// WithDate sets the date, dropping any time of day.
func (t *TaskDate) WithDate(date time.Time) *TaskDate {
	t.Date = truncateToDay(date)
	t.HasTime = false
	return t
}

// WithDateAndTime sets the date including its time of day.
func (t *TaskDate) WithDateAndTime(date time.Time) *TaskDate {
	t.Date = truncateToMinute(date)
	t.HasTime = true
	return t
}

// WithRepeater sets the repeater of this date. Passing nil makes the date
// non-repeating.
func (t *TaskDate) WithRepeater(repeater *Repeater) *TaskDate {
	t.Repeater = repeater
	return t
}

func (t *TaskDate) WithPreviousLineType(previousLineType PreviousLineType) *TaskDate {
	t.previousLineType = previousLineType
	return t
}

func (t *TaskDate) debug(p *debugPrinter) {
	p.StartType("TaskDate")
	switch t.Type {
	case TaskDateTypeScheduled:
		p.Field("type", "scheduled")
	case TaskDateTypeDeadline:
		p.Field("type", "deadline")
	}

	if t.HasTime {
		p.Field("date", t.Date.Format("2006-01-02 15:04"))
	} else {
		p.Field("date", t.Date.Format("2006-01-02"))
	}

	if t.Repeater != nil {
		p.Field("repeater", t.Repeater.String())
	}

	debugPreviousLineAware(p, t)
	p.EndType()
}

func (t *TaskDate) isBlock() {}

var _ BlockNode = (*TaskDate)(nil)

func truncateToDay(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

func truncateToMinute(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), 0, 0, date.Location())
}

// Scheduled gets the `SCHEDULED` date of this block, or nil if it does not
// have one.
func (b *Block) Scheduled() *TaskDate {
	return b.taskDate(TaskDateTypeScheduled)
}

// Deadline gets the `DEADLINE` date of this block, or nil if it does not have
// one.
func (b *Block) Deadline() *TaskDate {
	return b.taskDate(TaskDateTypeDeadline)
}

// SetScheduled sets the `SCHEDULED` date of this block, replacing the one that
// is already there. Passing nil removes the date.
func (b *Block) SetScheduled(date *TaskDate) {
	b.setTaskDate(TaskDateTypeScheduled, date)
}

// SetDeadline sets the `DEADLINE` date of this block, replacing the one that
// is already there. Passing nil removes the date.
func (b *Block) SetDeadline(date *TaskDate) {
	b.setTaskDate(TaskDateTypeDeadline, date)
}

func (b *Block) taskDate(dateType TaskDateType) *TaskDate {
	for node := b.FirstChild(); node != nil; node = node.NextSibling() {
		if date, ok := node.(*TaskDate); ok && date.Type == dateType {
			return date
		}
	}

	return nil
}

func (b *Block) setTaskDate(dateType TaskDateType, date *TaskDate) {
	existing := b.taskDate(dateType)

	if date == nil {
		if existing != nil {
			existing.RemoveSelf()
		}
		return
	}

	date.Type = dateType

	if existing != nil {
		b.ReplaceChild(existing, date)
		return
	}

	// Logseq keeps these dates directly after the content of the task, so
	// place the date after the properties, the first paragraph and any date
	// that is already there.
	var after Node
	seenParagraph := false
	for node := b.FirstChild(); node != nil; node = node.NextSibling() {
		if _, ok := node.(*Paragraph); ok {
			if seenParagraph {
				break
			}
			seenParagraph = true
		} else if _, ok := node.(*Properties); !ok {
			if _, ok := node.(*TaskDate); !ok {
				break
			}
		}

		after = node
	}

	if after == nil {
		b.PrependChild(date)
	} else {
		b.InsertChildAfter(date, after)
	}
}

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

// LogbookEntryClock is a time tracking entry, written by Logseq when the clock
// on a task is started and stopped:
//
//	CLOCK: [2023-06-26 Mon 17:25:56]--[2023-06-26 Mon 17:25:58] =>  00:00:02
//
// The duration at the end is derived from Start and End when the entry is
// written, so it can never drift from the times it sums up.
//
// The times are kept with second precision no matter how they were written. A
// graph where `:logbook/settings :with-second-support?` is off writes them,
// and the duration, without the seconds.
type LogbookEntryClock struct {
	baseNode

	// Start is when the clock was started.
	Start time.Time

	// End is when the clock was stopped. The zero time means the clock is
	// still running, which Logseq writes without the end and the duration.
	End time.Time
}

// NewLogbookEntryClock creates a clock entry that ran between two times. Pass
// the zero time as the end for a clock that is still running.
func NewLogbookEntryClock(start time.Time, end time.Time) *LogbookEntryClock {
	return &LogbookEntryClock{
		Start: truncateToSecond(start),
		End:   truncateToSecond(end),
	}
}

// WithStart sets when the clock was started.
func (t *LogbookEntryClock) WithStart(start time.Time) *LogbookEntryClock {
	t.Start = truncateToSecond(start)
	return t
}

// WithEnd sets when the clock was stopped. Passing the zero time makes the
// clock running again.
func (t *LogbookEntryClock) WithEnd(end time.Time) *LogbookEntryClock {
	t.End = truncateToSecond(end)
	return t
}

// IsRunning returns true if the clock has not been stopped.
func (t *LogbookEntryClock) IsRunning() bool {
	return t.End.IsZero()
}

// Duration is how long the clock ran, which is zero while it is running.
func (t *LogbookEntryClock) Duration() time.Duration {
	if t.IsRunning() {
		return 0
	}

	return t.End.Sub(t.Start)
}

func (t *LogbookEntryClock) debug(p *debugPrinter) {
	p.StartType("LogbookEntryClock")
	p.Field("start", t.Start.Format("2006-01-02 15:04:05"))
	if !t.IsRunning() {
		p.Field("end", t.End.Format("2006-01-02 15:04:05"))
	}
	p.EndType()
}

func (t *LogbookEntryClock) isLogbookEntry() {}

// LogbookEntryStateChange records that a task changed status, which Logseq
// writes when a repeating task is completed. The entry is written as
// `* State "DONE" from "TODO" [2023-06-26 Mon 17:25:56]`.
type LogbookEntryStateChange struct {
	baseNode

	// From is the status the task had. TaskStatusNone leaves the `from` part
	// out of the entry.
	From TaskStatus

	// To is the status the task changed to.
	To TaskStatus

	// Time is when the change happened.
	Time time.Time
}

func NewLogbookEntryStateChange(from TaskStatus, to TaskStatus, at time.Time) *LogbookEntryStateChange {
	return &LogbookEntryStateChange{
		From: from,
		To:   to,
		Time: truncateToSecond(at),
	}
}

// WithFrom sets the status the task had.
func (t *LogbookEntryStateChange) WithFrom(from TaskStatus) *LogbookEntryStateChange {
	t.From = from
	return t
}

// WithTo sets the status the task changed to.
func (t *LogbookEntryStateChange) WithTo(to TaskStatus) *LogbookEntryStateChange {
	t.To = to
	return t
}

// WithTime sets when the change happened.
func (t *LogbookEntryStateChange) WithTime(at time.Time) *LogbookEntryStateChange {
	t.Time = truncateToSecond(at)
	return t
}

func (t *LogbookEntryStateChange) debug(p *debugPrinter) {
	p.StartType("LogbookEntryStateChange")
	if t.From != TaskStatusNone {
		p.Field("from", t.From.String())
	}
	p.Field("to", t.To.String())
	p.Field("time", t.Time.Format("2006-01-02 15:04:05"))
	p.EndType()
}

func (t *LogbookEntryStateChange) isLogbookEntry() {}

func truncateToSecond(at time.Time) time.Time {
	if at.IsZero() {
		return at
	}

	return time.Date(at.Year(), at.Month(), at.Day(), at.Hour(), at.Minute(), at.Second(), 0, at.Location())
}

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
		if clock, ok := child.(*LogbookEntryClock); ok && clock.IsRunning() {
			clock.WithEnd(now)
			return true, nil
		}

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
