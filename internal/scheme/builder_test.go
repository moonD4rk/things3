package scheme

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// staticTokenFunc returns a token function that always yields the given token.
func staticTokenFunc(token string) func(context.Context) (string, error) {
	return func(context.Context) (string, error) { return token, nil }
}

// parseQuery extracts the query parameters from a Things URL.
func parseQuery(t *testing.T, thingsURL string) url.Values {
	t.Helper()
	parsed, err := url.Parse(thingsURL)
	require.NoError(t, err)
	return parsed.Query()
}

// Build must be idempotent: calling it twice (as Execute does internally,
// or on retry) must return byte-identical URLs even with a reminder set.
func TestBuildIdempotentWithReminder(t *testing.T) {
	s := New()
	when := time.Date(2025, time.January, 2, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name    string
		builder interface{ Build() (string, error) }
	}{
		{
			name:    "todo adder",
			builder: NewTodoAdder(s).Title("Meeting").When(when).Reminder(15, 0),
		},
		{
			name:    "project adder",
			builder: NewProjectAdder(s).Title("Project").When(when).Reminder(15, 0),
		},
		{
			name:    "todo updater",
			builder: NewTodoUpdater(s, staticTokenFunc("token"), "uuid-1").When(when).Reminder(15, 0),
		},
		{
			name:    "project updater",
			builder: NewProjectUpdater(s, staticTokenFunc("token"), "uuid-1").When(when).Reminder(15, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first, err := tt.builder.Build()
			require.NoError(t, err)
			second, err := tt.builder.Build()
			require.NoError(t, err)
			assert.Equal(t, first, second, "repeated Build calls must return identical URLs")

			params := parseQuery(t, second)
			assert.Equal(t, "2025-01-02@15:00", params.Get(KeyWhen),
				"reminder time must be appended exactly once")
		})
	}
}

// A reminder combined with someday or anytime has no concrete date to attach
// to, so Build must fail instead of emitting a when value Things cannot parse.
func TestBuildRejectsReminderWithoutConcreteDate(t *testing.T) {
	s := New()

	tests := []struct {
		name    string
		builder interface{ Build() (string, error) }
	}{
		{"todo adder someday", NewTodoAdder(s).Title("T").WhenSomeday().Reminder(9, 0)},
		{"todo adder anytime", NewTodoAdder(s).Title("T").WhenAnytime().Reminder(9, 0)},
		{"project adder someday", NewProjectAdder(s).Title("P").WhenSomeday().Reminder(9, 0)},
		{"project adder anytime", NewProjectAdder(s).Title("P").WhenAnytime().Reminder(9, 0)},
		{"todo updater someday", NewTodoUpdater(s, staticTokenFunc("token"), "id").WhenSomeday().Reminder(9, 0)},
		{"todo updater anytime", NewTodoUpdater(s, staticTokenFunc("token"), "id").WhenAnytime().Reminder(9, 0)},
		{"project updater someday", NewProjectUpdater(s, staticTokenFunc("token"), "id").WhenSomeday().Reminder(9, 0)},
		{"project updater anytime", NewProjectUpdater(s, staticTokenFunc("token"), "id").WhenAnytime().Reminder(9, 0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.builder.Build()
			assert.ErrorIs(t, err, ErrReminderNeedsDate)
		})
	}
}

// Evening carries a concrete scheduling slot, so reminders remain allowed.
func TestBuildAllowsReminderWithEvening(t *testing.T) {
	thingsURL, err := NewTodoAdder(New()).Title("T").WhenEvening().Reminder(18, 30).Build()
	require.NoError(t, err)
	assert.Equal(t, "evening@18:30", parseQuery(t, thingsURL).Get(KeyWhen))
}

// Length limits are documented in characters; multi-byte input must not be
// rejected at one third of the limit.
func TestLengthLimitsCountCharactersNotBytes(t *testing.T) {
	s := New()
	cjkTitle := strings.Repeat("中", 2000)  // 2000 runes, 6000 bytes
	cjkNotes := strings.Repeat("中", 10000) // 10000 runes, 30000 bytes

	t.Run("todo adder title", func(t *testing.T) {
		thingsURL, err := NewTodoAdder(s).Title(cjkTitle).Build()
		require.NoError(t, err)
		assert.Equal(t, cjkTitle, parseQuery(t, thingsURL).Get(KeyTitle))
	})

	t.Run("todo adder notes", func(t *testing.T) {
		_, err := NewTodoAdder(s).Title("T").Notes(cjkNotes).Build()
		require.NoError(t, err)
	})

	t.Run("titles are validated per title", func(t *testing.T) {
		thingsURL, err := NewTodoAdder(s).Titles(cjkTitle, cjkTitle, cjkTitle).Build()
		require.NoError(t, err)
		assert.Equal(t, strings.Repeat(cjkTitle+"\n", 2)+cjkTitle,
			parseQuery(t, thingsURL).Get(KeyTitles))
	})

	t.Run("batch todo title", func(t *testing.T) {
		item := newBatchTodoBuilder()
		item.Title(cjkTitle)
		require.NoError(t, item.err)
	})

	t.Run("rune limit still enforced", func(t *testing.T) {
		_, err := NewTodoAdder(s).Title(strings.Repeat("中", MaxTitleLength+1)).Build()
		require.ErrorIs(t, err, ErrTitleTooLong)

		_, err = NewTodoAdder(s).Titles(strings.Repeat("a", MaxTitleLength+1)).Build()
		assert.ErrorIs(t, err, ErrTitleTooLong)
	})
}

// Values containing the join separator would silently split into multiple
// items; builders must reject them at build time.
func TestSeparatorInjectionRejected(t *testing.T) {
	s := New()

	tests := []struct {
		name    string
		builder interface{ Build() (string, error) }
		wantErr error
	}{
		{"todo adder tag with comma", NewTodoAdder(s).Title("T").Tags("home,work"), ErrTagContainsComma},
		{"project adder tag with comma", NewProjectAdder(s).Title("P").Tags("a,b"), ErrTagContainsComma},
		{"todo updater add-tags with comma", NewTodoUpdater(s, staticTokenFunc("token"), "id").AddTags("a,b"), ErrTagContainsComma},
		{"checklist item with newline", NewTodoAdder(s).Title("T").ChecklistItems("line1\nline2"), ErrChecklistItemContainsNewline},
		{"prepend checklist item with newline", NewTodoUpdater(s, staticTokenFunc("token"), "id").PrependChecklistItems("a\nb"), ErrChecklistItemContainsNewline},
		{"append checklist item with newline", NewTodoUpdater(s, staticTokenFunc("token"), "id").AppendChecklistItems("a\nb"), ErrChecklistItemContainsNewline},
		{"titles with newline", NewTodoAdder(s).Titles("Task 1\nTask 2"), ErrTitleContainsNewline},
		{"project todos with newline", NewProjectAdder(s).Title("P").Todos("Task 1\nTask 2"), ErrTitleContainsNewline},
		{"show filter tag with comma", NewShowNavigator(s).List(ListToday).Filter("home,work"), ErrTagContainsComma},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.builder.Build()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// Separator-free values must keep working after the injection validation.
func TestSeparatorValidationAllowsCleanValues(t *testing.T) {
	s := New()

	thingsURL, err := NewTodoAdder(s).Title("T").Tags("home", "work").ChecklistItems("one", "two").Build()
	require.NoError(t, err)
	params := parseQuery(t, thingsURL)
	assert.Equal(t, "home,work", params.Get(KeyTags))
	assert.Equal(t, "one\ntwo", params.Get(KeyChecklistItems))

	thingsURL, err = NewShowNavigator(s).List(ListToday).Filter("home", "work").Build()
	require.NoError(t, err)
	assert.Equal(t, "home,work", parseQuery(t, thingsURL).Get(KeyFilter))
}
