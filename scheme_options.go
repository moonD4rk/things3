package things3

// schemeOption configures scheme behavior.
type schemeOption func(*scheme)

// withForeground configures the scheme to bring Things to foreground.
func withForeground() schemeOption {
	return func(s *scheme) {
		s.foreground = true
	}
}

// withBackground configures the scheme to run in background.
func withBackground() schemeOption {
	return func(s *scheme) {
		s.background = true
	}
}
