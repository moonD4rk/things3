package things3

import (
	"fmt"
	"strings"
	"time"
)

// attrStore abstracts attribute storage for URL params vs JSON attributes.
// URL builders store everything as strings, JSON builders preserve native types.
type attrStore interface {
	// SetString sets a string attribute.
	SetString(key, value string)
	// SetBool sets a boolean attribute.
	SetBool(key string, value bool)
	// SetStrings sets a string slice attribute.
	// For URL storage, values are joined with separator.
	// For JSON storage, values are stored as array.
	SetStrings(key string, values []string, separator string)
	// SetTime sets a time attribute in RFC3339 format.
	SetTime(key string, t time.Time)
	// SetDate sets a date attribute in yyyy-mm-dd format.
	SetDate(key string, year int, month time.Month, day int)
}

// attrBuilder is the interface that all URL scheme builders must implement.
// It provides access to the underlying attribute store and error handling.
//
// Implementations:
//   - URL Builders: TodoBuilder, ProjectBuilder, UpdateTodoBuilder, UpdateProjectBuilder
//   - JSON Builders: JSONBuilder, AuthJSONBuilder
//   - JSON Entry Builders: JSONTodoBuilder, JSONProjectBuilder (used within JSON batch operations)
type attrBuilder interface {
	getStore() attrStore
	setErr(err error)
}

// Parameter definition types for type-safe attribute setters.
// These define the validation rules and URL parameter keys for each attribute.

// strParam defines metadata for string URL parameters including validation rules.
type strParam struct {
	key    string // URL parameter key (e.g., "title", "notes")
	maxLen int    // Maximum length (0 = no limit)
	err    error  // Error to return if validation fails
}

// boolParam defines metadata for boolean URL parameters.
type boolParam struct {
	key string // URL parameter key
}

// strsParam defines metadata for string slice URL parameters.
type strsParam struct {
	key      string // URL parameter key
	sep      string // Separator for joining values (e.g., "," for tags, "\n" for checklist)
	maxCount int    // Maximum item count (0 = no limit)
	err      error  // Error to return if validation fails
}

// timeParam defines metadata for time URL parameters.
type timeParam struct {
	key string // URL parameter key
}

// dateParam defines metadata for date URL parameters.
type dateParam struct {
	key string // URL parameter key
}

// Predefined parameter definitions for Things URL scheme.
// These map builder methods to URL parameter keys with validation rules.
var (
	// String parameters
	titleParam        = strParam{key: keyTitle, maxLen: maxTitleLength, err: ErrTitleTooLong}
	notesParam        = strParam{key: keyNotes, maxLen: maxNotesLength, err: ErrNotesTooLong}
	deadlineParam     = strParam{key: keyDeadline}
	listParam         = strParam{key: keyList}
	listIDParam       = strParam{key: keyListID}
	headingParam      = strParam{key: keyHeading}
	headingIDParam    = strParam{key: keyHeadingID}
	areaParam         = strParam{key: keyArea}
	areaIDParam       = strParam{key: keyAreaID}
	prependNotesParam = strParam{key: keyPrependNotes}
	appendNotesParam  = strParam{key: keyAppendNotes}

	// Boolean parameters
	completedParam      = boolParam{key: keyCompleted}
	canceledParam       = boolParam{key: keyCanceled}
	revealParam         = boolParam{key: keyReveal}
	duplicateParam      = boolParam{key: keyDuplicate}
	showQuickEntryParam = boolParam{key: keyShowQuickEntry}

	// String slice parameters
	tagsParam             = strsParam{key: keyTags, sep: ","}
	addTagsParam          = strsParam{key: keyAddTags, sep: ","}
	checklistItemsParam   = strsParam{key: keyChecklistItems, sep: "\n", maxCount: maxChecklistItems, err: ErrTooManyChecklistItems}
	prependChecklistParam = strsParam{key: keyPrependChecklistItems, sep: "\n"}
	appendChecklistParam  = strsParam{key: keyAppendChecklistItems, sep: "\n"}

	// Time parameters
	creationDateParam   = timeParam{key: keyCreationDate}
	completionDateParam = timeParam{key: keyCompletionDate}

	// Date parameters
	whenParam = dateParam{key: keyWhen}
)

// Generic setter functions with type-safe parameter definitions.
// These provide a unified way to set attributes across all builder types.

// setStr sets a string attribute with optional length validation.
func setStr[T attrBuilder](b T, p strParam, value string) T {
	if p.maxLen > 0 && len(value) > p.maxLen {
		b.setErr(p.err)
		return b
	}
	b.getStore().SetString(p.key, value)
	return b
}

// setBool sets a boolean attribute.
func setBool[T attrBuilder](b T, p boolParam, value bool) T {
	b.getStore().SetBool(p.key, value)
	return b
}

// setStrs sets a string slice attribute with optional count validation.
func setStrs[T attrBuilder](b T, p strsParam, values []string) T {
	if p.maxCount > 0 && len(values) > p.maxCount {
		b.setErr(p.err)
		return b
	}
	b.getStore().SetStrings(p.key, values, p.sep)
	return b
}

// setTime sets a time attribute.
func setTime[T attrBuilder](b T, p timeParam, t time.Time) T {
	b.getStore().SetTime(p.key, t)
	return b
}

// setDate sets a date attribute.
func setDate[T attrBuilder](b T, p dateParam, year int, month time.Month, day int) T {
	b.getStore().SetDate(p.key, year, month, day)
	return b
}

// setWhenStr sets the when attribute using a when constant (internal use only).
func setWhenStr[T attrBuilder](b T, w when) T {
	b.getStore().SetString(keyWhen, string(w))
	return b
}

// setWhenTime sets the when attribute using a time.Time value.
// The time is formatted as yyyy-mm-dd for the Things URL scheme.
func setWhenTime[T attrBuilder](b T, t time.Time) T {
	b.getStore().SetDate(keyWhen, t.Year(), t.Month(), t.Day())
	return b
}

// setDeadlineTime sets the deadline attribute using a time.Time value.
// The time is formatted as yyyy-mm-dd for the Things URL scheme.
func setDeadlineTime[T attrBuilder](b T, t time.Time) T {
	formatted := fmt.Sprintf("%04d-%02d-%02d", t.Year(), int(t.Month()), t.Day())
	b.getStore().SetString(keyDeadline, formatted)
	return b
}

// reminderStore is implemented by builders that support reminder functionality.
type reminderStore interface {
	SetReminder(hour, minute int)
}

// setReminder sets the reminder time for builders that support it.
func setReminder[T attrBuilder](b T, hour, minute int) T {
	if hour < 0 || hour > 23 {
		b.setErr(ErrInvalidReminderTime)
		return b
	}
	if minute < 0 || minute > 59 {
		b.setErr(ErrInvalidReminderTime)
		return b
	}
	if store, ok := b.getStore().(reminderStore); ok {
		store.SetReminder(hour, minute)
	}
	return b
}

// urlAttrs stores attributes as URL query parameters (all strings).
type urlAttrs struct {
	params       map[string]string
	reminderHour *int // nil means not set
	reminderMin  *int // nil means not set
}

// SetString sets a string parameter.
func (u *urlAttrs) SetString(key, value string) {
	u.params[key] = value
}

// SetBool sets a boolean parameter as "true" or "false" string.
func (u *urlAttrs) SetBool(key string, value bool) {
	u.params[key] = fmt.Sprintf("%t", value)
}

// SetStrings joins values with separator and stores as string.
func (u *urlAttrs) SetStrings(key string, values []string, separator string) {
	u.params[key] = strings.Join(values, separator)
}

// SetTime formats time as RFC3339 string.
func (u *urlAttrs) SetTime(key string, t time.Time) {
	u.params[key] = t.Format(time.RFC3339)
}

// SetDate formats date as yyyy-mm-dd string.
func (u *urlAttrs) SetDate(key string, year int, month time.Month, day int) {
	u.params[key] = fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
}

// SetReminder sets the reminder time (hour and minute).
func (u *urlAttrs) SetReminder(hour, minute int) {
	u.reminderHour = &hour
	u.reminderMin = &minute
}

// FinalizeWhen returns the final "when" parameter value with reminder time appended.
// If reminder is set but no "when" value exists, defaults to "today".
// Format: "when@HH:MM" (e.g., "today@15:30", "2024-03-15@14:00")
func (u *urlAttrs) FinalizeWhen() {
	if u.reminderHour == nil {
		return
	}

	w, exists := u.params[keyWhen]
	if !exists || w == "" {
		w = "today" // default to today if no when specified
	}

	// Append reminder time in HH:MM format
	u.params[keyWhen] = fmt.Sprintf("%s@%02d:%02d", w, *u.reminderHour, *u.reminderMin)
}

// jsonAttrs stores attributes as JSON values (native types).
type jsonAttrs struct {
	attrs map[string]any
}

// SetString sets a string attribute.
func (j *jsonAttrs) SetString(key, value string) {
	j.attrs[key] = value
}

// SetBool sets a boolean attribute (native bool).
func (j *jsonAttrs) SetBool(key string, value bool) {
	j.attrs[key] = value
}

// SetStrings stores values as string slice (ignores separator).
func (j *jsonAttrs) SetStrings(key string, values []string, _ string) {
	j.attrs[key] = values
}

// SetTime formats time as RFC3339 string.
func (j *jsonAttrs) SetTime(key string, t time.Time) {
	j.attrs[key] = t.Format(time.RFC3339)
}

// SetDate formats date as yyyy-mm-dd string.
func (j *jsonAttrs) SetDate(key string, year int, month time.Month, day int) {
	j.attrs[key] = fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
}
