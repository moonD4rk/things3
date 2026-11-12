package things3

import (
	"encoding/json"
	"time"
)

// Item type discriminators, as rendered in JSON and YAML.
const (
	ItemTypeTodo    = "todo"
	ItemTypeProject = "project"
)

// Item is one row of a sidebar view: a todo or a project. The app's lists mix
// the two (a project scheduled for today sits in Today among the todos), so the
// views return this pair rather than forcing both into one type. Exactly one
// field is set.
type Item struct {
	Todo    *Todo
	Project *Project
}

// IsProject reports whether the item is a project rather than a todo.
func (it Item) IsProject() bool {
	return it.Project != nil
}

// Type returns ItemTypeProject or ItemTypeTodo.
func (it Item) Type() string {
	if it.Project != nil {
		return ItemTypeProject
	}
	return ItemTypeTodo
}

// UUID returns the item's UUID.
func (it Item) UUID() string {
	if it.Project != nil {
		return it.Project.UUID
	}
	return it.Todo.UUID
}

// Title returns the item's title.
func (it Item) Title() string {
	if it.Project != nil {
		return it.Project.Title
	}
	return it.Todo.Title
}

// Status returns the item's completion status.
func (it Item) Status() Status {
	if it.Project != nil {
		return it.Project.Status
	}
	return it.Todo.Status
}

// Tags returns the item's own tags.
func (it Item) Tags() []string {
	if it.Project != nil {
		return it.Project.Tags
	}
	return it.Todo.Tags
}

// AreaUUID returns the UUID of the area the item belongs to, or "".
func (it Item) AreaUUID() string {
	if it.Project != nil {
		return it.Project.AreaUUID
	}
	return it.Todo.AreaUUID
}

// StartDate returns the item's scheduled date, or nil.
func (it Item) StartDate() *time.Time {
	if it.Project != nil {
		return it.Project.StartDate
	}
	return it.Todo.StartDate
}

// Deadline returns the item's deadline, or nil.
func (it Item) Deadline() *time.Time {
	if it.Project != nil {
		return it.Project.Deadline
	}
	return it.Todo.Deadline
}

// StoppedAt returns when the item was completed or canceled, or nil while it
// is open.
func (it Item) StoppedAt() *time.Time {
	if it.Project != nil {
		return firstTime(it.Project.CompletedAt, it.Project.CanceledAt)
	}
	return firstTime(it.Todo.CompletedAt, it.Todo.CanceledAt)
}

// CreatedAt returns when the item was created.
func (it Item) CreatedAt() time.Time {
	if it.Project != nil {
		return it.Project.CreatedAt
	}
	return it.Todo.CreatedAt
}

// ModifiedAt returns when the item was last modified.
func (it Item) ModifiedAt() time.Time {
	if it.Project != nil {
		return it.Project.ModifiedAt
	}
	return it.Todo.ModifiedAt
}

// UpcomingDate returns the day the Upcoming view files the item under: its
// start date (for a repeating template, its next occurrence), or else its
// deadline. It is nil for an item with neither.
func (it Item) UpcomingDate() *time.Time {
	return firstTime(it.StartDate(), it.Deadline())
}

// MarshalJSON renders the todo or project with a leading "type" discriminator.
func (it Item) MarshalJSON() ([]byte, error) {
	var (
		raw []byte
		err error
	)
	if it.Project != nil {
		raw, err = json.Marshal(it.Project)
	} else {
		raw, err = json.Marshal(it.Todo)
	}
	if err != nil {
		return nil, err
	}
	// Splicing keeps the model's own field order, which a map would sort away.
	prefix := `{"type":"` + it.Type() + `",`
	return append([]byte(prefix), raw[1:]...), nil
}

// MarshalYAML mirrors MarshalJSON so YAML carries the same shape.
func (it Item) MarshalYAML() (any, error) {
	if it.Project != nil {
		return struct {
			Type    string `yaml:"type"`
			Project `yaml:",inline"`
		}{ItemTypeProject, *it.Project}, nil
	}
	return struct {
		Type string `yaml:"type"`
		Todo `yaml:",inline"`
	}{ItemTypeTodo, *it.Todo}, nil
}

func firstTime(times ...*time.Time) *time.Time {
	for _, t := range times {
		if t != nil {
			return t
		}
	}
	return nil
}
