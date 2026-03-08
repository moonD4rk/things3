package things3

import (
	"time"

	"github.com/moond4rk/things3/internal/database"
)

// =============================================================================
// Date Filter Value Field
// =============================================================================

// dateField identifies which date field a dateFilter operates on.
type dateField int

const (
	dateFieldStartDate dateField = iota
	dateFieldStopDate
	dateFieldDeadline
)

// =============================================================================
// statusFilter - Generic type-safe status filtering
// =============================================================================

// statusFilter provides type-safe status filtering for any query builder.
type statusFilter[T any] struct {
	query  *taskQuery
	parent T
}

// Incomplete filters for items with incomplete status.
func (f *statusFilter[T]) Incomplete() T {
	v := int(StatusIncomplete)
	f.query.filter.Status = &v
	return f.parent
}

// Completed filters for items with completed status.
func (f *statusFilter[T]) Completed() T {
	v := int(StatusCompleted)
	f.query.filter.Status = &v
	return f.parent
}

// Canceled filters for items with canceled status.
func (f *statusFilter[T]) Canceled() T {
	v := int(StatusCanceled)
	f.query.filter.Status = &v
	return f.parent
}

// Any clears the status filter to include items of any status.
func (f *statusFilter[T]) Any() T {
	f.query.filter.Status = nil
	return f.parent
}

// =============================================================================
// startFilter - Generic type-safe start bucket filtering
// =============================================================================

// startFilter provides type-safe start bucket filtering for any query builder.
type startFilter[T any] struct {
	query  *taskQuery
	parent T
}

// Inbox filters for items in the Inbox.
func (f *startFilter[T]) Inbox() T {
	v := int(StartInbox)
	f.query.filter.Start = &v
	return f.parent
}

// Anytime filters for items scheduled as Anytime.
func (f *startFilter[T]) Anytime() T {
	v := int(StartAnytime)
	f.query.filter.Start = &v
	return f.parent
}

// Someday filters for items scheduled as Someday.
func (f *startFilter[T]) Someday() T {
	v := int(StartSomeday)
	f.query.filter.Start = &v
	return f.parent
}

// =============================================================================
// dateFilter - Generic type-safe date filtering
// =============================================================================

// dateFilter provides type-safe date filtering for any query builder.
type dateFilter[T any] struct {
	query  *taskQuery
	parent T
	field  dateField
}

// Exists filters by whether the date exists (is not null).
func (f *dateFilter[T]) Exists(has bool) T {
	f.setFilter(&database.DateFilterValue{HasDate: &has})
	return f.parent
}

// Future filters for dates in the future (after today).
func (f *dateFilter[T]) Future() T {
	f.setFilter(&database.DateFilterValue{Relative: database.DateFuture})
	return f.parent
}

// Past filters for dates in the past (today or earlier).
func (f *dateFilter[T]) Past() T {
	f.setFilter(&database.DateFilterValue{Relative: database.DatePast})
	return f.parent
}

// On filters for a specific date (equals).
func (f *dateFilter[T]) On(date time.Time) T {
	f.setFilter(&database.DateFilterValue{Operator: "=", Date: &date})
	return f.parent
}

// Before filters for dates before the given date (exclusive).
func (f *dateFilter[T]) Before(date time.Time) T {
	f.setFilter(&database.DateFilterValue{Operator: "<", Date: &date})
	return f.parent
}

// OnOrBefore filters for dates on or before the given date (inclusive).
func (f *dateFilter[T]) OnOrBefore(date time.Time) T {
	f.setFilter(&database.DateFilterValue{Operator: "<=", Date: &date})
	return f.parent
}

// After filters for dates after the given date (exclusive).
func (f *dateFilter[T]) After(date time.Time) T {
	f.setFilter(&database.DateFilterValue{Operator: ">", Date: &date})
	return f.parent
}

// OnOrAfter filters for dates on or after the given date (inclusive).
func (f *dateFilter[T]) OnOrAfter(date time.Time) T {
	f.setFilter(&database.DateFilterValue{Operator: ">=", Date: &date})
	return f.parent
}

// setFilter sets the filter value on the appropriate field in the parent query.
func (f *dateFilter[T]) setFilter(v *database.DateFilterValue) {
	switch f.field {
	case dateFieldStartDate:
		f.query.filter.StartDateFilter = v
	case dateFieldStopDate:
		f.query.filter.StopDateFilter = v
	case dateFieldDeadline:
		f.query.filter.DeadlineFilter = v
	}
}
