package things3

// SchemeOption configures Scheme behavior.
type SchemeOption func(*Scheme)

// WithForeground configures the Scheme to bring Things to foreground
// when executing URL scheme operations.
// By default, operations run in background without stealing focus.
func WithForeground() SchemeOption {
	return func(s *Scheme) {
		s.foreground = true
	}
}
