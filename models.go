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
	StartDate    *string    `json:"start_date,omitempty"`    // ISO 8601 date
	Deadline     *string    `json:"deadline,omitempty"`      // ISO 8601 date
	ReminderTime *string    `json:"reminder_time,omitempty"` // HH:MM format
	StopDate     *time.Time `json:"stop_date,omitempty"`     // Completion/cancellation date
	Created      time.Time  `json:"created"`
	Modified     time.Time  `json:"modified"`

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
	UUID     string     `json:"uuid"`
	Type     string     `json:"type"` // Always "checklist-item"
	Title    string     `json:"title"`
	Status   string     `json:"status"` // "incomplete", "completed", or "canceled"
	StopDate *time.Time `json:"stop_date,omitempty"`
	Created  time.Time  `json:"created"`
	Modified time.Time  `json:"modified"`
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
