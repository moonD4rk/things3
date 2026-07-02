package things3

import (
	"context"
	"slices"
)

// Today returns the todos in the Things Today view: todos scheduled into Today,
// Someday todos whose scheduled date has arrived, and overdue-deadline todos,
// concatenated in the app's display order. Within the scheduled-today group,
// This Evening todos are placed after the rest, mirroring the app's Evening
// section. The result is never nil.
func (c *Client) Today(ctx context.Context) ([]Todo, error) {
	base := c.database.Todos()

	regular, err := base.
		StartDate().Exists(true).
		Start().Anytime().
		Status().Incomplete().
		OrderByTodayIndex().
		All(ctx)
	if err != nil {
		return nil, err
	}
	// This Evening todos form the bottom section of Today; move them after the
	// rest while preserving todayIndex order within each part.
	slices.SortStableFunc(regular, func(a, b Todo) int {
		switch {
		case a.Evening == b.Evening:
			return 0
		case a.Evening:
			return 1
		default:
			return -1
		}
	})

	scheduled, err := base.
		StartDate().Past().
		Start().Someday().
		Status().Incomplete().
		OrderByTodayIndex().
		All(ctx)
	if err != nil {
		return nil, err
	}

	overdue, err := base.
		deadlineSuppressed(false).
		StartDate().Exists(false).
		Deadline().Past().
		Status().Incomplete().
		All(ctx)
	if err != nil {
		return nil, err
	}

	todos := make([]Todo, 0, len(regular)+len(scheduled)+len(overdue))
	todos = append(todos, regular...)
	todos = append(todos, scheduled...)
	todos = append(todos, overdue...)
	return todos, nil
}
