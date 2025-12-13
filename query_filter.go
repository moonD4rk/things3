package things3

import "time"

// =============================================================================
// Date Filter Value
// =============================================================================

// dateFilterValue holds a parsed date filter configuration.
// Only one field should be set at a time.
type dateFilterValue struct {
	hasDate  *bool      // true/false for existence check
	relative string     // "future" or "past"
	operator string     // "=", "<", "<=", ">", ">="
	date     *time.Time // specific date for comparison
}

// dateField identifies which date field a DateFilter operates on.
type dateField int

const (
	dateFieldStartDate dateField = iota
	dateFieldStopDate
	dateFieldDeadline
)

// =============================================================================
// StatusFilter - Type-safe status filtering
// =============================================================================

// StatusFilter provides type-safe status filtering for TaskQuery.
// Use the terminal methods (Incomplete, Completed, Canceled) to set the filter
// and return to the parent TaskQuery for continued chaining.
type StatusFilter struct {
	query *TaskQuery
}

// Incomplete filters for tasks with incomplete status.
func (f *StatusFilter) Incomplete() *TaskQuery {
	s := StatusIncomplete
	f.query.status = &s
	return f.query
}

// Completed filters for tasks with completed status.
func (f *StatusFilter) Completed() *TaskQuery {
	s := StatusCompleted
	f.query.status = &s
	return f.query
}

// Canceled filters for tasks with canceled status.
func (f *StatusFilter) Canceled() *TaskQuery {
	s := StatusCanceled
	f.query.status = &s
	return f.query
}

// =============================================================================
// StartFilter - Type-safe start bucket filtering
// =============================================================================

// StartFilter provides type-safe start bucket filtering for TaskQuery.
// Use the terminal methods (Inbox, Anytime, Someday) to set the filter
// and return to the parent TaskQuery for continued chaining.
type StartFilter struct {
	query *TaskQuery
}

// Inbox filters for tasks in the Inbox.
func (f *StartFilter) Inbox() *TaskQuery {
	s := StartInbox
	f.query.start = &s
	return f.query
}

// Anytime filters for tasks scheduled as Anytime.
func (f *StartFilter) Anytime() *TaskQuery {
	s := StartAnytime
	f.query.start = &s
	return f.query
}

// Someday filters for tasks scheduled as Someday.
func (f *StartFilter) Someday() *TaskQuery {
	s := StartSomeday
	f.query.start = &s
	return f.query
}

// =============================================================================
// DateFilter - Type-safe date filtering
// =============================================================================

// DateFilter provides type-safe date filtering for TaskQuery.
// It is used for startDate, stopDate, and deadline fields.
// Use the terminal methods to set the filter and return to the parent TaskQuery.
type DateFilter struct {
	query *TaskQuery
	field dateField
}

// Exists filters by whether the date exists (is not null).
// Pass true to include only items with this date set.
// Pass false to include only items without this date set.
func (f *DateFilter) Exists(has bool) *TaskQuery {
	f.setFilter(&dateFilterValue{hasDate: &has})
	return f.query
}

// Future filters for dates in the future (after today).
func (f *DateFilter) Future() *TaskQuery {
	f.setFilter(&dateFilterValue{relative: dateFuture})
	return f.query
}

// Past filters for dates in the past (today or earlier).
func (f *DateFilter) Past() *TaskQuery {
	f.setFilter(&dateFilterValue{relative: datePast})
	return f.query
}

// On filters for a specific date (equals).
func (f *DateFilter) On(date time.Time) *TaskQuery {
	f.setFilter(&dateFilterValue{operator: "=", date: &date})
	return f.query
}

// Before filters for dates before the given date (exclusive).
func (f *DateFilter) Before(date time.Time) *TaskQuery {
	f.setFilter(&dateFilterValue{operator: "<", date: &date})
	return f.query
}

// OnOrBefore filters for dates on or before the given date (inclusive).
func (f *DateFilter) OnOrBefore(date time.Time) *TaskQuery {
	f.setFilter(&dateFilterValue{operator: "<=", date: &date})
	return f.query
}

// After filters for dates after the given date (exclusive).
func (f *DateFilter) After(date time.Time) *TaskQuery {
	f.setFilter(&dateFilterValue{operator: ">", date: &date})
	return f.query
}

// OnOrAfter filters for dates on or after the given date (inclusive).
func (f *DateFilter) OnOrAfter(date time.Time) *TaskQuery {
	f.setFilter(&dateFilterValue{operator: ">=", date: &date})
	return f.query
}

// setFilter sets the filter value on the appropriate field in the parent query.
func (f *DateFilter) setFilter(v *dateFilterValue) {
	switch f.field {
	case dateFieldStartDate:
		f.query.startDateFilter = v
	case dateFieldStopDate:
		f.query.stopDateFilter = v
	case dateFieldDeadline:
		f.query.deadlineFilter = v
	}
}

// =============================================================================
// TypeFilter - Type-safe task type filtering
// =============================================================================

// TypeFilter provides type-safe task type filtering for TaskQuery.
// Use the terminal methods (Todo, Project, Heading) to set the filter
// and return to the parent TaskQuery for continued chaining.
type TypeFilter struct {
	query *TaskQuery
}

// Todo filters for to-do items only.
func (f *TypeFilter) Todo() *TaskQuery {
	t := TaskTypeTodo
	f.query.taskType = &t
	return f.query
}

// Project filters for projects only.
func (f *TypeFilter) Project() *TaskQuery {
	t := TaskTypeProject
	f.query.taskType = &t
	return f.query
}

// Heading filters for headings only.
func (f *TypeFilter) Heading() *TaskQuery {
	t := TaskTypeHeading
	f.query.taskType = &t
	return f.query
}
