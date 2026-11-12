package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
	"github.com/moond4rk/things3/cmd/things3/internal/verify"
)

// Write-command flag names.
const (
	flagDryRun           = "dry-run"
	flagNoVerify         = "no-verify"
	flagTo               = "to"
	flagTitle            = "title"
	flagNotes            = "notes"
	flagClearNotes       = "clear-notes"
	flagPrependNotes     = "prepend-notes"
	flagAppendNotes      = "append-notes"
	flagDeadline         = "deadline"
	flagClearDeadline    = "clear-deadline"
	flagTags             = "tags"
	flagClearTags        = "clear-tags"
	flagAddTags          = "add-tags"
	flagWhen             = "when"
	flagReminder         = "reminder"
	flagProject          = "project"
	flagHeading          = "heading"
	flagChecklist        = "checklist"
	flagClearChecklist   = "clear-checklist"
	flagPrependChecklist = "prepend-checklist"
	flagAppendChecklist  = "append-checklist"
	flagTodos            = "todos"
	flagStatus           = "status"
)

// addWriteFlags adds the flags shared by every action command.
func addWriteFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(flagDryRun, false, "print the things:/// URL without executing it")
	cmd.Flags().Bool(flagNoVerify, false, "skip database verification after executing")
}

// urlBuilder is the common Build/Execute surface of every scheme builder.
type urlBuilder interface {
	Build() (string, error)
	Execute(ctx context.Context) error
}

// wrapExecError appends a macOS hint to a URL-scheme execution failure.
func wrapExecError(err error) error {
	return fmt.Errorf("%w (write commands require macOS with Things installed)", err)
}

// runWrite performs the shared write flow: dry-run prints the URL; otherwise it
// executes and, unless --no-verify, runs verifyFn and reports the result. An
// unverified send is still success (exit 0).
func runWrite(cmd *cobra.Command, action string, builder urlBuilder, verifyFn func(ctx context.Context) writeResult) error {
	format := resolveFormat(cmd)

	if dryRun, _ := cmd.Flags().GetBool(flagDryRun); dryRun {
		url, err := builder.Build()
		if err != nil {
			return err
		}
		return writeWriteResult(cmd.OutOrStdout(), &writeResult{Action: action, DryRun: true, URL: url}, format)
	}

	if err := builder.Execute(cmd.Context()); err != nil {
		return wrapExecError(err)
	}

	if noVerify, _ := cmd.Flags().GetBool(flagNoVerify); noVerify {
		return emitResult(cmd, &writeResult{Action: action, Verified: false, Message: "verification skipped"})
	}
	if verifyFn == nil {
		return emitResult(cmd, &writeResult{Action: action, Verified: true})
	}
	result := verifyFn(cmd.Context())
	return emitResult(cmd, &result)
}

// emitResult writes the result to stdout and, in text mode, the unverified
// caveat to stderr.
func emitResult(cmd *cobra.Command, r *writeResult) error {
	format := resolveFormat(cmd)
	// Surface a stderr caveat only when it adds detail beyond the stdout line
	// (ambiguity or an underlying error), not for the plain not-confirmed case.
	if format == formatText && !r.Verified && r.Message != "" && r.Message != genericUnverified {
		fmt.Fprintln(cmd.ErrOrStderr(), r.Message)
	}
	return writeWriteResult(cmd.OutOrStdout(), r, format)
}

// statusVerifier verifies a done/cancel status flip for the resolved item.
//
//nolint:dupl // intentional mirror of modifiedVerifier
func statusVerifier(action string, m resolve.Match, want things3.Status, client *things3.Client) func(context.Context) writeResult {
	return itemVerifier(action, m,
		func(ctx context.Context) (*things3.Todo, verify.Outcome, error) {
			return verify.TodoStatus(ctx, client, m.UUID(), want, verify.Options{})
		},
		func(ctx context.Context) (*things3.Project, verify.Outcome, error) {
			return verify.ProjectStatus(ctx, client, m.UUID(), want, verify.Options{})
		})
}

// modifiedVerifier verifies a schedule/move/edit against the resolved item,
// which is the state before the write.
//
//nolint:dupl // intentional mirror of statusVerifier
func modifiedVerifier(action string, m resolve.Match, client *things3.Client) func(context.Context) writeResult {
	return itemVerifier(action, m,
		func(ctx context.Context) (*things3.Todo, verify.Outcome, error) {
			return verify.TodoModified(ctx, client, m.Todo, verify.Options{})
		},
		func(ctx context.Context) (*things3.Project, verify.Outcome, error) {
			return verify.ProjectModified(ctx, client, m.Project, verify.Options{})
		})
}

// itemVerifier runs the matching todo/project verifier and builds the result.
func itemVerifier(action string, m resolve.Match,
	verifyTodo func(context.Context) (*things3.Todo, verify.Outcome, error),
	verifyProject func(context.Context) (*things3.Project, verify.Outcome, error),
) func(context.Context) writeResult {
	return func(ctx context.Context) writeResult {
		r := writeResult{Action: action, Type: string(m.Kind), UUID: m.UUID()}
		if m.Kind == resolve.KindProject {
			project, outcome, err := verifyProject(ctx)
			applyProjectOutcome(&r, project, outcome, err)
		} else {
			todo, outcome, err := verifyTodo(ctx)
			applyTodoOutcome(&r, todo, outcome, err)
		}
		return r
	}
}

func applyTodoOutcome(r *writeResult, todo *things3.Todo, outcome verify.Outcome, err error) {
	if outcome == verify.Confirmed && todo != nil {
		r.Verified = true
		r.Todo = todo
		r.UUID = todo.UUID
		return
	}
	r.Verified = false
	r.Message = unverifiedMessage(outcome, err)
}

func applyProjectOutcome(r *writeResult, project *things3.Project, outcome verify.Outcome, err error) {
	if outcome == verify.Confirmed && project != nil {
		r.Verified = true
		r.Project = project
		r.UUID = project.UUID
		return
	}
	r.Verified = false
	r.Message = unverifiedMessage(outcome, err)
}

// genericUnverified is the caveat for a send that simply was not confirmed in
// time, carrying no extra detail beyond the stdout line.
const genericUnverified = "sent to Things, not yet confirmed"

func unverifiedMessage(outcome verify.Outcome, err error) string {
	if outcome == verify.AmbiguousMatch {
		return "sent to Things; several same-title items exist, cannot identify the created one"
	}
	if err != nil {
		return genericUnverified + ": " + err.Error()
	}
	return genericUnverified
}

func isNotFound(err error) bool {
	_, ok := errors.AsType[*resolve.NotFoundError](err)
	return ok
}

func isWhenKeyword(s string) bool {
	switch strings.ToLower(s) {
	case nameToday, nameTomorrow, nameEvening, nameAnytime, nameSomeday:
		return true
	default:
		return false
	}
}

// splitTags splits a comma-separated tag string, trimming and dropping empties.
func splitTags(s string) []string {
	parts := strings.Split(s, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}
