package things3

import (
	"context"
	"os/exec"
)

// Show opens Things and shows the item with the given UUID.
func (d *DB) Show(ctx context.Context, uuid string) error {
	uri := NewScheme().Show().ID(uuid).Build()
	return exec.CommandContext(ctx, "open", uri).Run()
}

// Complete marks a task as complete using the Things URL scheme.
// Requires the URL scheme authentication token to be set in Things.
func (d *DB) Complete(ctx context.Context, uuid string) error {
	token, err := d.Token(ctx)
	if err != nil {
		return err
	}
	uri, err := NewScheme().WithToken(token).UpdateTodo(uuid).Completed(true).Build()
	if err != nil {
		return err
	}
	return exec.CommandContext(ctx, "open", uri).Run()
}

// OpenSearch opens Things and performs a search for the given query.
func (d *DB) OpenSearch(ctx context.Context, query string) error {
	uri := NewScheme().Search(query)
	return exec.CommandContext(ctx, "open", uri).Run()
}
