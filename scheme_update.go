package things3

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// UpdateTodoBuilder builds URLs for updating existing to-dos via the update command.
// Requires authentication token (obtained via AuthScheme).
type UpdateTodoBuilder struct {
	token  string
	id     string
	params map[string]string
	err    error
}

// Title replaces the to-do title.
func (b *UpdateTodoBuilder) Title(title string) *UpdateTodoBuilder {
	if len(title) > maxTitleLength {
		b.err = ErrTitleTooLong
		return b
	}
	b.params["title"] = title
	return b
}

// Notes replaces the to-do notes.
func (b *UpdateTodoBuilder) Notes(notes string) *UpdateTodoBuilder {
	if len(notes) > maxNotesLength {
		b.err = ErrNotesTooLong
		return b
	}
	b.params["notes"] = notes
	return b
}

// PrependNotes prepends text to existing notes.
func (b *UpdateTodoBuilder) PrependNotes(notes string) *UpdateTodoBuilder {
	b.params["prepend-notes"] = notes
	return b
}

// AppendNotes appends text to existing notes.
func (b *UpdateTodoBuilder) AppendNotes(notes string) *UpdateTodoBuilder {
	b.params["append-notes"] = notes
	return b
}

// When sets the scheduling date.
func (b *UpdateTodoBuilder) When(when When) *UpdateTodoBuilder {
	b.params["when"] = string(when)
	return b
}

// WhenDate sets a specific date for scheduling.
func (b *UpdateTodoBuilder) WhenDate(year int, month time.Month, day int) *UpdateTodoBuilder {
	b.params["when"] = fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
	return b
}

// Deadline sets the deadline date.
// Pass empty string to clear the deadline.
func (b *UpdateTodoBuilder) Deadline(date string) *UpdateTodoBuilder {
	b.params["deadline"] = date
	return b
}

// ClearDeadline removes the deadline.
func (b *UpdateTodoBuilder) ClearDeadline() *UpdateTodoBuilder {
	b.params["deadline"] = ""
	return b
}

// Tags replaces all tags.
func (b *UpdateTodoBuilder) Tags(tags ...string) *UpdateTodoBuilder {
	b.params["tags"] = strings.Join(tags, ",")
	return b
}

// AddTags adds tags without replacing existing ones.
func (b *UpdateTodoBuilder) AddTags(tags ...string) *UpdateTodoBuilder {
	b.params["add-tags"] = strings.Join(tags, ",")
	return b
}

// ChecklistItems replaces all checklist items.
func (b *UpdateTodoBuilder) ChecklistItems(items ...string) *UpdateTodoBuilder {
	if len(items) > maxChecklistItems {
		b.err = ErrTooManyChecklistItems
		return b
	}
	b.params["checklist-items"] = strings.Join(items, "\n")
	return b
}

// PrependChecklistItems prepends checklist items.
func (b *UpdateTodoBuilder) PrependChecklistItems(items ...string) *UpdateTodoBuilder {
	b.params["prepend-checklist-items"] = strings.Join(items, "\n")
	return b
}

// AppendChecklistItems appends checklist items.
func (b *UpdateTodoBuilder) AppendChecklistItems(items ...string) *UpdateTodoBuilder {
	b.params["append-checklist-items"] = strings.Join(items, "\n")
	return b
}

// List moves the to-do to a project or area by name.
func (b *UpdateTodoBuilder) List(name string) *UpdateTodoBuilder {
	b.params["list"] = name
	return b
}

// ListID moves the to-do to a project or area by UUID.
func (b *UpdateTodoBuilder) ListID(id string) *UpdateTodoBuilder {
	b.params["list-id"] = id
	return b
}

// Heading moves the to-do to a heading by name.
func (b *UpdateTodoBuilder) Heading(name string) *UpdateTodoBuilder {
	b.params["heading"] = name
	return b
}

// HeadingID moves the to-do to a heading by UUID.
func (b *UpdateTodoBuilder) HeadingID(id string) *UpdateTodoBuilder {
	b.params["heading-id"] = id
	return b
}

// Completed sets the completion status.
func (b *UpdateTodoBuilder) Completed(completed bool) *UpdateTodoBuilder {
	b.params["completed"] = fmt.Sprintf("%t", completed)
	return b
}

// Canceled sets the canceled status.
func (b *UpdateTodoBuilder) Canceled(canceled bool) *UpdateTodoBuilder {
	b.params["canceled"] = fmt.Sprintf("%t", canceled)
	return b
}

// Duplicate duplicates the to-do before updating.
func (b *UpdateTodoBuilder) Duplicate(duplicate bool) *UpdateTodoBuilder {
	b.params["duplicate"] = fmt.Sprintf("%t", duplicate)
	return b
}

// Reveal navigates to the to-do after updating.
func (b *UpdateTodoBuilder) Reveal(reveal bool) *UpdateTodoBuilder {
	b.params["reveal"] = fmt.Sprintf("%t", reveal)
	return b
}

// CreationDate sets the creation timestamp.
func (b *UpdateTodoBuilder) CreationDate(date time.Time) *UpdateTodoBuilder {
	b.params["creation-date"] = date.Format(time.RFC3339)
	return b
}

// CompletionDate sets the completion timestamp.
func (b *UpdateTodoBuilder) CompletionDate(date time.Time) *UpdateTodoBuilder {
	b.params["completion-date"] = date.Format(time.RFC3339)
	return b
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

	query := url.Values{}
	query.Set("id", b.id)
	query.Set("auth-token", b.token)
	for k, v := range b.params {
		query.Set(k, v)
	}

	return fmt.Sprintf("things:///%s?%s", CommandUpdate, query.Encode()), nil
}

// UpdateProjectBuilder builds URLs for updating existing projects via the update-project command.
// Requires authentication token (obtained via AuthScheme).
type UpdateProjectBuilder struct {
	token  string
	id     string
	params map[string]string
	err    error
}

// Title replaces the project title.
func (b *UpdateProjectBuilder) Title(title string) *UpdateProjectBuilder {
	if len(title) > maxTitleLength {
		b.err = ErrTitleTooLong
		return b
	}
	b.params["title"] = title
	return b
}

// Notes replaces the project notes.
func (b *UpdateProjectBuilder) Notes(notes string) *UpdateProjectBuilder {
	if len(notes) > maxNotesLength {
		b.err = ErrNotesTooLong
		return b
	}
	b.params["notes"] = notes
	return b
}

// PrependNotes prepends text to existing notes.
func (b *UpdateProjectBuilder) PrependNotes(notes string) *UpdateProjectBuilder {
	b.params["prepend-notes"] = notes
	return b
}

// AppendNotes appends text to existing notes.
func (b *UpdateProjectBuilder) AppendNotes(notes string) *UpdateProjectBuilder {
	b.params["append-notes"] = notes
	return b
}

// When sets the scheduling date.
func (b *UpdateProjectBuilder) When(when When) *UpdateProjectBuilder {
	b.params["when"] = string(when)
	return b
}

// WhenDate sets a specific date for scheduling.
func (b *UpdateProjectBuilder) WhenDate(year int, month time.Month, day int) *UpdateProjectBuilder {
	b.params["when"] = fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
	return b
}

// Deadline sets the deadline date.
func (b *UpdateProjectBuilder) Deadline(date string) *UpdateProjectBuilder {
	b.params["deadline"] = date
	return b
}

// ClearDeadline removes the deadline.
func (b *UpdateProjectBuilder) ClearDeadline() *UpdateProjectBuilder {
	b.params["deadline"] = ""
	return b
}

// Tags replaces all tags.
func (b *UpdateProjectBuilder) Tags(tags ...string) *UpdateProjectBuilder {
	b.params["tags"] = strings.Join(tags, ",")
	return b
}

// AddTags adds tags without replacing existing ones.
func (b *UpdateProjectBuilder) AddTags(tags ...string) *UpdateProjectBuilder {
	b.params["add-tags"] = strings.Join(tags, ",")
	return b
}

// Area moves the project to an area by name.
func (b *UpdateProjectBuilder) Area(name string) *UpdateProjectBuilder {
	b.params["area"] = name
	return b
}

// AreaID moves the project to an area by UUID.
func (b *UpdateProjectBuilder) AreaID(id string) *UpdateProjectBuilder {
	b.params["area-id"] = id
	return b
}

// Completed sets the completion status.
// Note: Setting completed=true is ignored unless all child to-dos
// are completed or canceled and all headings are archived.
func (b *UpdateProjectBuilder) Completed(completed bool) *UpdateProjectBuilder {
	b.params["completed"] = fmt.Sprintf("%t", completed)
	return b
}

// Canceled sets the canceled status.
func (b *UpdateProjectBuilder) Canceled(canceled bool) *UpdateProjectBuilder {
	b.params["canceled"] = fmt.Sprintf("%t", canceled)
	return b
}

// Reveal navigates to the project after updating.
func (b *UpdateProjectBuilder) Reveal(reveal bool) *UpdateProjectBuilder {
	b.params["reveal"] = fmt.Sprintf("%t", reveal)
	return b
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

	query := url.Values{}
	query.Set("id", b.id)
	query.Set("auth-token", b.token)
	for k, v := range b.params {
		query.Set(k, v)
	}

	return fmt.Sprintf("things:///%s?%s", CommandUpdateProject, query.Encode()), nil
}
