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

// withTaskFilter clones the parent builder, applies a mutation to the clone's
// filter, and returns the clone as T. Sub-filters hold this callback instead
// of the parent's mutable state, so every terminal call forks a fresh copy of
// the parent snapshot captured at factory time.
type withTaskFilter[T any] func(mutate func(*database.TaskFilter)) T

// =============================================================================
// statusFilter - Generic type-safe status filtering
// =============================================================================

// statusFilter provides type-safe status filtering for any query builder.
type statusFilter[T any] struct {
	with withTaskFilter[T]
}

// Incomplete filters for items with incomplete status.
func (f *statusFilter[T]) Incomplete() T {
	v := int(StatusIncomplete)
	return f.with(func(tf *database.TaskFilter) { tf.Status = &v })
}

// Completed filters for items with completed status.
func (f *statusFilter[T]) Completed() T {
	v := int(StatusCompleted)
	return f.with(func(tf *database.TaskFilter) { tf.Status = &v })
}

// Canceled filters for items with canceled status.
func (f *statusFilter[T]) Canceled() T {
	v := int(StatusCanceled)
	return f.with(func(tf *database.TaskFilter) { tf.Status = &v })
}

// Any clears the status filter to include items of any status.
func (f *statusFilter[T]) Any() T {
	return f.with(func(tf *database.TaskFilter) { tf.Status = nil })
}

// =============================================================================
// startFilter - Generic type-safe start bucket filtering
// =============================================================================

// startFilter provides type-safe start bucket filtering for any query builder.
type startFilter[T any] struct {
	with withTaskFilter[T]
}

// Inbox filters for items in the Inbox.
func (f *startFilter[T]) Inbox() T {
	v := int(StartInbox)
	return f.with(func(tf *database.TaskFilter) { tf.Start = &v })
}

// Anytime filters for items scheduled as Anytime.
func (f *startFilter[T]) Anytime() T {
	v := int(StartAnytime)
	return f.with(func(tf *database.TaskFilter) { tf.Start = &v })
}

// Someday filters for items scheduled as Someday.
func (f *startFilter[T]) Someday() T {
	v := int(StartSomeday)
	return f.with(func(tf *database.TaskFilter) { tf.Start = &v })
}

// =============================================================================
// dateFilter - Generic type-safe date filtering
// =============================================================================

// dateFilter provides type-safe date filtering for any query builder.
type dateFilter[T any] struct {
	with  withTaskFilter[T]
	field dateField
}

// Exists filters by whether the date exists (is not null).
func (f *dateFilter[T]) Exists(has bool) T {
	return f.set(&database.DateFilterValue{HasDate: &has})
}

// Future filters for dates in the future (after today).
func (f *dateFilter[T]) Future() T {
	return f.set(&database.DateFilterValue{Relative: database.DateFuture})
}

// Past filters for dates in the past (today or earlier).
func (f *dateFilter[T]) Past() T {
	return f.set(&database.DateFilterValue{Relative: database.DatePast})
}

// On filters for a specific date (equals).
func (f *dateFilter[T]) On(date time.Time) T {
	return f.set(&database.DateFilterValue{Operator: "=", Date: &date})
}

// Before filters for dates before the given date (exclusive).
func (f *dateFilter[T]) Before(date time.Time) T {
	return f.set(&database.DateFilterValue{Operator: "<", Date: &date})
}

// OnOrBefore filters for dates on or before the given date (inclusive).
func (f *dateFilter[T]) OnOrBefore(date time.Time) T {
	return f.set(&database.DateFilterValue{Operator: "<=", Date: &date})
}

// After filters for dates after the given date (exclusive).
func (f *dateFilter[T]) After(date time.Time) T {
	return f.set(&database.DateFilterValue{Operator: ">", Date: &date})
}

// OnOrAfter filters for dates on or after the given date (inclusive).
func (f *dateFilter[T]) OnOrAfter(date time.Time) T {
	return f.set(&database.DateFilterValue{Operator: ">=", Date: &date})
}

// set forks the parent with the filter value applied to the appropriate field.
func (f *dateFilter[T]) set(v *database.DateFilterValue) T {
	field := f.field
	return f.with(func(tf *database.TaskFilter) {
		switch field {
		case dateFieldStartDate:
			tf.StartDateFilter = v
		case dateFieldStopDate:
			tf.StopDateFilter = v
		case dateFieldDeadline:
			tf.DeadlineFilter = v
		}
	})
}
