package things3

// clientOptions holds the configuration options for the Client.
type clientOptions struct {
	databasePath string
	printSQL     bool
}

// Option is a functional option for configuring the Client.
type Option func(*clientOptions)

// WithDatabasePath sets a custom path to the Things database.
// If not set, the database path is discovered automatically.
func WithDatabasePath(path string) Option {
	return func(o *clientOptions) {
		o.databasePath = path
	}
}

// WithPrintSQL enables SQL query logging to stdout.
// Useful for debugging and understanding the queries being executed.
func WithPrintSQL(enabled bool) Option {
	return func(o *clientOptions) {
		o.printSQL = enabled
	}
}
