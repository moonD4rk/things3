package scheme

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// AttrStore abstracts attribute storage for URL params vs JSON attributes.
// URL builders store everything as strings, JSON builders preserve native types.
type AttrStore interface {
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

// AttrBuilder is the interface that all URL scheme builders must implement.
// It provides access to the underlying attribute store and error handling.
type AttrBuilder interface {
	GetStore() AttrStore
	SetErr(err error)
}

// ReminderStore is implemented by builders that support reminder functionality.
type ReminderStore interface {
	SetReminder(hour, minute int)
}

// Parameter definition types for type-safe attribute setters.

// StrParam defines metadata for string URL parameters including validation rules.
type StrParam struct {
	Key    string // URL parameter key (e.g., "title", "notes")
	MaxLen int    // Maximum length in characters/runes (0 = no limit)
	Err    error  // Error to return if validation fails
}

// BoolParam defines metadata for boolean URL parameters.
type BoolParam struct {
	Key string // URL parameter key
}

// StrsParam defines metadata for string slice URL parameters.
type StrsParam struct {
	Key      string // URL parameter key
	Sep      string // Separator for joining values (e.g., "," for tags, "\n" for checklist)
	MaxCount int    // Maximum item count (0 = no limit)
	Err      error  // Error to return if count validation fails
	SepErr   error  // Error to return when a value contains the separator
}

// TimeParam defines metadata for time URL parameters.
type TimeParam struct {
	Key string // URL parameter key
}

// DateParam defines metadata for date URL parameters.
type DateParam struct {
	Key string // URL parameter key
}

// Predefined parameter definitions for Things URL scheme.
// These map builder methods to URL parameter keys with validation rules.
var (
	// String parameters
	TitleParam        = StrParam{Key: KeyTitle, MaxLen: MaxTitleLength, Err: ErrTitleTooLong}
	NotesParam        = StrParam{Key: KeyNotes, MaxLen: MaxNotesLength, Err: ErrNotesTooLong}
	DeadlineParam     = StrParam{Key: KeyDeadline}
	ListParam         = StrParam{Key: KeyList}
	ListIDParam       = StrParam{Key: KeyListID}
	HeadingParam      = StrParam{Key: KeyHeading}
	HeadingIDParam    = StrParam{Key: KeyHeadingID}
	AreaParam         = StrParam{Key: KeyArea}
	AreaIDParam       = StrParam{Key: KeyAreaID}
	PrependNotesParam = StrParam{Key: KeyPrependNotes}
	AppendNotesParam  = StrParam{Key: KeyAppendNotes}

	// Boolean parameters
	CompletedParam      = BoolParam{Key: KeyCompleted}
	CanceledParam       = BoolParam{Key: KeyCanceled}
	RevealParam         = BoolParam{Key: KeyReveal}
	DuplicateParam      = BoolParam{Key: KeyDuplicate}
	ShowQuickEntryParam = BoolParam{Key: KeyShowQuickEntry}

	// String slice parameters
	TagsParam           = StrsParam{Key: KeyTags, Sep: ",", SepErr: ErrTagContainsComma}
	AddTagsParam        = StrsParam{Key: KeyAddTags, Sep: ",", SepErr: ErrTagContainsComma}
	ChecklistItemsParam = StrsParam{
		Key: KeyChecklistItems, Sep: "\n",
		MaxCount: MaxChecklistItems, Err: ErrTooManyChecklistItems,
		SepErr: ErrChecklistItemContainsNewline,
	}
	PrependChecklistParam = StrsParam{Key: KeyPrependChecklistItems, Sep: "\n", SepErr: ErrChecklistItemContainsNewline}
	AppendChecklistParam  = StrsParam{Key: KeyAppendChecklistItems, Sep: "\n", SepErr: ErrChecklistItemContainsNewline}
	TodosParam            = StrsParam{Key: KeyTodos, Sep: "\n", SepErr: ErrTitleContainsNewline}

	// Time parameters
	CreationDateParam   = TimeParam{Key: KeyCreationDate}
	CompletionDateParam = TimeParam{Key: KeyCompletionDate}

	// Date parameters
	WhenParam = DateParam{Key: KeyWhen}
)

// Validation errors for URL scheme parameters.
var (
	// ErrTitleTooLong is returned when title exceeds the character limit.
	ErrTitleTooLong = fmt.Errorf("things3: title exceeds %d character limit", MaxTitleLength)
	// ErrNotesTooLong is returned when notes exceed the character limit.
	ErrNotesTooLong = fmt.Errorf("things3: notes exceed %d character limit", MaxNotesLength)
	// ErrTooManyChecklistItems is returned when checklist exceeds the item limit.
	ErrTooManyChecklistItems = fmt.Errorf("things3: checklist exceeds %d item limit", MaxChecklistItems)
	// ErrTagContainsComma is returned when a tag contains a comma, the tag list separator.
	ErrTagContainsComma = errors.New("things3: tag must not contain a comma")
	// ErrChecklistItemContainsNewline is returned when a checklist item contains a newline, the item separator.
	ErrChecklistItemContainsNewline = errors.New("things3: checklist item must not contain a newline")
	// ErrTitleContainsNewline is returned when a title contains a newline, the title list separator.
	ErrTitleContainsNewline = errors.New("things3: title must not contain a newline")
)

// Generic setter functions with type-safe parameter definitions.

// SetStr sets a string attribute with optional length validation.
// Length limits are counted in Unicode characters, matching the limits
// documented by the Things URL scheme.
func SetStr[T AttrBuilder](b T, p StrParam, value string) T {
	if p.MaxLen > 0 && utf8.RuneCountInString(value) > p.MaxLen {
		b.SetErr(p.Err)
		return b
	}
	b.GetStore().SetString(p.Key, value)
	return b
}

// SetBool sets a boolean attribute.
func SetBool[T AttrBuilder](b T, p BoolParam, value bool) T {
	b.GetStore().SetBool(p.Key, value)
	return b
}

// SetStrs sets a string slice attribute with optional count validation.
// Values containing the separator are rejected because they would silently
// split into multiple items when joined for the URL scheme.
func SetStrs[T AttrBuilder](b T, p StrsParam, values []string) T {
	if p.MaxCount > 0 && len(values) > p.MaxCount {
		b.SetErr(p.Err)
		return b
	}
	if p.Sep != "" {
		for _, v := range values {
			if strings.Contains(v, p.Sep) {
				b.SetErr(p.SepErr)
				return b
			}
		}
	}
	b.GetStore().SetStrings(p.Key, values, p.Sep)
	return b
}

// SetTime sets a time attribute.
func SetTime[T AttrBuilder](b T, p TimeParam, t time.Time) T {
	b.GetStore().SetTime(p.Key, t)
	return b
}

// SetDate sets a date attribute.
func SetDate[T AttrBuilder](b T, p DateParam, year int, month time.Month, day int) T {
	b.GetStore().SetDate(p.Key, year, month, day)
	return b
}

// SetWhenStr sets the when attribute using a When constant.
func SetWhenStr[T AttrBuilder](b T, w When) T {
	b.GetStore().SetString(KeyWhen, string(w))
	return b
}

// SetWhenTime sets the when attribute using a time.Time value.
// The time is formatted as yyyy-mm-dd for the Things URL scheme.
// If the time is zero, the parameter is not set.
func SetWhenTime[T AttrBuilder](b T, t time.Time) T {
	if t.IsZero() {
		return b
	}
	b.GetStore().SetDate(KeyWhen, t.Year(), t.Month(), t.Day())
	return b
}

// SetDeadlineTime sets the deadline attribute using a time.Time value.
// The time is formatted as yyyy-mm-dd for the Things URL scheme.
// If the time is zero, the parameter is not set.
func SetDeadlineTime[T AttrBuilder](b T, t time.Time) T {
	if t.IsZero() {
		return b
	}
	b.GetStore().SetDate(KeyDeadline, t.Year(), t.Month(), t.Day())
	return b
}

// ErrInvalidReminderTime is returned when reminder hour or minute is out of range.
var ErrInvalidReminderTime = errors.New("things3: invalid reminder time (hour must be 0-23, minute must be 0-59)")

// ErrReminderNeedsDate is returned when a reminder is combined with a "when"
// value that has no concrete date. Things accepts reminder times only with a
// date, today, tomorrow, or evening; someday and anytime cannot carry one.
var ErrReminderNeedsDate = errors.New("things3: reminder requires a concrete date, today, tomorrow, or evening (not someday or anytime)")

// SetReminder sets the reminder time for builders that support it.
func SetReminder[T AttrBuilder](b T, hour, minute int) T {
	if hour < 0 || hour > 23 {
		b.SetErr(ErrInvalidReminderTime)
		return b
	}
	if minute < 0 || minute > 59 {
		b.SetErr(ErrInvalidReminderTime)
		return b
	}
	if store, ok := b.GetStore().(ReminderStore); ok {
		store.SetReminder(hour, minute)
	}
	return b
}

// URLAttrs stores attributes as URL query parameters (all strings).
type URLAttrs struct {
	Params       map[string]string
	ReminderHour *int // nil means not set
	ReminderMin  *int // nil means not set
}

// NewURLAttrs creates a new URLAttrs with initialized params map.
func NewURLAttrs() URLAttrs {
	return URLAttrs{Params: make(map[string]string)}
}

// SetString sets a string parameter.
func (u *URLAttrs) SetString(key, value string) {
	u.Params[key] = value
}

// SetBool sets a boolean parameter as "true" or "false" string.
func (u *URLAttrs) SetBool(key string, value bool) {
	u.Params[key] = strconv.FormatBool(value)
}

// SetStrings joins values with separator and stores as string.
func (u *URLAttrs) SetStrings(key string, values []string, separator string) {
	u.Params[key] = strings.Join(values, separator)
}

// SetTime formats time as RFC3339 string.
func (u *URLAttrs) SetTime(key string, t time.Time) {
	u.Params[key] = t.Format(time.RFC3339)
}

// SetDate formats date as yyyy-mm-dd string.
func (u *URLAttrs) SetDate(key string, year int, month time.Month, day int) {
	u.Params[key] = fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
}

// SetReminder sets the reminder time (hour and minute).
func (u *URLAttrs) SetReminder(hour, minute int) {
	u.ReminderHour = &hour
	u.ReminderMin = &minute
}

// QueryValues returns the URL query values for the stored parameters.
// When a reminder is set, its time is appended to the "when" value in
// "when@HH:MM" format (e.g., "today@15:30", "2024-03-15@14:00"),
// defaulting to "today" when no "when" value exists.
// The receiver is never mutated, so repeated calls yield identical results.
func (u *URLAttrs) QueryValues() (url.Values, error) {
	query := url.Values{}
	for k, v := range u.Params {
		query.Set(k, v)
	}

	if u.ReminderHour == nil {
		return query, nil
	}

	w := u.Params[KeyWhen]
	if w == "" {
		w = "today" // default to today if no when specified
	}
	if w == string(WhenSomeday) || w == string(WhenAnytime) {
		return nil, ErrReminderNeedsDate
	}

	query.Set(KeyWhen, fmt.Sprintf("%s@%02d:%02d", w, *u.ReminderHour, *u.ReminderMin))
	return query, nil
}

// JSONAttrs stores attributes as JSON values (native types).
type JSONAttrs struct {
	Attrs map[string]any
}

// NewJSONAttrs creates a new JSONAttrs with initialized attrs map.
func NewJSONAttrs() JSONAttrs {
	return JSONAttrs{Attrs: make(map[string]any)}
}

// SetString sets a string attribute.
func (j *JSONAttrs) SetString(key, value string) {
	j.Attrs[key] = value
}

// SetBool sets a boolean attribute (native bool).
func (j *JSONAttrs) SetBool(key string, value bool) {
	j.Attrs[key] = value
}

// SetStrings stores values as string slice (ignores separator).
func (j *JSONAttrs) SetStrings(key string, values []string, _ string) {
	j.Attrs[key] = values
}

// SetTime formats time as RFC3339 string.
func (j *JSONAttrs) SetTime(key string, t time.Time) {
	j.Attrs[key] = t.Format(time.RFC3339)
}

// SetDate formats date as yyyy-mm-dd string.
func (j *JSONAttrs) SetDate(key string, year int, month time.Month, day int) {
	j.Attrs[key] = fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
}

// EncodeQuery encodes url.Values for Things URL scheme.
// Things expects %20 for spaces, not + (which is standard form encoding).
// This is safe because original + characters are encoded as %2B by url.Values.Encode().
func EncodeQuery(query url.Values) string {
	return strings.ReplaceAll(query.Encode(), "+", "%20")
}
