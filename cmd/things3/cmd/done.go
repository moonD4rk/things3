package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

func newDoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "done <query>",
		Short:   "Complete a todo or project",
		GroupID: groupActions,
		Example: "  things3 done \"Buy milk\"\n  things3 done a1b2c3d4 --dry-run",
		Args:    cobra.ExactArgs(1),
		RunE:    withClient(runComplete("done", true, things3.StatusCompleted)),
	}
	addWriteFlags(cmd)
	return cmd
}

func newCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel <query>",
		Short:   "Cancel a todo or project",
		GroupID: groupActions,
		Example: "  things3 cancel \"Old task\"\n  things3 cancel a1b2c3d4 --dry-run",
		Args:    cobra.ExactArgs(1),
		RunE:    withClient(runComplete("cancel", false, things3.StatusCanceled)),
	}
	addWriteFlags(cmd)
	return cmd
}

// runComplete builds the shared body for done and cancel.
func runComplete(action string, complete bool, want things3.Status) func(*cobra.Command, []string, *things3.Client) error {
	return func(cmd *cobra.Command, args []string, client *things3.Client) error {
		match, err := resolve.ResolveOne(cmd.Context(), client, args[0])
		if err != nil {
			return fromResolveError(err)
		}

		var builder urlBuilder
		if match.Kind == resolve.KindProject {
			u := client.UpdateProject(match.UUID())
			if complete {
				u = u.Completed(true)
			} else {
				u = u.Canceled(true)
			}
			builder = u
		} else {
			u := client.UpdateTodo(match.UUID())
			if complete {
				u = u.Completed(true)
			} else {
				u = u.Canceled(true)
			}
			builder = u
		}
		return runWrite(cmd, action, builder, statusVerifier(action, match, want, client))
	}
}
