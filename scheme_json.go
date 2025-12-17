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

// JSONTodoBuilder builds a to-do entry for JSON batch operations.
// Unlike TodoBuilder which generates a complete URL, JSONTodoBuilder creates
// a JSON object that becomes part of a JSONBuilder or AuthJSONBuilder batch.
//
// Example:
//
//	Scheme.JSON().
//	    AddTodo(func(b *JSONTodoBuilder) {
//	        b.Title("Buy milk").Tags("shopping")
//	    }).
//	    Build()
type JSONTodoBuilder struct {
	item      JSONItem
	jsonAttrs jsonAttrs
	err       error
}

// getStore returns the attribute store for the builder.
func (t *JSONTodoBuilder) getStore() attrStore { return &t.jsonAttrs }

// setErr sets the error field for the builder.
func (t *JSONTodoBuilder) setErr(err error) { t.err = err }

// newJSONTodoBuilder creates a new JSONTodoBuilder for create operations.
func newJSONTodoBuilder() *JSONTodoBuilder {
	attrs := make(map[string]any)
	return &JSONTodoBuilder{
		item: JSONItem{
			Type:       JSONItemTypeTodo,
			Operation:  JSONOperationCreate,
			Attributes: attrs,
		},
		jsonAttrs: jsonAttrs{attrs: attrs},
	}
}

// newJSONTodoBuilderUpdate creates a new JSONTodoBuilder for update operations.
func newJSONTodoBuilderUpdate(id string) *JSONTodoBuilder {
	attrs := make(map[string]any)
	return &JSONTodoBuilder{
		item: JSONItem{
			Type:       JSONItemTypeTodo,
			Operation:  JSONOperationUpdate,
			ID:         id,
			Attributes: attrs,
		},
		jsonAttrs: jsonAttrs{attrs: attrs},
	}
}

// Title sets the to-do title.
func (t *JSONTodoBuilder) Title(title string) *JSONTodoBuilder {
	return setStr(t, titleParam, title)
}

// Notes sets the to-do notes.
func (t *JSONTodoBuilder) Notes(notes string) *JSONTodoBuilder {
	return setStr(t, notesParam, notes)
}

// When sets the scheduling date.
func (t *JSONTodoBuilder) When(when When) *JSONTodoBuilder {
	return setWhenStr(t, when)
}

// WhenDate sets a specific date for scheduling.
func (t *JSONTodoBuilder) WhenDate(year int, month time.Month, day int) *JSONTodoBuilder {
	return setDate(t, whenParam, year, month, day)
}

// Deadline sets the deadline date.
func (t *JSONTodoBuilder) Deadline(date string) *JSONTodoBuilder {
	return setStr(t, deadlineParam, date)
}

// Tags sets the tags for the to-do.
func (t *JSONTodoBuilder) Tags(tags ...string) *JSONTodoBuilder {
	return setStrs(t, tagsParam, tags)
}

// ChecklistItems sets the checklist items.
func (t *JSONTodoBuilder) ChecklistItems(items ...string) *JSONTodoBuilder {
	if len(items) > maxChecklistItems {
		t.err = ErrTooManyChecklistItems
		return t
	}
	checklistItems := make([]map[string]any, len(items))
	for i, item := range items {
		checklistItems[i] = map[string]any{
			"type":       "checklist-item",
			"attributes": map[string]any{keyTitle: item},
		}
	}
	t.item.Attributes[keyChecklistItems] = checklistItems
	return t
}

// List sets the target project or area by name.
func (t *JSONTodoBuilder) List(name string) *JSONTodoBuilder {
	return setStr(t, listParam, name)
}

// ListID sets the target project or area by UUID.
func (t *JSONTodoBuilder) ListID(id string) *JSONTodoBuilder {
	return setStr(t, listIDParam, id)
}

// Heading sets the target heading within a project by name.
func (t *JSONTodoBuilder) Heading(name string) *JSONTodoBuilder {
	return setStr(t, headingParam, name)
}

// Completed sets the completion status.
func (t *JSONTodoBuilder) Completed(completed bool) *JSONTodoBuilder {
	return setBool(t, completedParam, completed)
}

// Canceled sets the canceled status.
func (t *JSONTodoBuilder) Canceled(canceled bool) *JSONTodoBuilder {
	return setBool(t, canceledParam, canceled)
}

// CreationDate sets the creation timestamp.
func (t *JSONTodoBuilder) CreationDate(date time.Time) *JSONTodoBuilder {
	return setTime(t, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (t *JSONTodoBuilder) CompletionDate(date time.Time) *JSONTodoBuilder {
	return setTime(t, completionDateParam, date)
}

// PrependNotes prepends text to existing notes (update only).
func (t *JSONTodoBuilder) PrependNotes(notes string) *JSONTodoBuilder {
	return setStr(t, prependNotesParam, notes)
}

// AppendNotes appends text to existing notes (update only).
func (t *JSONTodoBuilder) AppendNotes(notes string) *JSONTodoBuilder {
	return setStr(t, appendNotesParam, notes)
}

// AddTags adds tags without replacing existing ones (update only).
func (t *JSONTodoBuilder) AddTags(tags ...string) *JSONTodoBuilder {
	return setStrs(t, addTagsParam, tags)
}

// build returns the JSON item and any error.
func (t *JSONTodoBuilder) build() (JSONItem, error) {
	return t.item, t.err
}

// JSONProjectBuilder builds a project entry for JSON batch operations.
// Unlike ProjectBuilder which generates a complete URL, JSONProjectBuilder creates
// a JSON object that becomes part of a JSONBuilder or AuthJSONBuilder batch.
type JSONProjectBuilder struct {
	item      JSONItem
	jsonAttrs jsonAttrs
	err       error
}

// getStore returns the attribute store for the builder.
func (p *JSONProjectBuilder) getStore() attrStore { return &p.jsonAttrs }

// setErr sets the error field for the builder.
func (p *JSONProjectBuilder) setErr(err error) { p.err = err }

// newJSONProjectBuilder creates a new JSONProjectBuilder for create operations.
func newJSONProjectBuilder() *JSONProjectBuilder {
	attrs := make(map[string]any)
	return &JSONProjectBuilder{
		item: JSONItem{
			Type:       JSONItemTypeProject,
			Operation:  JSONOperationCreate,
			Attributes: attrs,
		},
		jsonAttrs: jsonAttrs{attrs: attrs},
	}
}

// newJSONProjectBuilderUpdate creates a new JSONProjectBuilder for update operations.
func newJSONProjectBuilderUpdate(id string) *JSONProjectBuilder {
	attrs := make(map[string]any)
	return &JSONProjectBuilder{
		item: JSONItem{
			Type:       JSONItemTypeProject,
			Operation:  JSONOperationUpdate,
			ID:         id,
			Attributes: attrs,
		},
		jsonAttrs: jsonAttrs{attrs: attrs},
	}
}

// Title sets the project title.
func (p *JSONProjectBuilder) Title(title string) *JSONProjectBuilder {
	return setStr(p, titleParam, title)
}

// Notes sets the project notes.
func (p *JSONProjectBuilder) Notes(notes string) *JSONProjectBuilder {
	return setStr(p, notesParam, notes)
}

// When sets the scheduling date.
func (p *JSONProjectBuilder) When(when When) *JSONProjectBuilder {
	return setWhenStr(p, when)
}

// WhenDate sets a specific date for scheduling.
func (p *JSONProjectBuilder) WhenDate(year int, month time.Month, day int) *JSONProjectBuilder {
	return setDate(p, whenParam, year, month, day)
}

// Deadline sets the deadline date.
func (p *JSONProjectBuilder) Deadline(date string) *JSONProjectBuilder {
	return setStr(p, deadlineParam, date)
}

// Tags sets the tags for the project.
func (p *JSONProjectBuilder) Tags(tags ...string) *JSONProjectBuilder {
	return setStrs(p, tagsParam, tags)
}

// Area sets the parent area by name.
func (p *JSONProjectBuilder) Area(name string) *JSONProjectBuilder {
	return setStr(p, areaParam, name)
}

// AreaID sets the parent area by UUID.
func (p *JSONProjectBuilder) AreaID(id string) *JSONProjectBuilder {
	return setStr(p, areaIDParam, id)
}

// Completed sets the completion status.
func (p *JSONProjectBuilder) Completed(completed bool) *JSONProjectBuilder {
	return setBool(p, completedParam, completed)
}

// Canceled sets the canceled status.
func (p *JSONProjectBuilder) Canceled(canceled bool) *JSONProjectBuilder {
	return setBool(p, canceledParam, canceled)
}

// CreationDate sets the creation timestamp.
func (p *JSONProjectBuilder) CreationDate(date time.Time) *JSONProjectBuilder {
	return setTime(p, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (p *JSONProjectBuilder) CompletionDate(date time.Time) *JSONProjectBuilder {
	return setTime(p, completionDateParam, date)
}

// PrependNotes prepends text to existing notes (update only).
func (p *JSONProjectBuilder) PrependNotes(notes string) *JSONProjectBuilder {
	return setStr(p, prependNotesParam, notes)
}

// AppendNotes appends text to existing notes (update only).
func (p *JSONProjectBuilder) AppendNotes(notes string) *JSONProjectBuilder {
	return setStr(p, appendNotesParam, notes)
}

// AddTags adds tags without replacing existing ones (update only).
func (p *JSONProjectBuilder) AddTags(tags ...string) *JSONProjectBuilder {
	return setStrs(p, addTagsParam, tags)
}

// Todos sets the child to-do items.
func (p *JSONProjectBuilder) Todos(items ...*JSONTodoBuilder) *JSONProjectBuilder {
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
func (p *JSONProjectBuilder) build() (JSONItem, error) {
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
func (b *JSONBuilder) AddTodo(configure func(*JSONTodoBuilder)) *JSONBuilder {
	item := newJSONTodoBuilder()
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
func (b *JSONBuilder) AddProject(configure func(*JSONProjectBuilder)) *JSONBuilder {
	item := newJSONProjectBuilder()
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
	query.Set(keyData, string(jsonData))
	if b.reveal {
		query.Set(keyReveal, "true")
	}

	return fmt.Sprintf("things:///%s?%s", CommandJSON, encodeQuery(query)), nil
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
func (b *AuthJSONBuilder) AddTodo(configure func(*JSONTodoBuilder)) *AuthJSONBuilder {
	item := newJSONTodoBuilder()
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
func (b *AuthJSONBuilder) AddProject(configure func(*JSONProjectBuilder)) *AuthJSONBuilder {
	item := newJSONProjectBuilder()
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
func (b *AuthJSONBuilder) UpdateTodo(id string, configure func(*JSONTodoBuilder)) *AuthJSONBuilder {
	item := newJSONTodoBuilderUpdate(id)
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
func (b *AuthJSONBuilder) UpdateProject(id string, configure func(*JSONProjectBuilder)) *AuthJSONBuilder {
	item := newJSONProjectBuilderUpdate(id)
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
	query.Set(keyData, string(jsonData))
	if hasUpdates {
		query.Set(keyAuthToken, b.token)
	}
	if b.reveal {
		query.Set(keyReveal, "true")
	}

	return fmt.Sprintf("things:///%s?%s", CommandJSON, encodeQuery(query)), nil
}

// NewTodo creates a new JSONTodoBuilder for use with JSONBuilder.AddTodo.
// This is a convenience function for inline configuration.
func NewTodo() *JSONTodoBuilder {
	return newJSONTodoBuilder()
}

// NewProject creates a new JSONProjectBuilder for use with JSONBuilder.AddProject.
// This is a convenience function for inline configuration.
func NewProject() *JSONProjectBuilder {
	return newJSONProjectBuilder()
}

// Headings creates heading entries for a project's items.
// Used within JSONProjectBuilder.Todos to organize to-dos under headings.
func Headings(headings ...string) string {
	return strings.Join(headings, "\n")
}
