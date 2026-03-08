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
func scanTaskRow(rows *sql.Rows) (*TaskRow, error) {
	var s taskScanRow
	err := rows.Scan(
		&s.uuid, &s.typeStr, &s.trashed, &s.title, &s.statusStr,
		&s.areaUUID, &s.areaTitle, &s.projectUUID, &s.projectTitle,
		&s.headingUUID, &s.headingTitle, &s.notes, &s.tags, &s.start,
		&s.checklist, &s.startDate, &s.deadline, &s.reminderTime,
		&s.stopDate, &s.created, &s.modified, &s.index, &s.todayIndex,
	)
	if err != nil {
		return nil, err
	}
	return s.toTaskRow(), nil
}

// taskScanRow holds raw SQL scan targets for a task query.
type taskScanRow struct {
	uuid, title                                    string
	index, todayIndex                              int
	typeStr, statusStr                             sql.NullString
	trashed, tags, checklist                       sql.NullInt64
	areaUUID, areaTitle, projectUUID, projectTitle sql.NullString
	headingUUID, headingTitle, notes, start        sql.NullString
	startDate, deadline, reminderTime, stopDate    sql.NullString
	created, modified                              sql.NullString
}

// toTaskRow converts raw scan values into a TaskRow.
func (s *taskScanRow) toTaskRow() *TaskRow {
	row := &TaskRow{
		UUID:         s.uuid,
		Type:         nullStringValue(s.typeStr),
		Trashed:      nullBool(s.trashed),
		Title:        s.title,
		Status:       nullStringValue(s.statusStr),
		AreaUUID:     nullString(s.areaUUID),
		AreaTitle:    nullString(s.areaTitle),
		ProjectUUID:  nullString(s.projectUUID),
		ProjectTitle: nullString(s.projectTitle),
		HeadingUUID:  nullString(s.headingUUID),
		HeadingTitle: nullString(s.headingTitle),
		Notes:        nullStringValue(s.notes),
		HasTags:      nullBool(s.tags),
		Start:        nullStringValue(s.start),
		HasChecklist: nullBool(s.checklist),
		StartDate:    parseDate(s.startDate),
		Deadline:     parseDate(s.deadline),
		ReminderTime: parseTime(s.reminderTime),
		StopDate:     parseDateTime(s.stopDate),
		Created:      parseDateTimeValue(s.created),
		Modified:     parseDateTimeValue(s.modified),
		Index:        s.index,
		TodayIndex:   s.todayIndex,
	}
	return row
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

	row.HasTags = nullBool(tags)

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

	row.Shortcut = nullStringValue(shortcut)

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
	row.Created = parseDateTimeValue(created)
	row.Modified = parseDateTimeValue(modified)

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

// nullBool returns true if the value is valid and equals 1.
func nullBool(n sql.NullInt64) bool {
	return n.Valid && n.Int64 == 1
}

// parseDateTimeValue parses a datetime string, returning zero time on failure.
func parseDateTimeValue(s sql.NullString) time.Time {
	if t := parseDateTime(s); t != nil {
		return *t
	}
	return time.Time{}
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
