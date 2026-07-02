package resolve

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/thingstest"
)

func newFixtureClient(t *testing.T) *things3.Client {
	t.Helper()
	client, err := things3.NewClient(things3.WithDatabasePath(thingstest.DatabasePath(t)))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
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

	_, err = ResolveOne(ctx, c, "in Today")
	var ambiguous *AmbiguousError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("want *AmbiguousError, got %v", err)
	}
	if len(ambiguous.Matches) != len(matches) {
		t.Errorf("AmbiguousError should carry all %d candidates, got %d", len(matches), len(ambiguous.Matches))
	}
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
	_, err := ResolveOne(context.Background(), c, "zzz-no-such-item-zzz")
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
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

	_, err = Project(ctx, c, "Project in")
	var ambiguous *AmbiguousError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("want *AmbiguousError for %q, got %v", "Project in", err)
	}
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

	_, err = Area(ctx, c, "Area")
	var ambiguous *AmbiguousError
	if !errors.As(err, &ambiguous) {
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
	_, err = Heading(ctx, c, twoHeadingProject, "Heading")
	var ambiguous *AmbiguousError
	if !errors.As(err, &ambiguous) {
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
	var notFound *NotFoundError
	if _, err := Project(ctx, c, "zzznope"); !errors.As(err, &notFound) {
		t.Errorf("Project not-found should be *NotFoundError, got %v", err)
	}
	if _, err := Area(ctx, c, "zzznope"); !errors.As(err, &notFound) {
		t.Errorf("Area not-found should be *NotFoundError, got %v", err)
	}
	if _, err := Heading(ctx, c, thingstest.UUIDProject, "zzznope"); !errors.As(err, &notFound) {
		t.Errorf("Heading not-found should be *NotFoundError, got %v", err)
	}
}
