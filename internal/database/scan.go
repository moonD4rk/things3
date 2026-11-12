package database

import (
	"database/sql"
	"time"
)

// scanTaskRow scans a sql.Rows into a TaskRow.
func scanTaskRow(rows *sql.Rows) (*TaskRow, error) {
	var s taskScanRow
	err := rows.Scan(
		&s.uuid, &s.taskType, &s.trashed, &s.title, &s.status,
		&s.areaUUID, &s.areaTitle, &s.projectUUID, &s.projectTitle,
		&s.headingUUID, &s.headingTitle, &s.notes, &s.hasTags, &s.start,
		&s.hasChecklist, &s.startDate, &s.deadline, &s.reminderTime,
		&s.stopDate, &s.created, &s.modified, &s.index, &s.todayIndex,
		&s.startBucket, &s.repeating, &s.template,
	)
	if err != nil {
		return nil, err
	}
	return s.toTaskRow(), nil
}

// taskScanRow holds raw SQL scan targets for a task query.
type taskScanRow struct {
	uuid                                           string
	title, notes                                   sql.NullString
	taskType, status, start, startBucket           sql.NullInt64
	trashed, hasTags, hasChecklist                 sql.NullInt64
	repeating, template                            sql.NullInt64
	index, todayIndex                              sql.NullInt64
	areaUUID, areaTitle, projectUUID, projectTitle sql.NullString
	headingUUID, headingTitle                      sql.NullString
	startDate, deadline, reminderTime              sql.NullInt64
	stopDate, created, modified                    sql.NullFloat64
}

// toTaskRow converts raw scan values into a TaskRow.
func (s *taskScanRow) toTaskRow() *TaskRow {
	return &TaskRow{
		UUID:         s.uuid,
		Type:         int(s.taskType.Int64),
		Trashed:      nullBool(s.trashed),
		Title:        s.title.String,
		Status:       int(s.status.Int64),
		AreaUUID:     s.areaUUID.String,
		AreaTitle:    s.areaTitle.String,
		ProjectUUID:  s.projectUUID.String,
		ProjectTitle: s.projectTitle.String,
		HeadingUUID:  s.headingUUID.String,
		HeadingTitle: s.headingTitle.String,
		Notes:        s.notes.String,
		HasTags:      nullBool(s.hasTags),
		Start:        int(s.start.Int64),
		HasChecklist: nullBool(s.hasChecklist),
		StartDate:    thingsDatePtr(s.startDate),
		Deadline:     thingsDatePtr(s.deadline),
		ReminderTime: thingsClockPtr(s.reminderTime),
		StopDate:     unixTimePtr(s.stopDate),
		Created:      unixTimeValue(s.created),
		Modified:     unixTimeValue(s.modified),
		Index:        int(s.index.Int64),
		TodayIndex:   int(s.todayIndex.Int64),
		Evening:      s.startBucket.Valid && s.startBucket.Int64 == startBucketEvening,
		Repeating:    nullBool(s.repeating),
		Template:     nullBool(s.template),
	}
}

// scanAreaRow scans a sql.Rows into an AreaRow.
func scanAreaRow(rows *sql.Rows) (*AreaRow, error) {
	var row AreaRow
	var title sql.NullString
	var hasTags sql.NullInt64

	if err := rows.Scan(&row.UUID, &title, &hasTags); err != nil {
		return nil, err
	}

	row.Title = title.String
	row.HasTags = nullBool(hasTags)

	return &row, nil
}

// scanTagRow scans a sql.Rows into a TagRow.
func scanTagRow(rows *sql.Rows) (*TagRow, error) {
	var row TagRow
	var title, shortcut sql.NullString

	if err := rows.Scan(&row.UUID, &title, &shortcut); err != nil {
		return nil, err
	}

	row.Title = title.String
	row.Shortcut = shortcut.String

	return &row, nil
}

// scanChecklistItemRow scans a sql.Rows into a ChecklistItemRow.
func scanChecklistItemRow(rows *sql.Rows) (*ChecklistItemRow, error) {
	var row ChecklistItemRow
	var title sql.NullString
	var status sql.NullInt64
	var stopDate, created, modified sql.NullFloat64

	if err := rows.Scan(&row.UUID, &title, &status, &stopDate, &created, &modified); err != nil {
		return nil, err
	}

	row.Title = title.String
	row.Status = int(status.Int64)
	row.StopDate = unixTimePtr(stopDate)
	row.Created = unixTimeValue(created)
	row.Modified = unixTimeValue(modified)

	return &row, nil
}

// thingsDatePtr unpacks a nullable Things date; NULL and sentinels yield nil.
func thingsDatePtr(n sql.NullInt64) *time.Time {
	if !n.Valid || !isRealThingsDate(n.Int64) {
		return nil
	}
	t := thingsDateToTime(n.Int64)
	return &t
}

// thingsClockPtr unpacks a nullable reminder time. Zero is a real 00:00
// reminder; only NULL means "no reminder".
func thingsClockPtr(n sql.NullInt64) *time.Time {
	if !n.Valid || n.Int64 < 0 {
		return nil
	}
	t := thingsTimeToClock(n.Int64)
	return &t
}

// nullBool reports whether the value is valid and non-zero.
func nullBool(n sql.NullInt64) bool {
	return n.Valid && n.Int64 != 0
}

// unixTimePtr converts a nullable Unix epoch value to a local instant; NULL and
// zero yield nil.
func unixTimePtr(f sql.NullFloat64) *time.Time {
	if !f.Valid {
		return nil
	}
	t := unixToTime(f.Float64)
	if t.IsZero() {
		return nil
	}
	return &t
}

// unixTimeValue converts a nullable Unix epoch value to a local instant; NULL
// yields the zero time.
func unixTimeValue(f sql.NullFloat64) time.Time {
	if !f.Valid {
		return time.Time{}
	}
	return unixToTime(f.Float64)
}
