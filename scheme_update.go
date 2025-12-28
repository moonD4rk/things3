package things3

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// UpdateTodoBuilder builds URLs for updating existing to-dos via the update command.
// Requires authentication token (obtained via AuthScheme).
type UpdateTodoBuilder struct {
	scheme *Scheme
	token  string
	id     string
	attrs  urlAttrs
	err    error
}

// getStore returns the attribute store for the builder.
func (b *UpdateTodoBuilder) getStore() attrStore { return &b.attrs }

// setErr sets the error field for the builder.
func (b *UpdateTodoBuilder) setErr(err error) { b.err = err }

// Title replaces the to-do title.
func (b *UpdateTodoBuilder) Title(title string) *UpdateTodoBuilder {
	return setStr(b, titleParam, title)
}

// Notes replaces the to-do notes.
func (b *UpdateTodoBuilder) Notes(notes string) *UpdateTodoBuilder {
	return setStr(b, notesParam, notes)
}

// PrependNotes prepends text to existing notes.
func (b *UpdateTodoBuilder) PrependNotes(notes string) *UpdateTodoBuilder {
	return setStr(b, prependNotesParam, notes)
}

// AppendNotes appends text to existing notes.
func (b *UpdateTodoBuilder) AppendNotes(notes string) *UpdateTodoBuilder {
	return setStr(b, appendNotesParam, notes)
}

// When sets the scheduling date.
func (b *UpdateTodoBuilder) When(when When) *UpdateTodoBuilder {
	return setWhenStr(b, when)
}

// WhenDate sets a specific date for scheduling.
func (b *UpdateTodoBuilder) WhenDate(year int, month time.Month, day int) *UpdateTodoBuilder {
	return setDate(b, whenParam, year, month, day)
}

// Reminder sets a reminder time for the to-do.
// The reminder is combined with the scheduling date (When/WhenDate).
// If no scheduling date is set, defaults to "today".
// Hour must be 0-23, minute must be 0-59.
func (b *UpdateTodoBuilder) Reminder(hour, minute int) *UpdateTodoBuilder {
	return setReminder(b, hour, minute)
}

// Deadline sets the deadline date.
// Pass empty string to clear the deadline.
func (b *UpdateTodoBuilder) Deadline(date string) *UpdateTodoBuilder {
	return setStr(b, deadlineParam, date)
}

// ClearDeadline removes the deadline.
func (b *UpdateTodoBuilder) ClearDeadline() *UpdateTodoBuilder {
	b.attrs.SetString(keyDeadline, "")
	return b
}

// Tags replaces all tags.
func (b *UpdateTodoBuilder) Tags(tags ...string) *UpdateTodoBuilder {
	return setStrs(b, tagsParam, tags)
}

// AddTags adds tags without replacing existing ones.
func (b *UpdateTodoBuilder) AddTags(tags ...string) *UpdateTodoBuilder {
	return setStrs(b, addTagsParam, tags)
}

// ChecklistItems replaces all checklist items.
func (b *UpdateTodoBuilder) ChecklistItems(items ...string) *UpdateTodoBuilder {
	return setStrs(b, checklistItemsParam, items)
}

// PrependChecklistItems prepends checklist items.
func (b *UpdateTodoBuilder) PrependChecklistItems(items ...string) *UpdateTodoBuilder {
	return setStrs(b, prependChecklistParam, items)
}

// AppendChecklistItems appends checklist items.
func (b *UpdateTodoBuilder) AppendChecklistItems(items ...string) *UpdateTodoBuilder {
	return setStrs(b, appendChecklistParam, items)
}

// List moves the to-do to a project or area by name.
func (b *UpdateTodoBuilder) List(name string) *UpdateTodoBuilder {
	return setStr(b, listParam, name)
}

// ListID moves the to-do to a project or area by UUID.
func (b *UpdateTodoBuilder) ListID(id string) *UpdateTodoBuilder {
	return setStr(b, listIDParam, id)
}

// Heading moves the to-do to a heading by name.
func (b *UpdateTodoBuilder) Heading(name string) *UpdateTodoBuilder {
	return setStr(b, headingParam, name)
}

// HeadingID moves the to-do to a heading by UUID.
func (b *UpdateTodoBuilder) HeadingID(id string) *UpdateTodoBuilder {
	return setStr(b, headingIDParam, id)
}

// Completed sets the completion status.
func (b *UpdateTodoBuilder) Completed(completed bool) *UpdateTodoBuilder {
	return setBool(b, completedParam, completed)
}

// Canceled sets the canceled status.
func (b *UpdateTodoBuilder) Canceled(canceled bool) *UpdateTodoBuilder {
	return setBool(b, canceledParam, canceled)
}

// Duplicate duplicates the to-do before updating.
func (b *UpdateTodoBuilder) Duplicate(duplicate bool) *UpdateTodoBuilder {
	return setBool(b, duplicateParam, duplicate)
}

// Reveal navigates to the to-do after updating.
func (b *UpdateTodoBuilder) Reveal(reveal bool) *UpdateTodoBuilder {
	return setBool(b, revealParam, reveal)
}

// CreationDate sets the creation timestamp.
func (b *UpdateTodoBuilder) CreationDate(date time.Time) *UpdateTodoBuilder {
	return setTime(b, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (b *UpdateTodoBuilder) CompletionDate(date time.Time) *UpdateTodoBuilder {
	return setTime(b, completionDateParam, date)
}

// validate checks all builder requirements before building the URL.
func (b *UpdateTodoBuilder) validate() error {
	if b.err != nil {
		return b.err
	}
	if b.token == "" {
		return ErrEmptyToken
	}
	if b.id == "" {
		return ErrIDRequired
	}
	return nil
}

// Build returns the Things URL for updating the to-do.
func (b *UpdateTodoBuilder) Build() (string, error) {
	if err := b.validate(); err != nil {
		return "", err
	}

	// Finalize when parameter with reminder time if set
	b.attrs.FinalizeWhen()

	query := url.Values{}
	query.Set(keyID, b.id)
	query.Set(keyAuthToken, b.token)
	for k, v := range b.attrs.params {
		query.Set(k, v)
	}

	return fmt.Sprintf("things:///%s?%s", CommandUpdate, encodeQuery(query)), nil
}

// Execute builds and executes the update URL.
// Returns an error if the URL cannot be built or executed.
func (b *UpdateTodoBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.execute(ctx, uri)
}

// UpdateProjectBuilder builds URLs for updating existing projects via the update-project command.
// Requires authentication token (obtained via AuthScheme).
type UpdateProjectBuilder struct {
	scheme *Scheme
	token  string
	id     string
	attrs  urlAttrs
	err    error
}

// getStore returns the attribute store for the builder.
func (b *UpdateProjectBuilder) getStore() attrStore { return &b.attrs }

// setErr sets the error field for the builder.
func (b *UpdateProjectBuilder) setErr(err error) { b.err = err }

// Title replaces the project title.
func (b *UpdateProjectBuilder) Title(title string) *UpdateProjectBuilder {
	return setStr(b, titleParam, title)
}

// Notes replaces the project notes.
func (b *UpdateProjectBuilder) Notes(notes string) *UpdateProjectBuilder {
	return setStr(b, notesParam, notes)
}

// PrependNotes prepends text to existing notes.
func (b *UpdateProjectBuilder) PrependNotes(notes string) *UpdateProjectBuilder {
	return setStr(b, prependNotesParam, notes)
}

// AppendNotes appends text to existing notes.
func (b *UpdateProjectBuilder) AppendNotes(notes string) *UpdateProjectBuilder {
	return setStr(b, appendNotesParam, notes)
}

// When sets the scheduling date.
func (b *UpdateProjectBuilder) When(when When) *UpdateProjectBuilder {
	return setWhenStr(b, when)
}

// WhenDate sets a specific date for scheduling.
func (b *UpdateProjectBuilder) WhenDate(year int, month time.Month, day int) *UpdateProjectBuilder {
	return setDate(b, whenParam, year, month, day)
}

// Reminder sets a reminder time for the project.
// The reminder is combined with the scheduling date (When/WhenDate).
// If no scheduling date is set, defaults to "today".
// Hour must be 0-23, minute must be 0-59.
func (b *UpdateProjectBuilder) Reminder(hour, minute int) *UpdateProjectBuilder {
	return setReminder(b, hour, minute)
}

// Deadline sets the deadline date.
func (b *UpdateProjectBuilder) Deadline(date string) *UpdateProjectBuilder {
	return setStr(b, deadlineParam, date)
}

// ClearDeadline removes the deadline.
func (b *UpdateProjectBuilder) ClearDeadline() *UpdateProjectBuilder {
	b.attrs.SetString(keyDeadline, "")
	return b
}

// Tags replaces all tags.
func (b *UpdateProjectBuilder) Tags(tags ...string) *UpdateProjectBuilder {
	return setStrs(b, tagsParam, tags)
}

// AddTags adds tags without replacing existing ones.
func (b *UpdateProjectBuilder) AddTags(tags ...string) *UpdateProjectBuilder {
	return setStrs(b, addTagsParam, tags)
}

// Area moves the project to an area by name.
func (b *UpdateProjectBuilder) Area(name string) *UpdateProjectBuilder {
	return setStr(b, areaParam, name)
}

// AreaID moves the project to an area by UUID.
func (b *UpdateProjectBuilder) AreaID(id string) *UpdateProjectBuilder {
	return setStr(b, areaIDParam, id)
}

// Completed sets the completion status.
// Note: Setting completed=true is ignored unless all child to-dos
// are completed or canceled and all headings are archived.
func (b *UpdateProjectBuilder) Completed(completed bool) *UpdateProjectBuilder {
	return setBool(b, completedParam, completed)
}

// Canceled sets the canceled status.
func (b *UpdateProjectBuilder) Canceled(canceled bool) *UpdateProjectBuilder {
	return setBool(b, canceledParam, canceled)
}

// Reveal navigates to the project after updating.
func (b *UpdateProjectBuilder) Reveal(reveal bool) *UpdateProjectBuilder {
	return setBool(b, revealParam, reveal)
}

// validate checks all builder requirements before building the URL.
func (b *UpdateProjectBuilder) validate() error {
	if b.err != nil {
		return b.err
	}
	if b.token == "" {
		return ErrEmptyToken
	}
	if b.id == "" {
		return ErrIDRequired
	}
	return nil
}

// Build returns the Things URL for updating the project.
func (b *UpdateProjectBuilder) Build() (string, error) {
	if err := b.validate(); err != nil {
		return "", err
	}

	// Finalize when parameter with reminder time if set
	b.attrs.FinalizeWhen()

	query := url.Values{}
	query.Set(keyID, b.id)
	query.Set(keyAuthToken, b.token)
	for k, v := range b.attrs.params {
		query.Set(k, v)
	}

	return fmt.Sprintf("things:///%s?%s", CommandUpdateProject, encodeQuery(query)), nil
}

// Execute builds and executes the update URL.
// Returns an error if the URL cannot be built or executed.
func (b *UpdateProjectBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.execute(ctx, uri)
}
