package things3

import (
	"context"
	"slices"
	"time"
)

// Upcoming returns the todos in the Things Upcoming view: todos scheduled for a
// future date, plus each repeating task shown at its next occurrence, merged and
// sorted ascending by start date (todos without a start date sort last).
//
// The database stores only the NEXT occurrence of each repeating task, so a
// repeating task appears exactly once here, at that next occurrence; expanding a
// recurrence rule into its full future series is out of scope because the rule
// is persisted as an opaque binary blob.
//
// The result is never nil.
func (c *Client) Upcoming(ctx context.Context) ([]Todo, error) {
	base := c.database.Todos()

	scheduled, err := base.
		StartDate().Future().
		Start().Someday().
		Status().Incomplete().
		All(ctx)
	if err != nil {
		return nil, err
	}

	repeating, err := base.
		repeatingTemplates().
		StartDate().Future().
		Status().Incomplete().
		All(ctx)
	if err != nil {
		return nil, err
	}

	todos := make([]Todo, 0, len(scheduled)+len(repeating))
	todos = append(todos, scheduled...)
	todos = append(todos, repeating...)
	slices.SortStableFunc(todos, func(a, b Todo) int {
		return compareStartDateAsc(a.StartDate, b.StartDate)
	})
	return todos, nil
}

// compareStartDateAsc orders two start dates ascending, ranking a nil date last.
func compareStartDateAsc(a, b *time.Time) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	default:
		return a.Compare(*b)
	}
}
