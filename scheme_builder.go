package things3

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// TodoBuilder builds URLs for creating new to-dos via the add command.
type TodoBuilder struct {
	attrs urlAttrs
	err   error
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

// When sets the scheduling date.
func (b *TodoBuilder) When(when When) *TodoBuilder {
	return setWhenStr(b, when)
}

// WhenDate sets a specific date for scheduling.
func (b *TodoBuilder) WhenDate(year int, month time.Month, day int) *TodoBuilder {
	return setDate(b, whenParam, year, month, day)
}

// Deadline sets the deadline date in yyyy-mm-dd format.
func (b *TodoBuilder) Deadline(date string) *TodoBuilder {
	return setStr(b, deadlineParam, date)
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

	query := url.Values{}
	for k, v := range b.attrs.params {
		query.Set(k, v)
	}

	return fmt.Sprintf("things:///%s?%s", CommandAdd, query.Encode()), nil
}

// ProjectBuilder builds URLs for creating new projects via the add-project command.
type ProjectBuilder struct {
	attrs urlAttrs
	err   error
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

// When sets the scheduling date.
func (b *ProjectBuilder) When(when When) *ProjectBuilder {
	return setWhenStr(b, when)
}

// WhenDate sets a specific date for scheduling.
func (b *ProjectBuilder) WhenDate(year int, month time.Month, day int) *ProjectBuilder {
	return setDate(b, whenParam, year, month, day)
}

// Deadline sets the deadline date in yyyy-mm-dd format.
func (b *ProjectBuilder) Deadline(date string) *ProjectBuilder {
	return setStr(b, deadlineParam, date)
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

	query := url.Values{}
	for k, v := range b.attrs.params {
		query.Set(k, v)
	}

	return fmt.Sprintf("things:///%s?%s", CommandAddProject, query.Encode()), nil
}
