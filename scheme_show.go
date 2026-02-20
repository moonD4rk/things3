package things3

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// showBuilder builds URLs for navigating to items or lists via the show command.
type showBuilder struct {
	scheme *scheme
	params map[string]string
}

// ID sets the target item UUID or built-in list ID.
func (b *showBuilder) ID(id string) ShowNavigator {
	b.params[keyID] = id
	return b
}

// List sets the target to a built-in Things list.
func (b *showBuilder) List(list ListID) ShowNavigator {
	b.params[keyID] = string(list)
	return b
}

// Query searches for an area, project, or tag by name.
// Note: Tasks cannot be shown using query; use ID instead.
func (b *showBuilder) Query(query string) ShowNavigator {
	b.params[keyQuery] = query
	return b
}

// Filter filters the displayed items by tags.
func (b *showBuilder) Filter(tags ...string) ShowNavigator {
	b.params[keyFilter] = strings.Join(tags, ",")
	return b
}

// Build returns the Things URL for the show command.
func (b *showBuilder) Build() (string, error) {
	query := url.Values{}
	for k, v := range b.params {
		query.Set(k, v)
	}

	if len(query) == 0 {
		return fmt.Sprintf("things:///%s", CommandShow), nil
	}
	return fmt.Sprintf("things:///%s?%s", CommandShow, encodeQuery(query)), nil
}

// Execute builds and executes the show URL.
// By default, brings Things to foreground since the user wants to view content.
// Use WithBackground() option to run in background without stealing focus.
func (b *showBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.executeNavigation(ctx, uri)
}
