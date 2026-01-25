package things3

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// updateTodoBuilder builds URLs for updating existing to-dos via the update command.
// Requires authentication token (obtained via authScheme or Client).
type updateTodoBuilder struct {
	scheme    *scheme
	token     string
	tokenFunc func(context.Context) (string, error) // Optional lazy token loader
	id        string
	attrs     urlAttrs
	err       error
}

// getStore returns the attribute store for the builder.
func (b *updateTodoBuilder) getStore() attrStore { return &b.attrs }

// setErr sets the error field for the builder.
func (b *updateTodoBuilder) setErr(err error) { b.err = err }

// Title replaces the to-do title.
func (b *updateTodoBuilder) Title(title string) TodoUpdater {
	return setStr(b, titleParam, title)
}

// Notes replaces the to-do notes.
func (b *updateTodoBuilder) Notes(notes string) TodoUpdater {
	return setStr(b, notesParam, notes)
}

// PrependNotes prepends text to existing notes.
func (b *updateTodoBuilder) PrependNotes(notes string) TodoUpdater {
	return setStr(b, prependNotesParam, notes)
}

// AppendNotes appends text to existing notes.
func (b *updateTodoBuilder) AppendNotes(notes string) TodoUpdater {
	return setStr(b, appendNotesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *updateTodoBuilder) When(t time.Time) TodoUpdater {
	return setWhenTime(b, t)
}

// WhenEvening schedules the to-do for this evening.
func (b *updateTodoBuilder) WhenEvening() TodoUpdater {
	return setWhenStr(b, whenEvening)
}

// WhenAnytime schedules the to-do for anytime (no specific time).
func (b *updateTodoBuilder) WhenAnytime() TodoUpdater {
	return setWhenStr(b, whenAnytime)
}

// WhenSomeday schedules the to-do for someday (indefinite future).
func (b *updateTodoBuilder) WhenSomeday() TodoUpdater {
	return setWhenStr(b, whenSomeday)
}

// Reminder sets a reminder time for the to-do.
// The reminder is combined with the scheduling date (When).
// If no scheduling date is set, defaults to "today".
// Hour must be 0-23, minute must be 0-59.
func (b *updateTodoBuilder) Reminder(hour, minute int) TodoUpdater {
	return setReminder(b, hour, minute)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *updateTodoBuilder) Deadline(t time.Time) TodoUpdater {
	return setDeadlineTime(b, t)
}

// ClearDeadline removes the deadline.
func (b *updateTodoBuilder) ClearDeadline() TodoUpdater {
	b.attrs.SetString(keyDeadline, "")
	return b
}

// Tags replaces all tags.
func (b *updateTodoBuilder) Tags(tags ...string) TodoUpdater {
	return setStrs(b, tagsParam, tags)
}

// AddTags adds tags without replacing existing ones.
func (b *updateTodoBuilder) AddTags(tags ...string) TodoUpdater {
	return setStrs(b, addTagsParam, tags)
}

// ChecklistItems replaces all checklist items.
func (b *updateTodoBuilder) ChecklistItems(items ...string) TodoUpdater {
	return setStrs(b, checklistItemsParam, items)
}

// PrependChecklistItems prepends checklist items.
func (b *updateTodoBuilder) PrependChecklistItems(items ...string) TodoUpdater {
	return setStrs(b, prependChecklistParam, items)
}

// AppendChecklistItems appends checklist items.
func (b *updateTodoBuilder) AppendChecklistItems(items ...string) TodoUpdater {
	return setStrs(b, appendChecklistParam, items)
}

// List moves the to-do to a project or area by name.
func (b *updateTodoBuilder) List(name string) TodoUpdater {
	return setStr(b, listParam, name)
}

// ListID moves the to-do to a project or area by UUID.
func (b *updateTodoBuilder) ListID(id string) TodoUpdater {
	return setStr(b, listIDParam, id)
}

// Heading moves the to-do to a heading by name.
func (b *updateTodoBuilder) Heading(name string) TodoUpdater {
	return setStr(b, headingParam, name)
}

// HeadingID moves the to-do to a heading by UUID.
func (b *updateTodoBuilder) HeadingID(id string) TodoUpdater {
	return setStr(b, headingIDParam, id)
}

// Completed sets the completion status.
func (b *updateTodoBuilder) Completed(completed bool) TodoUpdater {
	return setBool(b, completedParam, completed)
}

// Canceled sets the canceled status.
func (b *updateTodoBuilder) Canceled(canceled bool) TodoUpdater {
	return setBool(b, canceledParam, canceled)
}

// Duplicate duplicates the to-do before updating.
func (b *updateTodoBuilder) Duplicate(duplicate bool) TodoUpdater {
	return setBool(b, duplicateParam, duplicate)
}

// Reveal navigates to the to-do after updating.
func (b *updateTodoBuilder) Reveal(reveal bool) TodoUpdater {
	return setBool(b, revealParam, reveal)
}

// CreationDate sets the creation timestamp.
func (b *updateTodoBuilder) CreationDate(date time.Time) TodoUpdater {
	return setTime(b, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (b *updateTodoBuilder) CompletionDate(date time.Time) TodoUpdater {
	return setTime(b, completionDateParam, date)
}

// validate checks all builder requirements before building the URL.
func (b *updateTodoBuilder) validate() error {
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
func (b *updateTodoBuilder) Build() (string, error) {
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
// If token is not set but tokenFunc is provided, it will fetch the token first.
func (b *updateTodoBuilder) Execute(ctx context.Context) error {
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

// updateProjectBuilder builds URLs for updating existing projects via the update-project command.
// Requires authentication token (obtained via authScheme or Client).
type updateProjectBuilder struct {
	scheme    *scheme
	token     string
	tokenFunc func(context.Context) (string, error) // Optional lazy token loader
	id        string
	attrs     urlAttrs
	err       error
}

// getStore returns the attribute store for the builder.
func (b *updateProjectBuilder) getStore() attrStore { return &b.attrs }

// setErr sets the error field for the builder.
func (b *updateProjectBuilder) setErr(err error) { b.err = err }

// Title replaces the project title.
func (b *updateProjectBuilder) Title(title string) ProjectUpdater {
	return setStr(b, titleParam, title)
}

// Notes replaces the project notes.
func (b *updateProjectBuilder) Notes(notes string) ProjectUpdater {
	return setStr(b, notesParam, notes)
}

// PrependNotes prepends text to existing notes.
func (b *updateProjectBuilder) PrependNotes(notes string) ProjectUpdater {
	return setStr(b, prependNotesParam, notes)
}

// AppendNotes appends text to existing notes.
func (b *updateProjectBuilder) AppendNotes(notes string) ProjectUpdater {
	return setStr(b, appendNotesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *updateProjectBuilder) When(t time.Time) ProjectUpdater {
	return setWhenTime(b, t)
}

// WhenEvening schedules the project for this evening.
func (b *updateProjectBuilder) WhenEvening() ProjectUpdater {
	return setWhenStr(b, whenEvening)
}

// WhenAnytime schedules the project for anytime (no specific time).
func (b *updateProjectBuilder) WhenAnytime() ProjectUpdater {
	return setWhenStr(b, whenAnytime)
}

// WhenSomeday schedules the project for someday (indefinite future).
func (b *updateProjectBuilder) WhenSomeday() ProjectUpdater {
	return setWhenStr(b, whenSomeday)
}

// Reminder sets a reminder time for the project.
// The reminder is combined with the scheduling date (When).
// If no scheduling date is set, defaults to "today".
// Hour must be 0-23, minute must be 0-59.
func (b *updateProjectBuilder) Reminder(hour, minute int) ProjectUpdater {
	return setReminder(b, hour, minute)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *updateProjectBuilder) Deadline(t time.Time) ProjectUpdater {
	return setDeadlineTime(b, t)
}

// ClearDeadline removes the deadline.
func (b *updateProjectBuilder) ClearDeadline() ProjectUpdater {
	b.attrs.SetString(keyDeadline, "")
	return b
}

// Tags replaces all tags.
func (b *updateProjectBuilder) Tags(tags ...string) ProjectUpdater {
	return setStrs(b, tagsParam, tags)
}

// AddTags adds tags without replacing existing ones.
func (b *updateProjectBuilder) AddTags(tags ...string) ProjectUpdater {
	return setStrs(b, addTagsParam, tags)
}

// Area moves the project to an area by name.
func (b *updateProjectBuilder) Area(name string) ProjectUpdater {
	return setStr(b, areaParam, name)
}

// AreaID moves the project to an area by UUID.
func (b *updateProjectBuilder) AreaID(id string) ProjectUpdater {
	return setStr(b, areaIDParam, id)
}

// Completed sets the completion status.
// Note: Setting completed=true is ignored unless all child to-dos
// are completed or canceled and all headings are archived.
func (b *updateProjectBuilder) Completed(completed bool) ProjectUpdater {
	return setBool(b, completedParam, completed)
}

// Canceled sets the canceled status.
func (b *updateProjectBuilder) Canceled(canceled bool) ProjectUpdater {
	return setBool(b, canceledParam, canceled)
}

// Reveal navigates to the project after updating.
func (b *updateProjectBuilder) Reveal(reveal bool) ProjectUpdater {
	return setBool(b, revealParam, reveal)
}

// validate checks all builder requirements before building the URL.
func (b *updateProjectBuilder) validate() error {
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
func (b *updateProjectBuilder) Build() (string, error) {
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
// If token is not set but tokenFunc is provided, it will fetch the token first.
func (b *updateProjectBuilder) Execute(ctx context.Context) error {
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
