package db

// Options holds the configuration options for the DB.
type Options struct {
	DatabasePath string
	PrintSQL     bool
}

// Option is a functional option for configuring the DB.
type Option func(*Options)

// WithPath sets a custom path to the Things database.
func WithPath(path string) Option {
	return func(opts *Options) {
		opts.DatabasePath = path
	}
}

// WithPrintSQL enables SQL query logging to stdout.
func WithPrintSQL(enabled bool) Option {
	return func(opts *Options) {
		opts.PrintSQL = enabled
	}
}
