package things3

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// TodoBuilder builds URLs for creating new to-dos via the add command.
type TodoBuilder struct {
	params map[string]string
	err    error
}

// Title sets the to-do title.
func (b *TodoBuilder) Title(title string) *TodoBuilder {
	if len(title) > maxTitleLength {
		b.err = ErrTitleTooLong
		return b
	}
	b.params["title"] = title
	return b
}

// Titles sets multiple to-do titles (creates multiple to-dos).
// Titles are newline-separated.
func (b *TodoBuilder) Titles(titles ...string) *TodoBuilder {
	combined := strings.Join(titles, "\n")
	if len(combined) > maxTitleLength {
		b.err = ErrTitleTooLong
		return b
	}
	b.params["titles"] = combined
	return b
}

// Notes sets the to-do notes/description.
func (b *TodoBuilder) Notes(notes string) *TodoBuilder {
	if len(notes) > maxNotesLength {
		b.err = ErrNotesTooLong
		return b
	}
	b.params["notes"] = notes
	return b
}

// When sets the scheduling date.
func (b *TodoBuilder) When(when When) *TodoBuilder {
	b.params["when"] = string(when)
	return b
}

// WhenDate sets a specific date for scheduling.
func (b *TodoBuilder) WhenDate(year int, month time.Month, day int) *TodoBuilder {
	b.params["when"] = fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
	return b
}

// Deadline sets the deadline date in yyyy-mm-dd format.
func (b *TodoBuilder) Deadline(date string) *TodoBuilder {
	b.params["deadline"] = date
	return b
}

// Tags sets the tags for the to-do.
// Tags must already exist in Things.
func (b *TodoBuilder) Tags(tags ...string) *TodoBuilder {
	b.params["tags"] = strings.Join(tags, ",")
	return b
}

// ChecklistItems sets the checklist items.
func (b *TodoBuilder) ChecklistItems(items ...string) *TodoBuilder {
	if len(items) > maxChecklistItems {
		b.err = ErrTooManyChecklistItems
		return b
	}
	b.params["checklist-items"] = strings.Join(items, "\n")
	return b
}

// List sets the target project or area by name.
func (b *TodoBuilder) List(name string) *TodoBuilder {
	b.params["list"] = name
	return b
}

// ListID sets the target project or area by UUID.
func (b *TodoBuilder) ListID(id string) *TodoBuilder {
	b.params["list-id"] = id
	return b
}

// Heading sets the target heading within a project by name.
func (b *TodoBuilder) Heading(name string) *TodoBuilder {
	b.params["heading"] = name
	return b
}

// HeadingID sets the target heading within a project by UUID.
func (b *TodoBuilder) HeadingID(id string) *TodoBuilder {
	b.params["heading-id"] = id
	return b
}

// Completed sets the completion status.
func (b *TodoBuilder) Completed(completed bool) *TodoBuilder {
	b.params["completed"] = fmt.Sprintf("%t", completed)
	return b
}

// Canceled sets the canceled status.
func (b *TodoBuilder) Canceled(canceled bool) *TodoBuilder {
	b.params["canceled"] = fmt.Sprintf("%t", canceled)
	return b
}

// ShowQuickEntry displays the quick entry dialog instead of adding directly.
func (b *TodoBuilder) ShowQuickEntry(show bool) *TodoBuilder {
	b.params["show-quick-entry"] = fmt.Sprintf("%t", show)
	return b
}

// Reveal navigates to the newly created to-do.
func (b *TodoBuilder) Reveal(reveal bool) *TodoBuilder {
	b.params["reveal"] = fmt.Sprintf("%t", reveal)
	return b
}

// CreationDate sets the creation timestamp.
// Future dates are ignored by Things.
func (b *TodoBuilder) CreationDate(date time.Time) *TodoBuilder {
	b.params["creation-date"] = date.Format(time.RFC3339)
	return b
}

// CompletionDate sets the completion timestamp.
// Future dates are ignored by Things.
func (b *TodoBuilder) CompletionDate(date time.Time) *TodoBuilder {
	b.params["completion-date"] = date.Format(time.RFC3339)
	return b
}

// Build returns the Things URL for creating the to-do.
func (b *TodoBuilder) Build() (string, error) {
	if b.err != nil {
		return "", b.err
	}

	query := url.Values{}
	for k, v := range b.params {
		query.Set(k, v)
	}

	return fmt.Sprintf("things:///%s?%s", CommandAdd, query.Encode()), nil
}

// ProjectBuilder builds URLs for creating new projects via the add-project command.
type ProjectBuilder struct {
	params map[string]string
	err    error
}

// Title sets the project title.
func (b *ProjectBuilder) Title(title string) *ProjectBuilder {
	if len(title) > maxTitleLength {
		b.err = ErrTitleTooLong
		return b
	}
	b.params["title"] = title
	return b
}

// Notes sets the project notes/description.
func (b *ProjectBuilder) Notes(notes string) *ProjectBuilder {
	if len(notes) > maxNotesLength {
		b.err = ErrNotesTooLong
		return b
	}
	b.params["notes"] = notes
	return b
}

// When sets the scheduling date.
func (b *ProjectBuilder) When(when When) *ProjectBuilder {
	b.params["when"] = string(when)
	return b
}

// WhenDate sets a specific date for scheduling.
func (b *ProjectBuilder) WhenDate(year int, month time.Month, day int) *ProjectBuilder {
	b.params["when"] = fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
	return b
}

// Deadline sets the deadline date in yyyy-mm-dd format.
func (b *ProjectBuilder) Deadline(date string) *ProjectBuilder {
	b.params["deadline"] = date
	return b
}

// Tags sets the tags for the project.
func (b *ProjectBuilder) Tags(tags ...string) *ProjectBuilder {
	b.params["tags"] = strings.Join(tags, ",")
	return b
}

// Area sets the parent area by name.
func (b *ProjectBuilder) Area(name string) *ProjectBuilder {
	b.params["area"] = name
	return b
}

// AreaID sets the parent area by UUID.
func (b *ProjectBuilder) AreaID(id string) *ProjectBuilder {
	b.params["area-id"] = id
	return b
}

// Todos sets the child to-do titles.
func (b *ProjectBuilder) Todos(titles ...string) *ProjectBuilder {
	b.params["to-dos"] = strings.Join(titles, "\n")
	return b
}

// Completed sets the completion status.
func (b *ProjectBuilder) Completed(completed bool) *ProjectBuilder {
	b.params["completed"] = fmt.Sprintf("%t", completed)
	return b
}

// Canceled sets the canceled status.
func (b *ProjectBuilder) Canceled(canceled bool) *ProjectBuilder {
	b.params["canceled"] = fmt.Sprintf("%t", canceled)
	return b
}

// Reveal navigates to the newly created project.
func (b *ProjectBuilder) Reveal(reveal bool) *ProjectBuilder {
	b.params["reveal"] = fmt.Sprintf("%t", reveal)
	return b
}

// CreationDate sets the creation timestamp.
func (b *ProjectBuilder) CreationDate(date time.Time) *ProjectBuilder {
	b.params["creation-date"] = date.Format(time.RFC3339)
	return b
}

// CompletionDate sets the completion timestamp.
func (b *ProjectBuilder) CompletionDate(date time.Time) *ProjectBuilder {
	b.params["completion-date"] = date.Format(time.RFC3339)
	return b
}

// Build returns the Things URL for creating the project.
func (b *ProjectBuilder) Build() (string, error) {
	if b.err != nil {
		return "", b.err
	}

	query := url.Values{}
	for k, v := range b.params {
		query.Set(k, v)
	}

	return fmt.Sprintf("things:///%s?%s", CommandAddProject, query.Encode()), nil
}
