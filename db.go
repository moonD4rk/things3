package things3

import (
	"context"

	idb "github.com/moond4rk/things3/internal/db"
)

// db provides read-only access to the Things 3 database.
type db struct {
	inner *idb.DB
}

// newDB creates a new Things 3 database connection.
func newDB(opts ...idb.Option) (*db, error) {
	inner, err := idb.Open(opts...)
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

// Token returns the Things URL scheme authentication token.
func (d *db) Token(ctx context.Context) (string, error) {
	return d.inner.AuthToken(ctx)
}

// =============================================================================
// Row-to-Model Conversion
// =============================================================================

// convertTaskRow converts an internal TaskRow to a public Task.
func convertTaskRow(r *idb.TaskRow) Task {
	task := Task{
		UUID:         r.UUID,
		Trashed:      r.Trashed,
		Title:        r.Title,
		Notes:        r.Notes,
		Start:        r.Start,
		AreaUUID:     r.AreaUUID,
		AreaTitle:    r.AreaTitle,
		ProjectUUID:  r.ProjectUUID,
		ProjectTitle: r.ProjectTitle,
		HeadingUUID:  r.HeadingUUID,
		HeadingTitle: r.HeadingTitle,
		StartDate:    r.StartDate,
		Deadline:     r.Deadline,
		ReminderTime: r.ReminderTime,
		StopDate:     r.StopDate,
		Created:      r.Created,
		Modified:     r.Modified,
		Index:        r.Index,
		TodayIndex:   r.TodayIndex,
	}

	// Convert type string to TaskType
	switch r.Type {
	case taskTypeStringTodo:
		task.Type = TaskTypeTodo
	case taskTypeStringProject:
		task.Type = TaskTypeProject
	case taskTypeStringHeading:
		task.Type = TaskTypeHeading
	}

	// Convert status string to Status
	switch r.Status {
	case statusStringIncomplete:
		task.Status = StatusIncomplete
	case statusStringCompleted:
		task.Status = StatusCompleted
	case statusStringCanceled:
		task.Status = StatusCanceled
	}

	// Mark if task has tags or checklist (actual items loaded separately)
	if r.HasTags {
		task.Tags = []string{}
	}
	if r.HasChecklist {
		task.Checklist = []ChecklistItem{}
	}

	return task
}

// convertAreaRow converts an internal AreaRow to a public Area.
func convertAreaRow(r idb.AreaRow) Area {
	area := Area{
		UUID:  r.UUID,
		Type:  "area",
		Title: r.Title,
	}
	if r.HasTags {
		area.Tags = []string{}
	}
	return area
}

// convertTagRow converts an internal TagRow to a public Tag.
func convertTagRow(r idb.TagRow) Tag {
	return Tag{
		UUID:     r.UUID,
		Type:     "tag",
		Title:    r.Title,
		Shortcut: r.Shortcut,
	}
}

// convertChecklistItemRows converts internal ChecklistItemRows to public ChecklistItems.
func convertChecklistItemRows(rows []idb.ChecklistItemRow) []ChecklistItem {
	items := make([]ChecklistItem, len(rows))
	for i, r := range rows {
		items[i] = ChecklistItem{
			UUID:     r.UUID,
			Type:     "checklist-item",
			Title:    r.Title,
			Status:   r.Status,
			StopDate: r.StopDate,
			Created:  r.Created,
			Modified: r.Modified,
		}
	}
	return items
}
