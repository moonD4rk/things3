package things3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLAttrs_SetString(t *testing.T) {
	attrs := &urlAttrs{params: make(map[string]string)}
	attrs.SetString("title", "Test Task")
	assert.Equal(t, "Test Task", attrs.params["title"])
}

func TestURLAttrs_SetBool(t *testing.T) {
	attrs := &urlAttrs{params: make(map[string]string)}

	attrs.SetBool("completed", true)
	assert.Equal(t, "true", attrs.params["completed"])

	attrs.SetBool("canceled", false)
	assert.Equal(t, "false", attrs.params["canceled"])
}

func TestURLAttrs_SetStrings(t *testing.T) {
	attrs := &urlAttrs{params: make(map[string]string)}
	attrs.SetStrings("tags", []string{"work", "urgent"}, ",")
	assert.Equal(t, "work,urgent", attrs.params["tags"])
}

func TestURLAttrs_SetTime(t *testing.T) {
	attrs := &urlAttrs{params: make(map[string]string)}
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	attrs.SetTime("creation-date", testTime)
	assert.Equal(t, "2024-06-15T10:30:00Z", attrs.params["creation-date"])
}

func TestURLAttrs_SetDate(t *testing.T) {
	attrs := &urlAttrs{params: make(map[string]string)}
	attrs.SetDate("when", 2024, time.June, 15)
	assert.Equal(t, "2024-06-15", attrs.params["when"])
}

func TestJSONAttrs_SetString(t *testing.T) {
	attrs := &jsonAttrs{attrs: make(map[string]any)}
	attrs.SetString("title", "Test Task")
	assert.Equal(t, "Test Task", attrs.attrs["title"])
}

func TestJSONAttrs_SetBool(t *testing.T) {
	attrs := &jsonAttrs{attrs: make(map[string]any)}

	attrs.SetBool("completed", true)
	assert.Equal(t, true, attrs.attrs["completed"])

	attrs.SetBool("canceled", false)
	assert.Equal(t, false, attrs.attrs["canceled"])
}

func TestJSONAttrs_SetStrings(t *testing.T) {
	attrs := &jsonAttrs{attrs: make(map[string]any)}
	attrs.SetStrings("tags", []string{"work", "urgent"}, ",")
	assert.Equal(t, []string{"work", "urgent"}, attrs.attrs["tags"])
}

func TestJSONAttrs_SetTime(t *testing.T) {
	attrs := &jsonAttrs{attrs: make(map[string]any)}
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	attrs.SetTime("creation-date", testTime)
	assert.Equal(t, "2024-06-15T10:30:00Z", attrs.attrs["creation-date"])
}

func TestJSONAttrs_SetDate(t *testing.T) {
	attrs := &jsonAttrs{attrs: make(map[string]any)}
	attrs.SetDate("when", 2024, time.June, 15)
	assert.Equal(t, "2024-06-15", attrs.attrs["when"])
}

// mockBuilder implements attrBuilder interface for testing generic setters.
type mockBuilder struct {
	attrs urlAttrs
	err   error
}

func newMockBuilder() *mockBuilder {
	return &mockBuilder{
		attrs: urlAttrs{params: make(map[string]string)},
	}
}

func (b *mockBuilder) getStore() attrStore { return &b.attrs }
func (b *mockBuilder) setErr(err error)    { b.err = err }

// mockJSONBuilder implements attrBuilder interface with JSON storage.
type mockJSONBuilder struct {
	attrs jsonAttrs
	err   error
}

func newMockJSONBuilder() *mockJSONBuilder {
	return &mockJSONBuilder{
		attrs: jsonAttrs{attrs: make(map[string]any)},
	}
}

func (b *mockJSONBuilder) getStore() attrStore { return &b.attrs }
func (b *mockJSONBuilder) setErr(err error)    { b.err = err }

func TestSetStr(t *testing.T) {
	t.Run("valid title", func(t *testing.T) {
		b := newMockBuilder()
		result := setStr(b, titleParam, "Test Title")
		assert.Same(t, b, result)
		require.NoError(t, b.err)
		assert.Equal(t, "Test Title", b.attrs.params["title"])
	})

	t.Run("title too long", func(t *testing.T) {
		b := newMockBuilder()
		longTitle := make([]byte, maxTitleLength+1)
		for i := range longTitle {
			longTitle[i] = 'a'
		}
		result := setStr(b, titleParam, string(longTitle))
		assert.Same(t, b, result)
		assert.ErrorIs(t, b.err, ErrTitleTooLong)
	})

	t.Run("valid notes", func(t *testing.T) {
		b := newMockBuilder()
		result := setStr(b, notesParam, "Test Notes")
		assert.Same(t, b, result)
		require.NoError(t, b.err)
		assert.Equal(t, "Test Notes", b.attrs.params["notes"])
	})

	t.Run("notes too long", func(t *testing.T) {
		b := newMockBuilder()
		longNotes := make([]byte, maxNotesLength+1)
		for i := range longNotes {
			longNotes[i] = 'a'
		}
		result := setStr(b, notesParam, string(longNotes))
		assert.Same(t, b, result)
		assert.ErrorIs(t, b.err, ErrNotesTooLong)
	})

	t.Run("deadline", func(t *testing.T) {
		b := newMockBuilder()
		result := setStr(b, deadlineParam, "2024-12-31")
		assert.Same(t, b, result)
		assert.Equal(t, "2024-12-31", b.attrs.params["deadline"])
	})

	t.Run("list", func(t *testing.T) {
		b := newMockBuilder()
		result := setStr(b, listParam, "Work Projects")
		assert.Same(t, b, result)
		assert.Equal(t, "Work Projects", b.attrs.params["list"])
	})

	t.Run("list-id", func(t *testing.T) {
		b := newMockBuilder()
		result := setStr(b, listIDParam, "abc-123")
		assert.Same(t, b, result)
		assert.Equal(t, "abc-123", b.attrs.params["list-id"])
	})

	t.Run("heading", func(t *testing.T) {
		b := newMockBuilder()
		result := setStr(b, headingParam, "Section 1")
		assert.Same(t, b, result)
		assert.Equal(t, "Section 1", b.attrs.params["heading"])
	})

	t.Run("heading-id", func(t *testing.T) {
		b := newMockBuilder()
		result := setStr(b, headingIDParam, "heading-123")
		assert.Same(t, b, result)
		assert.Equal(t, "heading-123", b.attrs.params["heading-id"])
	})

	t.Run("area", func(t *testing.T) {
		b := newMockBuilder()
		result := setStr(b, areaParam, "Personal")
		assert.Same(t, b, result)
		assert.Equal(t, "Personal", b.attrs.params["area"])
	})

	t.Run("area-id", func(t *testing.T) {
		b := newMockBuilder()
		result := setStr(b, areaIDParam, "area-456")
		assert.Same(t, b, result)
		assert.Equal(t, "area-456", b.attrs.params["area-id"])
	})

	t.Run("prepend-notes", func(t *testing.T) {
		b := newMockBuilder()
		result := setStr(b, prependNotesParam, "Prepended text")
		assert.Same(t, b, result)
		assert.Equal(t, "Prepended text", b.attrs.params["prepend-notes"])
	})

	t.Run("append-notes", func(t *testing.T) {
		b := newMockBuilder()
		result := setStr(b, appendNotesParam, "Appended text")
		assert.Same(t, b, result)
		assert.Equal(t, "Appended text", b.attrs.params["append-notes"])
	})
}

func TestSetBool(t *testing.T) {
	t.Run("URL attrs stores as string", func(t *testing.T) {
		b := newMockBuilder()
		result := setBool(b, completedParam, true)
		assert.Same(t, b, result)
		assert.Equal(t, "true", b.attrs.params["completed"])
	})

	t.Run("JSON attrs stores as bool", func(t *testing.T) {
		b := newMockJSONBuilder()
		result := setBool(b, completedParam, true)
		assert.Same(t, b, result)
		assert.Equal(t, true, b.attrs.attrs["completed"])
	})

	t.Run("canceled", func(t *testing.T) {
		b := newMockBuilder()
		result := setBool(b, canceledParam, true)
		assert.Same(t, b, result)
		assert.Equal(t, "true", b.attrs.params["canceled"])
	})

	t.Run("reveal", func(t *testing.T) {
		b := newMockBuilder()
		result := setBool(b, revealParam, true)
		assert.Same(t, b, result)
		assert.Equal(t, "true", b.attrs.params["reveal"])
	})
}

func TestSetStrs(t *testing.T) {
	t.Run("URL attrs joins with comma", func(t *testing.T) {
		b := newMockBuilder()
		result := setStrs(b, tagsParam, []string{"work", "urgent"})
		assert.Same(t, b, result)
		assert.Equal(t, "work,urgent", b.attrs.params["tags"])
	})

	t.Run("JSON attrs stores as array", func(t *testing.T) {
		b := newMockJSONBuilder()
		result := setStrs(b, tagsParam, []string{"work", "urgent"})
		assert.Same(t, b, result)
		assert.Equal(t, []string{"work", "urgent"}, b.attrs.attrs["tags"])
	})

	t.Run("add-tags", func(t *testing.T) {
		b := newMockBuilder()
		result := setStrs(b, addTagsParam, []string{"new-tag"})
		assert.Same(t, b, result)
		assert.Equal(t, "new-tag", b.attrs.params["add-tags"])
	})

	t.Run("checklist items valid", func(t *testing.T) {
		b := newMockBuilder()
		result := setStrs(b, checklistItemsParam, []string{"Item 1", "Item 2"})
		assert.Same(t, b, result)
		require.NoError(t, b.err)
		assert.Equal(t, "Item 1\nItem 2", b.attrs.params["checklist-items"])
	})

	t.Run("checklist items too many", func(t *testing.T) {
		b := newMockBuilder()
		items := make([]string, maxChecklistItems+1)
		for i := range items {
			items[i] = "item"
		}
		result := setStrs(b, checklistItemsParam, items)
		assert.Same(t, b, result)
		assert.ErrorIs(t, b.err, ErrTooManyChecklistItems)
	})
}

func TestSetTime(t *testing.T) {
	t.Run("creation-date", func(t *testing.T) {
		b := newMockBuilder()
		testTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
		result := setTime(b, creationDateParam, testTime)
		assert.Same(t, b, result)
		assert.Equal(t, "2024-01-15T09:00:00Z", b.attrs.params["creation-date"])
	})

	t.Run("completion-date", func(t *testing.T) {
		b := newMockBuilder()
		testTime := time.Date(2024, 1, 20, 17, 0, 0, 0, time.UTC)
		result := setTime(b, completionDateParam, testTime)
		assert.Same(t, b, result)
		assert.Equal(t, "2024-01-20T17:00:00Z", b.attrs.params["completion-date"])
	})
}

func TestSetDate(t *testing.T) {
	b := newMockBuilder()
	result := setDate(b, whenParam, 2024, time.December, 25)
	assert.Same(t, b, result)
	assert.Equal(t, "2024-12-25", b.attrs.params["when"])
}

func TestSetWhenStr(t *testing.T) {
	b := newMockBuilder()
	result := setWhenStr(b, whenEvening)
	assert.Same(t, b, result)
	assert.Equal(t, "evening", b.attrs.params["when"])
}

func TestSetWhenTime(t *testing.T) {
	b := newMockBuilder()
	testDate := time.Date(2025, 6, 15, 14, 30, 0, 0, time.Local)
	result := setWhenTime(b, testDate)
	assert.Same(t, b, result)
	assert.Equal(t, "2025-06-15", b.attrs.params["when"])
}

func TestSetDeadlineTime(t *testing.T) {
	b := newMockBuilder()
	testDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.Local)
	result := setDeadlineTime(b, testDate)
	assert.Same(t, b, result)
	assert.Equal(t, "2025-12-31", b.attrs.params["deadline"])
}

func TestSetWhenTime_ZeroValue(t *testing.T) {
	b := newMockBuilder()
	var zeroTime time.Time
	result := setWhenTime(b, zeroTime)
	assert.Same(t, b, result)
	_, exists := b.attrs.params["when"]
	assert.False(t, exists, "when parameter should not be set for zero time")
}

func TestSetDeadlineTime_ZeroValue(t *testing.T) {
	b := newMockBuilder()
	var zeroTime time.Time
	result := setDeadlineTime(b, zeroTime)
	assert.Same(t, b, result)
	_, exists := b.attrs.params["deadline"]
	assert.False(t, exists, "deadline parameter should not be set for zero time")
}
