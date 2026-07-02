package things3

import (
	"context"

	"github.com/moond4rk/things3/internal/database"
)

// tagQuery provides a fluent interface for building tag queries.
// Chainable methods are copy-on-write: each call returns a new builder, so a
// tagQuery can be forked into independent queries.
type tagQuery struct {
	database *db
	filter   database.TagFilter
}

// Tags creates a new tagQuery for querying tags.
func (d *db) Tags() *tagQuery {
	return &tagQuery{
		database: d,
	}
}

// clone returns a shallow copy of the query for copy-on-write chaining.
func (q *tagQuery) clone() *tagQuery {
	c := *q
	return &c
}

// WithUUID filters tags by UUID.
func (q *tagQuery) WithUUID(uuid string) TagQueryBuilder {
	c := q.clone()
	c.filter.UUID = &uuid
	return c
}

// WithTitle filters tags by title.
func (q *tagQuery) WithTitle(title string) TagQueryBuilder {
	c := q.clone()
	c.filter.Title = &title
	return c
}

// WithParent filters tags by parent tag UUID.
// Use this to find child tags of a specific parent tag.
func (q *tagQuery) WithParent(parentUUID string) TagQueryBuilder {
	c := q.clone()
	c.filter.ParentUUID = &parentUUID
	return c
}

// All executes the query and returns all matching tags.
// The result is never nil; an empty result encodes as a JSON array.
func (q *tagQuery) All(ctx context.Context) ([]Tag, error) {
	rows, err := q.database.inner.QueryTags(ctx, q.filter)
	if err != nil {
		return nil, err
	}

	tags := make([]Tag, len(rows))
	for i, row := range rows {
		tags[i] = convertTagRow(row)
	}

	return tags, nil
}

// First executes the query and returns the first matching tag.
func (q *tagQuery) First(ctx context.Context) (*Tag, error) {
	tags, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, ErrTagNotFound
	}
	return &tags[0], nil
}
