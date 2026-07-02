package things3

import "time"

// Todo represents an actionable todo item in Things 3.
type Todo struct {
	UUID   string      `json:"uuid"`
	Title  string      `json:"title"`
	Status Status      `json:"status"`
	Notes  string      `json:"notes,omitempty"`
	Start  StartBucket `json:"start"`

	// Relationships (empty string = no relationship)
	AreaUUID     string `json:"area_uuid,omitempty"`
	AreaTitle    string `json:"area_title,omitempty"`
	ProjectUUID  string `json:"project_uuid,omitempty"`
	ProjectTitle string `json:"project_title,omitempty"`
	HeadingUUID  string `json:"heading_uuid,omitempty"`
	HeadingTitle string `json:"heading_title,omitempty"`

	// Attributes
	Tags      []string        `json:"tags,omitempty"`
	Checklist []ChecklistItem `json:"checklist,omitempty"`

	// Dates (date only, no time component)
	StartDate *time.Time `json:"start_date,omitempty"`
	Deadline  *time.Time `json:"deadline,omitempty"`

	// Time (time only, date component is zero value)
	Reminder *time.Time `json:"reminder,omitempty"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	ModifiedAt  time.Time  `json:"modified_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CanceledAt  *time.Time `json:"canceled_at,omitempty"`

	Trashed bool `json:"trashed,omitempty"`

	// Evening reports whether the todo is in the This Evening section of Today.
	Evening bool `json:"evening,omitempty"`
	// Repeating reports whether the todo belongs to a repeating series, either a
	// generated instance or the template that schedules its next occurrence.
	Repeating bool `json:"repeating,omitempty"`
}

// Project represents a container for organizing todos in Things 3.
type Project struct {
	UUID   string      `json:"uuid"`
	Title  string      `json:"title"`
	Status Status      `json:"status"`
	Notes  string      `json:"notes,omitempty"`
	Start  StartBucket `json:"start"`

	// Relationships
	AreaUUID  string `json:"area_uuid,omitempty"`
	AreaTitle string `json:"area_title,omitempty"`

	// Attributes
	Tags []string `json:"tags,omitempty"`

	// Dates (date only, no time component)
	StartDate *time.Time `json:"start_date,omitempty"`
	Deadline  *time.Time `json:"deadline,omitempty"`

	// Time (time only, date component is zero value)
	Reminder *time.Time `json:"reminder,omitempty"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	ModifiedAt  time.Time  `json:"modified_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CanceledAt  *time.Time `json:"canceled_at,omitempty"`

	Trashed bool `json:"trashed,omitempty"`

	// Repeating reports whether the project belongs to a repeating series, either
	// a generated instance or the template that schedules its next occurrence.
	Repeating bool `json:"repeating,omitempty"`
}

// Heading represents a grouping label within a project.
// Headings are organizational only and carry no status, dates, or notes.
type Heading struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`

	// Parent project
	ProjectUUID  string `json:"project_uuid,omitempty"`
	ProjectTitle string `json:"project_title,omitempty"`
}

// Area represents a high-level responsibility area in Things 3.
type Area struct {
	UUID  string   `json:"uuid"`
	Title string   `json:"title"`
	Tags  []string `json:"tags,omitempty"`
}

// Tag represents a label for categorizing items in Things 3.
type Tag struct {
	UUID     string `json:"uuid"`
	Title    string `json:"title"`
	Shortcut string `json:"shortcut,omitempty"`
}

// ChecklistItem represents a sub-item within a todo.
type ChecklistItem struct {
	UUID   string `json:"uuid"`
	Title  string `json:"title"`
	Status Status `json:"status"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	ModifiedAt  time.Time  `json:"modified_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CanceledAt  *time.Time `json:"canceled_at,omitempty"`
}
