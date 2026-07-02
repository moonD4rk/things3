package cmd

import (
	"slices"
	"time"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewListCmd creates the list command with all view subcommands.
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <view>",
		Short: "List todos from various views",
	}

	cmd.AddCommand(
		newInboxCmd(),
		newTodayCmd(),
		newUpcomingCmd(),
		newAnytimeCmd(),
		newSomedayCmd(),
		newLogbookCmd(),
		newDeadlinesCmd(),
		newProjectsCmd(),
		newAreasCmd(),
		newTagsCmd(),
	)

	return cmd
}

func newInboxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inbox",
		Short: "List todos in the Inbox",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			todos, err := client.Todos().
				Start().Inbox().
				Status().Incomplete().
				All(cmd.Context())
			if err != nil {
				return err
			}
			return outputTodos(cmd, todos)
		},
	}
}

func newTodayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "today",
		Short: "List todos scheduled for today",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()

			// Regular Today tasks: have a start date and in Anytime bucket
			regularTodos, err := client.Todos().
				StartDate().Exists(true).
				Start().Anytime().
				Status().Incomplete().
				OrderByTodayIndex().
				All(ctx)
			if err != nil {
				return err
			}

			// Scheduled tasks from Someday with past start dates (yellow dot)
			scheduledTodos, err := client.Todos().
				StartDate().Past().
				Start().Someday().
				Status().Incomplete().
				OrderByTodayIndex().
				All(ctx)
			if err != nil {
				return err
			}

			// Overdue deadline tasks: no start date, deadline past, not suppressed
			overdueTodos, err := client.Todos().
				StartDate().Exists(false).
				Deadline().Past().
				DeadlineSuppressed(false).
				Status().Incomplete().
				All(ctx)
			if err != nil {
				return err
			}

			todos := make([]things3.Todo, 0, len(regularTodos)+len(scheduledTodos)+len(overdueTodos))
			todos = append(todos, regularTodos...)
			todos = append(todos, scheduledTodos...)
			todos = append(todos, overdueTodos...)

			return outputTodos(cmd, todos)
		},
	}
}

func newUpcomingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upcoming",
		Short: "List todos scheduled for future dates",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			todos, err := client.Todos().
				StartDate().Future().
				Start().Someday().
				Status().Incomplete().
				All(cmd.Context())
			if err != nil {
				return err
			}
			return outputTodos(cmd, todos)
		},
	}
}

func newAnytimeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "anytime",
		Short: "List todos in the Anytime list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			todos, err := client.Todos().
				Start().Anytime().
				Status().Incomplete().
				All(cmd.Context())
			if err != nil {
				return err
			}
			return outputTodos(cmd, todos)
		},
	}
}

func newSomedayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "someday",
		Short: "List todos in the Someday list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			todos, err := client.Todos().
				StartDate().Exists(false).
				Start().Someday().
				Status().Incomplete().
				All(cmd.Context())
			if err != nil {
				return err
			}
			return outputTodos(cmd, todos)
		},
	}
}

func newLogbookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logbook",
		Short: "List completed and canceled todos",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			days, _ := cmd.Flags().GetInt("days")

			q := client.Todos().
				Status().Any().
				StopDate().Exists(true)

			if days > 0 {
				since := time.Now().AddDate(0, 0, -days)
				q = q.StopDate().After(since)
			}

			todos, err := q.All(cmd.Context())
			if err != nil {
				return err
			}
			sortByStopTimeDesc(todos)
			return outputTodos(cmd, todos)
		},
	}

	cmd.Flags().IntP("days", "d", 30, "limit to recent N days (0 for all)")

	return cmd
}

// stopTime returns when a todo left the active list: completion time if set,
// otherwise cancellation time, otherwise nil.
func stopTime(t *things3.Todo) *time.Time {
	if t.CompletedAt != nil {
		return t.CompletedAt
	}
	return t.CanceledAt
}

// sortByStopTimeDesc orders todos by stop time descending so that --limit N
// yields the N most recent entries. Todos without a stop time sort last.
func sortByStopTimeDesc(todos []things3.Todo) {
	slices.SortStableFunc(todos, func(a, b things3.Todo) int {
		ta, tb := stopTime(&a), stopTime(&b)
		switch {
		case ta == nil && tb == nil:
			return 0
		case ta == nil:
			return 1
		case tb == nil:
			return -1
		default:
			return tb.Compare(*ta)
		}
	})
}

func newDeadlinesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deadlines",
		Short: "List todos with deadlines",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			todos, err := client.Todos().
				Deadline().Exists(true).
				Status().Incomplete().
				All(cmd.Context())
			if err != nil {
				return err
			}
			return outputTodos(cmd, todos)
		},
	}
}

func newProjectsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "projects",
		Short: "List all incomplete projects",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			projects, err := client.Projects().
				Status().Incomplete().
				All(cmd.Context())
			if err != nil {
				return err
			}
			return outputProjects(cmd, projects)
		},
	}
}

func newAreasCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "areas",
		Short: "List all areas",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			areas, err := client.Areas().All(cmd.Context())
			if err != nil {
				return err
			}
			return outputAreas(cmd, areas)
		},
	}
}

func newTagsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tags",
		Short: "List all tags",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tags, err := client.Tags().All(cmd.Context())
			if err != nil {
				return err
			}
			return outputTags(cmd, tags)
		},
	}
}
