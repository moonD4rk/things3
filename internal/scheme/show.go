package scheme

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// showBuilder builds URLs for navigating to items or lists via the show command.
type showBuilder struct {
	scheme *Scheme
	params map[string]string
	err    error
}

// NewShowNavigator creates a new ShowNavigator for navigation operations.
func NewShowNavigator(s *Scheme) ShowNavigator {
	return &showBuilder{scheme: s, params: make(map[string]string)}
}

// ID sets the target item UUID or built-in list ID.
func (b *showBuilder) ID(id string) ShowNavigator {
	b.params[KeyID] = id
	return b
}

// List sets the target to a built-in Things list.
func (b *showBuilder) List(list ListID) ShowNavigator {
	b.params[KeyID] = string(list)
	return b
}

// Query searches for an area, project, or tag by name.
// Note: Tasks cannot be shown using query; use ID instead.
func (b *showBuilder) Query(query string) ShowNavigator {
	b.params[KeyQuery] = query
	return b
}

// Filter filters the displayed items by tags.
// Tags are comma-separated in the URL, so a tag must not contain a comma.
func (b *showBuilder) Filter(tags ...string) ShowNavigator {
	for _, tag := range tags {
		if strings.Contains(tag, ",") {
			b.err = ErrTagContainsComma
			return b
		}
	}
	b.params[KeyFilter] = strings.Join(tags, ",")
	return b
}

// Build returns the Things URL for the show command.
func (b *showBuilder) Build() (string, error) {
	if b.err != nil {
		return "", b.err
	}

	query := url.Values{}
	for k, v := range b.params {
		query.Set(k, v)
	}

	if len(query) == 0 {
		return fmt.Sprintf("things:///%s", CommandShow), nil
	}
	return fmt.Sprintf("things:///%s?%s", CommandShow, EncodeQuery(query)), nil
}

// Execute builds and executes the show URL.
// By default, brings Things to foreground since the user wants to view content.
// Use WithBackground() option to run in background without stealing focus.
func (b *showBuilder) Execute(ctx context.Context) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	return b.scheme.ExecuteNavigation(ctx, uri)
}
