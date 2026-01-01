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

// DaysAgo returns the time n days before now.
// This is useful for filtering tasks by creation date.
//
// Example:
//
//	db.Tasks().CreatedAfter(things3.DaysAgo(7)).All(ctx) // tasks from last 7 days
func DaysAgo(n int) time.Time {
	return time.Now().AddDate(0, 0, -n)
}

// WeeksAgo returns the time n weeks before now.
// This is useful for filtering tasks by creation date.
//
// Example:
//
//	db.Tasks().CreatedAfter(things3.WeeksAgo(2)).All(ctx) // tasks from last 2 weeks
func WeeksAgo(n int) time.Time {
	return time.Now().AddDate(0, 0, -n*7)
}

// MonthsAgo returns the time n months before now.
// This is useful for filtering tasks by creation date.
//
// Example:
//
//	db.Tasks().CreatedAfter(things3.MonthsAgo(1)).All(ctx) // tasks from last month
func MonthsAgo(n int) time.Time {
	return time.Now().AddDate(0, -n, 0)
}

// YearsAgo returns the time n years before now.
// This is useful for filtering tasks by creation date.
//
// Example:
//
//	db.Tasks().CreatedAfter(things3.YearsAgo(1)).All(ctx) // tasks from last year
func YearsAgo(n int) time.Time {
	return time.Now().AddDate(-n, 0, 0)
}
