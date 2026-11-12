package database

import "time"

// Options holds the configuration options for the DB.
type Options struct {
	DatabasePath string
	PrintSQL     bool
	Now          func() time.Time
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

// WithNow sets the clock that decides what "today" means for relative date
// filters. A nil clock keeps the system clock.
func WithNow(now func() time.Time) Option {
	return func(opts *Options) {
		opts.Now = now
	}
}
