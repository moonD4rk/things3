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
// statusFilter - Type-safe status filtering
// =============================================================================

// statusFilter provides type-safe status filtering for taskQuery.
type statusFilter struct {
	query *taskQuery
}

// Incomplete filters for tasks with incomplete status.
func (f *statusFilter) Incomplete() TaskQueryBuilder {
	s := StatusIncomplete
	f.query.status = &s
	return f.query
}

// Completed filters for tasks with completed status.
func (f *statusFilter) Completed() TaskQueryBuilder {
	s := StatusCompleted
	f.query.status = &s
	return f.query
}

// Canceled filters for tasks with canceled status.
func (f *statusFilter) Canceled() TaskQueryBuilder {
	s := StatusCanceled
	f.query.status = &s
	return f.query
}

// Any clears the status filter to include tasks of any status.
func (f *statusFilter) Any() TaskQueryBuilder {
	f.query.status = nil
	return f.query
}

// =============================================================================
// startFilter - Type-safe start bucket filtering
// =============================================================================

// startFilter provides type-safe start bucket filtering for taskQuery.
type startFilter struct {
	query *taskQuery
}

// Inbox filters for tasks in the Inbox.
func (f *startFilter) Inbox() TaskQueryBuilder {
	s := StartInbox
	f.query.start = &s
	return f.query
}

// Anytime filters for tasks scheduled as Anytime.
func (f *startFilter) Anytime() TaskQueryBuilder {
	s := StartAnytime
	f.query.start = &s
	return f.query
}

// Someday filters for tasks scheduled as Someday.
func (f *startFilter) Someday() TaskQueryBuilder {
	s := StartSomeday
	f.query.start = &s
	return f.query
}

// =============================================================================
// dateFilter - Type-safe date filtering
// =============================================================================

// dateFilter provides type-safe date filtering for taskQuery.
type dateFilter struct {
	query *taskQuery
	field dateField
}

// Exists filters by whether the date exists (is not null).
func (f *dateFilter) Exists(has bool) TaskQueryBuilder {
	f.setFilter(&dateFilterValue{hasDate: &has})
	return f.query
}

// Future filters for dates in the future (after today).
func (f *dateFilter) Future() TaskQueryBuilder {
	f.setFilter(&dateFilterValue{relative: dateFuture})
	return f.query
}

// Past filters for dates in the past (today or earlier).
func (f *dateFilter) Past() TaskQueryBuilder {
	f.setFilter(&dateFilterValue{relative: datePast})
	return f.query
}

// On filters for a specific date (equals).
func (f *dateFilter) On(date time.Time) TaskQueryBuilder {
	f.setFilter(&dateFilterValue{operator: "=", date: &date})
	return f.query
}

// Before filters for dates before the given date (exclusive).
func (f *dateFilter) Before(date time.Time) TaskQueryBuilder {
	f.setFilter(&dateFilterValue{operator: "<", date: &date})
	return f.query
}

// OnOrBefore filters for dates on or before the given date (inclusive).
func (f *dateFilter) OnOrBefore(date time.Time) TaskQueryBuilder {
	f.setFilter(&dateFilterValue{operator: "<=", date: &date})
	return f.query
}

// After filters for dates after the given date (exclusive).
func (f *dateFilter) After(date time.Time) TaskQueryBuilder {
	f.setFilter(&dateFilterValue{operator: ">", date: &date})
	return f.query
}

// OnOrAfter filters for dates on or after the given date (inclusive).
func (f *dateFilter) OnOrAfter(date time.Time) TaskQueryBuilder {
	f.setFilter(&dateFilterValue{operator: ">=", date: &date})
	return f.query
}

// setFilter sets the filter value on the appropriate field in the parent query.
func (f *dateFilter) setFilter(v *dateFilterValue) {
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
// typeFilter - Type-safe task type filtering
// =============================================================================

// typeFilter provides type-safe task type filtering for taskQuery.
type typeFilter struct {
	query *taskQuery
}

// Todo filters for to-do items only.
func (f *typeFilter) Todo() TaskQueryBuilder {
	t := TaskTypeTodo
	f.query.taskType = &t
	return f.query
}

// Project filters for projects only.
func (f *typeFilter) Project() TaskQueryBuilder {
	t := TaskTypeProject
	f.query.taskType = &t
	return f.query
}

// Heading filters for headings only.
func (f *typeFilter) Heading() TaskQueryBuilder {
	t := TaskTypeHeading
	f.query.taskType = &t
	return f.query
}
