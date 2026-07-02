package things3

import (
	"context"

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

// Token returns the Things URL scheme authentication token.
func (d *db) Token(ctx context.Context) (string, error) {
	return d.inner.AuthToken(ctx)
}

// =============================================================================
// Row-to-Model Conversion
// =============================================================================

// convertTaskRowToTodo converts an internal TaskRow to a public Todo.
func convertTaskRowToTodo(r *database.TaskRow) Todo {
	todo := Todo{
		UUID:       r.UUID,
		Title:      r.Title,
		Notes:      r.Notes,
		StartDate:  r.StartDate,
		Deadline:   r.Deadline,
		Reminder:   r.ReminderTime,
		CreatedAt:  r.Created,
		ModifiedAt: r.Modified,
		Trashed:    r.Trashed,
		Evening:    r.Evening,
		Repeating:  r.Repeating,
	}

	// Convert status string to Status enum
	todo.Status = parseStatusFromString(r.Status)

	// Convert start string to StartBucket enum
	todo.Start = parseStartBucketFromString(r.Start)

	// Convert nullable relationship fields
	todo.AreaUUID = ptrToString(r.AreaUUID)
	todo.AreaTitle = ptrToString(r.AreaTitle)
	todo.ProjectUUID = ptrToString(r.ProjectUUID)
	todo.ProjectTitle = ptrToString(r.ProjectTitle)
	todo.HeadingUUID = ptrToString(r.HeadingUUID)
	todo.HeadingTitle = ptrToString(r.HeadingTitle)

	// Split stopDate into CompletedAt or CanceledAt based on Status
	if r.StopDate != nil {
		switch todo.Status {
		case StatusCompleted:
			todo.CompletedAt = r.StopDate
		case StatusCanceled:
			todo.CanceledAt = r.StopDate
		}
	}

	// Mark if task has tags (actual tag names loaded separately)
	if r.HasTags {
		todo.Tags = []string{}
	}

	return todo
}

// convertTaskRowToProject converts an internal TaskRow to a public Project.
func convertTaskRowToProject(r *database.TaskRow) Project {
	project := Project{
		UUID:       r.UUID,
		Title:      r.Title,
		Notes:      r.Notes,
		StartDate:  r.StartDate,
		Deadline:   r.Deadline,
		Reminder:   r.ReminderTime,
		CreatedAt:  r.Created,
		ModifiedAt: r.Modified,
		Trashed:    r.Trashed,
		Repeating:  r.Repeating,
	}

	project.Status = parseStatusFromString(r.Status)
	project.Start = parseStartBucketFromString(r.Start)

	project.AreaUUID = ptrToString(r.AreaUUID)
	project.AreaTitle = ptrToString(r.AreaTitle)

	if r.StopDate != nil {
		switch project.Status {
		case StatusCompleted:
			project.CompletedAt = r.StopDate
		case StatusCanceled:
			project.CanceledAt = r.StopDate
		}
	}

	if r.HasTags {
		project.Tags = []string{}
	}

	return project
}

// convertTaskRowToHeading converts an internal TaskRow to a public Heading.
func convertTaskRowToHeading(r *database.TaskRow) Heading {
	return Heading{
		UUID:         r.UUID,
		Title:        r.Title,
		ProjectUUID:  ptrToString(r.ProjectUUID),
		ProjectTitle: ptrToString(r.ProjectTitle),
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
			Status:     parseStatusFromString(r.Status),
			CreatedAt:  r.Created,
			ModifiedAt: r.Modified,
		}

		if r.StopDate != nil {
			switch item.Status {
			case StatusCompleted:
				item.CompletedAt = r.StopDate
			case StatusCanceled:
				item.CanceledAt = r.StopDate
			}
		}

		items[i] = item
	}
	return items
}

// =============================================================================
// Conversion Helpers
// =============================================================================

// ptrToString converts a nullable string pointer to a string value.
// Returns empty string if the pointer is nil.
func ptrToString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// parseStatusFromString converts a database status string to a Status enum.
func parseStatusFromString(s string) Status {
	switch s {
	case statusStringCompleted:
		return StatusCompleted
	case statusStringCanceled:
		return StatusCanceled
	default:
		return StatusIncomplete
	}
}

// parseStartBucketFromString converts a database start string to a StartBucket enum.
func parseStartBucketFromString(s string) StartBucket {
	switch s {
	case "Anytime":
		return StartAnytime
	case "Someday":
		return StartSomeday
	default:
		return StartInbox
	}
}
