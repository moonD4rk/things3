package things3

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// AddTodoBuilder builds URLs for creating new to-dos via the add command.
type AddTodoBuilder struct {
	scheme *Scheme
	attrs  urlAttrs
	err    error
}

// getStore returns the attribute store for the builder.
func (b *AddTodoBuilder) getStore() attrStore { return &b.attrs }

// setErr sets the error field for the builder.
func (b *AddTodoBuilder) setErr(err error) { b.err = err }

// Title sets the to-do title.
func (b *AddTodoBuilder) Title(title string) *AddTodoBuilder {
	return setStr(b, titleParam, title)
}

// Titles sets multiple to-do titles (creates multiple to-dos).
// Titles are newline-separated.
func (b *AddTodoBuilder) Titles(titles ...string) *AddTodoBuilder {
	combined := strings.Join(titles, "\n")
	if len(combined) > maxTitleLength {
		b.err = ErrTitleTooLong
		return b
	}
	b.attrs.SetString(keyTitles, combined)
	return b
}

// Notes sets the to-do notes/description.
func (b *AddTodoBuilder) Notes(notes string) *AddTodoBuilder {
	return setStr(b, notesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
//
// Example:
//
//	scheme.AddTodo().Title("Task").When(things3.Today())
//	scheme.AddTodo().Title("Task").When(time.Now().AddDate(0, 0, 7))
func (b *AddTodoBuilder) When(t time.Time) *AddTodoBuilder {
	return setWhenTime(b, t)
}

// WhenEvening schedules the to-do for this evening.
// This is a Things 3-specific concept that cannot be expressed as a date.
func (b *AddTodoBuilder) WhenEvening() *AddTodoBuilder {
	return setWhenStr(b, whenEvening)
}

// WhenAnytime schedules the to-do for anytime (no specific time).
// This is a Things 3-specific concept that cannot be expressed as a date.
func (b *AddTodoBuilder) WhenAnytime() *AddTodoBuilder {
	return setWhenStr(b, whenAnytime)
}

// WhenSomeday schedules the to-do for someday (indefinite future).
// This is a Things 3-specific concept that cannot be expressed as a date.
func (b *AddTodoBuilder) WhenSomeday() *AddTodoBuilder {
	return setWhenStr(b, whenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
//
// Example:
//
//	scheme.AddTodo().Title("Task").Deadline(time.Date(2025, 12, 31, 0, 0, 0, 0, time.Local))
func (b *AddTodoBuilder) Deadline(t time.Time) *AddTodoBuilder {
	return setDeadlineTime(b, t)
}

// Tags sets the tags for the to-do.
// Tags must already exist in Things.
func (b *AddTodoBuilder) Tags(tags ...string) *AddTodoBuilder {
	return setStrs(b, tagsParam, tags)
}

// ChecklistItems sets the checklist items.
func (b *AddTodoBuilder) ChecklistItems(items ...string) *AddTodoBuilder {
	return setStrs(b, checklistItemsParam, items)
}

// List sets the target project or area by name.
func (b *AddTodoBuilder) List(name string) *AddTodoBuilder {
	return setStr(b, listParam, name)
}

// ListID sets the target project or area by UUID.
func (b *AddTodoBuilder) ListID(id string) *AddTodoBuilder {
	return setStr(b, listIDParam, id)
}

// Heading sets the target heading within a project by name.
func (b *AddTodoBuilder) Heading(name string) *AddTodoBuilder {
	return setStr(b, headingParam, name)
}

// HeadingID sets the target heading within a project by UUID.
func (b *AddTodoBuilder) HeadingID(id string) *AddTodoBuilder {
	return setStr(b, headingIDParam, id)
}

// Completed sets the completion status.
func (b *AddTodoBuilder) Completed(completed bool) *AddTodoBuilder {
	return setBool(b, completedParam, completed)
}

// Canceled sets the canceled status.
func (b *AddTodoBuilder) Canceled(canceled bool) *AddTodoBuilder {
	return setBool(b, canceledParam, canceled)
}

// ShowQuickEntry displays the quick entry dialog instead of adding directly.
func (b *AddTodoBuilder) ShowQuickEntry(show bool) *AddTodoBuilder {
	return setBool(b, showQuickEntryParam, show)
}

// Reveal navigates to the newly created to-do.
func (b *AddTodoBuilder) Reveal(reveal bool) *AddTodoBuilder {
	return setBool(b, revealParam, reveal)
}

// Reminder sets a reminder time for the to-do.
// The reminder is combined with the scheduling date (When/WhenDate).
// If no scheduling date is set, defaults to "today".
// Hour must be 0-23, minute must be 0-59.
//
// Example:
//
//	scheme.AddTodo().Title("Meeting").When(WhenTomorrow).Reminder(14, 30) // tomorrow@14:30
//	scheme.AddTodo().Title("Call").Reminder(15, 0) // today@15:00 (defaults to today)
func (b *AddTodoBuilder) Reminder(hour, minute int) *AddTodoBuilder {
	return setReminder(b, hour, minute)
}

// CreationDate sets the creation timestamp.
// Future dates are ignored by Things.
func (b *AddTodoBuilder) CreationDate(date time.Time) *AddTodoBuilder {
	return setTime(b, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
// Future dates are ignored by Things.
func (b *AddTodoBuilder) CompletionDate(date time.Time) *AddTodoBuilder {
	return setTime(b, completionDateParam, date)
}

// Build returns the Things URL for creating the to-do.
func (b *AddTodoBuilder) Build() (string, error) {
	if b.err != nil {
		return "", b.err
	}

	// Finalize when parameter with reminder time if set
	b.attrs.FinalizeWhen()

	query := url.Values{}
	for k, v := range b.attrs.params {
		query.Set(k, v)
	}

	return fmt.Sprintf("things:///%s?%s", CommandAdd, encodeQuery(query)), nil
}

// Execute builds and executes the add URL.
// Returns an error if the URL cannot be built or executed.
func (b *AddTodoBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.execute(ctx, uri)
}

// AddProjectBuilder builds URLs for creating new projects via the add-project command.
type AddProjectBuilder struct {
	scheme *Scheme
	attrs  urlAttrs
	err    error
}

// getStore returns the attribute store for the builder.
func (b *AddProjectBuilder) getStore() attrStore { return &b.attrs }

// setErr sets the error field for the builder.
func (b *AddProjectBuilder) setErr(err error) { b.err = err }

// Title sets the project title.
func (b *AddProjectBuilder) Title(title string) *AddProjectBuilder {
	return setStr(b, titleParam, title)
}

// Notes sets the project notes/description.
func (b *AddProjectBuilder) Notes(notes string) *AddProjectBuilder {
	return setStr(b, notesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *AddProjectBuilder) When(t time.Time) *AddProjectBuilder {
	return setWhenTime(b, t)
}

// WhenEvening schedules the project for this evening.
func (b *AddProjectBuilder) WhenEvening() *AddProjectBuilder {
	return setWhenStr(b, whenEvening)
}

// WhenAnytime schedules the project for anytime (no specific time).
func (b *AddProjectBuilder) WhenAnytime() *AddProjectBuilder {
	return setWhenStr(b, whenAnytime)
}

// WhenSomeday schedules the project for someday (indefinite future).
func (b *AddProjectBuilder) WhenSomeday() *AddProjectBuilder {
	return setWhenStr(b, whenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *AddProjectBuilder) Deadline(t time.Time) *AddProjectBuilder {
	return setDeadlineTime(b, t)
}

// Tags sets the tags for the project.
func (b *AddProjectBuilder) Tags(tags ...string) *AddProjectBuilder {
	return setStrs(b, tagsParam, tags)
}

// Area sets the parent area by name.
func (b *AddProjectBuilder) Area(name string) *AddProjectBuilder {
	return setStr(b, areaParam, name)
}

// AreaID sets the parent area by UUID.
func (b *AddProjectBuilder) AreaID(id string) *AddProjectBuilder {
	return setStr(b, areaIDParam, id)
}

// Todos sets the child to-do titles.
func (b *AddProjectBuilder) Todos(titles ...string) *AddProjectBuilder {
	b.attrs.SetStrings(keyTodos, titles, "\n")
	return b
}

// Completed sets the completion status.
func (b *AddProjectBuilder) Completed(completed bool) *AddProjectBuilder {
	return setBool(b, completedParam, completed)
}

// Canceled sets the canceled status.
func (b *AddProjectBuilder) Canceled(canceled bool) *AddProjectBuilder {
	return setBool(b, canceledParam, canceled)
}

// Reveal navigates to the newly created project.
func (b *AddProjectBuilder) Reveal(reveal bool) *AddProjectBuilder {
	return setBool(b, revealParam, reveal)
}

// Reminder sets a reminder time for the project.
// The reminder is combined with the scheduling date (When).
// If no scheduling date is set, defaults to "today".
// Hour must be 0-23, minute must be 0-59.
func (b *AddProjectBuilder) Reminder(hour, minute int) *AddProjectBuilder {
	return setReminder(b, hour, minute)
}

// CreationDate sets the creation timestamp.
func (b *AddProjectBuilder) CreationDate(date time.Time) *AddProjectBuilder {
	return setTime(b, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (b *AddProjectBuilder) CompletionDate(date time.Time) *AddProjectBuilder {
	return setTime(b, completionDateParam, date)
}

// Build returns the Things URL for creating the project.
func (b *AddProjectBuilder) Build() (string, error) {
	if b.err != nil {
		return "", b.err
	}

	// Finalize when parameter with reminder time if set
	b.attrs.FinalizeWhen()

	query := url.Values{}
	for k, v := range b.attrs.params {
		query.Set(k, v)
	}

	return fmt.Sprintf("things:///%s?%s", CommandAddProject, encodeQuery(query)), nil
}

// Execute builds and executes the add-project URL.
// Returns an error if the URL cannot be built or executed.
func (b *AddProjectBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.execute(ctx, uri)
}
