package things3

import (
	"context"
	"time"

	"github.com/moond4rk/things3/internal/database"
)

// db provides read-only access to the Things 3 database.
type db struct {
	inner *database.DB
}

// newDB creates a new Things 3 database connection.
func newDB(opts ...database.Option) (*db, error) {
	inner, err := database.Open(opts...)
	if err != nil {
		return nil, err
	}
	return &db{inner: inner}, nil
}

// Close closes the database connection.
func (d *db) Close() error {
	return d.inner.Close()
}

// Filepath returns the path to the Things database file.
func (d *db) Filepath() string {
	return d.inner.Filepath()
}

// Now returns the current time on the database clock.
func (d *db) Now() time.Time {
	return d.inner.Now()
}

// Token returns the Things URL scheme authentication token.
func (d *db) Token(ctx context.Context) (string, error) {
	return d.inner.AuthToken(ctx)
}

// items runs one mixed query and converts its rows, in query order, to Items.
func (d *db) items(ctx context.Context, f *database.TaskFilter) ([]Item, error) {
	return d.itemsOf(ctx, f)
}

// itemsOf concatenates the results of several mixed queries, in argument order.
// The result is never nil.
func (d *db) itemsOf(ctx context.Context, filters ...*database.TaskFilter) ([]Item, error) {
	items := []Item{}
	for _, f := range filters {
		rows, err := d.inner.QueryTasks(ctx, f)
		if err != nil {
			return nil, err
		}
		for i := range rows {
			tags, err := d.tagsOf(ctx, &rows[i])
			if err != nil {
				return nil, err
			}
			if rows[i].Type == int(taskTypeProject) {
				project := convertTaskRowToProject(&rows[i])
				project.Tags = tags
				items = append(items, Item{Project: &project})
				continue
			}
			todo := convertTaskRowToTodo(&rows[i])
			todo.Tags = tags
			items = append(items, Item{Todo: &todo})
		}
	}
	return items, nil
}

// tagsOf loads a row's tag titles; a row without tags costs no query.
func (d *db) tagsOf(ctx context.Context, r *database.TaskRow) ([]string, error) {
	if !r.HasTags {
		return nil, nil
	}
	return d.inner.TagsOfTask(ctx, r.UUID)
}

// =============================================================================
// Row-to-Model Conversion
// =============================================================================

// convertTaskRowToTodo converts an internal TaskRow to a public Todo.
func convertTaskRowToTodo(r *database.TaskRow) Todo {
	todo := Todo{
		UUID:         r.UUID,
		Title:        r.Title,
		Status:       Status(r.Status),
		Notes:        r.Notes,
		Start:        StartBucket(r.Start),
		AreaUUID:     r.AreaUUID,
		AreaTitle:    r.AreaTitle,
		ProjectUUID:  r.ProjectUUID,
		ProjectTitle: r.ProjectTitle,
		HeadingUUID:  r.HeadingUUID,
		HeadingTitle: r.HeadingTitle,
		StartDate:    r.StartDate,
		Deadline:     r.Deadline,
		Reminder:     r.ReminderTime,
		CreatedAt:    r.Created,
		ModifiedAt:   r.Modified,
		Trashed:      r.Trashed,
		Evening:      r.Evening,
		Repeating:    r.Repeating,
		Template:     r.Template,
		Repeat:       convertRepeat(r),
	}
	todo.CompletedAt, todo.CanceledAt = splitStopDate(todo.Status, r.StopDate)
	return todo
}

// convertTaskRowToProject converts an internal TaskRow to a public Project.
func convertTaskRowToProject(r *database.TaskRow) Project {
	project := Project{
		UUID:       r.UUID,
		Title:      r.Title,
		Status:     Status(r.Status),
		Notes:      r.Notes,
		Start:      StartBucket(r.Start),
		AreaUUID:   r.AreaUUID,
		AreaTitle:  r.AreaTitle,
		StartDate:  r.StartDate,
		Deadline:   r.Deadline,
		Reminder:   r.ReminderTime,
		CreatedAt:  r.Created,
		ModifiedAt: r.Modified,
		Trashed:    r.Trashed,
		Repeating:  r.Repeating,
		Template:   r.Template,
		Repeat:     convertRepeat(r),
	}
	project.CompletedAt, project.CanceledAt = splitStopDate(project.Status, r.StopDate)
	return project
}

// convertTaskRowToHeading converts an internal TaskRow to a public Heading.
func convertTaskRowToHeading(r *database.TaskRow) Heading {
	return Heading{
		UUID:         r.UUID,
		Title:        r.Title,
		ProjectUUID:  r.ProjectUUID,
		ProjectTitle: r.ProjectTitle,
	}
}

// convertAreaRow converts an internal AreaRow to a public Area.
func convertAreaRow(r database.AreaRow) Area {
	area := Area{
		UUID:  r.UUID,
		Title: r.Title,
	}
	if r.HasTags {
		area.Tags = []string{}
	}
	return area
}

// convertTagRow converts an internal TagRow to a public Tag.
func convertTagRow(r database.TagRow) Tag {
	return Tag{
		UUID:     r.UUID,
		Title:    r.Title,
		Shortcut: r.Shortcut,
	}
}

// convertChecklistItemRows converts internal ChecklistItemRows to public ChecklistItems.
func convertChecklistItemRows(rows []database.ChecklistItemRow) []ChecklistItem {
	items := make([]ChecklistItem, len(rows))
	for i, r := range rows {
		item := ChecklistItem{
			UUID:       r.UUID,
			Title:      r.Title,
			Status:     Status(r.Status),
			CreatedAt:  r.Created,
			ModifiedAt: r.Modified,
		}
		item.CompletedAt, item.CanceledAt = splitStopDate(item.Status, r.StopDate)
		items[i] = item
	}
	return items
}

// splitStopDate assigns a stop date to the field matching the closed status.
func splitStopDate(status Status, stop *time.Time) (completedAt, canceledAt *time.Time) {
	switch status {
	case StatusCompleted:
		return stop, nil
	case StatusCanceled:
		return nil, stop
	default:
		return nil, nil
	}
}
