package change

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
	"github.com/moond4rk/things3/thingstest"
)

const (
	todoInToday = thingstest.UUIDTodoInToday // scheduled 2021-03-28
	todoAnytime = "QqhVksfbsAVaNnwB1x3CuD"   // no scheduled day
	project     = thingstest.UUIDProject
)

func clientAt(t *testing.T, path string) *things3.Client {
	t.Helper()
	client, err := things3.NewClient(things3.WithDatabasePath(path), things3.WithClock(thingstest.Now))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func target(t *testing.T, c *things3.Client, query string) resolve.Match {
	t.Helper()
	m, err := resolve.ResolveTarget(context.Background(), c, query, resolve.PreferOpen)
	if err != nil {
		t.Fatalf("resolve %q: %v", query, err)
	}
	return m
}

// built applies a change and returns the query of the URL it would execute.
func built(t *testing.T, b Builder, err error) url.Values {
	t.Helper()
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	raw, err := b.Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	return u.Query()
}

func isInputError(err error) bool {
	_, ok := errors.AsType[*InputError](err)
	return ok
}

func TestEditAttributes(t *testing.T) {
	c := clientAt(t, thingstest.DatabasePath(t))
	m := target(t, c, todoInToday)
	deadline := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name  string
		edit  Edit
		param string
		want  string
	}{
		{"title", Edit{Title: new("New title")}, "title", "New title"},
		{"notes replace", Edit{Notes: new("fresh")}, "notes", "fresh"},
		{"notes clear", Edit{ClearNotes: true}, "notes", ""},
		{"notes prepend", Edit{PrependNotes: "first"}, "prepend-notes", "first"},
		{"notes append", Edit{AppendNotes: "more"}, "append-notes", "more"},
		{"deadline", Edit{Deadline: &deadline}, "deadline", "2026-08-01"},
		{"deadline clear", Edit{ClearDeadline: true}, "deadline", ""},
		{"tags replace", Edit{Tags: []string{"a", "b"}}, "tags", "a,b"},
		{"tags clear", Edit{ClearTags: true}, "tags", ""},
		{"tags add", Edit{AddTags: []string{"urgent"}}, "add-tags", "urgent"},
		{"checklist replace", Edit{Checklist: []string{"one", "two"}}, "checklist-items", "one\ntwo"},
		{"checklist clear", Edit{ClearChecklist: true}, "checklist-items", ""},
		{"checklist prepend", Edit{PrependChecklist: []string{"top"}}, "prepend-checklist-items", "top"},
		{"checklist append", Edit{AppendChecklist: []string{"x", "y"}}, "append-checklist-items", "x\ny"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.edit.Apply(c, m)
			q := built(t, b, err)
			got, sent := q[tc.param]
			if !sent || got[0] != tc.want {
				t.Errorf("%s = %q (sent %v), want %q", tc.param, got, sent, tc.want)
			}
			if q.Get("id") != todoInToday {
				t.Errorf("id = %q", q.Get("id"))
			}
		})
	}
}

// The URL scheme carries a reminder inside "when", so setting one re-sends the
// day the item already has rather than letting it default to today.
func TestEditReminderKeepsTheScheduledDay(t *testing.T) {
	path := thingstest.DatabasePath(t)
	c := clientAt(t, path)
	reminder := &Clock{Hour: 9, Minute: 30}

	b, err := (&Edit{Reminder: reminder}).Apply(c, target(t, c, todoInToday))
	if got := built(t, b, err).Get("when"); got != "2021-03-28@09:30" {
		t.Errorf("when = %q, want the existing day with the reminder", got)
	}

	_, err = (&Edit{Reminder: reminder}).Apply(c, target(t, c, todoAnytime))
	if !isInputError(err) {
		t.Errorf("a reminder on an unscheduled item must be rejected, got %v", err)
	}

	// A This Evening todo must stay in This Evening.
	raw, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer raw.Close()
	if _, execErr := raw.ExecContext(t.Context(), "UPDATE TMTask SET startBucket = 1 WHERE uuid = ?", todoInToday); execErr != nil {
		t.Fatalf("mark evening: %v", execErr)
	}
	b, err = (&Edit{Reminder: reminder}).Apply(c, target(t, c, todoInToday))
	if got := built(t, b, err).Get("when"); got != "evening@09:30" {
		t.Errorf("when = %q, want evening@09:30", got)
	}
}

func TestEditValidation(t *testing.T) {
	c := clientAt(t, thingstest.DatabasePath(t))
	todo, proj := target(t, c, todoInToday), target(t, c, project)
	day := time.Now()

	cases := []struct {
		name string
		edit Edit
		m    resolve.Match
	}{
		{"empty", Edit{}, todo},
		{"notes and clear", Edit{Notes: new("x"), ClearNotes: true}, todo},
		{"deadline and clear", Edit{Deadline: &day, ClearDeadline: true}, todo},
		{"tags and clear", Edit{Tags: []string{"a"}, ClearTags: true}, todo},
		{"checklist and clear", Edit{Checklist: []string{"a"}, ClearChecklist: true}, todo},
		{"checklist on a project", Edit{AppendChecklist: []string{"a"}}, proj},
		{"clear checklist on a project", Edit{ClearChecklist: true}, proj},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.edit.Apply(c, tc.m); !isInputError(err) {
				t.Errorf("want *InputError, got %v", err)
			}
		})
	}

	b, err := (&Edit{Title: new("Renamed"), ClearTags: true}).Apply(c, proj)
	if q := built(t, b, err); q.Get("title") != "Renamed" || q.Get("id") != project {
		t.Errorf("a project accepts the shared attributes, got %v", q)
	}
}

func TestSchedule(t *testing.T) {
	c := clientAt(t, thingstest.DatabasePath(t))
	todo, proj := target(t, c, todoInToday), target(t, c, project)

	for _, m := range []resolve.Match{todo, proj} {
		b, err := Schedule{When: "2026-08-01"}.Apply(c, m)
		if got := built(t, b, err).Get("when"); got != "2026-08-01" {
			t.Errorf("%s when = %q", m.Kind, got)
		}
		b, err = Schedule{When: "evening", Reminder: &Clock{Hour: 18}}.Apply(c, m)
		if got := built(t, b, err).Get("when"); got != "evening@18:00" {
			t.Errorf("%s when with reminder = %q", m.Kind, got)
		}
	}

	if _, err := (Schedule{When: "tomorow"}).Apply(c, todo); !isInputError(err) {
		t.Errorf("an unknown when must be an *InputError, got %v", err)
	}

	// A reminder needs a day; the library refuses the combination when building.
	b, err := Schedule{When: "someday", Reminder: &Clock{Hour: 9}}.Apply(c, todo)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if _, err := b.Build(); !errors.Is(err, things3.ErrReminderNeedsDate) {
		t.Errorf("someday with a reminder should fail to build, got %v", err)
	}
}

func TestParse(t *testing.T) {
	if clock, err := ParseClock("07:05"); err != nil || clock != (Clock{Hour: 7, Minute: 5}) {
		t.Errorf("ParseClock = %+v, %v", clock, err)
	}
	if _, err := ParseClock("25:00"); !isInputError(err) {
		t.Errorf("ParseClock should reject 25:00, got %v", err)
	}
	if d, err := ParseDate("deadline", "2026-08-01"); err != nil || d.Day() != 1 || d.Month() != time.August {
		t.Errorf("ParseDate = %v, %v", d, err)
	}
	if _, err := ParseDate("deadline", "next week"); !isInputError(err) {
		t.Errorf("ParseDate should reject free text, got %v", err)
	}
}
