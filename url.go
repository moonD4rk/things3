package things3

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
)

// URL builds a Things URL scheme URL.
// See https://culturedcode.com/things/help/url-scheme/ for details.
func (c *Client) URL(ctx context.Context, cmd URLCommand, params map[string]string) (string, error) {
	// For update commands, we need an auth token
	if cmd == URLCommandUpdate || cmd == URLCommandUpdateProject {
		token, err := c.Token(ctx)
		if err != nil {
			return "", err
		}
		if params == nil {
			params = make(map[string]string)
		}
		params["auth-token"] = token
	}

	// Build query string
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}

	return fmt.Sprintf("things:///%s?%s", cmd, query.Encode()), nil
}

// ShowURL returns a URL to show an item in Things.
func (c *Client) ShowURL(uuid string) string {
	query := url.Values{}
	query.Set("id", uuid)
	return fmt.Sprintf("things:///%s?%s", URLCommandShow, query.Encode())
}

// Show opens Things and shows the item with the given UUID.
func (c *Client) Show(ctx context.Context, uuid string) error {
	uri := c.ShowURL(uuid)
	return exec.CommandContext(ctx, "open", uri).Run()
}

// Complete marks a task as complete using the Things URL scheme.
// Requires the URL scheme authentication token to be set in Things.
func (c *Client) Complete(ctx context.Context, uuid string) error {
	uri, err := c.URL(ctx, URLCommandUpdate, map[string]string{
		"id":        uuid,
		"completed": "true",
	})
	if err != nil {
		return err
	}
	return exec.CommandContext(ctx, "open", uri).Run()
}

// Link returns a things:// URL that shows the item.
// Alias for ShowURL for backwards compatibility.
func (c *Client) Link(uuid string) string {
	return c.ShowURL(uuid)
}

// AddTodoURL returns a URL to add a new to-do with the given parameters.
func (c *Client) AddTodoURL(params map[string]string) string {
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	return fmt.Sprintf("things:///%s?%s", URLCommandAdd, query.Encode())
}

// AddProjectURL returns a URL to add a new project with the given parameters.
func (c *Client) AddProjectURL(params map[string]string) string {
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	return fmt.Sprintf("things:///%s?%s", URLCommandAddProject, query.Encode())
}

// SearchURL returns a URL to search for the given query in Things.
func (c *Client) SearchURL(query string) string {
	q := url.Values{}
	q.Set("query", query)
	return fmt.Sprintf("things:///%s?%s", URLCommandSearch, q.Encode())
}

// OpenSearch opens Things and performs a search for the given query.
func (c *Client) OpenSearch(ctx context.Context, query string) error {
	uri := c.SearchURL(query)
	return exec.CommandContext(ctx, "open", uri).Run()
}
