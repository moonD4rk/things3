package things3

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// addTodoBuilder builds URLs for creating new to-dos via the add command.
type addTodoBuilder struct {
	scheme *scheme
	attrs  urlAttrs
	err    error
}

// getStore returns the attribute store for the builder.
func (b *addTodoBuilder) getStore() attrStore { return &b.attrs }

// setErr sets the error field for the builder.
func (b *addTodoBuilder) setErr(err error) { b.err = err }

// Title sets the to-do title.
func (b *addTodoBuilder) Title(title string) TodoAdder {
	return setStr(b, titleParam, title)
}

// Titles sets multiple to-do titles (creates multiple to-dos).
// Titles are newline-separated.
func (b *addTodoBuilder) Titles(titles ...string) TodoAdder {
	combined := strings.Join(titles, "\n")
	if len(combined) > maxTitleLength {
		b.err = ErrTitleTooLong
		return b
	}
	b.attrs.SetString(keyTitles, combined)
	return b
}

// Notes sets the to-do notes/description.
func (b *addTodoBuilder) Notes(notes string) TodoAdder {
	return setStr(b, notesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
//
// Example:
//
//	scheme.AddTodo().Title("Task").When(things3.Today())
//	scheme.AddTodo().Title("Task").When(time.Now().AddDate(0, 0, 7))
func (b *addTodoBuilder) When(t time.Time) TodoAdder {
	return setWhenTime(b, t)
}

// WhenEvening schedules the to-do for this evening.
// This is a Things 3-specific concept that cannot be expressed as a date.
func (b *addTodoBuilder) WhenEvening() TodoAdder {
	return setWhenStr(b, whenEvening)
}

// WhenAnytime schedules the to-do for anytime (no specific time).
// This is a Things 3-specific concept that cannot be expressed as a date.
func (b *addTodoBuilder) WhenAnytime() TodoAdder {
	return setWhenStr(b, whenAnytime)
}

// WhenSomeday schedules the to-do for someday (indefinite future).
// This is a Things 3-specific concept that cannot be expressed as a date.
func (b *addTodoBuilder) WhenSomeday() TodoAdder {
	return setWhenStr(b, whenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
//
// Example:
//
//	scheme.AddTodo().Title("Task").Deadline(time.Date(2025, 12, 31, 0, 0, 0, 0, time.Local))
func (b *addTodoBuilder) Deadline(t time.Time) TodoAdder {
	return setDeadlineTime(b, t)
}

// Tags sets the tags for the to-do.
// Tags must already exist in Things.
func (b *addTodoBuilder) Tags(tags ...string) TodoAdder {
	return setStrs(b, tagsParam, tags)
}

// ChecklistItems sets the checklist items.
func (b *addTodoBuilder) ChecklistItems(items ...string) TodoAdder {
	return setStrs(b, checklistItemsParam, items)
}

// List sets the target project or area by name.
func (b *addTodoBuilder) List(name string) TodoAdder {
	return setStr(b, listParam, name)
}

// ListID sets the target project or area by UUID.
func (b *addTodoBuilder) ListID(id string) TodoAdder {
	return setStr(b, listIDParam, id)
}

// Heading sets the target heading within a project by name.
func (b *addTodoBuilder) Heading(name string) TodoAdder {
	return setStr(b, headingParam, name)
}

// HeadingID sets the target heading within a project by UUID.
func (b *addTodoBuilder) HeadingID(id string) TodoAdder {
	return setStr(b, headingIDParam, id)
}

// Completed sets the completion status.
func (b *addTodoBuilder) Completed(completed bool) TodoAdder {
	return setBool(b, completedParam, completed)
}

// Canceled sets the canceled status.
func (b *addTodoBuilder) Canceled(canceled bool) TodoAdder {
	return setBool(b, canceledParam, canceled)
}

// ShowQuickEntry displays the quick entry dialog instead of adding directly.
func (b *addTodoBuilder) ShowQuickEntry(show bool) TodoAdder {
	return setBool(b, showQuickEntryParam, show)
}

// Reveal navigates to the newly created to-do.
func (b *addTodoBuilder) Reveal(reveal bool) TodoAdder {
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
func (b *addTodoBuilder) Reminder(hour, minute int) TodoAdder {
	return setReminder(b, hour, minute)
}

// CreationDate sets the creation timestamp.
// Future dates are ignored by Things.
func (b *addTodoBuilder) CreationDate(date time.Time) TodoAdder {
	return setTime(b, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
// Future dates are ignored by Things.
func (b *addTodoBuilder) CompletionDate(date time.Time) TodoAdder {
	return setTime(b, completionDateParam, date)
}

// Build returns the Things URL for creating the to-do.
func (b *addTodoBuilder) Build() (string, error) {
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
func (b *addTodoBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.execute(ctx, uri)
}

// addProjectBuilder builds URLs for creating new projects via the add-project command.
type addProjectBuilder struct {
	scheme *scheme
	attrs  urlAttrs
	err    error
}

// getStore returns the attribute store for the builder.
func (b *addProjectBuilder) getStore() attrStore { return &b.attrs }

// setErr sets the error field for the builder.
func (b *addProjectBuilder) setErr(err error) { b.err = err }

// Title sets the project title.
func (b *addProjectBuilder) Title(title string) ProjectAdder {
	return setStr(b, titleParam, title)
}

// Notes sets the project notes/description.
func (b *addProjectBuilder) Notes(notes string) ProjectAdder {
	return setStr(b, notesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *addProjectBuilder) When(t time.Time) ProjectAdder {
	return setWhenTime(b, t)
}

// WhenEvening schedules the project for this evening.
func (b *addProjectBuilder) WhenEvening() ProjectAdder {
	return setWhenStr(b, whenEvening)
}

// WhenAnytime schedules the project for anytime (no specific time).
func (b *addProjectBuilder) WhenAnytime() ProjectAdder {
	return setWhenStr(b, whenAnytime)
}

// WhenSomeday schedules the project for someday (indefinite future).
func (b *addProjectBuilder) WhenSomeday() ProjectAdder {
	return setWhenStr(b, whenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *addProjectBuilder) Deadline(t time.Time) ProjectAdder {
	return setDeadlineTime(b, t)
}

// Tags sets the tags for the project.
func (b *addProjectBuilder) Tags(tags ...string) ProjectAdder {
	return setStrs(b, tagsParam, tags)
}

// Area sets the parent area by name.
func (b *addProjectBuilder) Area(name string) ProjectAdder {
	return setStr(b, areaParam, name)
}

// AreaID sets the parent area by UUID.
func (b *addProjectBuilder) AreaID(id string) ProjectAdder {
	return setStr(b, areaIDParam, id)
}

// Todos sets the child to-do titles.
func (b *addProjectBuilder) Todos(titles ...string) ProjectAdder {
	b.attrs.SetStrings(keyTodos, titles, "\n")
	return b
}

// Completed sets the completion status.
func (b *addProjectBuilder) Completed(completed bool) ProjectAdder {
	return setBool(b, completedParam, completed)
}

// Canceled sets the canceled status.
func (b *addProjectBuilder) Canceled(canceled bool) ProjectAdder {
	return setBool(b, canceledParam, canceled)
}

// Reveal navigates to the newly created project.
func (b *addProjectBuilder) Reveal(reveal bool) ProjectAdder {
	return setBool(b, revealParam, reveal)
}

// Reminder sets a reminder time for the project.
// The reminder is combined with the scheduling date (When).
// If no scheduling date is set, defaults to "today".
// Hour must be 0-23, minute must be 0-59.
func (b *addProjectBuilder) Reminder(hour, minute int) ProjectAdder {
	return setReminder(b, hour, minute)
}

// CreationDate sets the creation timestamp.
func (b *addProjectBuilder) CreationDate(date time.Time) ProjectAdder {
	return setTime(b, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (b *addProjectBuilder) CompletionDate(date time.Time) ProjectAdder {
	return setTime(b, completionDateParam, date)
}

// Build returns the Things URL for creating the project.
func (b *addProjectBuilder) Build() (string, error) {
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
func (b *addProjectBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.execute(ctx, uri)
}
