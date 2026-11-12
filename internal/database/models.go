package database

import "time"

// TaskRow represents a row from a task query result. Type, Status, and Start
// carry the raw TMTask integers. Project and area are the effective containers:
// a todo under a heading reports the heading's project, and anything inside a
// project reports that project's area. Empty strings mean "none".
type TaskRow struct {
	UUID         string
	Type         int
	Trashed      bool
	Title        string
	Status       int
	AreaUUID     string
	AreaTitle    string
	ProjectUUID  string
	ProjectTitle string
	HeadingUUID  string
	HeadingTitle string
	Notes        string
	HasTags      bool
	Start        int
	HasChecklist bool
	StartDate    *time.Time
	Deadline     *time.Time
	ReminderTime *time.Time
	StopDate     *time.Time
	Created      time.Time
	Modified     time.Time
	Index        int
	TodayIndex   int
	Evening      bool
	Repeating    bool
	Template     bool
	// Repeat is the rule of the series the task belongs to: its own for a
	// template, its template's for a generated instance.
	Repeat *RepeatRule
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

// ChecklistItemRow represents a row from a checklist item query result. Status
// carries the raw TMChecklistItem integer.
type ChecklistItemRow struct {
	UUID     string
	Title    string
	Status   int
	StopDate *time.Time
	Created  time.Time
	Modified time.Time
}
