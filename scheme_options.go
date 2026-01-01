package things3

// SchemeOption configures Scheme behavior.
type SchemeOption func(*Scheme)

// WithForeground configures the Scheme to bring Things to foreground
// when executing create/update operations (Todo, Project, JSON, Update*).
//
// By default, create/update operations run in background without stealing focus.
// Use this option when you want Things to become the active window after creation.
//
// Example:
//
//	scheme := things3.NewScheme(things3.WithForeground())
//	scheme.Todo().Title("Buy milk").Execute(ctx)  // Things comes to foreground
func WithForeground() SchemeOption {
	return func(s *Scheme) {
		s.foreground = true
	}
}

// WithBackground configures the Scheme to run navigation operations
// (Show, Search, ShowBuilder) in the background without stealing focus.
//
// By default, navigation operations bring Things to foreground since
// the user typically wants to view the content.
// Use this option for programmatic navigation where focus change is undesired.
//
// Example:
//
//	scheme := things3.NewScheme(things3.WithBackground())
//	scheme.Show(ctx, "uuid")  // Things stays in background
func WithBackground() SchemeOption {
	return func(s *Scheme) {
		s.background = true
	}
}
