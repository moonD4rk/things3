package scheme

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLAttrs_SetString(t *testing.T) {
	attrs := &URLAttrs{Params: make(map[string]string)}
	attrs.SetString("title", "Test Task")
	assert.Equal(t, "Test Task", attrs.Params["title"])
}

func TestURLAttrs_SetBool(t *testing.T) {
	attrs := &URLAttrs{Params: make(map[string]string)}

	attrs.SetBool("completed", true)
	assert.Equal(t, "true", attrs.Params["completed"])

	attrs.SetBool("canceled", false)
	assert.Equal(t, "false", attrs.Params["canceled"])
}

func TestURLAttrs_SetStrings(t *testing.T) {
	attrs := &URLAttrs{Params: make(map[string]string)}
	attrs.SetStrings("tags", []string{"work", "urgent"}, ",")
	assert.Equal(t, "work,urgent", attrs.Params["tags"])
}

func TestURLAttrs_SetTime(t *testing.T) {
	attrs := &URLAttrs{Params: make(map[string]string)}
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	attrs.SetTime("creation-date", testTime)
	assert.Equal(t, "2024-06-15T10:30:00Z", attrs.Params["creation-date"])
}

func TestURLAttrs_SetDate(t *testing.T) {
	attrs := &URLAttrs{Params: make(map[string]string)}
	attrs.SetDate("when", 2024, time.June, 15)
	assert.Equal(t, "2024-06-15", attrs.Params["when"])
}

func TestJSONAttrs_SetString(t *testing.T) {
	attrs := &JSONAttrs{Attrs: make(map[string]any)}
	attrs.SetString("title", "Test Task")
	assert.Equal(t, "Test Task", attrs.Attrs["title"])
}

func TestJSONAttrs_SetBool(t *testing.T) {
	attrs := &JSONAttrs{Attrs: make(map[string]any)}

	attrs.SetBool("completed", true)
	assert.Equal(t, true, attrs.Attrs["completed"])

	attrs.SetBool("canceled", false)
	assert.Equal(t, false, attrs.Attrs["canceled"])
}

func TestJSONAttrs_SetStrings(t *testing.T) {
	attrs := &JSONAttrs{Attrs: make(map[string]any)}
	attrs.SetStrings("tags", []string{"work", "urgent"}, ",")
	assert.Equal(t, []string{"work", "urgent"}, attrs.Attrs["tags"])
}

func TestJSONAttrs_SetTime(t *testing.T) {
	attrs := &JSONAttrs{Attrs: make(map[string]any)}
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	attrs.SetTime("creation-date", testTime)
	assert.Equal(t, "2024-06-15T10:30:00Z", attrs.Attrs["creation-date"])
}

func TestJSONAttrs_SetDate(t *testing.T) {
	attrs := &JSONAttrs{Attrs: make(map[string]any)}
	attrs.SetDate("when", 2024, time.June, 15)
	assert.Equal(t, "2024-06-15", attrs.Attrs["when"])
}

// mockBuilder implements AttrBuilder interface for testing generic setters.
type mockBuilder struct {
	attrs URLAttrs
	err   error
}

func newMockBuilder() *mockBuilder {
	return &mockBuilder{attrs: NewURLAttrs()}
}

func (b *mockBuilder) GetStore() AttrStore { return &b.attrs }
func (b *mockBuilder) SetErr(err error)    { b.err = err }

// mockJSONBuilder implements AttrBuilder interface with JSON storage.
type mockJSONBuilder struct {
	attrs JSONAttrs
	err   error
}

func newMockJSONBuilder() *mockJSONBuilder {
	return &mockJSONBuilder{attrs: NewJSONAttrs()}
}

func (b *mockJSONBuilder) GetStore() AttrStore { return &b.attrs }
func (b *mockJSONBuilder) SetErr(err error)    { b.err = err }

func TestSetStr(t *testing.T) {
	t.Run("valid title", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStr(b, TitleParam, "Test Title")
		assert.Same(t, b, result)
		require.NoError(t, b.err)
		assert.Equal(t, "Test Title", b.attrs.Params["title"])
	})

	t.Run("title too long", func(t *testing.T) {
		b := newMockBuilder()
		longTitle := make([]byte, MaxTitleLength+1)
		for i := range longTitle {
			longTitle[i] = 'a'
		}
		result := SetStr(b, TitleParam, string(longTitle))
		assert.Same(t, b, result)
		assert.ErrorIs(t, b.err, ErrTitleTooLong)
	})

	t.Run("valid notes", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStr(b, NotesParam, "Test Notes")
		assert.Same(t, b, result)
		require.NoError(t, b.err)
		assert.Equal(t, "Test Notes", b.attrs.Params["notes"])
	})

	t.Run("notes too long", func(t *testing.T) {
		b := newMockBuilder()
		longNotes := make([]byte, MaxNotesLength+1)
		for i := range longNotes {
			longNotes[i] = 'a'
		}
		result := SetStr(b, NotesParam, string(longNotes))
		assert.Same(t, b, result)
		assert.ErrorIs(t, b.err, ErrNotesTooLong)
	})

	t.Run("deadline", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStr(b, DeadlineParam, "2024-12-31")
		assert.Same(t, b, result)
		assert.Equal(t, "2024-12-31", b.attrs.Params["deadline"])
	})

	t.Run("list", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStr(b, ListParam, "Work Projects")
		assert.Same(t, b, result)
		assert.Equal(t, "Work Projects", b.attrs.Params["list"])
	})

	t.Run("list-id", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStr(b, ListIDParam, "abc-123")
		assert.Same(t, b, result)
		assert.Equal(t, "abc-123", b.attrs.Params["list-id"])
	})

	t.Run("heading", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStr(b, HeadingParam, "Section 1")
		assert.Same(t, b, result)
		assert.Equal(t, "Section 1", b.attrs.Params["heading"])
	})

	t.Run("heading-id", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStr(b, HeadingIDParam, "heading-123")
		assert.Same(t, b, result)
		assert.Equal(t, "heading-123", b.attrs.Params["heading-id"])
	})

	t.Run("area", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStr(b, AreaParam, "Personal")
		assert.Same(t, b, result)
		assert.Equal(t, "Personal", b.attrs.Params["area"])
	})

	t.Run("area-id", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStr(b, AreaIDParam, "area-456")
		assert.Same(t, b, result)
		assert.Equal(t, "area-456", b.attrs.Params["area-id"])
	})

	t.Run("prepend-notes", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStr(b, PrependNotesParam, "Prepended text")
		assert.Same(t, b, result)
		assert.Equal(t, "Prepended text", b.attrs.Params["prepend-notes"])
	})

	t.Run("append-notes", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStr(b, AppendNotesParam, "Appended text")
		assert.Same(t, b, result)
		assert.Equal(t, "Appended text", b.attrs.Params["append-notes"])
	})
}

func TestSetBool(t *testing.T) {
	t.Run("URL attrs stores as string", func(t *testing.T) {
		b := newMockBuilder()
		result := SetBool(b, CompletedParam, true)
		assert.Same(t, b, result)
		assert.Equal(t, "true", b.attrs.Params["completed"])
	})

	t.Run("JSON attrs stores as bool", func(t *testing.T) {
		b := newMockJSONBuilder()
		result := SetBool(b, CompletedParam, true)
		assert.Same(t, b, result)
		assert.Equal(t, true, b.attrs.Attrs["completed"])
	})

	t.Run("canceled", func(t *testing.T) {
		b := newMockBuilder()
		result := SetBool(b, CanceledParam, true)
		assert.Same(t, b, result)
		assert.Equal(t, "true", b.attrs.Params["canceled"])
	})

	t.Run("reveal", func(t *testing.T) {
		b := newMockBuilder()
		result := SetBool(b, RevealParam, true)
		assert.Same(t, b, result)
		assert.Equal(t, "true", b.attrs.Params["reveal"])
	})
}

func TestSetStrs(t *testing.T) {
	t.Run("URL attrs joins with comma", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStrs(b, TagsParam, []string{"work", "urgent"})
		assert.Same(t, b, result)
		assert.Equal(t, "work,urgent", b.attrs.Params["tags"])
	})

	t.Run("JSON attrs stores as array", func(t *testing.T) {
		b := newMockJSONBuilder()
		result := SetStrs(b, TagsParam, []string{"work", "urgent"})
		assert.Same(t, b, result)
		assert.Equal(t, []string{"work", "urgent"}, b.attrs.Attrs["tags"])
	})

	t.Run("add-tags", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStrs(b, AddTagsParam, []string{"new-tag"})
		assert.Same(t, b, result)
		assert.Equal(t, "new-tag", b.attrs.Params["add-tags"])
	})

	t.Run("checklist items valid", func(t *testing.T) {
		b := newMockBuilder()
		result := SetStrs(b, ChecklistItemsParam, []string{"Item 1", "Item 2"})
		assert.Same(t, b, result)
		require.NoError(t, b.err)
		assert.Equal(t, "Item 1\nItem 2", b.attrs.Params["checklist-items"])
	})

	t.Run("checklist items too many", func(t *testing.T) {
		b := newMockBuilder()
		items := make([]string, MaxChecklistItems+1)
		for i := range items {
			items[i] = "item"
		}
		result := SetStrs(b, ChecklistItemsParam, items)
		assert.Same(t, b, result)
		assert.ErrorIs(t, b.err, ErrTooManyChecklistItems)
	})
}

func TestSetTime(t *testing.T) {
	t.Run("creation-date", func(t *testing.T) {
		b := newMockBuilder()
		testTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
		result := SetTime(b, CreationDateParam, testTime)
		assert.Same(t, b, result)
		assert.Equal(t, "2024-01-15T09:00:00Z", b.attrs.Params["creation-date"])
	})

	t.Run("completion-date", func(t *testing.T) {
		b := newMockBuilder()
		testTime := time.Date(2024, 1, 20, 17, 0, 0, 0, time.UTC)
		result := SetTime(b, CompletionDateParam, testTime)
		assert.Same(t, b, result)
		assert.Equal(t, "2024-01-20T17:00:00Z", b.attrs.Params["completion-date"])
	})
}

func TestSetDate(t *testing.T) {
	b := newMockBuilder()
	result := SetDate(b, WhenParam, 2024, time.December, 25)
	assert.Same(t, b, result)
	assert.Equal(t, "2024-12-25", b.attrs.Params["when"])
}

func TestSetWhenStr(t *testing.T) {
	b := newMockBuilder()
	result := SetWhenStr(b, WhenEvening)
	assert.Same(t, b, result)
	assert.Equal(t, "evening", b.attrs.Params["when"])
}

func TestSetWhenTime(t *testing.T) {
	b := newMockBuilder()
	testDate := time.Date(2025, 6, 15, 14, 30, 0, 0, time.Local)
	result := SetWhenTime(b, testDate)
	assert.Same(t, b, result)
	assert.Equal(t, "2025-06-15", b.attrs.Params["when"])
}

func TestSetDeadlineTime(t *testing.T) {
	b := newMockBuilder()
	testDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.Local)
	result := SetDeadlineTime(b, testDate)
	assert.Same(t, b, result)
	assert.Equal(t, "2025-12-31", b.attrs.Params["deadline"])
}

func TestSetWhenTime_ZeroValue(t *testing.T) {
	b := newMockBuilder()
	var zeroTime time.Time
	result := SetWhenTime(b, zeroTime)
	assert.Same(t, b, result)
	_, exists := b.attrs.Params["when"]
	assert.False(t, exists, "when parameter should not be set for zero time")
}

func TestSetDeadlineTime_ZeroValue(t *testing.T) {
	b := newMockBuilder()
	var zeroTime time.Time
	result := SetDeadlineTime(b, zeroTime)
	assert.Same(t, b, result)
	_, exists := b.attrs.Params["deadline"]
	assert.False(t, exists, "deadline parameter should not be set for zero time")
}

func TestEncodeQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		contains []string
		excludes []string
	}{
		{
			name:     "space encoding",
			input:    map[string]string{"title": "Hello World"},
			contains: []string{"Hello%20World"},
			excludes: []string{"Hello+World"},
		},
		{
			name:     "plus encoding",
			input:    map[string]string{"title": "C++"},
			contains: []string{"C%2B%2B"},
			excludes: []string{},
		},
		{
			name:     "mixed space and plus",
			input:    map[string]string{"title": "C++ tutorial"},
			contains: []string{"%2B%2B", "%20"},
			excludes: []string{},
		},
		{
			name:     "special characters",
			input:    map[string]string{"title": "test@example.com"},
			contains: []string{"%40"}, // @ is encoded as %40
			excludes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := url.Values{}
			for k, v := range tt.input {
				query.Set(k, v)
			}
			encoded := EncodeQuery(query)

			for _, c := range tt.contains {
				assert.Contains(t, encoded, c, "encoded query should contain %s", c)
			}
			for _, e := range tt.excludes {
				assert.NotContains(t, encoded, e, "encoded query should not contain %s", e)
			}
		})
	}
}
