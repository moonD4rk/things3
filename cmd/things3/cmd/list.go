package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewListCmd creates the list command with all view subcommands.
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <view>",
		Short: "List tasks from various views",
	}

	// Register view subcommands
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
		Short: "List tasks in the Inbox",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Inbox(cmd.Context())
			if err != nil {
				return err
			}
			return outputTasks(cmd, tasks)
		},
	}
}

func newTodayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "today",
		Short: "List tasks scheduled for today",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Today(cmd.Context())
			if err != nil {
				return err
			}
			return outputTasks(cmd, tasks)
		},
	}
}

func newUpcomingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upcoming",
		Short: "List tasks scheduled for future dates",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Upcoming(cmd.Context())
			if err != nil {
				return err
			}
			return outputTasks(cmd, tasks)
		},
	}
}

func newAnytimeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "anytime",
		Short: "List tasks in the Anytime list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Anytime(cmd.Context())
			if err != nil {
				return err
			}
			return outputTasks(cmd, tasks)
		},
	}
}

func newSomedayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "someday",
		Short: "List tasks in the Someday list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Someday(cmd.Context())
			if err != nil {
				return err
			}
			return outputTasks(cmd, tasks)
		},
	}
}

func newLogbookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logbook",
		Short: "List completed and canceled tasks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			days, _ := cmd.Flags().GetInt("days")

			var tasks []things3.Task
			if days == 0 {
				// No limit, get all
				tasks, err = client.Logbook(cmd.Context())
			} else {
				// Filter by stop date within N days
				since := time.Now().AddDate(0, 0, -days)
				tasks, err = client.Tasks().
					StopDate().After(since).
					Status().Any().
					ContextTrashed(false).
					All(cmd.Context())
			}
			if err != nil {
				return err
			}

			return outputTasks(cmd, tasks)
		},
	}

	cmd.Flags().IntP("days", "d", 30, "limit to recent N days (0 for all)")

	return cmd
}

func newDeadlinesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deadlines",
		Short: "List tasks with deadlines",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Deadlines(cmd.Context())
			if err != nil {
				return err
			}
			return outputTasks(cmd, tasks)
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

			tasks, err := client.Projects(cmd.Context())
			if err != nil {
				return err
			}
			return outputTasks(cmd, tasks)
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
