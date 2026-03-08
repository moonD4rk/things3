package db

import (
	"database/sql"
	"time"
)

// Time format constants used by Things 3 database.
const (
	dateFormat     = "2006-01-02"
	dateTimeFormat = "2006-01-02 15:04:05"
	timeFormat     = "15:04"
)

// scanTaskRow scans a sql.Rows into a TaskRow.
//
//nolint:gocyclo,funlen // Complexity is inherent to scanning many fields
func scanTaskRow(rows *sql.Rows) (*TaskRow, error) {
	var row TaskRow
	var typeStr, statusStr sql.NullString
	var trashed, tags, checklist sql.NullInt64
	var areaUUID, areaTitle, projectUUID, projectTitle sql.NullString
	var headingUUID, headingTitle, notes, start sql.NullString
	var startDate, deadline, reminderTime, stopDate sql.NullString
	var created, modified sql.NullString

	err := rows.Scan(
		&row.UUID,
		&typeStr,
		&trashed,
		&row.Title,
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
		&row.Index,
		&row.TodayIndex,
	)
	if err != nil {
		return nil, err
	}

	row.Type = nullStringValue(typeStr)
	row.Status = nullStringValue(statusStr)
	row.Trashed = trashed.Valid && trashed.Int64 == 1
	row.Notes = nullStringValue(notes)
	row.Start = nullStringValue(start)
	row.AreaUUID = nullString(areaUUID)
	row.AreaTitle = nullString(areaTitle)
	row.ProjectUUID = nullString(projectUUID)
	row.ProjectTitle = nullString(projectTitle)
	row.HeadingUUID = nullString(headingUUID)
	row.HeadingTitle = nullString(headingTitle)
	row.StartDate = parseDate(startDate)
	row.Deadline = parseDate(deadline)
	row.ReminderTime = parseTime(reminderTime)
	row.StopDate = parseDateTime(stopDate)
	if t := parseDateTime(created); t != nil {
		row.Created = *t
	}
	if t := parseDateTime(modified); t != nil {
		row.Modified = *t
	}
	row.HasTags = tags.Valid && tags.Int64 == 1
	row.HasChecklist = checklist.Valid && checklist.Int64 == 1

	return &row, nil
}

// scanAreaRow scans a sql.Rows into an AreaRow.
func scanAreaRow(rows *sql.Rows) (*AreaRow, error) {
	var row AreaRow
	var typeStr sql.NullString
	var tags sql.NullInt64

	err := rows.Scan(&row.UUID, &typeStr, &row.Title, &tags)
	if err != nil {
		return nil, err
	}

	row.HasTags = tags.Valid && tags.Int64 == 1

	return &row, nil
}

// scanTagRow scans a sql.Rows into a TagRow.
func scanTagRow(rows *sql.Rows) (*TagRow, error) {
	var row TagRow
	var typeStr, shortcut sql.NullString

	err := rows.Scan(&row.UUID, &typeStr, &row.Title, &shortcut)
	if err != nil {
		return nil, err
	}

	if shortcut.Valid {
		row.Shortcut = shortcut.String
	}

	return &row, nil
}

// scanChecklistItemRow scans a sql.Rows into a ChecklistItemRow.
func scanChecklistItemRow(rows *sql.Rows) (*ChecklistItemRow, error) {
	var row ChecklistItemRow
	var typeStr, stopDate sql.NullString
	var created, modified sql.NullString

	err := rows.Scan(&row.Title, &row.Status, &stopDate, &typeStr, &row.UUID, &created, &modified)
	if err != nil {
		return nil, err
	}

	row.StopDate = parseDate(stopDate)
	if t := parseDateTime(created); t != nil {
		row.Created = *t
	}
	if t := parseDateTime(modified); t != nil {
		row.Modified = *t
	}

	return &row, nil
}

// parseDate parses a date string in "2006-01-02" format.
// Returns nil if the string is empty or invalid.
func parseDate(s sql.NullString) *time.Time {
	if !s.Valid || s.String == "" {
		return nil
	}
	t, err := time.Parse(dateFormat, s.String)
	if err != nil {
		return nil
	}
	return &t
}

// parseDateTime parses a datetime string in "2006-01-02 15:04:05" format.
// Returns nil if the string is empty or invalid.
func parseDateTime(s sql.NullString) *time.Time {
	if !s.Valid || s.String == "" {
		return nil
	}
	t, err := time.Parse(dateTimeFormat, s.String)
	if err != nil {
		return nil
	}
	return &t
}

// parseTime parses a time string in "15:04" format.
// Returns nil if the string is empty or invalid.
func parseTime(s sql.NullString) *time.Time {
	if !s.Valid || s.String == "" {
		return nil
	}
	t, err := time.Parse(timeFormat, s.String)
	if err != nil {
		return nil
	}
	return &t
}

// nullString returns nil if NULL, otherwise returns pointer to string.
func nullString(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

// nullStringValue returns empty string if NULL, otherwise returns the string value.
func nullStringValue(s sql.NullString) string {
	if !s.Valid {
		return ""
	}
	return s.String
}
