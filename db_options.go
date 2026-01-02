package things3

// dbOptions holds the configuration options for the DB.
type dbOptions struct {
	databasePath string
	printSQL     bool
}

// DBOption is a functional option for configuring the DB.
type DBOption func(*dbOptions)

// WithDBPath sets a custom path to the Things database.
// If not set, the database path is discovered automatically.
func WithDBPath(path string) DBOption {
	return func(opts *dbOptions) {
		opts.databasePath = path
	}
}

// withDBPrintSQL enables SQL query logging to stdout (internal use by Client).
func withDBPrintSQL(enabled bool) DBOption {
	return func(opts *dbOptions) {
		opts.printSQL = enabled
	}
}
