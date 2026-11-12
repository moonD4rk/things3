package resolve

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/thingstest"
)

// Fixture rows the resolver tests lean on.
const (
	uuidTrashedTodo      = "Tmt3omxtu61wAoTaSLweQy"
	uuidTemplate         = "N1PJHsbjct4mb1bhcs7aHa" // "Repeating To-Do", the series
	uuidTemplateInstance = "K9bx7h1xCJdevvyWardZDq" // "Repeating To-Do", today's copy
	uuidCompletedInToday = "56dtXSk3A373M6n4eqGyr3"
)

func newFixtureClient(t *testing.T) *things3.Client {
	t.Helper()
	return newClientAt(t, thingstest.DatabasePath(t))
}

func newClientAt(t *testing.T, path string) *things3.Client {
	t.Helper()
	client, err := things3.NewClient(things3.WithDatabasePath(path), things3.WithClock(thingstest.Now))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// mutateFixture returns a fixture copy with the statements applied.
func mutateFixture(t *testing.T, statements ...string) string {
	t.Helper()
	path := thingstest.DatabasePath(t)
	raw, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer raw.Close()
	for _, stmt := range statements {
		if _, err := raw.ExecContext(t.Context(), stmt); err != nil {
			t.Fatalf("mutate fixture: %v", err)
		}
	}
	return path
}

func containsUUID(matches []Match, uuid string) bool {
	for _, m := range matches {
		if m.UUID() == uuid {
			return true
		}
	}
	return false
}

func TestResolveExactUUID(t *testing.T) {
	c := newFixtureClient(t)
	ctx := context.Background()

	matches, err := Resolve(ctx, c, thingstest.UUIDTodoInToday)
	if err != nil {
		t.Fatalf("Resolve todo: %v", err)
	}
	if len(matches) != 1 || matches[0].Kind != KindTodo || matches[0].UUID() != thingstest.UUIDTodoInToday {
		t.Fatalf("want single todo %s, got %+v", thingstest.UUIDTodoInToday, matches)
	}

	pmatches, err := Resolve(ctx, c, thingstest.UUIDProject)
	if err != nil {
		t.Fatalf("Resolve project: %v", err)
	}
	if len(pmatches) != 1 || pmatches[0].Kind != KindProject {
		t.Fatalf("want single project match, got %+v", pmatches)
	}
}

func TestResolveUUIDPrefixAndLenGuard(t *testing.T) {
	c := newFixtureClient(t)
	ctx := context.Background()
	uuid := thingstest.UUIDTodoInToday

	matches, err := Resolve(ctx, c, uuid[:8])
	if err != nil {
		t.Fatalf("Resolve prefix: %v", err)
	}
	if !containsUUID(matches, uuid) {
		t.Errorf("8-char prefix %q should resolve to %s", uuid[:8], uuid)
	}

	// A 3-char query skips prefix matching; the head of this UUID is not a
	// title substring in the fixture, so it resolves to nothing.
	short, err := Resolve(ctx, c, uuid[:3])
	if err != nil {
		t.Fatalf("Resolve short: %v", err)
	}
	if containsUUID(short, uuid) {
		t.Errorf("3-char query %q must not match UUID %s (len guard)", uuid[:3], uuid)
	}
}

func TestResolveExactTitleBeatsSubstring(t *testing.T) {
	c := newFixtureClient(t)
	ctx := context.Background()

	matches, err := Resolve(ctx, c, "To-Do in Today")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("exact title should yield 1 match, got %d: %+v", len(matches), matches)
	}
	if matches[0].Title() != "To-Do in Today" {
		t.Errorf("want title %q, got %q", "To-Do in Today", matches[0].Title())
	}

	// The substring query alone matches more than one, proving the exact-title
	// tier discarded the rest.
	sub, err := c.Todos().WithTitle("To-Do in Today").Status().Any().All(ctx)
	if err != nil {
		t.Fatalf("WithTitle: %v", err)
	}
	if len(sub) < 2 {
		t.Fatalf("expected substring query to match >= 2 todos, got %d (fixture drift?)", len(sub))
	}
}

func TestResolveAmbiguousAndRanking(t *testing.T) {
	c := newFixtureClient(t)
	ctx := context.Background()

	matches, err := Resolve(ctx, c, "in Today")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(matches) < 3 {
		t.Fatalf("expected several matches for %q, got %d", "in Today", len(matches))
	}
	seenClosed := false
	for _, m := range matches {
		if !m.Open() {
			seenClosed = true
			continue
		}
		if seenClosed {
			t.Errorf("open item %q ranked after a closed item", m.Title())
		}
	}

	// Among equally-open matches, todos rank before projects.
	lastOpenTodo, firstOpenProject := -1, -1
	for i, m := range matches {
		if !m.Open() {
			continue
		}
		switch m.Kind {
		case KindTodo:
			lastOpenTodo = i
		case KindProject:
			if firstOpenProject < 0 {
				firstOpenProject = i
			}
		}
	}
	if firstOpenProject >= 0 && lastOpenTodo >= 0 && firstOpenProject < lastOpenTodo {
		t.Errorf("an open project ranked before an open todo (want todos first)")
	}

	// Several open items share the substring, so it stays ambiguous, but the
	// closed ones are no longer candidates for the caller to wade through.
	_, err = ResolveOne(ctx, c, "in Today", PreferOpen)
	ambiguous, ok := errors.AsType[*AmbiguousError](err)
	if !ok {
		t.Fatalf("want *AmbiguousError, got %v", err)
	}
	open := keep(matches, Match.Open)
	if len(open) < 2 || len(open) == len(matches) {
		t.Fatalf("fixture drift: want several open and some closed matches, got %d open of %d", len(open), len(matches))
	}
	if len(ambiguous.Matches) != len(open) {
		t.Errorf("AmbiguousError should carry the %d open candidates, got %d", len(open), len(ambiguous.Matches))
	}
}

// One open item among same-titled closed ones is what a repeating task looks
// like after a few completions; naming it must not be ambiguous.
func TestResolveOnePrefersTheSingleOpenMatch(t *testing.T) {
	path := mutateFixture(t,
		"UPDATE TMTask SET title = 'To-Do in Today' WHERE uuid = '"+uuidCompletedInToday+"'")
	c := newClientAt(t, path)
	ctx := context.Background()

	matches, err := Resolve(ctx, c, "To-Do in Today")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("want the open and the completed todo, got %d", len(matches))
	}

	open, err := ResolveOne(ctx, c, "To-Do in Today", PreferOpen)
	if err != nil {
		t.Fatalf("PreferOpen: %v", err)
	}
	if open.UUID() != thingstest.UUIDTodoInToday {
		t.Errorf("PreferOpen picked %s, want the open todo", open.UUID())
	}

	closed, err := ResolveOne(ctx, c, "To-Do in Today", PreferClosed)
	if err != nil {
		t.Fatalf("PreferClosed: %v", err)
	}
	if closed.UUID() != uuidCompletedInToday {
		t.Errorf("PreferClosed picked %s, want the completed todo", closed.UUID())
	}
}

// A lone match is returned whatever its status: completing a completed item is
// the caller's business, not an ambiguity.
func TestResolveOneKeepsALoneNonPreferredMatch(t *testing.T) {
	c := newFixtureClient(t)
	m, err := ResolveOne(context.Background(), c, "Completed To-Do in Today", PreferOpen)
	if err != nil {
		t.Fatalf("ResolveOne: %v", err)
	}
	if m.UUID() != uuidCompletedInToday {
		t.Errorf("got %s, want the completed todo", m.UUID())
	}
}

// Views print the UUIDs of trashed items and repeating templates, so those
// UUIDs must resolve; the URL scheme cannot change either, so writes refuse.
func TestResolveReachesTrashedAndTemplatesByUUIDOnly(t *testing.T) {
	c := newFixtureClient(t)
	ctx := context.Background()

	for name, tc := range map[string]struct {
		uuid     string
		trashed  bool
		template bool
	}{
		"trashed todo":       {uuid: uuidTrashedTodo, trashed: true},
		"repeating template": {uuid: uuidTemplate, template: true},
	} {
		t.Run(name, func(t *testing.T) {
			for _, query := range []string{tc.uuid, tc.uuid[:8]} {
				m, err := ResolveOne(ctx, c, query, PreferOpen)
				if err != nil {
					t.Fatalf("ResolveOne(%q): %v", query, err)
				}
				if m.UUID() != tc.uuid || m.Trashed() != tc.trashed || m.Template() != tc.template || m.Live() {
					t.Errorf("ResolveOne(%q) = %s trashed=%v template=%v", query, m.UUID(), m.Trashed(), m.Template())
				}

				_, err = ResolveTarget(ctx, c, query, PreferOpen)
				readOnly, ok := errors.AsType[*ReadOnlyError](err)
				if !ok {
					t.Fatalf("ResolveTarget(%q): want *ReadOnlyError, got %v", query, err)
				}
				if readOnly.Match.UUID() != tc.uuid {
					t.Errorf("ReadOnlyError names %s, want %s", readOnly.Match.UUID(), tc.uuid)
				}
			}
		})
	}

	// By title the template stays hidden: the title names today's copy.
	m, err := ResolveTarget(ctx, c, "Repeating To-Do", PreferOpen)
	if err != nil {
		t.Fatalf("ResolveTarget by title: %v", err)
	}
	if m.UUID() != uuidTemplateInstance {
		t.Errorf("title resolved to %s, want the live instance %s", m.UUID(), uuidTemplateInstance)
	}
	if _, err := ResolveOne(ctx, c, "Another Deleted Todo", PreferOpen); !isNotFound(err) {
		t.Errorf("a trashed item must not match by title, got %v", err)
	}
}

func isNotFound(err error) bool {
	_, ok := errors.AsType[*NotFoundError](err)
	return ok
}

func TestResolveWildcardsAreLiteral(t *testing.T) {
	c := newFixtureClient(t)
	ctx := context.Background()
	// No fixture title contains % or _, so LIKE wildcards must not match all.
	for _, q := range []string{"%", "_"} {
		matches, err := Resolve(ctx, c, q)
		if err != nil {
			t.Fatalf("Resolve(%q): %v", q, err)
		}
		if len(matches) != 0 {
			t.Errorf("query %q treated as wildcard: matched %d, want 0", q, len(matches))
		}
	}
}

func TestResolveOneNotFound(t *testing.T) {
	c := newFixtureClient(t)
	_, err := ResolveOne(context.Background(), c, "zzz-no-such-item-zzz", PreferOpen)
	if !isNotFound(err) {
		t.Fatalf("want *NotFoundError, got %v", err)
	}
}

func TestProjectResolve(t *testing.T) {
	c := newFixtureClient(t)
	ctx := context.Background()

	p, err := Project(ctx, c, "Project in Today")
	if err != nil {
		t.Fatalf("Project exact title: %v", err)
	}
	if p.Title != "Project in Today" {
		t.Errorf("want %q, got %q", "Project in Today", p.Title)
	}

	if _, err = Project(ctx, c, "Project in"); !isAmbiguous(err) {
		t.Fatalf("want *AmbiguousError for %q, got %v", "Project in", err)
	}
}

func isAmbiguous(err error) bool {
	_, ok := errors.AsType[*AmbiguousError](err)
	return ok
}

func TestAreaResolve(t *testing.T) {
	c := newFixtureClient(t)
	ctx := context.Background()

	a, err := Area(ctx, c, thingstest.UUIDArea)
	if err != nil {
		t.Fatalf("Area by UUID: %v", err)
	}
	if a.UUID != thingstest.UUIDArea {
		t.Errorf("want area %s, got %s", thingstest.UUIDArea, a.UUID)
	}

	byTitle, err := Area(ctx, c, "Area 1")
	if err != nil {
		t.Fatalf("Area by title: %v", err)
	}
	if byTitle.Title != "Area 1" {
		t.Errorf("want Area 1, got %q", byTitle.Title)
	}

	if _, err = Area(ctx, c, "Area"); !isAmbiguous(err) {
		t.Fatalf("want *AmbiguousError for %q, got %v", "Area", err)
	}
}

func TestHeadingResolve(t *testing.T) {
	c := newFixtureClient(t)
	ctx := context.Background()

	// UUIDProject has exactly one heading, "Heading".
	h, err := Heading(ctx, c, thingstest.UUIDProject, "Heading")
	if err != nil {
		t.Fatalf("Heading: %v", err)
	}
	if h.Title != "Heading" {
		t.Errorf("want Heading, got %q", h.Title)
	}

	// A project with two headings, matched by a shared substring -> ambiguous.
	const twoHeadingProject = "TCozQqXVbB2TJkXXXQj2H9"
	if _, err = Heading(ctx, c, twoHeadingProject, "Heading"); !isAmbiguous(err) {
		t.Fatalf("want *AmbiguousError for two-heading project, got %v", err)
	}
}

func TestResolveExactTitleCaseInsensitive(t *testing.T) {
	c := newFixtureClient(t)
	matches, err := Resolve(context.Background(), c, "to-do in today")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(matches) != 1 || !strings.EqualFold(matches[0].Title(), "To-Do in Today") {
		t.Fatalf("case-insensitive exact title should collapse to the single match, got %+v", matches)
	}
}

func TestNarrowResolveNotFound(t *testing.T) {
	c := newFixtureClient(t)
	ctx := context.Background()
	if _, err := Project(ctx, c, "zzznope"); !isNotFound(err) {
		t.Errorf("Project not-found should be *NotFoundError, got %v", err)
	}
	if _, err := Area(ctx, c, "zzznope"); !isNotFound(err) {
		t.Errorf("Area not-found should be *NotFoundError, got %v", err)
	}
	if _, err := Heading(ctx, c, thingstest.UUIDProject, "zzznope"); !isNotFound(err) {
		t.Errorf("Heading not-found should be *NotFoundError, got %v", err)
	}
}
