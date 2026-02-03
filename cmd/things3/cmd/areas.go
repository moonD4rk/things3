package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewAreasCmd creates the areas command.
func NewAreasCmd() *cobra.Command {
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
