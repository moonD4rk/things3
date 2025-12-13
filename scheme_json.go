package things3

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// JSONOperation represents the operation type for a JSON item.
type JSONOperation string

const (
	// JSONOperationCreate creates a new item.
	JSONOperationCreate JSONOperation = "create"
	// JSONOperationUpdate updates an existing item.
	JSONOperationUpdate JSONOperation = "update"
)

// JSONItemType represents the type of item in a JSON operation.
type JSONItemType string

const (
	// JSONItemTypeTodo represents a to-do item.
	JSONItemTypeTodo JSONItemType = "to-do"
	// JSONItemTypeProject represents a project item.
	JSONItemTypeProject JSONItemType = "project"
)

// JSONItem represents a single item in a JSON batch operation.
type JSONItem struct {
	Type       JSONItemType   `json:"type"`
	Operation  JSONOperation  `json:"operation,omitempty"`
	ID         string         `json:"id,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// JSONTodoItem builds a to-do item for JSON batch operations.
type JSONTodoItem struct {
	item JSONItem
	err  error
}

// newJSONTodoItem creates a new JSONTodoItem for create operations.
func newJSONTodoItem() *JSONTodoItem {
	return &JSONTodoItem{
		item: JSONItem{
			Type:       JSONItemTypeTodo,
			Operation:  JSONOperationCreate,
			Attributes: make(map[string]any),
		},
	}
}

// newJSONTodoItemUpdate creates a new JSONTodoItem for update operations.
func newJSONTodoItemUpdate(id string) *JSONTodoItem {
	return &JSONTodoItem{
		item: JSONItem{
			Type:       JSONItemTypeTodo,
			Operation:  JSONOperationUpdate,
			ID:         id,
			Attributes: make(map[string]any),
		},
	}
}

// Title sets the to-do title.
func (t *JSONTodoItem) Title(title string) *JSONTodoItem {
	if len(title) > maxTitleLength {
		t.err = ErrTitleTooLong
		return t
	}
	t.item.Attributes["title"] = title
	return t
}

// Notes sets the to-do notes.
func (t *JSONTodoItem) Notes(notes string) *JSONTodoItem {
	if len(notes) > maxNotesLength {
		t.err = ErrNotesTooLong
		return t
	}
	t.item.Attributes["notes"] = notes
	return t
}

// When sets the scheduling date.
func (t *JSONTodoItem) When(when When) *JSONTodoItem {
	t.item.Attributes["when"] = string(when)
	return t
}

// WhenDate sets a specific date for scheduling.
func (t *JSONTodoItem) WhenDate(year int, month time.Month, day int) *JSONTodoItem {
	t.item.Attributes["when"] = fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
	return t
}

// Deadline sets the deadline date.
func (t *JSONTodoItem) Deadline(date string) *JSONTodoItem {
	t.item.Attributes["deadline"] = date
	return t
}

// Tags sets the tags for the to-do.
func (t *JSONTodoItem) Tags(tags ...string) *JSONTodoItem {
	t.item.Attributes["tags"] = tags
	return t
}

// ChecklistItems sets the checklist items.
func (t *JSONTodoItem) ChecklistItems(items ...string) *JSONTodoItem {
	if len(items) > maxChecklistItems {
		t.err = ErrTooManyChecklistItems
		return t
	}
	checklistItems := make([]map[string]any, len(items))
	for i, item := range items {
		checklistItems[i] = map[string]any{
			"type":       "checklist-item",
			"attributes": map[string]any{"title": item},
		}
	}
	t.item.Attributes["checklist-items"] = checklistItems
	return t
}

// List sets the target project or area by name.
func (t *JSONTodoItem) List(name string) *JSONTodoItem {
	t.item.Attributes["list"] = name
	return t
}

// ListID sets the target project or area by UUID.
func (t *JSONTodoItem) ListID(id string) *JSONTodoItem {
	t.item.Attributes["list-id"] = id
	return t
}

// Heading sets the target heading within a project by name.
func (t *JSONTodoItem) Heading(name string) *JSONTodoItem {
	t.item.Attributes["heading"] = name
	return t
}

// Completed sets the completion status.
func (t *JSONTodoItem) Completed(completed bool) *JSONTodoItem {
	t.item.Attributes["completed"] = completed
	return t
}

// Canceled sets the canceled status.
func (t *JSONTodoItem) Canceled(canceled bool) *JSONTodoItem {
	t.item.Attributes["canceled"] = canceled
	return t
}

// CreationDate sets the creation timestamp.
func (t *JSONTodoItem) CreationDate(date time.Time) *JSONTodoItem {
	t.item.Attributes["creation-date"] = date.Format(time.RFC3339)
	return t
}

// CompletionDate sets the completion timestamp.
func (t *JSONTodoItem) CompletionDate(date time.Time) *JSONTodoItem {
	t.item.Attributes["completion-date"] = date.Format(time.RFC3339)
	return t
}

// PrependNotes prepends text to existing notes (update only).
func (t *JSONTodoItem) PrependNotes(notes string) *JSONTodoItem {
	t.item.Attributes["prepend-notes"] = notes
	return t
}

// AppendNotes appends text to existing notes (update only).
func (t *JSONTodoItem) AppendNotes(notes string) *JSONTodoItem {
	t.item.Attributes["append-notes"] = notes
	return t
}

// AddTags adds tags without replacing existing ones (update only).
func (t *JSONTodoItem) AddTags(tags ...string) *JSONTodoItem {
	t.item.Attributes["add-tags"] = tags
	return t
}

// build returns the JSON item and any error.
func (t *JSONTodoItem) build() (JSONItem, error) {
	return t.item, t.err
}

// JSONProjectItem builds a project item for JSON batch operations.
type JSONProjectItem struct {
	item JSONItem
	err  error
}

// newJSONProjectItem creates a new JSONProjectItem for create operations.
func newJSONProjectItem() *JSONProjectItem {
	return &JSONProjectItem{
		item: JSONItem{
			Type:       JSONItemTypeProject,
			Operation:  JSONOperationCreate,
			Attributes: make(map[string]any),
		},
	}
}

// newJSONProjectItemUpdate creates a new JSONProjectItem for update operations.
func newJSONProjectItemUpdate(id string) *JSONProjectItem {
	return &JSONProjectItem{
		item: JSONItem{
			Type:       JSONItemTypeProject,
			Operation:  JSONOperationUpdate,
			ID:         id,
			Attributes: make(map[string]any),
		},
	}
}

// Title sets the project title.
func (p *JSONProjectItem) Title(title string) *JSONProjectItem {
	if len(title) > maxTitleLength {
		p.err = ErrTitleTooLong
		return p
	}
	p.item.Attributes["title"] = title
	return p
}

// Notes sets the project notes.
func (p *JSONProjectItem) Notes(notes string) *JSONProjectItem {
	if len(notes) > maxNotesLength {
		p.err = ErrNotesTooLong
		return p
	}
	p.item.Attributes["notes"] = notes
	return p
}

// When sets the scheduling date.
func (p *JSONProjectItem) When(when When) *JSONProjectItem {
	p.item.Attributes["when"] = string(when)
	return p
}

// WhenDate sets a specific date for scheduling.
func (p *JSONProjectItem) WhenDate(year int, month time.Month, day int) *JSONProjectItem {
	p.item.Attributes["when"] = fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
	return p
}

// Deadline sets the deadline date.
func (p *JSONProjectItem) Deadline(date string) *JSONProjectItem {
	p.item.Attributes["deadline"] = date
	return p
}

// Tags sets the tags for the project.
func (p *JSONProjectItem) Tags(tags ...string) *JSONProjectItem {
	p.item.Attributes["tags"] = tags
	return p
}

// Area sets the parent area by name.
func (p *JSONProjectItem) Area(name string) *JSONProjectItem {
	p.item.Attributes["area"] = name
	return p
}

// AreaID sets the parent area by UUID.
func (p *JSONProjectItem) AreaID(id string) *JSONProjectItem {
	p.item.Attributes["area-id"] = id
	return p
}

// Completed sets the completion status.
func (p *JSONProjectItem) Completed(completed bool) *JSONProjectItem {
	p.item.Attributes["completed"] = completed
	return p
}

// Canceled sets the canceled status.
func (p *JSONProjectItem) Canceled(canceled bool) *JSONProjectItem {
	p.item.Attributes["canceled"] = canceled
	return p
}

// CreationDate sets the creation timestamp.
func (p *JSONProjectItem) CreationDate(date time.Time) *JSONProjectItem {
	p.item.Attributes["creation-date"] = date.Format(time.RFC3339)
	return p
}

// CompletionDate sets the completion timestamp.
func (p *JSONProjectItem) CompletionDate(date time.Time) *JSONProjectItem {
	p.item.Attributes["completion-date"] = date.Format(time.RFC3339)
	return p
}

// PrependNotes prepends text to existing notes (update only).
func (p *JSONProjectItem) PrependNotes(notes string) *JSONProjectItem {
	p.item.Attributes["prepend-notes"] = notes
	return p
}

// AppendNotes appends text to existing notes (update only).
func (p *JSONProjectItem) AppendNotes(notes string) *JSONProjectItem {
	p.item.Attributes["append-notes"] = notes
	return p
}

// AddTags adds tags without replacing existing ones (update only).
func (p *JSONProjectItem) AddTags(tags ...string) *JSONProjectItem {
	p.item.Attributes["add-tags"] = tags
	return p
}

// Todos sets the child to-do items.
func (p *JSONProjectItem) Todos(items ...*JSONTodoItem) *JSONProjectItem {
	todos := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if item.err != nil {
			p.err = item.err
			return p
		}
		todos = append(todos, map[string]any{
			"type":       "to-do",
			"attributes": item.item.Attributes,
		})
	}
	p.item.Attributes["items"] = todos
	return p
}

// build returns the JSON item and any error.
func (p *JSONProjectItem) build() (JSONItem, error) {
	return p.item, p.err
}

// JSONBuilder builds URLs for batch create operations via the json command.
// Does not support update operations; use AuthJSONBuilder for updates.
type JSONBuilder struct {
	items  []JSONItem
	reveal bool
	err    error
}

// AddTodo adds a to-do creation to the batch.
func (b *JSONBuilder) AddTodo(configure func(*JSONTodoItem)) *JSONBuilder {
	item := newJSONTodoItem()
	configure(item)
	built, err := item.build()
	if err != nil {
		b.err = err
		return b
	}
	b.items = append(b.items, built)
	return b
}

// AddProject adds a project creation to the batch.
func (b *JSONBuilder) AddProject(configure func(*JSONProjectItem)) *JSONBuilder {
	item := newJSONProjectItem()
	configure(item)
	built, err := item.build()
	if err != nil {
		b.err = err
		return b
	}
	b.items = append(b.items, built)
	return b
}

// Reveal navigates to the first created item after processing.
func (b *JSONBuilder) Reveal(reveal bool) *JSONBuilder {
	b.reveal = reveal
	return b
}

// Build returns the Things URL for the JSON batch operation.
func (b *JSONBuilder) Build() (string, error) {
	if b.err != nil {
		return "", b.err
	}
	if len(b.items) == 0 {
		return "", ErrNoJSONItems
	}

	data := make([]map[string]any, len(b.items))
	for i, item := range b.items {
		data[i] = map[string]any{
			"type":       string(item.Type),
			"attributes": item.Attributes,
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("things3: failed to marshal JSON: %w", err)
	}

	query := url.Values{}
	query.Set("data", string(jsonData))
	if b.reveal {
		query.Set("reveal", "true")
	}

	return fmt.Sprintf("things:///%s?%s", CommandJSON, query.Encode()), nil
}

// AuthJSONBuilder builds URLs for batch operations including updates via the json command.
// Requires authentication token for update operations.
type AuthJSONBuilder struct {
	token  string
	items  []JSONItem
	reveal bool
	err    error
}

// AddTodo adds a to-do creation to the batch.
func (b *AuthJSONBuilder) AddTodo(configure func(*JSONTodoItem)) *AuthJSONBuilder {
	item := newJSONTodoItem()
	configure(item)
	built, err := item.build()
	if err != nil {
		b.err = err
		return b
	}
	b.items = append(b.items, built)
	return b
}

// AddProject adds a project creation to the batch.
func (b *AuthJSONBuilder) AddProject(configure func(*JSONProjectItem)) *AuthJSONBuilder {
	item := newJSONProjectItem()
	configure(item)
	built, err := item.build()
	if err != nil {
		b.err = err
		return b
	}
	b.items = append(b.items, built)
	return b
}

// UpdateTodo adds a to-do update to the batch.
func (b *AuthJSONBuilder) UpdateTodo(id string, configure func(*JSONTodoItem)) *AuthJSONBuilder {
	item := newJSONTodoItemUpdate(id)
	configure(item)
	built, err := item.build()
	if err != nil {
		b.err = err
		return b
	}
	b.items = append(b.items, built)
	return b
}

// UpdateProject adds a project update to the batch.
func (b *AuthJSONBuilder) UpdateProject(id string, configure func(*JSONProjectItem)) *AuthJSONBuilder {
	item := newJSONProjectItemUpdate(id)
	configure(item)
	built, err := item.build()
	if err != nil {
		b.err = err
		return b
	}
	b.items = append(b.items, built)
	return b
}

// Reveal navigates to the first item after processing.
func (b *AuthJSONBuilder) Reveal(reveal bool) *AuthJSONBuilder {
	b.reveal = reveal
	return b
}

// Build returns the Things URL for the JSON batch operation.
func (b *AuthJSONBuilder) Build() (string, error) {
	if b.err != nil {
		return "", b.err
	}
	if b.token == "" {
		return "", ErrEmptyToken
	}
	if len(b.items) == 0 {
		return "", ErrNoJSONItems
	}

	// Check if any items are updates
	hasUpdates := false
	for _, item := range b.items {
		if item.Operation == JSONOperationUpdate {
			hasUpdates = true
			break
		}
	}

	data := make([]map[string]any, len(b.items))
	for i, item := range b.items {
		entry := map[string]any{
			"type":       string(item.Type),
			"attributes": item.Attributes,
		}
		if item.Operation == JSONOperationUpdate {
			entry["operation"] = "update"
			entry["id"] = item.ID
		}
		data[i] = entry
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("things3: failed to marshal JSON: %w", err)
	}

	query := url.Values{}
	query.Set("data", string(jsonData))
	if hasUpdates {
		query.Set("auth-token", b.token)
	}
	if b.reveal {
		query.Set("reveal", "true")
	}

	return fmt.Sprintf("things:///%s?%s", CommandJSON, query.Encode()), nil
}

// NewTodo creates a new JSONTodoItem for use with JSONBuilder.AddTodo.
// This is a convenience function for inline configuration.
func NewTodo() *JSONTodoItem {
	return newJSONTodoItem()
}

// NewProject creates a new JSONProjectItem for use with JSONBuilder.AddProject.
// This is a convenience function for inline configuration.
func NewProject() *JSONProjectItem {
	return newJSONProjectItem()
}

// Headings creates heading entries for a project's items.
// Used within JSONProjectItem.Todos to organize to-dos under headings.
func Headings(headings ...string) string {
	return strings.Join(headings, "\n")
}
