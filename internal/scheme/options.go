package scheme

// Option configures Scheme behavior.
type Option func(*Scheme)

// WithForeground configures the scheme to bring Things to foreground.
func WithForeground() Option {
	return func(s *Scheme) {
		s.foreground = true
	}
}

// WithBackground configures the scheme to run in background.
func WithBackground() Option {
	return func(s *Scheme) {
		s.background = true
	}
}
