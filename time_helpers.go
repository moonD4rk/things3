package things3

import (
	"fmt"
	"time"
)

// Today returns today's date at midnight (00:00:00) in local timezone.
// This is useful for scheduling todos with When().
//
// Example:
//
//	client.AddTodo().Title("Morning task").When(things3.Today())
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Tomorrow returns tomorrow's date at midnight (00:00:00) in local timezone.
// This is useful for scheduling todos with When().
//
// Example:
//
//	client.AddTodo().Title("Todo for tomorrow").When(things3.Tomorrow())
func Tomorrow() time.Time {
	return Today().AddDate(0, 0, 1)
}

// DaysAgo returns the time n days before now.
// This is useful for filtering todos by creation date.
//
// Example:
//
//	client.Todos().CreatedAfter(things3.DaysAgo(7)).All(ctx) // todos from last 7 days
func DaysAgo(n int) time.Time {
	return time.Now().AddDate(0, 0, -n)
}

// WeeksAgo returns the time n weeks before now.
// This is useful for filtering todos by creation date.
//
// Example:
//
//	client.Todos().CreatedAfter(things3.WeeksAgo(2)).All(ctx) // todos from last 2 weeks
func WeeksAgo(n int) time.Time {
	return time.Now().AddDate(0, 0, -n*7)
}

// MonthsAgo returns the time n months before now.
// This is useful for filtering todos by creation date.
//
// Example:
//
//	client.Todos().CreatedAfter(things3.MonthsAgo(1)).All(ctx) // todos from last month
func MonthsAgo(n int) time.Time {
	return time.Now().AddDate(0, -n, 0)
}

// YearsAgo returns the time n years before now.
// This is useful for filtering todos by creation date.
//
// Example:
//
//	client.Todos().CreatedAfter(things3.YearsAgo(1)).All(ctx) // todos from last year
func YearsAgo(n int) time.Time {
	return time.Now().AddDate(-n, 0, 0)
}

// When keyword strings accepted by ParseWhen and ApplyWhen.
const (
	whenKeywordToday    = "today"
	whenKeywordTomorrow = "tomorrow"
	whenKeywordEvening  = "evening"
	whenKeywordAnytime  = "anytime"
	whenKeywordSomeday  = "someday"
)

// WhenScheduler is implemented by builders that support scheduling.
// All builder types (TodoAdder, ProjectAdder, TodoUpdater,
// ProjectUpdater, BatchTodoConfigurator, BatchProjectConfigurator) satisfy this interface.
type WhenScheduler[T any] interface {
	When(t time.Time) T
	WhenEvening() T
	WhenAnytime() T
	WhenSomeday() T
}

// ParseWhen parses a when string and applies scheduling to a builder.
// Supports:
//   - "today": schedules for today
//   - "tomorrow": schedules for tomorrow
//   - "evening": schedules for this evening
//   - "anytime": removes specific scheduling (anytime)
//   - "someday": schedules for someday (indefinite future)
//   - "yyyy-mm-dd": schedules for specific date
//
// Unrecognized input returns the builder unchanged along with a descriptive
// error. Use ApplyWhen to silently ignore invalid input instead.
//
// Example:
//
//	todo := client.AddTodo().Title("Buy milk")
//	todo, err := things3.ParseWhen(todo, "2024-12-25")
func ParseWhen[T WhenScheduler[T]](b T, when string) (T, error) {
	switch when {
	case whenKeywordToday:
		return b.When(Today()), nil
	case whenKeywordTomorrow:
		return b.When(Tomorrow()), nil
	case whenKeywordEvening:
		return b.WhenEvening(), nil
	case whenKeywordAnytime:
		return b.WhenAnytime(), nil
	case whenKeywordSomeday:
		return b.WhenSomeday(), nil
	default:
		if t, err := time.Parse(time.DateOnly, when); err == nil {
			return b.When(t), nil
		}
		return b, fmt.Errorf(
			"things3: unrecognized when value %q (expected today, tomorrow, evening, anytime, someday, or yyyy-mm-dd)",
			when,
		)
	}
}

// ApplyWhen is the error-ignoring variant of ParseWhen: it applies the parsed
// scheduling to the builder and silently returns the builder unchanged when
// the input is not recognized. Use ParseWhen to detect invalid input.
//
// Example:
//
//	todo := client.AddTodo().Title("Buy milk")
//	todo = things3.ApplyWhen(todo, "today")
//	todo = things3.ApplyWhen(todo, "2024-12-25")
func ApplyWhen[T WhenScheduler[T]](b T, when string) T {
	result, _ := ParseWhen(b, when)
	return result
}
