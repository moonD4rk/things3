package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
	"github.com/moond4rk/things3/cmd/things3/internal/verify"
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <title>",
		Short:   "Add a todo",
		GroupID: groupActions,
		Example: `  things3 add "Buy milk" --when today
  things3 add "Email Bob" --project Work --tags urgent
  things3 add "Plan trip" --checklist Flights --checklist Hotel`,
		Args: cobra.ExactArgs(1),
		RunE: withClient(runAddTodo),
	}
	cmd.Flags().String(flagNotes, "", "notes")
	cmd.Flags().String(flagWhen, "", "when: today|tomorrow|evening|anytime|someday|YYYY-MM-DD")
	cmd.Flags().String(flagDeadline, "", "deadline (YYYY-MM-DD)")
	cmd.Flags().String(flagReminder, "", "reminder time (HH:MM)")
	cmd.Flags().String(flagProject, "", "place in a project (name, prefix, or UUID)")
	cmd.Flags().String(flagArea, "", "place in an area (name, prefix, or UUID)")
	cmd.Flags().String(flagHeading, "", "place under a heading (requires --project)")
	cmd.Flags().String(flagTags, "", "tags (comma-separated)")
	cmd.Flags().StringArray(flagChecklist, nil, "checklist item (repeatable)")
	cmd.MarkFlagsMutuallyExclusive(flagProject, flagArea)
	cmd.MarkFlagsMutuallyExclusive(flagArea, flagHeading)
	addWriteFlags(cmd)
	cmd.AddCommand(newAddProjectCmd())
	return cmd
}

func runAddTodo(cmd *cobra.Command, args []string, client *things3.Client) error {
	ctx := cmd.Context()
	f := cmd.Flags()
	title := args[0]
	builder := client.AddTodo().Title(title)

	if v, _ := f.GetString(flagNotes); v != "" {
		builder = builder.Notes(v)
	}
	if f.Changed(flagWhen) {
		v, _ := f.GetString(flagWhen)
		b, err := things3.ParseWhen(builder, v)
		if err != nil {
			return err
		}
		builder = b
	}
	if f.Changed(flagDeadline) {
		v, _ := f.GetString(flagDeadline)
		d, err := time.Parse(time.DateOnly, v)
		if err != nil {
			return fmt.Errorf("invalid --deadline %q: use YYYY-MM-DD", v)
		}
		builder = builder.Deadline(d)
	}
	if f.Changed(flagReminder) {
		v, _ := f.GetString(flagReminder)
		hour, minute, err := parseHHMM(v)
		if err != nil {
			return err
		}
		builder = builder.Reminder(hour, minute)
	}
	if v, _ := f.GetString(flagTags); v != "" {
		builder = builder.Tags(splitTags(v)...)
	}
	if items, _ := f.GetStringArray(flagChecklist); len(items) > 0 {
		builder = builder.ChecklistItems(items...)
	}

	builder, err := placeTodo(ctx, client, builder, f)
	if err != nil {
		return err
	}

	t0 := time.Now().Add(-2 * time.Second)
	return runWrite(cmd, actionAdd, builder, func(ctx context.Context) writeResult {
		r := writeResult{Action: actionAdd, Type: typeTodo}
		todo, outcome, addErr := verify.AddedTodo(ctx, client, title, t0, verify.Options{})
		applyTodoOutcome(&r, todo, outcome, addErr)
		return r
	})
}

// placeTodo resolves the --project/--area/--heading destination flags onto the
// adder. --heading requires --project.
func placeTodo(ctx context.Context, client *things3.Client, builder things3.TodoAdder, f *pflag.FlagSet) (things3.TodoAdder, error) {
	heading, _ := f.GetString(flagHeading)
	project, _ := f.GetString(flagProject)
	area, _ := f.GetString(flagArea)

	switch {
	case heading != "":
		if project == "" {
			return builder, errors.New("--heading requires --project")
		}
		p, err := resolve.Project(ctx, client, project)
		if err != nil {
			return builder, fromResolveError(err)
		}
		h, err := resolve.Heading(ctx, client, p.UUID, heading)
		if err != nil {
			return builder, fromResolveError(err)
		}
		return builder.ListID(p.UUID).HeadingID(h.UUID), nil
	case project != "":
		p, err := resolve.Project(ctx, client, project)
		if err != nil {
			return builder, fromResolveError(err)
		}
		return builder.ListID(p.UUID), nil
	case area != "":
		a, err := resolve.Area(ctx, client, area)
		if err != nil {
			return builder, fromResolveError(err)
		}
		return builder.ListID(a.UUID), nil
	default:
		return builder, nil
	}
}

func newAddProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project <title>",
		Short:   "Add a project",
		Example: "  things3 add project \"Website\" --area Work\n  things3 add project \"Trip\" --todos Flights --todos Hotel",
		Args:    cobra.ExactArgs(1),
		RunE:    withClient(runAddProject),
	}
	cmd.Flags().String(flagNotes, "", "notes")
	cmd.Flags().String(flagWhen, "", "when: today|tomorrow|evening|anytime|someday|YYYY-MM-DD")
	cmd.Flags().String(flagDeadline, "", "deadline (YYYY-MM-DD)")
	cmd.Flags().String(flagArea, "", "place in an area (name, prefix, or UUID)")
	cmd.Flags().StringArray(flagTodos, nil, "initial todo (repeatable)")
	addWriteFlags(cmd)
	return cmd
}

func runAddProject(cmd *cobra.Command, args []string, client *things3.Client) error {
	ctx := cmd.Context()
	f := cmd.Flags()
	title := args[0]
	builder := client.AddProject().Title(title)

	if v, _ := f.GetString(flagNotes); v != "" {
		builder = builder.Notes(v)
	}
	if f.Changed(flagWhen) {
		v, _ := f.GetString(flagWhen)
		b, err := things3.ParseWhen(builder, v)
		if err != nil {
			return err
		}
		builder = b
	}
	if f.Changed(flagDeadline) {
		v, _ := f.GetString(flagDeadline)
		d, err := time.Parse(time.DateOnly, v)
		if err != nil {
			return fmt.Errorf("invalid --deadline %q: use YYYY-MM-DD", v)
		}
		builder = builder.Deadline(d)
	}
	if v, _ := f.GetString(flagArea); v != "" {
		a, err := resolve.Area(ctx, client, v)
		if err != nil {
			return fromResolveError(err)
		}
		builder = builder.AreaID(a.UUID)
	}
	if todos, _ := f.GetStringArray(flagTodos); len(todos) > 0 {
		builder = builder.Todos(todos...)
	}

	t0 := time.Now().Add(-2 * time.Second)
	return runWrite(cmd, actionAdd, builder, func(ctx context.Context) writeResult {
		r := writeResult{Action: actionAdd, Type: typeProject}
		project, outcome, addErr := verify.AddedProject(ctx, client, title, t0, verify.Options{})
		applyProjectOutcome(&r, project, outcome, addErr)
		return r
	})
}
