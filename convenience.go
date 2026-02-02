package things3

import (
	"context"
	"errors"
	"sort"
	"time"
)

// Todos returns all incomplete to-do items.
func (d *db) Todos(ctx context.Context) ([]Task, error) {
	return d.Tasks().
		Type().Todo().
		Status().Incomplete().
		ContextTrashed(false).
		All(ctx)
}

// Projects returns all incomplete projects.
func (d *db) Projects(ctx context.Context) ([]Task, error) {
	return d.Tasks().
		Type().Project().
		Status().Incomplete().
		ContextTrashed(false).
		All(ctx)
}

// Inbox returns all tasks in the Inbox.
func (d *db) Inbox(ctx context.Context) ([]Task, error) {
	return d.Tasks().
		Start().Inbox().
		Status().Incomplete().
		ContextTrashed(false).
		All(ctx)
}

// Today returns tasks that would appear in Today view.
// This includes:
// - Tasks with a start date set to today or earlier and in Anytime
// - Scheduled tasks from Someday with past start dates (yellow dot tasks)
// - Overdue tasks with deadlines that haven't been suppressed
func (d *db) Today(ctx context.Context) ([]Task, error) {
	// Regular Today tasks
	regularTasks, err := d.Tasks().
		StartDate().Exists(true).
		Start().Anytime().
		Status().Incomplete().
		ContextTrashed(false).
		OrderByTodayIndex().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Unconfirmed scheduled tasks (yellow dot)
	scheduledTasks, err := d.Tasks().
		StartDate().Past().
		Start().Someday().
		Status().Incomplete().
		ContextTrashed(false).
		OrderByTodayIndex().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Unconfirmed overdue tasks
	overdueTasks, err := d.Tasks().
		StartDate().Exists(false).
		Deadline().Past().
		WithDeadlineSuppressed(false).
		Status().Incomplete().
		ContextTrashed(false).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Combine all tasks
	result := make([]Task, 0, len(regularTasks)+len(scheduledTasks)+len(overdueTasks))
	result = append(result, regularTasks...)
	result = append(result, scheduledTasks...)
	result = append(result, overdueTasks...)

	// Sort by today_index and start_date
	sort.Slice(result, func(i, j int) bool {
		if result[i].TodayIndex != result[j].TodayIndex {
			return result[i].TodayIndex < result[j].TodayIndex
		}
		return comparePtrTime(result[i].StartDate, result[j].StartDate)
	})

	return result, nil
}

// Upcoming returns tasks scheduled for future dates.
func (d *db) Upcoming(ctx context.Context) ([]Task, error) {
	return d.Tasks().
		StartDate().Future().
		Start().Someday().
		Status().Incomplete().
		ContextTrashed(false).
		All(ctx)
}

// Anytime returns tasks in the Anytime list.
func (d *db) Anytime(ctx context.Context) ([]Task, error) {
	return d.Tasks().
		Start().Anytime().
		Status().Incomplete().
		ContextTrashed(false).
		All(ctx)
}

// Someday returns tasks in the Someday list (without a start date).
func (d *db) Someday(ctx context.Context) ([]Task, error) {
	return d.Tasks().
		StartDate().Exists(false).
		Start().Someday().
		Status().Incomplete().
		ContextTrashed(false).
		All(ctx)
}

// Logbook returns completed and canceled tasks, sorted by stop date.
func (d *db) Logbook(ctx context.Context) ([]Task, error) {
	completed, err := d.Completed(ctx)
	if err != nil {
		return nil, err
	}

	canceled, err := d.Canceled(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Task, 0, len(completed)+len(canceled))
	result = append(result, completed...)
	result = append(result, canceled...)

	// Sort by stop_date (newest first)
	sort.Slice(result, func(i, j int) bool {
		return comparePtrTimeDesc(result[i].StopDate, result[j].StopDate)
	})

	return result, nil
}

// Trash returns trashed tasks.
func (d *db) Trash(ctx context.Context) ([]Task, error) {
	return d.Tasks().
		Trashed(true).
		Status().Any().
		All(ctx)
}

// Completed returns completed tasks.
func (d *db) Completed(ctx context.Context) ([]Task, error) {
	return d.Tasks().
		Status().Completed().
		ContextTrashed(false).
		All(ctx)
}

// Canceled returns canceled tasks.
func (d *db) Canceled(ctx context.Context) ([]Task, error) {
	return d.Tasks().
		Status().Canceled().
		ContextTrashed(false).
		All(ctx)
}

// Deadlines returns tasks with deadlines, sorted by deadline.
func (d *db) Deadlines(ctx context.Context) ([]Task, error) {
	tasks, err := d.Tasks().
		Deadline().Exists(true).
		Status().Incomplete().
		ContextTrashed(false).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Sort by deadline
	sort.Slice(tasks, func(i, j int) bool {
		return comparePtrTime(tasks[i].Deadline, tasks[j].Deadline)
	})

	return tasks, nil
}

// CreatedWithin returns tasks created after the specified time.
// Example: db.CreatedWithin(ctx, things3.DaysAgo(7))
func (d *db) CreatedWithin(ctx context.Context, since time.Time) ([]Task, error) {
	if since.IsZero() {
		return nil, ErrInvalidParameter
	}

	tasks, err := d.Tasks().
		CreatedAfter(since).
		Status().Incomplete().
		ContextTrashed(false).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Sort by created date (newest first)
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Created.After(tasks[j].Created)
	})

	return tasks, nil
}

// Search searches for tasks matching the query.
// Searches in task title, notes, and area title.
func (d *db) Search(ctx context.Context, query string) ([]Task, error) {
	return d.Tasks().
		Search(query).
		Status().Incomplete().
		ContextTrashed(false).
		All(ctx)
}

// Get retrieves an object by UUID.
// Returns a Task, Area, or Tag depending on what is found.
// Returns nil if not found.
func (d *db) Get(ctx context.Context, uuid string) (any, error) {
	// Try to find as task
	task, err := d.Tasks().WithUUID(uuid).First(ctx)
	if err == nil {
		return task, nil
	}
	if !errors.Is(err, ErrTaskNotFound) {
		return nil, err
	}

	// Try to find as area
	area, err := d.Areas().WithUUID(uuid).First(ctx)
	if err == nil {
		return area, nil
	}
	if !errors.Is(err, ErrAreaNotFound) {
		return nil, err
	}

	// Try to find as tag
	tags, err := d.Tags().All(ctx)
	if err != nil {
		return nil, err
	}
	for _, tag := range tags {
		if tag.UUID == uuid {
			return &tag, nil
		}
	}

	return nil, nil
}

// ChecklistItems returns the checklist items for a to-do.
func (d *db) ChecklistItems(ctx context.Context, todoUUID string) ([]ChecklistItem, error) {
	return d.getChecklistItems(ctx, todoUUID)
}
