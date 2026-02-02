package things3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// db provides read-only access to the Things 3 database.
type db struct {
	sqlDB      *sql.DB
	filepath   string
	printSQL   bool
	queryCount int
}

// newDB creates a new Things 3 database connection.
// Options can be provided to configure the database behavior.
func newDB(opts ...dbOption) (*db, error) {
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
	sqlDB, err := openDatabase(filepath)
	if err != nil {
		return nil, err
	}

	// Validate database version
	if err := validateDatabaseVersion(sqlDB); err != nil {
		sqlDB.Close()
		return nil, err
	}

	return &db{
		sqlDB:    sqlDB,
		filepath: filepath,
		printSQL: options.printSQL,
	}, nil
}

// Close closes the database connection.
func (d *db) Close() error {
	if d.sqlDB != nil {
		return d.sqlDB.Close()
	}
	return nil
}

// Filepath returns the path to the Things database file.
func (d *db) Filepath() string {
	return d.filepath
}

// executeQuery executes a SQL query and returns the results.
func (d *db) executeQuery(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
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

	return d.sqlDB.QueryContext(ctx, query, args...)
}

// executeQueryRow executes a SQL query that returns a single row.
func (d *db) executeQueryRow(ctx context.Context, query string, args ...any) *sql.Row {
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

	return d.sqlDB.QueryRowContext(ctx, query, args...)
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
	task.Notes = nullStringValue(notes)
	task.Start = nullStringValue(start)
	task.AreaUUID = nullString(areaUUID)
	task.AreaTitle = nullString(areaTitle)
	task.ProjectUUID = nullString(projectUUID)
	task.ProjectTitle = nullString(projectTitle)
	task.HeadingUUID = nullString(headingUUID)
	task.HeadingTitle = nullString(headingTitle)
	task.StartDate = parseDate(startDate)
	task.Deadline = parseDate(deadline)
	task.ReminderTime = parseTime(reminderTime)
	task.StopDate = parseDateTime(stopDate)
	if t := parseDateTime(created); t != nil {
		task.Created = *t
	}
	if t := parseDateTime(modified); t != nil {
		task.Modified = *t
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
	item.StopDate = parseDate(stopDate)
	if t := parseDateTime(created); t != nil {
		item.Created = *t
	}
	if t := parseDateTime(modified); t != nil {
		item.Modified = *t
	}

	return &item, nil
}

// getTagsOfTask returns the tag titles for a task.
func (d *db) getTagsOfTask(ctx context.Context, taskUUID string) ([]string, error) {
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
func (d *db) getTagsOfArea(ctx context.Context, areaUUID string) ([]string, error) {
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
func (d *db) getChecklistItems(ctx context.Context, todoUUID string) ([]ChecklistItem, error) {
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
func (d *db) Token(ctx context.Context) (string, error) {
	query := buildAuthTokenSQL()
	var token sql.NullString
	if err := d.executeQueryRow(ctx, query).Scan(&token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrAuthTokenNotFound
		}
		return "", err
	}
	if !token.Valid {
		return "", ErrAuthTokenNotFound
	}
	return token.String, nil
}
