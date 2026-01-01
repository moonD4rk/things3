package things3

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// TodoBuilder builds URLs for creating new to-dos via the add command.
type TodoBuilder struct {
	scheme *Scheme
	attrs  urlAttrs
	err    error
}

// getStore returns the attribute store for the builder.
func (b *TodoBuilder) getStore() attrStore { return &b.attrs }

// setErr sets the error field for the builder.
func (b *TodoBuilder) setErr(err error) { b.err = err }

// Title sets the to-do title.
func (b *TodoBuilder) Title(title string) *TodoBuilder {
	return setStr(b, titleParam, title)
}

// Titles sets multiple to-do titles (creates multiple to-dos).
// Titles are newline-separated.
func (b *TodoBuilder) Titles(titles ...string) *TodoBuilder {
	combined := strings.Join(titles, "\n")
	if len(combined) > maxTitleLength {
		b.err = ErrTitleTooLong
		return b
	}
	b.attrs.SetString(keyTitles, combined)
	return b
}

// Notes sets the to-do notes/description.
func (b *TodoBuilder) Notes(notes string) *TodoBuilder {
	return setStr(b, notesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
//
// Example:
//
//	scheme.Todo().Title("Task").When(things3.Today())
//	scheme.Todo().Title("Task").When(time.Now().AddDate(0, 0, 7))
func (b *TodoBuilder) When(t time.Time) *TodoBuilder {
	return setWhenTime(b, t)
}

// WhenEvening schedules the to-do for this evening.
// This is a Things 3-specific concept that cannot be expressed as a date.
func (b *TodoBuilder) WhenEvening() *TodoBuilder {
	return setWhenStr(b, whenEvening)
}

// WhenAnytime schedules the to-do for anytime (no specific time).
// This is a Things 3-specific concept that cannot be expressed as a date.
func (b *TodoBuilder) WhenAnytime() *TodoBuilder {
	return setWhenStr(b, whenAnytime)
}

// WhenSomeday schedules the to-do for someday (indefinite future).
// This is a Things 3-specific concept that cannot be expressed as a date.
func (b *TodoBuilder) WhenSomeday() *TodoBuilder {
	return setWhenStr(b, whenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
//
// Example:
//
//	scheme.Todo().Title("Task").Deadline(time.Date(2025, 12, 31, 0, 0, 0, 0, time.Local))
func (b *TodoBuilder) Deadline(t time.Time) *TodoBuilder {
	return setDeadlineTime(b, t)
}

// Tags sets the tags for the to-do.
// Tags must already exist in Things.
func (b *TodoBuilder) Tags(tags ...string) *TodoBuilder {
	return setStrs(b, tagsParam, tags)
}

// ChecklistItems sets the checklist items.
func (b *TodoBuilder) ChecklistItems(items ...string) *TodoBuilder {
	return setStrs(b, checklistItemsParam, items)
}

// List sets the target project or area by name.
func (b *TodoBuilder) List(name string) *TodoBuilder {
	return setStr(b, listParam, name)
}

// ListID sets the target project or area by UUID.
func (b *TodoBuilder) ListID(id string) *TodoBuilder {
	return setStr(b, listIDParam, id)
}

// Heading sets the target heading within a project by name.
func (b *TodoBuilder) Heading(name string) *TodoBuilder {
	return setStr(b, headingParam, name)
}

// HeadingID sets the target heading within a project by UUID.
func (b *TodoBuilder) HeadingID(id string) *TodoBuilder {
	return setStr(b, headingIDParam, id)
}

// Completed sets the completion status.
func (b *TodoBuilder) Completed(completed bool) *TodoBuilder {
	return setBool(b, completedParam, completed)
}

// Canceled sets the canceled status.
func (b *TodoBuilder) Canceled(canceled bool) *TodoBuilder {
	return setBool(b, canceledParam, canceled)
}

// ShowQuickEntry displays the quick entry dialog instead of adding directly.
func (b *TodoBuilder) ShowQuickEntry(show bool) *TodoBuilder {
	return setBool(b, showQuickEntryParam, show)
}

// Reveal navigates to the newly created to-do.
func (b *TodoBuilder) Reveal(reveal bool) *TodoBuilder {
	return setBool(b, revealParam, reveal)
}

// Reminder sets a reminder time for the to-do.
// The reminder is combined with the scheduling date (When/WhenDate).
// If no scheduling date is set, defaults to "today".
// Hour must be 0-23, minute must be 0-59.
//
// Example:
//
//	scheme.Todo().Title("Meeting").When(WhenTomorrow).Reminder(14, 30) // tomorrow@14:30
//	scheme.Todo().Title("Call").Reminder(15, 0) // today@15:00 (defaults to today)
func (b *TodoBuilder) Reminder(hour, minute int) *TodoBuilder {
	return setReminder(b, hour, minute)
}

// CreationDate sets the creation timestamp.
// Future dates are ignored by Things.
func (b *TodoBuilder) CreationDate(date time.Time) *TodoBuilder {
	return setTime(b, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
// Future dates are ignored by Things.
func (b *TodoBuilder) CompletionDate(date time.Time) *TodoBuilder {
	return setTime(b, completionDateParam, date)
}

// Build returns the Things URL for creating the to-do.
func (b *TodoBuilder) Build() (string, error) {
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
func (b *TodoBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.execute(ctx, uri)
}

// ProjectBuilder builds URLs for creating new projects via the add-project command.
type ProjectBuilder struct {
	scheme *Scheme
	attrs  urlAttrs
	err    error
}

// getStore returns the attribute store for the builder.
func (b *ProjectBuilder) getStore() attrStore { return &b.attrs }

// setErr sets the error field for the builder.
func (b *ProjectBuilder) setErr(err error) { b.err = err }

// Title sets the project title.
func (b *ProjectBuilder) Title(title string) *ProjectBuilder {
	return setStr(b, titleParam, title)
}

// Notes sets the project notes/description.
func (b *ProjectBuilder) Notes(notes string) *ProjectBuilder {
	return setStr(b, notesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *ProjectBuilder) When(t time.Time) *ProjectBuilder {
	return setWhenTime(b, t)
}

// WhenEvening schedules the project for this evening.
func (b *ProjectBuilder) WhenEvening() *ProjectBuilder {
	return setWhenStr(b, whenEvening)
}

// WhenAnytime schedules the project for anytime (no specific time).
func (b *ProjectBuilder) WhenAnytime() *ProjectBuilder {
	return setWhenStr(b, whenAnytime)
}

// WhenSomeday schedules the project for someday (indefinite future).
func (b *ProjectBuilder) WhenSomeday() *ProjectBuilder {
	return setWhenStr(b, whenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *ProjectBuilder) Deadline(t time.Time) *ProjectBuilder {
	return setDeadlineTime(b, t)
}

// Tags sets the tags for the project.
func (b *ProjectBuilder) Tags(tags ...string) *ProjectBuilder {
	return setStrs(b, tagsParam, tags)
}

// Area sets the parent area by name.
func (b *ProjectBuilder) Area(name string) *ProjectBuilder {
	return setStr(b, areaParam, name)
}

// AreaID sets the parent area by UUID.
func (b *ProjectBuilder) AreaID(id string) *ProjectBuilder {
	return setStr(b, areaIDParam, id)
}

// Todos sets the child to-do titles.
func (b *ProjectBuilder) Todos(titles ...string) *ProjectBuilder {
	b.attrs.SetStrings(keyTodos, titles, "\n")
	return b
}

// Completed sets the completion status.
func (b *ProjectBuilder) Completed(completed bool) *ProjectBuilder {
	return setBool(b, completedParam, completed)
}

// Canceled sets the canceled status.
func (b *ProjectBuilder) Canceled(canceled bool) *ProjectBuilder {
	return setBool(b, canceledParam, canceled)
}

// Reveal navigates to the newly created project.
func (b *ProjectBuilder) Reveal(reveal bool) *ProjectBuilder {
	return setBool(b, revealParam, reveal)
}

// Reminder sets a reminder time for the project.
// The reminder is combined with the scheduling date (When).
// If no scheduling date is set, defaults to "today".
// Hour must be 0-23, minute must be 0-59.
func (b *ProjectBuilder) Reminder(hour, minute int) *ProjectBuilder {
	return setReminder(b, hour, minute)
}

// CreationDate sets the creation timestamp.
func (b *ProjectBuilder) CreationDate(date time.Time) *ProjectBuilder {
	return setTime(b, creationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (b *ProjectBuilder) CompletionDate(date time.Time) *ProjectBuilder {
	return setTime(b, completionDateParam, date)
}

// Build returns the Things URL for creating the project.
func (b *ProjectBuilder) Build() (string, error) {
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
func (b *ProjectBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.execute(ctx, uri)
}
