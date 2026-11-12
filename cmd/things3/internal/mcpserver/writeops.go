package mcpserver

import (
	"context"
	"errors"
	"time"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/change"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
	"github.com/moond4rk/things3/cmd/things3/internal/verify"
)

// guardBand is the pre-write window a create verification looks back over, so a
// freshly created item is recovered by title without matching stale rows.
const guardBand = 2 * time.Second

// registerWrite registers the six write tools; it runs only for a non-read-only
// server, so a read-only client cannot even list them.
func (s *Server) registerWrite(r *registrar) {
	regTool(r, "add_todo", descAddTodo, s.handleAddTodo)
	regTool(r, "add_project", descAddProject, s.handleAddProject)
	regTool(r, "complete", descComplete, s.handleComplete)
	regTool(r, "schedule", descSchedule, s.handleSchedule)
	regTool(r, "move", descMove, s.handleMove)
	regTool(r, "edit", descEdit, s.handleEdit)
	regTool(r, "open", descOpen, s.handleOpen)
}

// runWrite executes a scheme builder under the write mutex, then verifies. An
// execution failure rides the envelope as execution_failed; verification then
// reports verified true or false. Only a nil verifyFn (navigation) skips it.
func (s *Server) runWrite(ctx context.Context, builder URLBuilder, verifyFn func(context.Context) WriteResult) WriteResult {
	defer s.lockWrites()()
	if err := s.cfg.Execute(ctx, builder); err != nil {
		return WriteResult{Success: false, Error: writeExecError(err)}
	}
	if verifyFn == nil {
		return WriteResult{Success: true, Verified: true}
	}
	return verifyFn(ctx)
}

// writeExecError classifies a failed write. A URL-scheme input-validation error is
// the caller's to fix (invalid_input); anything else - a missing auth token, a
// database read, or an osascript failure - is environmental (execution_failed).
// The scheme resolves the auth token inside URL building, so a token failure must
// stay execution_failed rather than masquerade as bad input.
func writeExecError(err error) *ToolError {
	if isInputError(err) {
		return invalidInput(err.Error())
	}
	return executionFailed(err)
}

// inputErrors are the URL-scheme validation sentinels that mean the request itself
// is malformed, as opposed to a token or transport failure.
var inputErrors = []error{
	things3.ErrTitleTooLong,
	things3.ErrNotesTooLong,
	things3.ErrTooManyChecklistItems,
	things3.ErrTitleContainsNewline,
	things3.ErrTagContainsComma,
	things3.ErrChecklistItemContainsNewline,
	things3.ErrInvalidReminderTime,
	things3.ErrReminderNeedsDate,
}

// isInputError reports whether err is a URL-scheme input-validation failure.
func isInputError(err error) bool {
	for _, target := range inputErrors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// verifyAddedTodo confirms a freshly created todo by title.
func (s *Server) verifyAddedTodo(ctx context.Context, title string, t0 time.Time) WriteResult {
	todo, outcome, err := verify.AddedTodo(ctx, s.client, title, t0, s.cfg.Verify)
	return todoOutcome(todo, outcome, err)
}

// verifyAddedProject confirms a freshly created project by title.
func (s *Server) verifyAddedProject(ctx context.Context, title string, t0 time.Time) WriteResult {
	project, outcome, err := verify.AddedProject(ctx, s.client, title, t0, s.cfg.Verify)
	return projectOutcome(project, outcome, err)
}

// verifyStatus confirms a status flip for the resolved todo or project.
func (s *Server) verifyStatus(ctx context.Context, uuid, kind string, want things3.Status) WriteResult {
	if kind == typeProject {
		project, outcome, err := verify.ProjectStatus(ctx, s.client, uuid, want, s.cfg.Verify)
		return projectOutcome(project, outcome, err)
	}
	todo, outcome, err := verify.TodoStatus(ctx, s.client, uuid, want, s.cfg.Verify)
	return todoOutcome(todo, outcome, err)
}

// verifyModified confirms a schedule/move/edit landed, against the resolved
// item as the state before the write.
func (s *Server) verifyModified(ctx context.Context, m resolve.Match) WriteResult {
	if m.Kind == resolve.KindProject {
		project, outcome, err := verify.ProjectModified(ctx, s.client, m.Project, s.cfg.Verify)
		return projectOutcome(project, outcome, err)
	}
	todo, outcome, err := verify.TodoModified(ctx, s.client, m.Todo, s.cfg.Verify)
	return todoOutcome(todo, outcome, err)
}

// todoOutcome maps a todo verification into a WriteResult. An unconfirmed send is
// still success, matching the CLI's exit-0 semantics.
func todoOutcome(todo *things3.Todo, outcome verify.Outcome, err error) WriteResult {
	if outcome == verify.Confirmed && todo != nil {
		item := todoItem(todo)
		return WriteResult{Success: true, Verified: true, Item: &item}
	}
	return WriteResult{Success: true, Verified: false, Message: unverifiedMessage(outcome, err)}
}

// projectOutcome mirrors todoOutcome for a project.
func projectOutcome(project *things3.Project, outcome verify.Outcome, err error) WriteResult {
	if outcome == verify.Confirmed && project != nil {
		item := projectItem(project)
		return WriteResult{Success: true, Verified: true, Item: &item}
	}
	return WriteResult{Success: true, Verified: false, Message: unverifiedMessage(outcome, err)}
}

// unverifiedMessage explains why a sent write was not confirmed in time.
func unverifiedMessage(outcome verify.Outcome, err error) string {
	switch {
	case outcome == verify.AmbiguousMatch:
		return "sent to Things; several same-title items exist, cannot identify the created one"
	case err != nil:
		return "sent to Things, not yet confirmed: " + err.Error()
	default:
		return "sent to Things, not yet confirmed"
	}
}

// writeError wraps a structured error as a failed write envelope.
func writeError(te *ToolError) WriteResult {
	return WriteResult{Success: false, Error: te}
}

// changeError maps an error from the shared change rules into a ToolError. A
// request the rules reject is the caller's to fix.
func changeError(err error) *ToolError {
	if _, ok := errors.AsType[*change.InputError](err); ok {
		return invalidInput(err.Error())
	}
	return writeExecError(err)
}

// parseDate parses a YYYY-MM-DD deadline, returning a structured invalid_input error.
func parseDate(value string) (time.Time, *ToolError) {
	d, err := change.ParseDate("deadline", value)
	if err != nil {
		return time.Time{}, invalidInput(err.Error())
	}
	return d, nil
}

// parseReminder parses an HH:MM reminder time into hour and minute.
func parseReminder(value string) (hour, minute int, te *ToolError) {
	clock, err := change.ParseClock(value)
	if err != nil {
		return 0, 0, invalidInput(err.Error())
	}
	return clock.Hour, clock.Minute, nil
}

// now returns the current time; the guard-band anchor for create verification.
func now() time.Time { return time.Now() }
