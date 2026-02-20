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

// batchTodoBuilder builds a to-do entry for batch operations.
// Unlike addTodoBuilder which generates a complete URL, batchTodoBuilder creates
// a JSON object that becomes part of a batchBuilder or authBatchBuilder batch.
//
// Example:
//
//	Scheme.Batch().
//	    AddTodo(func(b BatchTodoConfigurator) {
//	        b.Title("Buy milk").Tags("shopping")
//	    }).
//	    Build()
type batchTodoBuilder struct {
	item      JSONItem
	jsonAttrs jsonAttrs
	err       error
}

// getStore returns the attribute store for the builder.
func (t *batchTodoBuilder) getStore() attrStore { return &t.jsonAttrs }

// setErr sets the error field for the builder.
func (t *batchTodoBuilder) setErr(err error) { t.err = err }

// newBatchTodoBuilder creates a new batchTodoBuilder for create operations.
func newBatchTodoBuilder() *batchTodoBuilder {
	attrs := make(map[string]any)
	return &batchTodoBuilder{
		item: JSONItem{
			Type:       JSONItemTypeTodo,
			Operation:  JSONOperationCreate,
			Attributes: attrs,
		},
		jsonAttrs: jsonAttrs{attrs: attrs},
	}
}

// newBatchTodoBuilderUpdate creates a new batchTodoBuilder for update operations.
func newBatchTodoBuilderUpdate(id string) *batchTodoBuilder {
	attrs := make(map[string]any)
	return &batchTodoBuilder{
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
func (t *batchTodoBuilder) Title(title string) BatchTodoConfigurator {
	return setStr(t, titleParam, title)
}

// Notes sets the to-do notes.
func (t *batchTodoBuilder) Notes(notes string) BatchTodoConfigurator {
	return setStr(t, notesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (t *batchTodoBuilder) When(tm time.Time) BatchTodoConfigurator {
	return setWhenTime(t, tm)
}

// WhenEvening schedules the to-do for this evening.
func (t *batchTodoBuilder) WhenEvening() BatchTodoConfigurator {
	return setWhenStr(t, whenEvening)
}

// WhenAnytime schedules the to-do for anytime (no specific time).
func (t *batchTodoBuilder) WhenAnytime() BatchTodoConfigurator {
	return setWhenStr(t, whenAnytime)
}

// WhenSomeday schedules the to-do for someday (indefinite future).
func (t *batchTodoBuilder) WhenSomeday() BatchTodoConfigurator {
	return setWhenStr(t, whenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (t *batchTodoBuilder) Deadline(tm time.Time) BatchTodoConfigurator {
	return setDeadlineTime(t, tm)
}

// Tags sets the tags for the to-do.
func (t *batchTodoBuilder) Tags(tags ...string) BatchTodoConfigurator {
	return setStrs(t, tagsParam, tags)
}

// ChecklistItems sets the checklist items.
func (t *batchTodoBuilder) ChecklistItems(items ...string) BatchTodoConfigurator {
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
func (t *batchTodoBuilder) List(name string) BatchTodoConfigurator {
	return setStr(t, listParam, name)
}

// ListID sets the target project or area by UUID.
func (t *batchTodoBuilder) ListID(id string) BatchTodoConfigurator {
	return setStr(t, listIDParam, id)
}

// Heading sets the target heading within a project by name.
func (t *batchTodoBuilder) Heading(name string) BatchTodoConfigurator {
	return setStr(t, headingParam, name)
}

// Completed sets the completion status.
func (t *batchTodoBuilder) Completed(completed bool) BatchTodoConfigurator {
	return setBool(t, completedParam, completed)
}

// Canceled sets the canceled status.
func (t *batchTodoBuilder) Canceled(canceled bool) BatchTodoConfigurator {
	return setBool(t, canceledParam, canceled)
}

// CreationDate sets the creation timestamp.
func (t *batchTodoBuilder) CreationDate(date time.Time) BatchTodoConfigurator {
	return setTime(t, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (t *batchTodoBuilder) CompletionDate(date time.Time) BatchTodoConfigurator {
	return setTime(t, completionDateParam, date)
}

// PrependNotes prepends text to existing notes (update only).
func (t *batchTodoBuilder) PrependNotes(notes string) BatchTodoConfigurator {
	return setStr(t, prependNotesParam, notes)
}

// AppendNotes appends text to existing notes (update only).
func (t *batchTodoBuilder) AppendNotes(notes string) BatchTodoConfigurator {
	return setStr(t, appendNotesParam, notes)
}

// AddTags adds tags without replacing existing ones (update only).
func (t *batchTodoBuilder) AddTags(tags ...string) BatchTodoConfigurator {
	return setStrs(t, addTagsParam, tags)
}

// build returns the JSON item and any error.
func (t *batchTodoBuilder) build() (JSONItem, error) {
	return t.item, t.err
}

// batchProjectBuilder builds a project entry for batch operations.
// Unlike addProjectBuilder which generates a complete URL, batchProjectBuilder creates
// a JSON object that becomes part of a batchBuilder or authBatchBuilder batch.
type batchProjectBuilder struct {
	item      JSONItem
	jsonAttrs jsonAttrs
	err       error
}

// getStore returns the attribute store for the builder.
func (p *batchProjectBuilder) getStore() attrStore { return &p.jsonAttrs }

// setErr sets the error field for the builder.
func (p *batchProjectBuilder) setErr(err error) { p.err = err }

// newBatchProjectBuilder creates a new batchProjectBuilder for create operations.
func newBatchProjectBuilder() *batchProjectBuilder {
	attrs := make(map[string]any)
	return &batchProjectBuilder{
		item: JSONItem{
			Type:       JSONItemTypeProject,
			Operation:  JSONOperationCreate,
			Attributes: attrs,
		},
		jsonAttrs: jsonAttrs{attrs: attrs},
	}
}

// newBatchProjectBuilderUpdate creates a new batchProjectBuilder for update operations.
func newBatchProjectBuilderUpdate(id string) *batchProjectBuilder {
	attrs := make(map[string]any)
	return &batchProjectBuilder{
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
func (p *batchProjectBuilder) Title(title string) BatchProjectConfigurator {
	return setStr(p, titleParam, title)
}

// Notes sets the project notes.
func (p *batchProjectBuilder) Notes(notes string) BatchProjectConfigurator {
	return setStr(p, notesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (p *batchProjectBuilder) When(t time.Time) BatchProjectConfigurator {
	return setWhenTime(p, t)
}

// WhenEvening schedules the project for this evening.
func (p *batchProjectBuilder) WhenEvening() BatchProjectConfigurator {
	return setWhenStr(p, whenEvening)
}

// WhenAnytime schedules the project for anytime (no specific time).
func (p *batchProjectBuilder) WhenAnytime() BatchProjectConfigurator {
	return setWhenStr(p, whenAnytime)
}

// WhenSomeday schedules the project for someday (indefinite future).
func (p *batchProjectBuilder) WhenSomeday() BatchProjectConfigurator {
	return setWhenStr(p, whenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (p *batchProjectBuilder) Deadline(t time.Time) BatchProjectConfigurator {
	return setDeadlineTime(p, t)
}

// Tags sets the tags for the project.
func (p *batchProjectBuilder) Tags(tags ...string) BatchProjectConfigurator {
	return setStrs(p, tagsParam, tags)
}

// Area sets the parent area by name.
func (p *batchProjectBuilder) Area(name string) BatchProjectConfigurator {
	return setStr(p, areaParam, name)
}

// AreaID sets the parent area by UUID.
func (p *batchProjectBuilder) AreaID(id string) BatchProjectConfigurator {
	return setStr(p, areaIDParam, id)
}

// Completed sets the completion status.
func (p *batchProjectBuilder) Completed(completed bool) BatchProjectConfigurator {
	return setBool(p, completedParam, completed)
}

// Canceled sets the canceled status.
func (p *batchProjectBuilder) Canceled(canceled bool) BatchProjectConfigurator {
	return setBool(p, canceledParam, canceled)
}

// CreationDate sets the creation timestamp.
func (p *batchProjectBuilder) CreationDate(date time.Time) BatchProjectConfigurator {
	return setTime(p, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (p *batchProjectBuilder) CompletionDate(date time.Time) BatchProjectConfigurator {
	return setTime(p, completionDateParam, date)
}

// PrependNotes prepends text to existing notes (update only).
func (p *batchProjectBuilder) PrependNotes(notes string) BatchProjectConfigurator {
	return setStr(p, prependNotesParam, notes)
}

// AppendNotes appends text to existing notes (update only).
func (p *batchProjectBuilder) AppendNotes(notes string) BatchProjectConfigurator {
	return setStr(p, appendNotesParam, notes)
}

// AddTags adds tags without replacing existing ones (update only).
func (p *batchProjectBuilder) AddTags(tags ...string) BatchProjectConfigurator {
	return setStrs(p, addTagsParam, tags)
}

// Todos sets the child to-do items using configuration functions.
func (p *batchProjectBuilder) Todos(configs ...func(BatchTodoConfigurator)) BatchProjectConfigurator {
	todos := make([]map[string]any, 0, len(configs))
	for _, configure := range configs {
		item := newBatchTodoBuilder()
		configure(item)
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
func (p *batchProjectBuilder) build() (JSONItem, error) {
	return p.item, p.err
}

// batchBuilder builds URLs for batch create operations via the json command.
// Does not support update operations; use authBatchBuilder for updates.
type batchBuilder struct {
	scheme *scheme
	items  []JSONItem
	reveal bool
	err    error
}

// AddTodo adds a to-do creation to the batch.
func (b *batchBuilder) AddTodo(configure func(BatchTodoConfigurator)) BatchCreator {
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
func (b *batchBuilder) AddProject(configure func(BatchProjectConfigurator)) BatchCreator {
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
func (b *batchBuilder) Reveal(reveal bool) BatchCreator {
	b.reveal = reveal
	return b
}

// Build returns the Things URL for the JSON batch operation.
func (b *batchBuilder) Build() (string, error) {
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
func (b *batchBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.execute(ctx, uri)
}

// authBatchBuilder builds URLs for batch operations including updates via the json command.
// Requires authentication token for update operations.
type authBatchBuilder struct {
	scheme    *scheme
	token     string
	tokenFunc func(context.Context) (string, error) // Optional lazy token loader
	items     []JSONItem
	reveal    bool
	err       error
}

// AddTodo adds a to-do creation to the batch.
func (b *authBatchBuilder) AddTodo(configure func(BatchTodoConfigurator)) AuthBatchCreator {
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
func (b *authBatchBuilder) AddProject(configure func(BatchProjectConfigurator)) AuthBatchCreator {
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
func (b *authBatchBuilder) UpdateTodo(id string, configure func(BatchTodoConfigurator)) AuthBatchCreator {
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
func (b *authBatchBuilder) UpdateProject(id string, configure func(BatchProjectConfigurator)) AuthBatchCreator {
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
func (b *authBatchBuilder) Reveal(reveal bool) AuthBatchCreator {
	b.reveal = reveal
	return b
}

// Build returns the Things URL for the JSON batch operation.
// If token is not set but tokenFunc is provided, it will fetch the token using context.Background().
// For explicit context control, use Execute() instead.
func (b *authBatchBuilder) Build() (string, error) {
	if b.err != nil {
		return "", b.err
	}
	// Lazy load token if needed
	if b.token == "" && b.tokenFunc != nil {
		t, err := b.tokenFunc(context.Background())
		if err != nil {
			return "", err
		}
		b.token = t
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
// If token is not set but tokenFunc is provided, it will fetch the token first.
func (b *authBatchBuilder) Execute(ctx context.Context) error {
	// Lazy load token if needed
	if b.token == "" && b.tokenFunc != nil {
		token, err := b.tokenFunc(ctx)
		if err != nil {
			return err
		}
		b.token = token
	}
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.execute(ctx, uri)
}

// Headings creates heading entries for a project's items.
// Used within batchProjectBuilder.Todos to organize to-dos under headings.
func Headings(headings ...string) string {
	return strings.Join(headings, "\n")
}
