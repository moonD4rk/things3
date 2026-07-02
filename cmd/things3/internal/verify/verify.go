// Package verify confirms that a fire-and-forget URL-scheme write actually
// landed in the Things database. It is cobra-free so it can be unit-tested
// against the fixture with an injectable clock.
package verify

import (
	"context"
	"strings"
	"time"

	"github.com/moond4rk/things3"
)

// Default polling bounds.
const (
	DefaultBudget   = 2 * time.Second
	DefaultInterval = 150 * time.Millisecond
)

// Options tunes a verification poll. The zero value uses the defaults.
type Options struct {
	Budget   time.Duration
	Interval time.Duration
	// Sleep waits d or returns early if ctx is done. Nil uses a real timer;
	// tests inject a deterministic implementation.
	Sleep func(ctx context.Context, d time.Duration) error
}

func (o Options) budget() time.Duration {
	if o.Budget > 0 {
		return o.Budget
	}
	return DefaultBudget
}

func (o Options) interval() time.Duration {
	if o.Interval > 0 {
		return o.Interval
	}
	return DefaultInterval
}

func (o Options) sleep(ctx context.Context, d time.Duration) error {
	if o.Sleep != nil {
		return o.Sleep(ctx, d)
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// Outcome describes how a verification ended.
type Outcome int

const (
	// Confirmed means the write was observed in the database.
	Confirmed Outcome = iota
	// Unverified means the budget elapsed without confirmation.
	Unverified
	// AmbiguousMatch means add-verification found several same-title candidates.
	AmbiguousMatch
)

// Poll runs check immediately, then every Interval until check returns true,
// the Budget elapses, or ctx is canceled. check errors do not abort the loop;
// the last error is returned alongside confirmed=false.
func Poll(ctx context.Context, opts Options, check func(ctx context.Context) (bool, error)) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, opts.budget())
	defer cancel()

	var lastErr error
	for {
		ok, err := check(ctx)
		switch {
		case err != nil:
			lastErr = err
		case ok:
			return true, nil
		}
		if err := opts.sleep(ctx, opts.interval()); err != nil {
			return false, lastErr
		}
	}
}

// AddedTodo recovers a freshly created todo by title. t0 is a pre-write guard
// band (callers pass time.Now().Add(-2s) before Execute). Exactly one same-title
// hit is Confirmed (re-fetched with checklist); several is AmbiguousMatch (never
// guessed); none by budget end is Unverified.
func AddedTodo(ctx context.Context, c *things3.Client, title string, t0 time.Time, opts Options) (*things3.Todo, Outcome, error) {
	var candidate *things3.Todo
	outcome := Unverified
	confirmed, lastErr := Poll(ctx, opts, func(ctx context.Context) (bool, error) {
		todos, err := c.Todos().WithTitle(title).CreatedAfter(t0).Status().Any().All(ctx)
		if err != nil {
			return false, err
		}
		matches := exactTitle(todos, title)
		switch len(matches) {
		case 0:
			return false, nil
		case 1:
			candidate = &matches[0]
			return true, nil
		default:
			outcome = AmbiguousMatch
			return true, nil
		}
	})
	if outcome == AmbiguousMatch {
		return nil, AmbiguousMatch, lastErr
	}
	if !confirmed || candidate == nil {
		return nil, Unverified, lastErr
	}
	full, err := c.Todos().WithUUID(candidate.UUID).Status().Any().IncludeChecklist().First(ctx)
	if err != nil {
		return candidate, Confirmed, err
	}
	return full, Confirmed, nil
}

// AddedProject recovers a freshly created project by title, mirroring AddedTodo.
func AddedProject(ctx context.Context, c *things3.Client, title string, t0 time.Time, opts Options) (*things3.Project, Outcome, error) {
	var candidate *things3.Project
	outcome := Unverified
	confirmed, lastErr := Poll(ctx, opts, func(ctx context.Context) (bool, error) {
		projects, err := c.Projects().WithTitle(title).CreatedAfter(t0).Status().Any().All(ctx)
		if err != nil {
			return false, err
		}
		matches := make([]things3.Project, 0, len(projects))
		for i := range projects {
			if strings.EqualFold(projects[i].Title, title) {
				matches = append(matches, projects[i])
			}
		}
		switch len(matches) {
		case 0:
			return false, nil
		case 1:
			candidate = &matches[0]
			return true, nil
		default:
			outcome = AmbiguousMatch
			return true, nil
		}
	})
	if outcome == AmbiguousMatch {
		return nil, AmbiguousMatch, lastErr
	}
	if !confirmed || candidate == nil {
		return nil, Unverified, lastErr
	}
	return candidate, Confirmed, nil
}

// TodoStatus confirms a todo reached the wanted status (done/cancel).
func TodoStatus(ctx context.Context, c *things3.Client, uuid string, want things3.Status, opts Options) (*things3.Todo, Outcome, error) {
	var found *things3.Todo
	confirmed, lastErr := Poll(ctx, opts, func(ctx context.Context) (bool, error) {
		todo, err := c.Todos().WithUUID(uuid).Status().Any().First(ctx)
		if err != nil {
			return false, err
		}
		if todo.Status == want {
			found = todo
			return true, nil
		}
		return false, nil
	})
	return resolveOutcome(found, confirmed, lastErr)
}

// ProjectStatus confirms a project reached the wanted status.
func ProjectStatus(
	ctx context.Context, c *things3.Client, uuid string, want things3.Status, opts Options,
) (*things3.Project, Outcome, error) {
	var found *things3.Project
	confirmed, lastErr := Poll(ctx, opts, func(ctx context.Context) (bool, error) {
		project, err := c.Projects().WithUUID(uuid).Status().Any().First(ctx)
		if err != nil {
			return false, err
		}
		if project.Status == want {
			found = project
			return true, nil
		}
		return false, nil
	})
	if confirmed && found != nil {
		return found, Confirmed, nil
	}
	return nil, Unverified, lastErr
}

// TodoModified confirms a todo's ModifiedAt advanced past baseline
// (schedule/move/edit). baseline is the resolved item's own ModifiedAt.
func TodoModified(ctx context.Context, c *things3.Client, uuid string, baseline time.Time, opts Options) (*things3.Todo, Outcome, error) {
	var found *things3.Todo
	confirmed, lastErr := Poll(ctx, opts, func(ctx context.Context) (bool, error) {
		todo, err := c.Todos().WithUUID(uuid).Status().Any().First(ctx)
		if err != nil {
			return false, err
		}
		if todo.ModifiedAt.After(baseline) {
			found = todo
			return true, nil
		}
		return false, nil
	})
	return resolveOutcome(found, confirmed, lastErr)
}

// ProjectModified confirms a project's ModifiedAt advanced past baseline.
func ProjectModified(
	ctx context.Context, c *things3.Client, uuid string, baseline time.Time, opts Options,
) (*things3.Project, Outcome, error) {
	var found *things3.Project
	confirmed, lastErr := Poll(ctx, opts, func(ctx context.Context) (bool, error) {
		project, err := c.Projects().WithUUID(uuid).Status().Any().First(ctx)
		if err != nil {
			return false, err
		}
		if project.ModifiedAt.After(baseline) {
			found = project
			return true, nil
		}
		return false, nil
	})
	if confirmed && found != nil {
		return found, Confirmed, nil
	}
	return nil, Unverified, lastErr
}

func resolveOutcome(found *things3.Todo, confirmed bool, lastErr error) (*things3.Todo, Outcome, error) {
	if confirmed && found != nil {
		return found, Confirmed, nil
	}
	return nil, Unverified, lastErr
}

func exactTitle(todos []things3.Todo, title string) []things3.Todo {
	matches := make([]things3.Todo, 0, len(todos))
	for i := range todos {
		if strings.EqualFold(todos[i].Title, title) {
			matches = append(matches, todos[i])
		}
	}
	return matches
}
