package things3

import "time"

// Today returns today's date at midnight (00:00:00) in local timezone.
// This is useful for scheduling tasks with When().
//
// Example:
//
//	scheme.Todo().Title("Morning task").When(things3.Today())
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Tomorrow returns tomorrow's date at midnight (00:00:00) in local timezone.
// This is useful for scheduling tasks with When().
//
// Example:
//
//	scheme.Todo().Title("Task for tomorrow").When(things3.Tomorrow())
func Tomorrow() time.Time {
	return Today().AddDate(0, 0, 1)
}
