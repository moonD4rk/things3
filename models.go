package things3

import "time"

// Status string constants for ChecklistItem.
const (
	statusStringIncomplete = "incomplete"
	statusStringCompleted  = "completed"
	statusStringCanceled   = "canceled"
)

// Task represents a task in Things 3, which can be a to-do, project, or heading.
type Task struct {
	UUID   string   `json:"uuid"`
	Type   TaskType `json:"type"`
	Title  string   `json:"title"`
	Status Status   `json:"status"`
	Notes  string   `json:"notes,omitempty"`
	Start  string   `json:"start,omitempty"` // "Inbox", "Anytime", or "Someday"

	// Trashed indicates whether the task is in the trash.
	Trashed bool `json:"trashed,omitempty"`

	// Relationships
	AreaUUID     *string `json:"area,omitempty"`
	AreaTitle    *string `json:"area_title,omitempty"`
	ProjectUUID  *string `json:"project,omitempty"`
	ProjectTitle *string `json:"project_title,omitempty"`
	HeadingUUID  *string `json:"heading,omitempty"`
	HeadingTitle *string `json:"heading_title,omitempty"`

	// Dates
	// All date/time fields are converted from SQLite string formats to time.Time.
	//
	// StartDate: scheduled start date.
	//   Database: "2024-01-15" (date only, format "YYYY-MM-DD")
	//   Parsed:   time.Time with zero time component
	StartDate *time.Time `json:"start_date,omitempty"`
	// Deadline: task deadline date.
	//   Database: "2024-01-15" (date only, format "YYYY-MM-DD")
	//   Parsed:   time.Time with zero time component
	Deadline *time.Time `json:"deadline,omitempty"`
	// ReminderTime: time-only reminder (date component is zero value).
	//   Database: "14:30" (time only, format "HH:MM")
	//   Parsed:   time.Time with zero date (0000-01-01), only Hour/Minute meaningful
	ReminderTime *time.Time `json:"reminder_time,omitempty"`
	// StopDate: completion or cancellation timestamp.
	//   Database: "2024-01-15 10:30:45" (datetime, format "YYYY-MM-DD HH:MM:SS")
	//   Parsed:   time.Time with full date and time
	StopDate *time.Time `json:"stop_date,omitempty"`
	// Created: task creation timestamp.
	//   Database: "2024-01-15 10:30:45" (datetime, format "YYYY-MM-DD HH:MM:SS")
	//   Parsed:   time.Time with full date and time
	Created time.Time `json:"created"`
	// Modified: last modification timestamp.
	//   Database: "2024-01-15 10:30:45" (datetime, format "YYYY-MM-DD HH:MM:SS")
	//   Parsed:   time.Time with full date and time
	Modified time.Time `json:"modified"`

	// Index values for ordering
	Index      int `json:"index"`
	TodayIndex int `json:"today_index"`

	// Nested items (populated when include_items=true)
	Tags      []string        `json:"tags,omitempty"`
	Checklist []ChecklistItem `json:"checklist,omitempty"`
	Items     []Task          `json:"items,omitempty"` // For projects and headings
}

// IsTodo returns true if the task is a to-do item.
func (t *Task) IsTodo() bool {
	return t.Type == TaskTypeTodo
}

// IsProject returns true if the task is a project.
func (t *Task) IsProject() bool {
	return t.Type == TaskTypeProject
}

// IsHeading returns true if the task is a heading.
func (t *Task) IsHeading() bool {
	return t.Type == TaskTypeHeading
}

// IsIncomplete returns true if the task status is incomplete.
func (t *Task) IsIncomplete() bool {
	return t.Status == StatusIncomplete
}

// IsCompleted returns true if the task status is completed.
func (t *Task) IsCompleted() bool {
	return t.Status == StatusCompleted
}

// IsCanceled returns true if the task status is canceled.
func (t *Task) IsCanceled() bool {
	return t.Status == StatusCanceled
}

// HasTags returns true if the task has any tags.
func (t *Task) HasTags() bool {
	return len(t.Tags) > 0
}

// HasChecklist returns true if the task has a checklist.
func (t *Task) HasChecklist() bool {
	return len(t.Checklist) > 0
}

// Area represents an area in Things 3.
type Area struct {
	UUID  string `json:"uuid"`
	Type  string `json:"type"` // Always "area"
	Title string `json:"title"`

	// Nested items (populated when include_items=true)
	Tags  []string `json:"tags,omitempty"`
	Items []Task   `json:"items,omitempty"`
}

// Tag represents a tag in Things 3.
type Tag struct {
	UUID     string `json:"uuid"`
	Type     string `json:"type"` // Always "tag"
	Title    string `json:"title"`
	Shortcut string `json:"shortcut,omitempty"`

	// Nested items (populated when include_items=true)
	Items []any `json:"items,omitempty"` // Can contain Area or Task
}

// ChecklistItem represents a checklist item within a to-do.
type ChecklistItem struct {
	UUID   string `json:"uuid"`
	Type   string `json:"type"` // Always "checklist-item"
	Title  string `json:"title"`
	Status string `json:"status"` // "incomplete", "completed", or "canceled"
	// StopDate: completion date.
	//   Database: "2024-01-15" (date only, format "YYYY-MM-DD")
	//   Parsed:   time.Time with zero time component
	StopDate *time.Time `json:"stop_date,omitempty"`
	// Created: item creation timestamp.
	//   Database: "2024-01-15 10:30:45" (datetime, format "YYYY-MM-DD HH:MM:SS")
	//   Parsed:   time.Time with full date and time
	Created time.Time `json:"created"`
	// Modified: last modification timestamp.
	//   Database: "2024-01-15 10:30:45" (datetime, format "YYYY-MM-DD HH:MM:SS")
	//   Parsed:   time.Time with full date and time
	Modified time.Time `json:"modified"`
}

// IsIncomplete returns true if the checklist item is incomplete.
func (c *ChecklistItem) IsIncomplete() bool {
	return c.Status == statusStringIncomplete
}

// IsCompleted returns true if the checklist item is completed.
func (c *ChecklistItem) IsCompleted() bool {
	return c.Status == statusStringCompleted
}

// IsCanceled returns true if the checklist item is canceled.
func (c *ChecklistItem) IsCanceled() bool {
	return c.Status == statusStringCanceled
}
