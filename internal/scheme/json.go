package scheme

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// batchTodoBuilder builds a todo entry for batch operations.
type batchTodoBuilder struct {
	item      JSONItem
	jsonAttrs JSONAttrs
	err       error
}

// GetStore returns the attribute store for the builder.
func (t *batchTodoBuilder) GetStore() AttrStore { return &t.jsonAttrs }

// SetErr sets the error field for the builder.
func (t *batchTodoBuilder) SetErr(err error) { t.err = err }

// newBatchTodoBuilder creates a new batchTodoBuilder for create operations.
func newBatchTodoBuilder() *batchTodoBuilder {
	attrs := make(map[string]any)
	return &batchTodoBuilder{
		item: JSONItem{
			Type:       JSONItemTypeTodo,
			Operation:  JSONOperationCreate,
			Attributes: attrs,
		},
		jsonAttrs: JSONAttrs{Attrs: attrs},
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
		jsonAttrs: JSONAttrs{Attrs: attrs},
	}
}

// Title sets the todo title.
func (t *batchTodoBuilder) Title(title string) BatchTodoConfigurator {
	return SetStr(t, TitleParam, title)
}

// Notes sets the todo notes.
func (t *batchTodoBuilder) Notes(notes string) BatchTodoConfigurator {
	return SetStr(t, NotesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (t *batchTodoBuilder) When(tm time.Time) BatchTodoConfigurator {
	return SetWhenTime(t, tm)
}

// WhenEvening schedules the todo for this evening.
func (t *batchTodoBuilder) WhenEvening() BatchTodoConfigurator {
	return SetWhenStr(t, WhenEvening)
}

// WhenAnytime schedules the todo for anytime (no specific time).
func (t *batchTodoBuilder) WhenAnytime() BatchTodoConfigurator {
	return SetWhenStr(t, WhenAnytime)
}

// WhenSomeday schedules the todo for someday (indefinite future).
func (t *batchTodoBuilder) WhenSomeday() BatchTodoConfigurator {
	return SetWhenStr(t, WhenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (t *batchTodoBuilder) Deadline(tm time.Time) BatchTodoConfigurator {
	return SetDeadlineTime(t, tm)
}

// Tags sets the tags for the todo.
func (t *batchTodoBuilder) Tags(tags ...string) BatchTodoConfigurator {
	return SetStrs(t, TagsParam, tags)
}

// ChecklistItems sets the checklist items.
func (t *batchTodoBuilder) ChecklistItems(items ...string) BatchTodoConfigurator {
	if len(items) > MaxChecklistItems {
		t.err = ErrTooManyChecklistItems
		return t
	}
	checklistItems := make([]map[string]any, len(items))
	for i, item := range items {
		checklistItems[i] = map[string]any{
			KeyType:       "checklist-item",
			KeyAttributes: map[string]any{KeyTitle: item},
		}
	}
	t.item.Attributes[KeyChecklistItems] = checklistItems
	return t
}

// List sets the target project or area by name.
func (t *batchTodoBuilder) List(name string) BatchTodoConfigurator {
	return SetStr(t, ListParam, name)
}

// ListID sets the target project or area by UUID.
func (t *batchTodoBuilder) ListID(id string) BatchTodoConfigurator {
	return SetStr(t, ListIDParam, id)
}

// Heading sets the target heading within a project by name.
func (t *batchTodoBuilder) Heading(name string) BatchTodoConfigurator {
	return SetStr(t, HeadingParam, name)
}

// Completed sets the completion status.
func (t *batchTodoBuilder) Completed(completed bool) BatchTodoConfigurator {
	return SetBool(t, CompletedParam, completed)
}

// Canceled sets the canceled status.
func (t *batchTodoBuilder) Canceled(canceled bool) BatchTodoConfigurator {
	return SetBool(t, CanceledParam, canceled)
}

// CreationDate sets the creation timestamp.
func (t *batchTodoBuilder) CreationDate(date time.Time) BatchTodoConfigurator {
	return SetTime(t, CreationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (t *batchTodoBuilder) CompletionDate(date time.Time) BatchTodoConfigurator {
	return SetTime(t, CompletionDateParam, date)
}

// PrependNotes prepends text to existing notes (update only).
func (t *batchTodoBuilder) PrependNotes(notes string) BatchTodoConfigurator {
	return SetStr(t, PrependNotesParam, notes)
}

// AppendNotes appends text to existing notes (update only).
func (t *batchTodoBuilder) AppendNotes(notes string) BatchTodoConfigurator {
	return SetStr(t, AppendNotesParam, notes)
}

// AddTags adds tags without replacing existing ones (update only).
func (t *batchTodoBuilder) AddTags(tags ...string) BatchTodoConfigurator {
	return SetStrs(t, AddTagsParam, tags)
}

// build returns the JSON item and any error.
func (t *batchTodoBuilder) build() (JSONItem, error) {
	return t.item, t.err
}

// batchProjectBuilder builds a project entry for batch operations.
type batchProjectBuilder struct {
	item      JSONItem
	jsonAttrs JSONAttrs
	err       error
}

// GetStore returns the attribute store for the builder.
func (p *batchProjectBuilder) GetStore() AttrStore { return &p.jsonAttrs }

// SetErr sets the error field for the builder.
func (p *batchProjectBuilder) SetErr(err error) { p.err = err }

// newBatchProjectBuilder creates a new batchProjectBuilder for create operations.
func newBatchProjectBuilder() *batchProjectBuilder {
	attrs := make(map[string]any)
	return &batchProjectBuilder{
		item: JSONItem{
			Type:       JSONItemTypeProject,
			Operation:  JSONOperationCreate,
			Attributes: attrs,
		},
		jsonAttrs: JSONAttrs{Attrs: attrs},
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
		jsonAttrs: JSONAttrs{Attrs: attrs},
	}
}

// Title sets the project title.
func (p *batchProjectBuilder) Title(title string) BatchProjectConfigurator {
	return SetStr(p, TitleParam, title)
}

// Notes sets the project notes.
func (p *batchProjectBuilder) Notes(notes string) BatchProjectConfigurator {
	return SetStr(p, NotesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (p *batchProjectBuilder) When(t time.Time) BatchProjectConfigurator {
	return SetWhenTime(p, t)
}

// WhenEvening schedules the project for this evening.
func (p *batchProjectBuilder) WhenEvening() BatchProjectConfigurator {
	return SetWhenStr(p, WhenEvening)
}

// WhenAnytime schedules the project for anytime (no specific time).
func (p *batchProjectBuilder) WhenAnytime() BatchProjectConfigurator {
	return SetWhenStr(p, WhenAnytime)
}

// WhenSomeday schedules the project for someday (indefinite future).
func (p *batchProjectBuilder) WhenSomeday() BatchProjectConfigurator {
	return SetWhenStr(p, WhenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (p *batchProjectBuilder) Deadline(t time.Time) BatchProjectConfigurator {
	return SetDeadlineTime(p, t)
}

// Tags sets the tags for the project.
func (p *batchProjectBuilder) Tags(tags ...string) BatchProjectConfigurator {
	return SetStrs(p, TagsParam, tags)
}

// Area sets the parent area by name.
func (p *batchProjectBuilder) Area(name string) BatchProjectConfigurator {
	return SetStr(p, AreaParam, name)
}

// AreaID sets the parent area by UUID.
func (p *batchProjectBuilder) AreaID(id string) BatchProjectConfigurator {
	return SetStr(p, AreaIDParam, id)
}

// Completed sets the completion status.
func (p *batchProjectBuilder) Completed(completed bool) BatchProjectConfigurator {
	return SetBool(p, CompletedParam, completed)
}

// Canceled sets the canceled status.
func (p *batchProjectBuilder) Canceled(canceled bool) BatchProjectConfigurator {
	return SetBool(p, CanceledParam, canceled)
}

// CreationDate sets the creation timestamp.
func (p *batchProjectBuilder) CreationDate(date time.Time) BatchProjectConfigurator {
	return SetTime(p, CreationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (p *batchProjectBuilder) CompletionDate(date time.Time) BatchProjectConfigurator {
	return SetTime(p, CompletionDateParam, date)
}

// PrependNotes prepends text to existing notes (update only).
func (p *batchProjectBuilder) PrependNotes(notes string) BatchProjectConfigurator {
	return SetStr(p, PrependNotesParam, notes)
}

// AppendNotes appends text to existing notes (update only).
func (p *batchProjectBuilder) AppendNotes(notes string) BatchProjectConfigurator {
	return SetStr(p, AppendNotesParam, notes)
}

// AddTags adds tags without replacing existing ones (update only).
func (p *batchProjectBuilder) AddTags(tags ...string) BatchProjectConfigurator {
	return SetStrs(p, AddTagsParam, tags)
}

// Todos sets the child todo items using configuration functions.
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
			KeyType:       "to-do",
			KeyAttributes: item.item.Attributes,
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
	scheme *Scheme
	items  []JSONItem
	reveal bool
	err    error
}

// NewBatch creates a new BatchCreator for batch create operations.
func NewBatch(s *Scheme) BatchCreator {
	return &batchBuilder{scheme: s, items: make([]JSONItem, 0)}
}

// AddTodo adds a todo creation to the batch.
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
			KeyType:       string(item.Type),
			KeyAttributes: item.Attributes,
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("things3: failed to marshal JSON: %w", err)
	}

	query := url.Values{}
	query.Set(KeyData, string(jsonData))
	if b.reveal {
		query.Set(KeyReveal, "true")
	}

	return fmt.Sprintf("things:///%s?%s", CommandJSON, EncodeQuery(query)), nil
}

// Execute builds and executes the JSON batch URL.
// Returns an error if the URL cannot be built or executed.
func (b *batchBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.Execute(ctx, uri)
}

// authBatchBuilder builds URLs for batch operations including updates via the json command.
// Requires authentication token for update operations.
type authBatchBuilder struct {
	scheme    *Scheme
	token     string
	tokenFunc func(context.Context) (string, error) // Optional lazy token loader
	items     []JSONItem
	reveal    bool
	err       error
}

// NewAuthBatch creates a new AuthBatchCreator for batch operations including updates.
func NewAuthBatch(s *Scheme, tokenFunc func(context.Context) (string, error)) AuthBatchCreator {
	return &authBatchBuilder{
		scheme:    s,
		tokenFunc: tokenFunc,
		items:     make([]JSONItem, 0),
	}
}

// AddTodo adds a todo creation to the batch.
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

// UpdateTodo adds a todo update to the batch.
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

// hasUpdates reports whether any item in the batch is an update operation.
func (b *authBatchBuilder) hasUpdates() bool {
	for _, item := range b.items {
		if item.Operation == JSONOperationUpdate {
			return true
		}
	}
	return false
}

// build resolves the auth token (only when the batch contains updates)
// and constructs the JSON batch URL.
func (b *authBatchBuilder) build(ctx context.Context) (string, error) {
	if b.err != nil {
		return "", b.err
	}
	if len(b.items) == 0 {
		return "", ErrNoJSONItems
	}

	// The auth-token parameter is only required for update operations,
	// so create-only batches never need a token.
	hasUpdates := b.hasUpdates()
	if hasUpdates {
		if err := resolveToken(ctx, &b.token, b.tokenFunc); err != nil {
			return "", err
		}
	}

	data := make([]map[string]any, len(b.items))
	for i, item := range b.items {
		entry := map[string]any{
			KeyType:       string(item.Type),
			KeyAttributes: item.Attributes,
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
	query.Set(KeyData, string(jsonData))
	if hasUpdates {
		query.Set(KeyAuthToken, b.token)
	}
	if b.reveal {
		query.Set(KeyReveal, "true")
	}

	return fmt.Sprintf("things:///%s?%s", CommandJSON, EncodeQuery(query)), nil
}

// Build returns the Things URL for the JSON batch operation.
// A token is required only when the batch contains update operations.
// If the token has not been resolved yet, the token function runs without a
// caller context; use Execute for context-aware token loading.
func (b *authBatchBuilder) Build() (string, error) {
	return b.build(context.Background())
}

// Execute builds and executes the JSON batch URL.
// Returns an error if the URL cannot be built or executed.
// The auth token is resolved at most once, using the provided context,
// and only when the batch contains update operations.
func (b *authBatchBuilder) Execute(ctx context.Context) error {
	uri, err := b.build(ctx)
	if err != nil {
		return err
	}
	return b.scheme.Execute(ctx, uri)
}
