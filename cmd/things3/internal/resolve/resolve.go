// Package resolve turns human queries - UUIDs, UUID prefixes, or titles - into
// the exact todos and projects they name. It is deliberately cobra-free so it
// can be unit-tested against the fixture and later promoted into the library.
package resolve

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/moond4rk/things3"
)

// Kind distinguishes the item types a query can resolve to.
type Kind string

const (
	// KindTodo marks a match that is a todo.
	KindTodo Kind = "todo"
	// KindProject marks a match that is a project.
	KindProject Kind = "project"
	// KindArea marks a match that is an area.
	KindArea Kind = "area"
	// KindHeading marks a match that is a heading.
	KindHeading Kind = "heading"
)

// uuidPrefixMinLen guards prefix matching so short words do not collide with the
// head of a random UUID.
const uuidPrefixMinLen = 4

// Preference steers which match wins when a query names several items.
type Preference int

const (
	// PreferOpen picks the single open item among the matches. It is the right
	// reading of every lookup and write except reopening: "done Buy milk" means
	// the open "Buy milk", however many completed ones the Logbook holds.
	PreferOpen Preference = iota
	// PreferClosed picks the single closed item; it is the reading of "reopen".
	PreferClosed
)

// Match is a single resolved item. Exactly one pointer is set, per Kind.
type Match struct {
	Kind    Kind
	Todo    *things3.Todo
	Project *things3.Project
	Area    *things3.Area
	Heading *things3.Heading
}

// UUID returns the matched item's UUID.
func (m Match) UUID() string {
	switch m.Kind {
	case KindProject:
		return m.Project.UUID
	case KindArea:
		return m.Area.UUID
	case KindHeading:
		return m.Heading.UUID
	default:
		return m.Todo.UUID
	}
}

// Title returns the matched item's title.
func (m Match) Title() string {
	switch m.Kind {
	case KindProject:
		return m.Project.Title
	case KindArea:
		return m.Area.Title
	case KindHeading:
		return m.Heading.Title
	default:
		return m.Todo.Title
	}
}

// Open reports whether the matched item is still actionable. Areas and headings
// have no status and are always considered open.
func (m Match) Open() bool {
	switch m.Kind {
	case KindProject:
		return m.Project.Status.IsOpen()
	case KindArea, KindHeading:
		return true
	default:
		return m.Todo.Status.IsOpen()
	}
}

// Trashed reports whether the matched item is in the Trash.
func (m Match) Trashed() bool {
	switch m.Kind {
	case KindProject:
		return m.Project.Trashed
	case KindTodo:
		return m.Todo.Trashed
	default:
		return false
	}
}

// Template reports whether the matched item is a repeating template rather than
// an actionable item.
func (m Match) Template() bool {
	switch m.Kind {
	case KindProject:
		return m.Project.Template
	case KindTodo:
		return m.Todo.Template
	default:
		return false
	}
}

// Live reports whether the matched item shows up in the app's working lists:
// neither trashed nor a repeating template.
func (m Match) Live() bool {
	return !m.Trashed() && !m.Template()
}

// NotFoundError reports that a query matched no items.
type NotFoundError struct{ Query string }

func (e *NotFoundError) Error() string { return fmt.Sprintf("no item matches %q", e.Query) }

// AmbiguousError reports that a query matched more than one item.
type AmbiguousError struct {
	Query   string
	Matches []Match
}

func (e *AmbiguousError) Error() string {
	return fmt.Sprintf("query %q matches %d items", e.Query, len(e.Matches))
}

// ReadOnlyError reports that a query resolved to an item the URL scheme cannot
// modify. The item exists and can still be shown.
type ReadOnlyError struct {
	Match  Match
	Reason string
}

func (e *ReadOnlyError) Error() string {
	return fmt.Sprintf("%q %s", e.Match.Title(), e.Reason)
}

// Resolve finds todos and projects matching query via tiered matching: exact
// UUID, then UUID prefix (>= 4 chars), then exact title (case-insensitive),
// then title substring. The first tier with any hit wins; later tiers are not
// consulted. The UUID tiers also reach trashed items and repeating templates,
// so every UUID a view prints resolves; the title tiers match live items only.
func Resolve(ctx context.Context, c *things3.Client, query string) ([]Match, error) {
	matches, err := lookup(ctx, identityTodos(c).WithUUID(query), identityProjects(c).WithUUID(query))
	if err != nil {
		return nil, err
	}
	if len(matches) > 0 {
		return rank(matches), nil
	}

	if len(query) >= uuidPrefixMinLen {
		matches, err = lookup(ctx, identityTodos(c).WithUUIDPrefix(query), identityProjects(c).WithUUIDPrefix(query))
		if err != nil {
			return nil, err
		}
		if len(matches) > 0 {
			return rank(matches), nil
		}
	}

	// WithTitle is an escaped substring match, a superset of exact-title.
	all, err := lookup(ctx, c.Todos().WithTitle(query).Status().Any(), c.Projects().WithTitle(query).Status().Any())
	if err != nil {
		return nil, err
	}
	if exact := exactTitle(all, query); len(exact) > 0 {
		return rank(exact), nil
	}
	return rank(all), nil
}

// ResolveOne resolves a lookup to exactly one item or a typed error.
func ResolveOne(ctx context.Context, c *things3.Client, query string, prefer Preference) (Match, error) {
	matches, err := Resolve(ctx, c, query)
	if err != nil {
		return Match{}, err
	}
	return pick(query, matches, prefer)
}

// ResolveTarget resolves the target of a write: exactly one item, and one the
// URL scheme is able to modify.
func ResolveTarget(ctx context.Context, c *things3.Client, query string, prefer Preference) (Match, error) {
	m, err := ResolveOne(ctx, c, query, prefer)
	if err != nil {
		return Match{}, err
	}
	switch {
	case m.Trashed():
		return Match{}, &ReadOnlyError{Match: m, Reason: "is in the Trash; restore it in Things before changing it"}
	case m.Template():
		return Match{}, &ReadOnlyError{Match: m, Reason: "is a repeating series, which only the Things app itself can change"}
	default:
		return m, nil
	}
}

// Project resolves a destination query to exactly one live project (UUID,
// prefix, exact title, then title substring), preferring an open one.
func Project(ctx context.Context, c *things3.Client, q string) (*things3.Project, error) {
	live := c.Projects().Status().Any()
	projects, err := live.WithUUID(q).All(ctx)
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 && len(q) >= uuidPrefixMinLen {
		if projects, err = live.WithUUIDPrefix(q).All(ctx); err != nil {
			return nil, err
		}
	}
	matches := toMatches(nil, projects)
	if len(matches) == 0 {
		titled, terr := live.WithTitle(q).All(ctx)
		if terr != nil {
			return nil, terr
		}
		matches = toMatches(nil, titled)
		if exact := exactTitle(matches, q); len(exact) > 0 {
			matches = exact
		}
	}
	m, err := pick(q, rank(matches), PreferOpen)
	if err != nil {
		return nil, err
	}
	return m.Project, nil
}

// Area resolves a query to exactly one area. Areas are few, so all are fetched
// once and matched client-side (AreaQueryBuilder.WithTitle is exact-equality).
func Area(ctx context.Context, c *things3.Client, q string) (*things3.Area, error) {
	areas, err := c.Areas().All(ctx)
	if err != nil {
		return nil, err
	}
	hits := tieredFilter(areas, q,
		func(a things3.Area) string { return a.UUID },
		func(a things3.Area) string { return a.Title })
	matches := make([]Match, len(hits))
	for i := range hits {
		matches[i] = Match{Kind: KindArea, Area: &hits[i]}
	}
	m, err := pick(q, matches, PreferOpen)
	if err != nil {
		return nil, err
	}
	return m.Area, nil
}

// Heading resolves a query to exactly one heading within the given project.
func Heading(ctx context.Context, c *things3.Client, projectUUID, q string) (*things3.Heading, error) {
	headings, err := c.Headings().InProject(projectUUID).All(ctx)
	if err != nil {
		return nil, err
	}
	hits := tieredFilter(headings, q,
		func(h things3.Heading) string { return h.UUID },
		func(h things3.Heading) string { return h.Title })
	matches := make([]Match, len(hits))
	for i := range hits {
		matches[i] = Match{Kind: KindHeading, Heading: &hits[i]}
	}
	m, err := pick(q, matches, PreferOpen)
	if err != nil {
		return nil, err
	}
	return m.Heading, nil
}

// identityTodos is the scope of a lookup by UUID: every todo that exists,
// including the trashed ones and repeating templates the views print.
func identityTodos(c *things3.Client) things3.TodoQueryBuilder {
	return c.Todos().Status().Any().IncludeTrashed().IncludeTemplates()
}

// identityProjects mirrors identityTodos for projects.
func identityProjects(c *things3.Client) things3.ProjectQueryBuilder {
	return c.Projects().Status().Any().IncludeTrashed().IncludeTemplates()
}

func lookup(ctx context.Context, todoQuery things3.TodoQueryBuilder, projectQuery things3.ProjectQueryBuilder) ([]Match, error) {
	todos, err := todoQuery.All(ctx)
	if err != nil {
		return nil, err
	}
	projects, err := projectQuery.All(ctx)
	if err != nil {
		return nil, err
	}
	return toMatches(todos, projects), nil
}

func toMatches(todos []things3.Todo, projects []things3.Project) []Match {
	matches := make([]Match, 0, len(todos)+len(projects))
	for i := range todos {
		matches = append(matches, Match{Kind: KindTodo, Todo: &todos[i]})
	}
	for i := range projects {
		matches = append(matches, Match{Kind: KindProject, Project: &projects[i]})
	}
	return matches
}

func exactTitle(matches []Match, title string) []Match {
	return keep(matches, func(m Match) bool { return strings.EqualFold(m.Title(), title) })
}

func keep(matches []Match, want func(Match) bool) []Match {
	kept := make([]Match, 0, len(matches))
	for _, m := range matches {
		if want(m) {
			kept = append(kept, m)
		}
	}
	return kept
}

// rank stably orders matches: live before trashed or template, open before
// closed, then todos before projects, preserving the original query order for
// otherwise-equal items.
func rank(matches []Match) []Match {
	slices.SortStableFunc(matches, func(a, b Match) int {
		if c := boolFirst(a.Live(), b.Live()); c != 0 {
			return c
		}
		if c := boolFirst(a.Open(), b.Open()); c != 0 {
			return c
		}
		return boolFirst(a.Kind == KindTodo, b.Kind == KindTodo)
	})
	return matches
}

// boolFirst orders true before false.
func boolFirst(a, b bool) int {
	switch {
	case a == b:
		return 0
	case a:
		return -1
	default:
		return 1
	}
}

// pick reduces matches to exactly one. Several matches narrow first to the live
// ones, then to the preferred status; a narrowing that would leave nothing is
// skipped, so a lone trashed or closed match is still returned.
func pick(query string, matches []Match, prefer Preference) (Match, error) {
	if len(matches) == 0 {
		return Match{}, &NotFoundError{Query: query}
	}
	preferred := func(m Match) bool { return m.Open() == (prefer == PreferOpen) }
	for _, want := range []func(Match) bool{Match.Live, preferred} {
		if kept := keep(matches, want); len(kept) > 0 {
			matches = kept
		}
	}
	if len(matches) > 1 {
		return Match{}, &AmbiguousError{Query: query, Matches: matches}
	}
	return matches[0], nil
}

// tieredFilter applies the client-side tiers exact UUID, then title EqualFold,
// then title substring, returning the first non-empty tier.
func tieredFilter[T any](items []T, q string, uuidOf, titleOf func(T) string) []T {
	var byUUID, byExact, bySubstr []T
	lower := strings.ToLower(q)
	for i := range items {
		switch {
		case uuidOf(items[i]) == q:
			byUUID = append(byUUID, items[i])
		case strings.EqualFold(titleOf(items[i]), q):
			byExact = append(byExact, items[i])
		case strings.Contains(strings.ToLower(titleOf(items[i])), lower):
			bySubstr = append(bySubstr, items[i])
		}
	}
	switch {
	case len(byUUID) > 0:
		return byUUID
	case len(byExact) > 0:
		return byExact
	default:
		return bySubstr
	}
}
