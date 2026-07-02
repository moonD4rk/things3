package scheme

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// addTodoBuilder builds URLs for creating new todos via the add command.
type addTodoBuilder struct {
	scheme *Scheme
	attrs  URLAttrs
	err    error
}

// NewTodoAdder creates a new TodoAdder for creating a new todo.
func NewTodoAdder(s *Scheme) TodoAdder {
	return &addTodoBuilder{scheme: s, attrs: NewURLAttrs()}
}

// GetStore returns the attribute store for the builder.
func (b *addTodoBuilder) GetStore() AttrStore { return &b.attrs }

// SetErr sets the error field for the builder.
func (b *addTodoBuilder) SetErr(err error) { b.err = err }

// Title sets the todo title.
func (b *addTodoBuilder) Title(title string) TodoAdder {
	return SetStr(b, TitleParam, title)
}

// Titles sets multiple todo titles (creates multiple todos).
// Titles are newline-separated in the URL, so each title must not contain
// a newline and each title is limited to MaxTitleLength characters.
func (b *addTodoBuilder) Titles(titles ...string) TodoAdder {
	for _, title := range titles {
		if utf8.RuneCountInString(title) > MaxTitleLength {
			b.err = ErrTitleTooLong
			return b
		}
		if strings.Contains(title, "\n") {
			b.err = ErrTitleContainsNewline
			return b
		}
	}
	b.attrs.SetString(KeyTitles, strings.Join(titles, "\n"))
	return b
}

// Notes sets the todo notes/description.
func (b *addTodoBuilder) Notes(notes string) TodoAdder {
	return SetStr(b, NotesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *addTodoBuilder) When(t time.Time) TodoAdder {
	return SetWhenTime(b, t)
}

// WhenEvening schedules the todo for this evening.
// This is a Things 3-specific concept that cannot be expressed as a date.
func (b *addTodoBuilder) WhenEvening() TodoAdder {
	return SetWhenStr(b, WhenEvening)
}

// WhenAnytime schedules the todo for anytime (no specific time).
// This is a Things 3-specific concept that cannot be expressed as a date.
func (b *addTodoBuilder) WhenAnytime() TodoAdder {
	return SetWhenStr(b, WhenAnytime)
}

// WhenSomeday schedules the todo for someday (indefinite future).
// This is a Things 3-specific concept that cannot be expressed as a date.
func (b *addTodoBuilder) WhenSomeday() TodoAdder {
	return SetWhenStr(b, WhenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *addTodoBuilder) Deadline(t time.Time) TodoAdder {
	return SetDeadlineTime(b, t)
}

// Tags sets the tags for the todo.
// Tags must already exist in Things.
func (b *addTodoBuilder) Tags(tags ...string) TodoAdder {
	return SetStrs(b, TagsParam, tags)
}

// ChecklistItems sets the checklist items.
func (b *addTodoBuilder) ChecklistItems(items ...string) TodoAdder {
	return SetStrs(b, ChecklistItemsParam, items)
}

// List sets the target project or area by name.
func (b *addTodoBuilder) List(name string) TodoAdder {
	return SetStr(b, ListParam, name)
}

// ListID sets the target project or area by UUID.
func (b *addTodoBuilder) ListID(id string) TodoAdder {
	return SetStr(b, ListIDParam, id)
}

// Heading sets the target heading within a project by name.
func (b *addTodoBuilder) Heading(name string) TodoAdder {
	return SetStr(b, HeadingParam, name)
}

// HeadingID sets the target heading within a project by UUID.
func (b *addTodoBuilder) HeadingID(id string) TodoAdder {
	return SetStr(b, HeadingIDParam, id)
}

// Completed sets the completion status.
func (b *addTodoBuilder) Completed(completed bool) TodoAdder {
	return SetBool(b, CompletedParam, completed)
}

// Canceled sets the canceled status.
func (b *addTodoBuilder) Canceled(canceled bool) TodoAdder {
	return SetBool(b, CanceledParam, canceled)
}

// ShowQuickEntry displays the quick entry dialog instead of adding directly.
func (b *addTodoBuilder) ShowQuickEntry(show bool) TodoAdder {
	return SetBool(b, ShowQuickEntryParam, show)
}

// Reveal navigates to the newly created todo.
func (b *addTodoBuilder) Reveal(reveal bool) TodoAdder {
	return SetBool(b, RevealParam, reveal)
}

// Reminder sets a reminder time for the todo.
// The reminder is combined with the scheduling date (When/WhenEvening).
// If no scheduling date is set, defaults to "today".
// Hour must be 0-23, minute must be 0-59.
// Combining a reminder with WhenSomeday or WhenAnytime is a build error.
func (b *addTodoBuilder) Reminder(hour, minute int) TodoAdder {
	return SetReminder(b, hour, minute)
}

// CreationDate sets the creation timestamp.
// Future dates are ignored by Things.
func (b *addTodoBuilder) CreationDate(date time.Time) TodoAdder {
	return SetTime(b, CreationDateParam, date)
}

// CompletionDate sets the completion timestamp.
// Future dates are ignored by Things.
func (b *addTodoBuilder) CompletionDate(date time.Time) TodoAdder {
	return SetTime(b, CompletionDateParam, date)
}

// Build returns the Things URL for creating the todo.
// Build is pure: it never mutates the builder, so it can be called
// repeatedly (including via Execute) with identical results.
func (b *addTodoBuilder) Build() (string, error) {
	if b.err != nil {
		return "", b.err
	}

	query, err := b.attrs.QueryValues()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("things:///%s?%s", CommandAdd, EncodeQuery(query)), nil
}

// Execute builds and executes the add URL.
// Returns an error if the URL cannot be built or executed.
func (b *addTodoBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.Execute(ctx, uri)
}

// addProjectBuilder builds URLs for creating new projects via the add-project command.
type addProjectBuilder struct {
	scheme *Scheme
	attrs  URLAttrs
	err    error
}

// NewProjectAdder creates a new ProjectAdder for creating a new project.
func NewProjectAdder(s *Scheme) ProjectAdder {
	return &addProjectBuilder{scheme: s, attrs: NewURLAttrs()}
}

// GetStore returns the attribute store for the builder.
func (b *addProjectBuilder) GetStore() AttrStore { return &b.attrs }

// SetErr sets the error field for the builder.
func (b *addProjectBuilder) SetErr(err error) { b.err = err }

// Title sets the project title.
func (b *addProjectBuilder) Title(title string) ProjectAdder {
	return SetStr(b, TitleParam, title)
}

// Notes sets the project notes/description.
func (b *addProjectBuilder) Notes(notes string) ProjectAdder {
	return SetStr(b, NotesParam, notes)
}

// When sets the scheduling date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *addProjectBuilder) When(t time.Time) ProjectAdder {
	return SetWhenTime(b, t)
}

// WhenEvening schedules the project for this evening.
func (b *addProjectBuilder) WhenEvening() ProjectAdder {
	return SetWhenStr(b, WhenEvening)
}

// WhenAnytime schedules the project for anytime (no specific time).
func (b *addProjectBuilder) WhenAnytime() ProjectAdder {
	return SetWhenStr(b, WhenAnytime)
}

// WhenSomeday schedules the project for someday (indefinite future).
func (b *addProjectBuilder) WhenSomeday() ProjectAdder {
	return SetWhenStr(b, WhenSomeday)
}

// Deadline sets the deadline date using a time.Time value.
// The date portion is used; time-of-day is ignored.
func (b *addProjectBuilder) Deadline(t time.Time) ProjectAdder {
	return SetDeadlineTime(b, t)
}

// Tags sets the tags for the project.
func (b *addProjectBuilder) Tags(tags ...string) ProjectAdder {
	return SetStrs(b, TagsParam, tags)
}

// Area sets the parent area by name.
func (b *addProjectBuilder) Area(name string) ProjectAdder {
	return SetStr(b, AreaParam, name)
}

// AreaID sets the parent area by UUID.
func (b *addProjectBuilder) AreaID(id string) ProjectAdder {
	return SetStr(b, AreaIDParam, id)
}

// Todos sets the child todo titles.
// Titles are newline-separated in the URL, so each title must not contain a newline.
func (b *addProjectBuilder) Todos(titles ...string) ProjectAdder {
	return SetStrs(b, TodosParam, titles)
}

// Completed sets the completion status.
func (b *addProjectBuilder) Completed(completed bool) ProjectAdder {
	return SetBool(b, CompletedParam, completed)
}

// Canceled sets the canceled status.
func (b *addProjectBuilder) Canceled(canceled bool) ProjectAdder {
	return SetBool(b, CanceledParam, canceled)
}

// Reveal navigates to the newly created project.
func (b *addProjectBuilder) Reveal(reveal bool) ProjectAdder {
	return SetBool(b, RevealParam, reveal)
}

// Reminder sets a reminder time for the project.
// The reminder is combined with the scheduling date (When).
// If no scheduling date is set, defaults to "today".
// Hour must be 0-23, minute must be 0-59.
// Combining a reminder with WhenSomeday or WhenAnytime is a build error.
func (b *addProjectBuilder) Reminder(hour, minute int) ProjectAdder {
	return SetReminder(b, hour, minute)
}

// CreationDate sets the creation timestamp.
func (b *addProjectBuilder) CreationDate(date time.Time) ProjectAdder {
	return SetTime(b, CreationDateParam, date)
}

// CompletionDate sets the completion timestamp.
func (b *addProjectBuilder) CompletionDate(date time.Time) ProjectAdder {
	return SetTime(b, CompletionDateParam, date)
}

// Build returns the Things URL for creating the project.
// Build is pure: it never mutates the builder, so it can be called
// repeatedly (including via Execute) with identical results.
func (b *addProjectBuilder) Build() (string, error) {
	if b.err != nil {
		return "", b.err
	}

	query, err := b.attrs.QueryValues()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("things:///%s?%s", CommandAddProject, EncodeQuery(query)), nil
}

// Execute builds and executes the add-project URL.
// Returns an error if the URL cannot be built or executed.
func (b *addProjectBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.Execute(ctx, uri)
}
