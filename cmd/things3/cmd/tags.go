package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewTagsCmd creates the tags command.
func NewTagsCmd() *cobra.Command {
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
