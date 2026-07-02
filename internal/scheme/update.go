package scheme

import (
	"context"
	"fmt"
	"time"
)

// resolveToken ensures a non-empty auth token is cached in *token.
// A successful fetch is cached and reused by later Build calls; a failed or
// empty fetch returns an error immediately and is retried on the next Build.
func resolveToken(ctx context.Context, token *string, tokenFunc func(context.Context) (string, error)) error {
	if *token != "" {
		return nil
	}
	if tokenFunc == nil {
		return ErrEmptyToken
	}
	t, err := tokenFunc(ctx)
	if err != nil {
		return err
	}
	if t == "" {
		return ErrEmptyToken
	}
	*token = t
	return nil
}

// buildUpdateURL constructs an update URL from already-validated inputs.
// It is pure: the attribute store is never mutated.
func buildUpdateURL(id, token string, attrs *URLAttrs, command Command) (string, error) {
	query, err := attrs.QueryValues()
	if err != nil {
		return "", err
	}
	query.Set(KeyID, id)
	query.Set(KeyAuthToken, token)

	return fmt.Sprintf("things:///%s?%s", command, EncodeQuery(query)), nil
}

// updateTodoBuilder builds URLs for updating existing todos via the update command.
// Requires authentication token (obtained via authScheme or Client).
type updateTodoBuilder struct {
	scheme    *Scheme
	token     string
	tokenFunc func(context.Context) (string, error) // Optional lazy token loader
	id        string
	attrs     URLAttrs
	err       error
}

// NewTodoUpdater creates a new TodoUpdater for updating an existing todo.
func NewTodoUpdater(s *Scheme, tokenFunc func(context.Context) (string, error), id string) TodoUpdater {
	return &updateTodoBuilder{
		scheme:    s,
		tokenFunc: tokenFunc,
		id:        id,
		attrs:     NewURLAttrs(),
	}
}

// GetStore returns the attribute store for the builder.
func (b *updateTodoBuilder) GetStore() AttrStore { return &b.attrs }

// SetErr sets the error field for the builder.
func (b *updateTodoBuilder) SetErr(err error) { b.err = err }

// Title replaces the todo title.
func (b *updateTodoBuilder) Title(title string) TodoUpdater {
	return SetStr(b, TitleParam, title)
}

// Notes replaces the todo notes.
func (b *updateTodoBuilder) Notes(notes string) TodoUpdater {
	return SetStr(b, NotesParam, notes)
}

// PrependNotes prepends text to existing notes.
func (b *updateTodoBuilder) PrependNotes(notes string) TodoUpdater {
	return SetStr(b, PrependNotesParam, notes)
}

// AppendNotes appends text to existing notes.
func (b *updateTodoBuilder) AppendNotes(notes string) TodoUpdater {
	return SetStr(b, AppendNotesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *updateTodoBuilder) When(t time.Time) TodoUpdater {
	return SetWhenTime(b, t)
}

// WhenEvening schedules the todo for this evening.
func (b *updateTodoBuilder) WhenEvening() TodoUpdater {
	return SetWhenStr(b, WhenEvening)
}

// WhenAnytime schedules the todo for anytime (no specific time).
func (b *updateTodoBuilder) WhenAnytime() TodoUpdater {
	return SetWhenStr(b, WhenAnytime)
}

// WhenSomeday schedules the todo for someday (indefinite future).
func (b *updateTodoBuilder) WhenSomeday() TodoUpdater {
	return SetWhenStr(b, WhenSomeday)
}

// Reminder sets a reminder time for the todo.
// The reminder is combined with the scheduling date (When).
// If no scheduling date is set, defaults to "today".
// Hour must be 0-23, minute must be 0-59.
// Combining a reminder with WhenSomeday or WhenAnytime is a build error.
func (b *updateTodoBuilder) Reminder(hour, minute int) TodoUpdater {
	return SetReminder(b, hour, minute)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *updateTodoBuilder) Deadline(t time.Time) TodoUpdater {
	return SetDeadlineTime(b, t)
}

// ClearDeadline removes the deadline.
func (b *updateTodoBuilder) ClearDeadline() TodoUpdater {
	b.attrs.SetString(KeyDeadline, "")
	return b
}

// Tags replaces all tags.
func (b *updateTodoBuilder) Tags(tags ...string) TodoUpdater {
	return SetStrs(b, TagsParam, tags)
}

// AddTags adds tags without replacing existing ones.
func (b *updateTodoBuilder) AddTags(tags ...string) TodoUpdater {
	return SetStrs(b, AddTagsParam, tags)
}

// ChecklistItems replaces all checklist items.
func (b *updateTodoBuilder) ChecklistItems(items ...string) TodoUpdater {
	return SetStrs(b, ChecklistItemsParam, items)
}

// PrependChecklistItems prepends checklist items.
func (b *updateTodoBuilder) PrependChecklistItems(items ...string) TodoUpdater {
	return SetStrs(b, PrependChecklistParam, items)
}

// AppendChecklistItems appends checklist items.
func (b *updateTodoBuilder) AppendChecklistItems(items ...string) TodoUpdater {
	return SetStrs(b, AppendChecklistParam, items)
}

// List moves the todo to a project or area by name.
func (b *updateTodoBuilder) List(name string) TodoUpdater {
	return SetStr(b, ListParam, name)
}

// ListID moves the todo to a project or area by UUID.
func (b *updateTodoBuilder) ListID(id string) TodoUpdater {
	return SetStr(b, ListIDParam, id)
}

// Heading moves the todo to a heading by name.
func (b *updateTodoBuilder) Heading(name string) TodoUpdater {
	return SetStr(b, HeadingParam, name)
}

// HeadingID moves the todo to a heading by UUID.
func (b *updateTodoBuilder) HeadingID(id string) TodoUpdater {
	return SetStr(b, HeadingIDParam, id)
}

// Completed sets the completion status.
func (b *updateTodoBuilder) Completed(completed bool) TodoUpdater {
	return SetBool(b, CompletedParam, completed)
}

// Canceled sets the canceled status.
func (b *updateTodoBuilder) Canceled(canceled bool) TodoUpdater {
	return SetBool(b, CanceledParam, canceled)
}

// Duplicate duplicates the todo before updating.
func (b *updateTodoBuilder) Duplicate(duplicate bool) TodoUpdater {
	return SetBool(b, DuplicateParam, duplicate)
}

// Reveal navigates to the todo after updating.
func (b *updateTodoBuilder) Reveal(reveal bool) TodoUpdater {
	return SetBool(b, RevealParam, reveal)
}

// CreationDate sets the creation timestamp.
func (b *updateTodoBuilder) CreationDate(date time.Time) TodoUpdater {
	return SetTime(b, CreationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (b *updateTodoBuilder) CompletionDate(date time.Time) TodoUpdater {
	return SetTime(b, CompletionDateParam, date)
}

// validate checks all builder requirements before building the URL.
func (b *updateTodoBuilder) validate() error {
	if b.err != nil {
		return b.err
	}
	if b.id == "" {
		return ErrIDRequired
	}
	return nil
}

// build validates the builder, resolves the auth token once using ctx,
// and constructs the update URL.
func (b *updateTodoBuilder) build(ctx context.Context) (string, error) {
	if err := b.validate(); err != nil {
		return "", err
	}
	if err := resolveToken(ctx, &b.token, b.tokenFunc); err != nil {
		return "", err
	}
	return buildUpdateURL(b.id, b.token, &b.attrs, CommandUpdate)
}

// Build returns the Things URL for updating the todo.
// If the token has not been resolved yet, the token function runs without a
// caller context; use Execute for context-aware token loading.
func (b *updateTodoBuilder) Build() (string, error) {
	return b.build(context.Background())
}

// Execute builds and executes the update URL.
// Returns an error if the URL cannot be built or executed.
// The auth token is resolved at most once, using the provided context.
func (b *updateTodoBuilder) Execute(ctx context.Context) error {
	uri, err := b.build(ctx)
	if err != nil {
		return err
	}
	return b.scheme.Execute(ctx, uri)
}

// updateProjectBuilder builds URLs for updating existing projects via the update-project command.
// Requires authentication token (obtained via authScheme or Client).
type updateProjectBuilder struct {
	scheme    *Scheme
	token     string
	tokenFunc func(context.Context) (string, error) // Optional lazy token loader
	id        string
	attrs     URLAttrs
	err       error
}

// NewProjectUpdater creates a new ProjectUpdater for updating an existing project.
func NewProjectUpdater(s *Scheme, tokenFunc func(context.Context) (string, error), id string) ProjectUpdater {
	return &updateProjectBuilder{
		scheme:    s,
		tokenFunc: tokenFunc,
		id:        id,
		attrs:     NewURLAttrs(),
	}
}

// GetStore returns the attribute store for the builder.
func (b *updateProjectBuilder) GetStore() AttrStore { return &b.attrs }

// SetErr sets the error field for the builder.
func (b *updateProjectBuilder) SetErr(err error) { b.err = err }

// Title replaces the project title.
func (b *updateProjectBuilder) Title(title string) ProjectUpdater {
	return SetStr(b, TitleParam, title)
}

// Notes replaces the project notes.
func (b *updateProjectBuilder) Notes(notes string) ProjectUpdater {
	return SetStr(b, NotesParam, notes)
}

// PrependNotes prepends text to existing notes.
func (b *updateProjectBuilder) PrependNotes(notes string) ProjectUpdater {
	return SetStr(b, PrependNotesParam, notes)
}

// AppendNotes appends text to existing notes.
func (b *updateProjectBuilder) AppendNotes(notes string) ProjectUpdater {
	return SetStr(b, AppendNotesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *updateProjectBuilder) When(t time.Time) ProjectUpdater {
	return SetWhenTime(b, t)
}

// WhenEvening schedules the project for this evening.
func (b *updateProjectBuilder) WhenEvening() ProjectUpdater {
	return SetWhenStr(b, WhenEvening)
}

// WhenAnytime schedules the project for anytime (no specific time).
func (b *updateProjectBuilder) WhenAnytime() ProjectUpdater {
	return SetWhenStr(b, WhenAnytime)
}

// WhenSomeday schedules the project for someday (indefinite future).
func (b *updateProjectBuilder) WhenSomeday() ProjectUpdater {
	return SetWhenStr(b, WhenSomeday)
}

// Reminder sets a reminder time for the project.
// The reminder is combined with the scheduling date (When).
// If no scheduling date is set, defaults to "today".
// Hour must be 0-23, minute must be 0-59.
// Combining a reminder with WhenSomeday or WhenAnytime is a build error.
func (b *updateProjectBuilder) Reminder(hour, minute int) ProjectUpdater {
	return SetReminder(b, hour, minute)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *updateProjectBuilder) Deadline(t time.Time) ProjectUpdater {
	return SetDeadlineTime(b, t)
}

// ClearDeadline removes the deadline.
func (b *updateProjectBuilder) ClearDeadline() ProjectUpdater {
	b.attrs.SetString(KeyDeadline, "")
	return b
}

// Tags replaces all tags.
func (b *updateProjectBuilder) Tags(tags ...string) ProjectUpdater {
	return SetStrs(b, TagsParam, tags)
}

// AddTags adds tags without replacing existing ones.
func (b *updateProjectBuilder) AddTags(tags ...string) ProjectUpdater {
	return SetStrs(b, AddTagsParam, tags)
}

// Area moves the project to an area by name.
func (b *updateProjectBuilder) Area(name string) ProjectUpdater {
	return SetStr(b, AreaParam, name)
}

// AreaID moves the project to an area by UUID.
func (b *updateProjectBuilder) AreaID(id string) ProjectUpdater {
	return SetStr(b, AreaIDParam, id)
}

// Completed sets the completion status.
// Note: Setting completed=true is ignored unless all child todos
// are completed or canceled and all headings are archived.
func (b *updateProjectBuilder) Completed(completed bool) ProjectUpdater {
	return SetBool(b, CompletedParam, completed)
}

// Canceled sets the canceled status.
func (b *updateProjectBuilder) Canceled(canceled bool) ProjectUpdater {
	return SetBool(b, CanceledParam, canceled)
}

// Reveal navigates to the project after updating.
func (b *updateProjectBuilder) Reveal(reveal bool) ProjectUpdater {
	return SetBool(b, RevealParam, reveal)
}

// validate checks all builder requirements before building the URL.
func (b *updateProjectBuilder) validate() error {
	if b.err != nil {
		return b.err
	}
	if b.id == "" {
		return ErrIDRequired
	}
	return nil
}

// build validates the builder, resolves the auth token once using ctx,
// and constructs the update URL.
func (b *updateProjectBuilder) build(ctx context.Context) (string, error) {
	if err := b.validate(); err != nil {
		return "", err
	}
	if err := resolveToken(ctx, &b.token, b.tokenFunc); err != nil {
		return "", err
	}
	return buildUpdateURL(b.id, b.token, &b.attrs, CommandUpdateProject)
}

// Build returns the Things URL for updating the project.
// If the token has not been resolved yet, the token function runs without a
// caller context; use Execute for context-aware token loading.
func (b *updateProjectBuilder) Build() (string, error) {
	return b.build(context.Background())
}

// Execute builds and executes the update URL.
// Returns an error if the URL cannot be built or executed.
// The auth token is resolved at most once, using the provided context.
func (b *updateProjectBuilder) Execute(ctx context.Context) error {
	uri, err := b.build(ctx)
	if err != nil {
		return err
	}
	return b.scheme.Execute(ctx, uri)
}
