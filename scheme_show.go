package things3

import (
	"fmt"
	"net/url"
	"strings"
)

// ShowBuilder builds URLs for navigating to items or lists via the show command.
type ShowBuilder struct {
	params map[string]string
}

// ID sets the target item UUID or built-in list ID.
func (b *ShowBuilder) ID(id string) *ShowBuilder {
	b.params["id"] = id
	return b
}

// List sets the target to a built-in Things list.
func (b *ShowBuilder) List(list ListID) *ShowBuilder {
	b.params["id"] = string(list)
	return b
}

// Query searches for an area, project, or tag by name.
// Note: Tasks cannot be shown using query; use ID instead.
func (b *ShowBuilder) Query(query string) *ShowBuilder {
	b.params["query"] = query
	return b
}

// Filter filters the displayed items by tags.
func (b *ShowBuilder) Filter(tags ...string) *ShowBuilder {
	b.params["filter"] = strings.Join(tags, ",")
	return b
}

// Build returns the Things URL for the show command.
func (b *ShowBuilder) Build() string {
	query := url.Values{}
	for k, v := range b.params {
		query.Set(k, v)
	}

	if len(query) == 0 {
		return fmt.Sprintf("things:///%s", CommandShow)
	}
	return fmt.Sprintf("things:///%s?%s", CommandShow, query.Encode())
}
