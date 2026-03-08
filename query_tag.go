package things3

import (
	"context"

	"github.com/moond4rk/things3/internal/database"
)

// tagQuery provides a fluent interface for building tag queries.
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

// WithUUID filters tags by UUID.
func (q *tagQuery) WithUUID(uuid string) TagQueryBuilder {
	q.filter.UUID = &uuid
	return q
}

// WithTitle filters tags by title.
func (q *tagQuery) WithTitle(title string) TagQueryBuilder {
	q.filter.Title = &title
	return q
}

// WithParent filters tags by parent tag UUID.
// Use this to find child tags of a specific parent tag.
func (q *tagQuery) WithParent(parentUUID string) TagQueryBuilder {
	q.filter.ParentUUID = &parentUUID
	return q
}

// All executes the query and returns all matching tags.
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
