package things3

import (
	"time"

	idb "github.com/moond4rk/things3/internal/db"
)

// =============================================================================
// Date Filter Value Field
// =============================================================================

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
	v := int(StatusIncomplete)
	f.query.filter.Status = &v
	return f.query
}

// Completed filters for tasks with completed status.
func (f *statusFilter) Completed() TaskQueryBuilder {
	v := int(StatusCompleted)
	f.query.filter.Status = &v
	return f.query
}

// Canceled filters for tasks with canceled status.
func (f *statusFilter) Canceled() TaskQueryBuilder {
	v := int(StatusCanceled)
	f.query.filter.Status = &v
	return f.query
}

// Any clears the status filter to include tasks of any status.
func (f *statusFilter) Any() TaskQueryBuilder {
	f.query.filter.Status = nil
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
	v := int(StartInbox)
	f.query.filter.Start = &v
	return f.query
}

// Anytime filters for tasks scheduled as Anytime.
func (f *startFilter) Anytime() TaskQueryBuilder {
	v := int(StartAnytime)
	f.query.filter.Start = &v
	return f.query
}

// Someday filters for tasks scheduled as Someday.
func (f *startFilter) Someday() TaskQueryBuilder {
	v := int(StartSomeday)
	f.query.filter.Start = &v
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
	f.setFilter(&idb.DateFilterValue{HasDate: &has})
	return f.query
}

// Future filters for dates in the future (after today).
func (f *dateFilter) Future() TaskQueryBuilder {
	f.setFilter(&idb.DateFilterValue{Relative: idb.DateFuture})
	return f.query
}

// Past filters for dates in the past (today or earlier).
func (f *dateFilter) Past() TaskQueryBuilder {
	f.setFilter(&idb.DateFilterValue{Relative: idb.DatePast})
	return f.query
}

// On filters for a specific date (equals).
func (f *dateFilter) On(date time.Time) TaskQueryBuilder {
	f.setFilter(&idb.DateFilterValue{Operator: "=", Date: &date})
	return f.query
}

// Before filters for dates before the given date (exclusive).
func (f *dateFilter) Before(date time.Time) TaskQueryBuilder {
	f.setFilter(&idb.DateFilterValue{Operator: "<", Date: &date})
	return f.query
}

// OnOrBefore filters for dates on or before the given date (inclusive).
func (f *dateFilter) OnOrBefore(date time.Time) TaskQueryBuilder {
	f.setFilter(&idb.DateFilterValue{Operator: "<=", Date: &date})
	return f.query
}

// After filters for dates after the given date (exclusive).
func (f *dateFilter) After(date time.Time) TaskQueryBuilder {
	f.setFilter(&idb.DateFilterValue{Operator: ">", Date: &date})
	return f.query
}

// OnOrAfter filters for dates on or after the given date (inclusive).
func (f *dateFilter) OnOrAfter(date time.Time) TaskQueryBuilder {
	f.setFilter(&idb.DateFilterValue{Operator: ">=", Date: &date})
	return f.query
}

// setFilter sets the filter value on the appropriate field in the parent query.
func (f *dateFilter) setFilter(v *idb.DateFilterValue) {
	switch f.field {
	case dateFieldStartDate:
		f.query.filter.StartDateFilter = v
	case dateFieldStopDate:
		f.query.filter.StopDateFilter = v
	case dateFieldDeadline:
		f.query.filter.DeadlineFilter = v
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
	v := int(TaskTypeTodo)
	f.query.filter.TaskType = &v
	return f.query
}

// Project filters for projects only.
func (f *typeFilter) Project() TaskQueryBuilder {
	v := int(TaskTypeProject)
	f.query.filter.TaskType = &v
	return f.query
}

// Heading filters for headings only.
func (f *typeFilter) Heading() TaskQueryBuilder {
	v := int(TaskTypeHeading)
	f.query.filter.TaskType = &v
	return f.query
}
