package things3

import (
	"context"
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

// BatchTodoBuilder builds a to-do entry for batch operations.
// Unlike AddTodoBuilder which generates a complete URL, BatchTodoBuilder creates
// a JSON object that becomes part of a BatchBuilder or AuthBatchBuilder batch.
//
// Example:
//
//	Scheme.Batch().
//	    AddTodo(func(b *BatchTodoBuilder) {
//	        b.Title("Buy milk").Tags("shopping")
//	    }).
//	    Build()
type BatchTodoBuilder struct {
	item      JSONItem
	jsonAttrs jsonAttrs
	err       error
}

// getStore returns the attribute store for the builder.
func (t *BatchTodoBuilder) getStore() attrStore { return &t.jsonAttrs }

// setErr sets the error field for the builder.
func (t *BatchTodoBuilder) setErr(err error) { t.err = err }

// newBatchTodoBuilder creates a new BatchTodoBuilder for create operations.
func newBatchTodoBuilder() *BatchTodoBuilder {
	attrs := make(map[string]any)
	return &BatchTodoBuilder{
		item: JSONItem{
			Type:       JSONItemTypeTodo,
			Operation:  JSONOperationCreate,
			Attributes: attrs,
		},
		jsonAttrs: jsonAttrs{attrs: attrs},
	}
}

// newBatchTodoBuilderUpdate creates a new BatchTodoBuilder for update operations.
func newBatchTodoBuilderUpdate(id string) *BatchTodoBuilder {
	attrs := make(map[string]any)
	return &BatchTodoBuilder{
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
func (t *BatchTodoBuilder) Title(title string) *BatchTodoBuilder {
	return setStr(t, titleParam, title)
}

// Notes sets the to-do notes.
func (t *BatchTodoBuilder) Notes(notes string) *BatchTodoBuilder {
	return setStr(t, notesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (t *BatchTodoBuilder) When(tm time.Time) *BatchTodoBuilder {
	return setWhenTime(t, tm)
}

// WhenEvening schedules the to-do for this evening.
func (t *BatchTodoBuilder) WhenEvening() *BatchTodoBuilder {
	return setWhenStr(t, whenEvening)
}

// WhenAnytime schedules the to-do for anytime (no specific time).
func (t *BatchTodoBuilder) WhenAnytime() *BatchTodoBuilder {
	return setWhenStr(t, whenAnytime)
}

// WhenSomeday schedules the to-do for someday (indefinite future).
func (t *BatchTodoBuilder) WhenSomeday() *BatchTodoBuilder {
	return setWhenStr(t, whenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (t *BatchTodoBuilder) Deadline(tm time.Time) *BatchTodoBuilder {
	return setDeadlineTime(t, tm)
}

// Tags sets the tags for the to-do.
func (t *BatchTodoBuilder) Tags(tags ...string) *BatchTodoBuilder {
	return setStrs(t, tagsParam, tags)
}

// ChecklistItems sets the checklist items.
func (t *BatchTodoBuilder) ChecklistItems(items ...string) *BatchTodoBuilder {
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
func (t *BatchTodoBuilder) List(name string) *BatchTodoBuilder {
	return setStr(t, listParam, name)
}

// ListID sets the target project or area by UUID.
func (t *BatchTodoBuilder) ListID(id string) *BatchTodoBuilder {
	return setStr(t, listIDParam, id)
}

// Heading sets the target heading within a project by name.
func (t *BatchTodoBuilder) Heading(name string) *BatchTodoBuilder {
	return setStr(t, headingParam, name)
}

// Completed sets the completion status.
func (t *BatchTodoBuilder) Completed(completed bool) *BatchTodoBuilder {
	return setBool(t, completedParam, completed)
}

// Canceled sets the canceled status.
func (t *BatchTodoBuilder) Canceled(canceled bool) *BatchTodoBuilder {
	return setBool(t, canceledParam, canceled)
}

// CreationDate sets the creation timestamp.
func (t *BatchTodoBuilder) CreationDate(date time.Time) *BatchTodoBuilder {
	return setTime(t, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (t *BatchTodoBuilder) CompletionDate(date time.Time) *BatchTodoBuilder {
	return setTime(t, completionDateParam, date)
}

// PrependNotes prepends text to existing notes (update only).
func (t *BatchTodoBuilder) PrependNotes(notes string) *BatchTodoBuilder {
	return setStr(t, prependNotesParam, notes)
}

// AppendNotes appends text to existing notes (update only).
func (t *BatchTodoBuilder) AppendNotes(notes string) *BatchTodoBuilder {
	return setStr(t, appendNotesParam, notes)
}

// AddTags adds tags without replacing existing ones (update only).
func (t *BatchTodoBuilder) AddTags(tags ...string) *BatchTodoBuilder {
	return setStrs(t, addTagsParam, tags)
}

// build returns the JSON item and any error.
func (t *BatchTodoBuilder) build() (JSONItem, error) {
	return t.item, t.err
}

// BatchProjectBuilder builds a project entry for batch operations.
// Unlike AddProjectBuilder which generates a complete URL, BatchProjectBuilder creates
// a JSON object that becomes part of a BatchBuilder or AuthBatchBuilder batch.
type BatchProjectBuilder struct {
	item      JSONItem
	jsonAttrs jsonAttrs
	err       error
}

// getStore returns the attribute store for the builder.
func (p *BatchProjectBuilder) getStore() attrStore { return &p.jsonAttrs }

// setErr sets the error field for the builder.
func (p *BatchProjectBuilder) setErr(err error) { p.err = err }

// newBatchProjectBuilder creates a new BatchProjectBuilder for create operations.
func newBatchProjectBuilder() *BatchProjectBuilder {
	attrs := make(map[string]any)
	return &BatchProjectBuilder{
		item: JSONItem{
			Type:       JSONItemTypeProject,
			Operation:  JSONOperationCreate,
			Attributes: attrs,
		},
		jsonAttrs: jsonAttrs{attrs: attrs},
	}
}

// newBatchProjectBuilderUpdate creates a new BatchProjectBuilder for update operations.
func newBatchProjectBuilderUpdate(id string) *BatchProjectBuilder {
	attrs := make(map[string]any)
	return &BatchProjectBuilder{
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
func (p *BatchProjectBuilder) Title(title string) *BatchProjectBuilder {
	return setStr(p, titleParam, title)
}

// Notes sets the project notes.
func (p *BatchProjectBuilder) Notes(notes string) *BatchProjectBuilder {
	return setStr(p, notesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (p *BatchProjectBuilder) When(t time.Time) *BatchProjectBuilder {
	return setWhenTime(p, t)
}

// WhenEvening schedules the project for this evening.
func (p *BatchProjectBuilder) WhenEvening() *BatchProjectBuilder {
	return setWhenStr(p, whenEvening)
}

// WhenAnytime schedules the project for anytime (no specific time).
func (p *BatchProjectBuilder) WhenAnytime() *BatchProjectBuilder {
	return setWhenStr(p, whenAnytime)
}

// WhenSomeday schedules the project for someday (indefinite future).
func (p *BatchProjectBuilder) WhenSomeday() *BatchProjectBuilder {
	return setWhenStr(p, whenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (p *BatchProjectBuilder) Deadline(t time.Time) *BatchProjectBuilder {
	return setDeadlineTime(p, t)
}

// Tags sets the tags for the project.
func (p *BatchProjectBuilder) Tags(tags ...string) *BatchProjectBuilder {
	return setStrs(p, tagsParam, tags)
}

// Area sets the parent area by name.
func (p *BatchProjectBuilder) Area(name string) *BatchProjectBuilder {
	return setStr(p, areaParam, name)
}

// AreaID sets the parent area by UUID.
func (p *BatchProjectBuilder) AreaID(id string) *BatchProjectBuilder {
	return setStr(p, areaIDParam, id)
}

// Completed sets the completion status.
func (p *BatchProjectBuilder) Completed(completed bool) *BatchProjectBuilder {
	return setBool(p, completedParam, completed)
}

// Canceled sets the canceled status.
func (p *BatchProjectBuilder) Canceled(canceled bool) *BatchProjectBuilder {
	return setBool(p, canceledParam, canceled)
}

// CreationDate sets the creation timestamp.
func (p *BatchProjectBuilder) CreationDate(date time.Time) *BatchProjectBuilder {
	return setTime(p, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (p *BatchProjectBuilder) CompletionDate(date time.Time) *BatchProjectBuilder {
	return setTime(p, completionDateParam, date)
}

// PrependNotes prepends text to existing notes (update only).
func (p *BatchProjectBuilder) PrependNotes(notes string) *BatchProjectBuilder {
	return setStr(p, prependNotesParam, notes)
}

// AppendNotes appends text to existing notes (update only).
func (p *BatchProjectBuilder) AppendNotes(notes string) *BatchProjectBuilder {
	return setStr(p, appendNotesParam, notes)
}

// AddTags adds tags without replacing existing ones (update only).
func (p *BatchProjectBuilder) AddTags(tags ...string) *BatchProjectBuilder {
	return setStrs(p, addTagsParam, tags)
}

// Todos sets the child to-do items.
func (p *BatchProjectBuilder) Todos(items ...*BatchTodoBuilder) *BatchProjectBuilder {
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
func (p *BatchProjectBuilder) build() (JSONItem, error) {
	return p.item, p.err
}

// BatchBuilder builds URLs for batch create operations via the json command.
// Does not support update operations; use AuthBatchBuilder for updates.
type BatchBuilder struct {
	scheme *Scheme
	items  []JSONItem
	reveal bool
	err    error
}

// AddTodo adds a to-do creation to the batch.
func (b *BatchBuilder) AddTodo(configure func(*BatchTodoBuilder)) *BatchBuilder {
	item := newBatchTodoBuilder()
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
func (b *BatchBuilder) AddProject(configure func(*BatchProjectBuilder)) *BatchBuilder {
	item := newBatchProjectBuilder()
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
func (b *BatchBuilder) Reveal(reveal bool) *BatchBuilder {
	b.reveal = reveal
	return b
}

// Build returns the Things URL for the JSON batch operation.
func (b *BatchBuilder) Build() (string, error) {
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

// Execute builds and executes the JSON batch URL.
// Returns an error if the URL cannot be built or executed.
func (b *BatchBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.execute(ctx, uri)
}

// AuthBatchBuilder builds URLs for batch operations including updates via the json command.
// Requires authentication token for update operations.
type AuthBatchBuilder struct {
	scheme *Scheme
	token  string
	items  []JSONItem
	reveal bool
	err    error
}

// AddTodo adds a to-do creation to the batch.
func (b *AuthBatchBuilder) AddTodo(configure func(*BatchTodoBuilder)) *AuthBatchBuilder {
	item := newBatchTodoBuilder()
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
func (b *AuthBatchBuilder) AddProject(configure func(*BatchProjectBuilder)) *AuthBatchBuilder {
	item := newBatchProjectBuilder()
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
func (b *AuthBatchBuilder) UpdateTodo(id string, configure func(*BatchTodoBuilder)) *AuthBatchBuilder {
	item := newBatchTodoBuilderUpdate(id)
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
func (b *AuthBatchBuilder) UpdateProject(id string, configure func(*BatchProjectBuilder)) *AuthBatchBuilder {
	item := newBatchProjectBuilderUpdate(id)
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
func (b *AuthBatchBuilder) Reveal(reveal bool) *AuthBatchBuilder {
	b.reveal = reveal
	return b
}

// Build returns the Things URL for the JSON batch operation.
func (b *AuthBatchBuilder) Build() (string, error) {
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

// Execute builds and executes the JSON batch URL.
// Returns an error if the URL cannot be built or executed.
func (b *AuthBatchBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.execute(ctx, uri)
}

// NewTodo creates a new BatchTodoBuilder for use with JSONBuilder.AddTodo.
// This is a convenience function for inline configuration.
func NewTodo() *BatchTodoBuilder {
	return newBatchTodoBuilder()
}

// NewProject creates a new BatchProjectBuilder for use with JSONBuilder.AddProject.
// This is a convenience function for inline configuration.
func NewProject() *BatchProjectBuilder {
	return newBatchProjectBuilder()
}

// Headings creates heading entries for a project's items.
// Used within BatchProjectBuilder.Todos to organize to-dos under headings.
func Headings(headings ...string) string {
	return strings.Join(headings, "\n")
}
