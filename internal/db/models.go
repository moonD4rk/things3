package db

import "time"

// TaskRow represents a row from a task query result.
type TaskRow struct {
	UUID         string
	Type         string // "to-do", "project", "heading"
	Trashed      bool
	Title        string
	Status       string // "incomplete", "completed", "canceled"
	AreaUUID     *string
	AreaTitle    *string
	ProjectUUID  *string
	ProjectTitle *string
	HeadingUUID  *string
	HeadingTitle *string
	Notes        string
	HasTags      bool
	Start        string // "Inbox", "Anytime", "Someday"
	HasChecklist bool
	StartDate    *time.Time
	Deadline     *time.Time
	ReminderTime *time.Time
	StopDate     *time.Time
	Created      time.Time
	Modified     time.Time
	Index        int
	TodayIndex   int
}

// AreaRow represents a row from an area query result.
type AreaRow struct {
	UUID    string
	Title   string
	HasTags bool
}

// TagRow represents a row from a tag query result.
type TagRow struct {
	UUID     string
	Title    string
	Shortcut string
}

// ChecklistItemRow represents a row from a checklist item query result.
type ChecklistItemRow struct {
	UUID     string
	Title    string
	Status   string // "incomplete", "completed", "canceled"
	StopDate *time.Time
	Created  time.Time
	Modified time.Time
}
