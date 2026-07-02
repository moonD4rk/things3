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

const (
	// uuidPrefixMinLen guards prefix matching so short words do not collide
	// with the head of a random UUID.
	uuidPrefixMinLen = 4
	// titleQueryLimit caps title scans; resolution is interactive, not a report.
	titleQueryLimit = 200
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

// Resolve finds todos and projects matching query via tiered matching: exact
// UUID, then UUID prefix (>= 4 chars), then exact title (case-insensitive),
// then title substring. The first tier with any hit wins; later tiers are not
// consulted. Trashed items never match (library queries exclude them).
func Resolve(ctx context.Context, c *things3.Client, query string) ([]Match, error) {
	// Tier 1: exact UUID.
	matches, err := queryByUUID(ctx, c, query)
	if err != nil {
		return nil, err
	}
	if len(matches) > 0 {
		return rank(matches), nil
	}

	// Tier 2: UUID prefix.
	if len(query) >= uuidPrefixMinLen {
		matches, err = queryByUUIDPrefix(ctx, c, query)
		if err != nil {
			return nil, err
		}
		if len(matches) > 0 {
			return rank(matches), nil
		}
	}

	// Tiers 3 and 4: one title query per type (WithTitle is escaped substring
	// LIKE, a superset of exact-title). Exact-title matches, if any, win.
	todos, err := c.Todos().WithTitle(query).Status().Any().Limit(titleQueryLimit).All(ctx)
	if err != nil {
		return nil, err
	}
	projects, err := c.Projects().WithTitle(query).Status().Any().Limit(titleQueryLimit).All(ctx)
	if err != nil {
		return nil, err
	}
	all := toMatches(todos, projects)
	exact := make([]Match, 0, len(all))
	for _, m := range all {
		if strings.EqualFold(m.Title(), query) {
			exact = append(exact, m)
		}
	}
	if len(exact) > 0 {
		return rank(exact), nil
	}
	return rank(all), nil
}

// ResolveOne resolves a write target: exactly one match or a typed error.
func ResolveOne(ctx context.Context, c *things3.Client, query string) (Match, error) {
	matches, err := Resolve(ctx, c, query)
	if err != nil {
		return Match{}, err
	}
	return one(query, matches)
}

// Project resolves a query to exactly one project (UUID, prefix, then title).
func Project(ctx context.Context, c *things3.Client, q string) (*things3.Project, error) {
	projects, err := c.Projects().WithUUID(q).Status().Any().All(ctx)
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 && len(q) >= uuidPrefixMinLen {
		if projects, err = c.Projects().WithUUIDPrefix(q).Status().Any().All(ctx); err != nil {
			return nil, err
		}
	}
	if len(projects) == 0 {
		titled, terr := c.Projects().WithTitle(q).Status().Any().Limit(titleQueryLimit).All(ctx)
		if terr != nil {
			return nil, terr
		}
		projects = filterExactTitleProjects(titled, q)
		if len(projects) == 0 {
			projects = titled
		}
	}
	m, err := one(q, rank(toMatches(nil, projects)))
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
	hits := filterAreas(areas, q)
	matches := make([]Match, len(hits))
	for i := range hits {
		matches[i] = Match{Kind: KindArea, Area: &hits[i]}
	}
	m, err := one(q, matches)
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
	hits := filterHeadings(headings, q)
	matches := make([]Match, len(hits))
	for i := range hits {
		matches[i] = Match{Kind: KindHeading, Heading: &hits[i]}
	}
	m, err := one(q, matches)
	if err != nil {
		return nil, err
	}
	return m.Heading, nil
}

func queryByUUID(ctx context.Context, c *things3.Client, q string) ([]Match, error) {
	todos, err := c.Todos().WithUUID(q).Status().Any().All(ctx)
	if err != nil {
		return nil, err
	}
	projects, err := c.Projects().WithUUID(q).Status().Any().All(ctx)
	if err != nil {
		return nil, err
	}
	return toMatches(todos, projects), nil
}

func queryByUUIDPrefix(ctx context.Context, c *things3.Client, q string) ([]Match, error) {
	todos, err := c.Todos().WithUUIDPrefix(q).Status().Any().All(ctx)
	if err != nil {
		return nil, err
	}
	projects, err := c.Projects().WithUUIDPrefix(q).Status().Any().All(ctx)
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

// rank stably orders matches: open before closed, then todos before projects,
// preserving the original query order for otherwise-equal items.
func rank(matches []Match) []Match {
	slices.SortStableFunc(matches, func(a, b Match) int {
		if a.Open() != b.Open() {
			if a.Open() {
				return -1
			}
			return 1
		}
		if a.Kind != b.Kind {
			if a.Kind == KindTodo {
				return -1
			}
			return 1
		}
		return 0
	})
	return matches
}

func one(query string, matches []Match) (Match, error) {
	switch len(matches) {
	case 0:
		return Match{}, &NotFoundError{Query: query}
	case 1:
		return matches[0], nil
	default:
		return Match{}, &AmbiguousError{Query: query, Matches: matches}
	}
}

func filterExactTitleProjects(projects []things3.Project, q string) []things3.Project {
	exact := make([]things3.Project, 0, len(projects))
	for i := range projects {
		if strings.EqualFold(projects[i].Title, q) {
			exact = append(exact, projects[i])
		}
	}
	return exact
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

func filterAreas(areas []things3.Area, q string) []things3.Area {
	return tieredFilter(areas, q,
		func(a things3.Area) string { return a.UUID },
		func(a things3.Area) string { return a.Title })
}

func filterHeadings(headings []things3.Heading, q string) []things3.Heading {
	return tieredFilter(headings, q,
		func(h things3.Heading) string { return h.UUID },
		func(h things3.Heading) string { return h.Title })
}
