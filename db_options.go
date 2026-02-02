package things3

// dbOptions holds the configuration options for the db.
type dbOptions struct {
	databasePath string
	printSQL     bool
}

// dbOption is a functional option for configuring the db.
type dbOption func(*dbOptions)

// withDBPath sets a custom path to the Things database.
func withDBPath(path string) dbOption {
	return func(opts *dbOptions) {
		opts.databasePath = path
	}
}

// withDBPrintSQL enables SQL query logging to stdout.
func withDBPrintSQL(enabled bool) dbOption {
	return func(opts *dbOptions) {
		opts.printSQL = enabled
	}
}
