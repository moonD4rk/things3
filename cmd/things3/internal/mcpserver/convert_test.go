package mcpserver

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/moond4rk/things3/thingstest"
)

// TestTruncateNotes proves list/search notes are bounded to notesLimit runes
// without splitting a multibyte codepoint, that the cut is flagged, and that
// short or empty notes are left untouched and unflagged.
func TestTruncateNotes(t *testing.T) {
	long := strings.Repeat("字", 250) // 250 CJK runes, 750 bytes
	items := []Item{
		{Notes: long},
		{Notes: "short"},
		{Notes: ""},
	}
	truncateNotes(items)

	if got := utf8.RuneCountInString(items[0].Notes); got != notesLimit {
		t.Errorf("long note = %d runes, want %d", got, notesLimit)
	}
	if !utf8.ValidString(items[0].Notes) {
		t.Error("truncation split a codepoint")
	}
	if !items[0].NotesTruncated {
		t.Error("long note must be flagged truncated")
	}
	if items[1].Notes != "short" || items[1].NotesTruncated {
		t.Errorf("short note changed: %q flagged=%v", items[1].Notes, items[1].NotesTruncated)
	}
	if items[2].NotesTruncated {
		t.Error("empty note must not be flagged")
	}
}

// TestNotesTruncation proves the list-truncates / get-keeps-full split end to
// end: a long note on a real row is shortened and flagged in list and search
// results but returned whole by get.
func TestNotesTruncation(t *testing.T) {
	path := thingstest.DatabasePath(t)
	srv := serverOnPath(t, path, Config{})

	inbox := listTodos(t, srv, ListTodosInput{View: "inbox", Limit: 100})
	if len(inbox.Items) == 0 {
		t.Fatal("fixture has no inbox todo to annotate")
	}
	target := inbox.Items[0].UUID
	const noteRunes = 250
	execFixture(t, path, "UPDATE TMTask SET notes = ? WHERE uuid = ?", strings.Repeat("字", noteRunes), target)

	// List truncates and flags.
	listed := findItem(t, listTodos(t, srv, ListTodosInput{View: "inbox", Limit: 100}).Items, target)
	if got := utf8.RuneCountInString(listed.Notes); got != notesLimit {
		t.Errorf("listed note = %d runes, want %d", got, notesLimit)
	}
	if !listed.NotesTruncated {
		t.Error("listed note must be flagged truncated")
	}

	// Search truncates the same way.
	hit := findItem(t, searchItems(t, srv, "字"), target)
	if utf8.RuneCountInString(hit.Notes) != notesLimit || !hit.NotesTruncated {
		t.Errorf("search hit note = %d runes flagged=%v, want %d truncated",
			utf8.RuneCountInString(hit.Notes), hit.NotesTruncated, notesLimit)
	}

	// get keeps the full note, never flagged.
	_, got, err := srv.handleGet(context.Background(), nil, GetInput{ID: target})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Item == nil {
		t.Fatal("get returned no item")
	}
	if n := utf8.RuneCountInString(got.Item.Notes); n != noteRunes {
		t.Errorf("get note = %d runes, want full %d", n, noteRunes)
	}
	if got.Item.NotesTruncated {
		t.Error("get must never flag truncation")
	}
}

// findItem returns the item with the given UUID, failing if it is absent.
func findItem(t *testing.T, items []Item, uuid string) Item {
	t.Helper()
	for i := range items {
		if items[i].UUID == uuid {
			return items[i]
		}
	}
	t.Fatalf("item %s not found in %d results", uuid, len(items))
	return Item{}
}

// searchItems runs the search tool and returns its items.
func searchItems(t *testing.T, srv *Server, query string) []Item {
	t.Helper()
	_, page, err := srv.handleSearch(context.Background(), nil, SearchInput{Query: query, Limit: 100})
	if err != nil {
		t.Fatalf("search %q: %v", query, err)
	}
	return page.Items
}
