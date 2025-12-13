package things3

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DB provides read-only access to the Things 3 database.
type DB struct {
	db         *sql.DB
	filepath   string
	printSQL   bool
	queryCount int
}

// NewDB creates a new Things 3 database connection.
// Options can be provided to configure the database behavior.
func NewDB(opts ...DBOption) (*DB, error) {
	options := &dbOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Discover database path
	filepath, err := discoverDatabasePath(options.databasePath)
	if err != nil {
		return nil, err
	}

	// Open database connection
	db, err := openDatabase(filepath)
	if err != nil {
		return nil, err
	}

	// Validate database version
	if err := validateDatabaseVersion(db); err != nil {
		db.Close()
		return nil, err
	}

	return &DB{
		db:       db,
		filepath: filepath,
		printSQL: options.printSQL,
	}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// Filepath returns the path to the Things database file.
func (d *DB) Filepath() string {
	return d.filepath
}

// executeQuery executes a SQL query and returns the results.
func (d *DB) executeQuery(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if d.printSQL {
		d.queryCount++
		fmt.Printf("/* Query %d */\n", d.queryCount)
		if len(args) > 0 {
			fmt.Printf("/* Parameters: %v */\n", args)
		}
		fmt.Println()
		fmt.Println(query)
		fmt.Println()
	}

	return d.db.QueryContext(ctx, query, args...)
}

// executeQueryRow executes a SQL query that returns a single row.
func (d *DB) executeQueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	if d.printSQL {
		d.queryCount++
		fmt.Printf("/* Query %d */\n", d.queryCount)
		if len(args) > 0 {
			fmt.Printf("/* Parameters: %v */\n", args)
		}
		fmt.Println()
		fmt.Println(query)
		fmt.Println()
	}

	return d.db.QueryRowContext(ctx, query, args...)
}

// scanTask scans a row into a Task struct.
//
//nolint:gocyclo,funlen // Complexity is inherent to scanning many fields
func scanTask(rows *sql.Rows) (*Task, error) {
	var task Task
	var typeStr, statusStr sql.NullString
	var trashed, tags, checklist sql.NullInt64
	var areaUUID, areaTitle, projectUUID, projectTitle sql.NullString
	var headingUUID, headingTitle, notes, start sql.NullString
	var startDate, deadline, reminderTime, stopDate sql.NullString
	var created, modified sql.NullString

	err := rows.Scan(
		&task.UUID,
		&typeStr,
		&trashed,
		&task.Title,
		&statusStr,
		&areaUUID,
		&areaTitle,
		&projectUUID,
		&projectTitle,
		&headingUUID,
		&headingTitle,
		&notes,
		&tags,
		&start,
		&checklist,
		&startDate,
		&deadline,
		&reminderTime,
		&stopDate,
		&created,
		&modified,
		&task.Index,
		&task.TodayIndex,
	)
	if err != nil {
		return nil, err
	}

	// Convert type string to TaskType
	switch typeStr.String {
	case "to-do":
		task.Type = TaskTypeTodo
	case "project":
		task.Type = TaskTypeProject
	case "heading":
		task.Type = TaskTypeHeading
	}

	// Convert status string to Status
	switch statusStr.String {
	case "incomplete":
		task.Status = StatusIncomplete
	case "completed":
		task.Status = StatusCompleted
	case "canceled":
		task.Status = StatusCanceled
	}

	// Set boolean fields
	task.Trashed = trashed.Valid && trashed.Int64 == 1

	// Set optional string fields
	if notes.Valid {
		task.Notes = notes.String
	}
	if start.Valid {
		task.Start = start.String
	}
	if areaUUID.Valid {
		task.AreaUUID = &areaUUID.String
	}
	if areaTitle.Valid {
		task.AreaTitle = &areaTitle.String
	}
	if projectUUID.Valid {
		task.ProjectUUID = &projectUUID.String
	}
	if projectTitle.Valid {
		task.ProjectTitle = &projectTitle.String
	}
	if headingUUID.Valid {
		task.HeadingUUID = &headingUUID.String
	}
	if headingTitle.Valid {
		task.HeadingTitle = &headingTitle.String
	}
	if startDate.Valid && startDate.String != "" {
		task.StartDate = &startDate.String
	}
	if deadline.Valid && deadline.String != "" {
		task.Deadline = &deadline.String
	}
	if reminderTime.Valid && reminderTime.String != "" {
		task.ReminderTime = &reminderTime.String
	}
	if stopDate.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", stopDate.String); err == nil {
			task.StopDate = &t
		}
	}
	if created.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", created.String); err == nil {
			task.Created = t
		}
	}
	if modified.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", modified.String); err == nil {
			task.Modified = t
		}
	}

	// Mark if task has tags or checklist (actual items loaded separately)
	if tags.Valid && tags.Int64 == 1 {
		task.Tags = []string{} // Will be populated later
	}
	if checklist.Valid && checklist.Int64 == 1 {
		task.Checklist = []ChecklistItem{} // Will be populated later
	}

	return &task, nil
}

// scanArea scans a row into an Area struct.
func scanArea(rows *sql.Rows) (*Area, error) {
	var area Area
	var typeStr sql.NullString
	var tags sql.NullInt64

	err := rows.Scan(
		&area.UUID,
		&typeStr,
		&area.Title,
		&tags,
	)
	if err != nil {
		return nil, err
	}

	area.Type = "area"
	if tags.Valid && tags.Int64 == 1 {
		area.Tags = []string{} // Will be populated later
	}

	return &area, nil
}

// scanTag scans a row into a Tag struct.
func scanTag(rows *sql.Rows) (*Tag, error) {
	var tag Tag
	var typeStr, shortcut sql.NullString

	err := rows.Scan(
		&tag.UUID,
		&typeStr,
		&tag.Title,
		&shortcut,
	)
	if err != nil {
		return nil, err
	}

	tag.Type = "tag"
	if shortcut.Valid {
		tag.Shortcut = shortcut.String
	}

	return &tag, nil
}

// scanChecklistItem scans a row into a ChecklistItem struct.
func scanChecklistItem(rows *sql.Rows) (*ChecklistItem, error) {
	var item ChecklistItem
	var typeStr, stopDate sql.NullString
	var created, modified sql.NullString

	err := rows.Scan(
		&item.Title,
		&item.Status,
		&stopDate,
		&typeStr,
		&item.UUID,
		&created,
		&modified,
	)
	if err != nil {
		return nil, err
	}

	item.Type = "checklist-item"
	if stopDate.Valid {
		if t, err := time.Parse("2006-01-02", stopDate.String); err == nil {
			item.StopDate = &t
		}
	}
	if created.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", created.String); err == nil {
			item.Created = t
		}
	}
	if modified.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", modified.String); err == nil {
			item.Modified = t
		}
	}

	return &item, nil
}

// getTagsOfTask returns the tag titles for a task.
func (d *DB) getTagsOfTask(ctx context.Context, taskUUID string) ([]string, error) {
	query := buildTagsOfTaskSQL()
	rows, err := d.executeQuery(ctx, query, taskUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			return nil, err
		}
		tags = append(tags, title)
	}

	return tags, rows.Err()
}

// getTagsOfArea returns the tag titles for an area.
func (d *DB) getTagsOfArea(ctx context.Context, areaUUID string) ([]string, error) {
	query := buildTagsOfAreaSQL()
	rows, err := d.executeQuery(ctx, query, areaUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			return nil, err
		}
		tags = append(tags, title)
	}

	return tags, rows.Err()
}

// getChecklistItems returns the checklist items for a to-do.
func (d *DB) getChecklistItems(ctx context.Context, todoUUID string) ([]ChecklistItem, error) {
	query := buildChecklistItemsSQL()
	rows, err := d.executeQuery(ctx, query, todoUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ChecklistItem
	for rows.Next() {
		item, err := scanChecklistItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}

	return items, rows.Err()
}

// Token returns the Things URL scheme authentication token.
func (d *DB) Token(ctx context.Context) (string, error) {
	query := buildAuthTokenSQL()
	var token sql.NullString
	if err := d.executeQueryRow(ctx, query).Scan(&token); err != nil {
		if err == sql.ErrNoRows {
			return "", ErrAuthTokenNotFound
		}
		return "", err
	}
	if !token.Valid {
		return "", ErrAuthTokenNotFound
	}
	return token.String, nil
}
