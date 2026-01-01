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

// WhenScheduler is implemented by builders that support scheduling.
// All builder types (TodoBuilder, ProjectBuilder, UpdateTodoBuilder,
// UpdateProjectBuilder, JSONTodoBuilder, JSONProjectBuilder) satisfy this interface.
type WhenScheduler[T any] interface {
	When(t time.Time) T
	WhenEvening() T
	WhenAnytime() T
	WhenSomeday() T
}

// ApplyWhen parses a when string and applies scheduling to a builder.
// Supports:
//   - "today": schedules for today
//   - "tomorrow": schedules for tomorrow
//   - "evening": schedules for this evening
//   - "anytime": removes specific scheduling (anytime)
//   - "someday": schedules for someday (indefinite future)
//   - "yyyy-mm-dd": schedules for specific date
//
// Returns the builder unchanged if the format is not recognized.
//
// Example:
//
//	todo := scheme.Todo().Title("Task")
//	todo = things3.ApplyWhen(todo, "today")
//	todo = things3.ApplyWhen(todo, "2024-12-25")
func ApplyWhen[T WhenScheduler[T]](b T, when string) T {
	switch when {
	case "today":
		return b.When(Today())
	case "tomorrow":
		return b.When(Tomorrow())
	case "evening":
		return b.WhenEvening()
	case "anytime":
		return b.WhenAnytime()
	case "someday":
		return b.WhenSomeday()
	default:
		if t, err := time.Parse(time.DateOnly, when); err == nil {
			return b.When(t)
		}
		return b
	}
}
