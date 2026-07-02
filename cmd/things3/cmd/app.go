package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// Flag names shared across commands.
const (
	flagText  = "text"
	flagJSON  = "json"
	flagYAML  = "yaml"
	flagLimit = "limit"
	flagDB    = "db"
)

// withClient wraps a command body with database client lifecycle management:
// it opens a client (honoring --db over THINGSDB over auto-discovery), passes
// it to run, and closes it afterward.
func withClient(run func(cmd *cobra.Command, args []string, client *things3.Client) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var opts []things3.ClientOption
		if dbPath, _ := cmd.Flags().GetString(flagDB); dbPath != "" {
			opts = append(opts, things3.WithDatabasePath(dbPath))
		}
		client, err := things3.NewClient(opts...)
		if err != nil {
			return err
		}
		defer func() { _ = client.Close() }()
		return run(cmd, args, client)
	}
}
