package things3

import (
	"context"

	"github.com/moond4rk/things3/internal/database"
)

// areaQuery provides a fluent interface for building area queries.
// Chainable methods are copy-on-write: each call returns a new builder, so an
// areaQuery can be forked into independent queries.
type areaQuery struct {
	database *db
	filter   database.AreaFilter
}

// Areas creates a new areaQuery for querying areas.
func (d *db) Areas() *areaQuery {
	return &areaQuery{
		database: d,
	}
}

// clone returns a shallow copy of the query for copy-on-write chaining.
func (q *areaQuery) clone() *areaQuery {
	c := *q
	return &c
}

// WithUUID filters areas by UUID.
func (q *areaQuery) WithUUID(uuid string) AreaQueryBuilder {
	c := q.clone()
	c.filter.UUID = &uuid
	return c
}

// WithTitle filters areas by title.
func (q *areaQuery) WithTitle(title string) AreaQueryBuilder {
	c := q.clone()
	c.filter.Title = &title
	return c
}

// Visible filters areas by visibility status.
// Pass true to include only visible areas.
// Pass false to include only hidden areas.
func (q *areaQuery) Visible(visible bool) AreaQueryBuilder {
	c := q.clone()
	c.filter.Visible = &visible
	return c
}

// InTag filters areas by a specific tag title.
func (q *areaQuery) InTag(title string) AreaQueryBuilder {
	c := q.clone()
	c.filter.TagTitle = &title
	return c
}

// HasTag filters areas by whether they have any tags.
func (q *areaQuery) HasTag(has bool) AreaQueryBuilder {
	c := q.clone()
	c.filter.HasTag = &has
	return c
}

// All executes the query and returns all matching areas.
// The result is never nil; an empty result encodes as a JSON array.
func (q *areaQuery) All(ctx context.Context) ([]Area, error) {
	rows, err := q.database.inner.QueryAreas(ctx, q.filter)
	if err != nil {
		return nil, err
	}

	areas := make([]Area, 0, len(rows))
	for _, row := range rows {
		area := convertAreaRow(row)

		// Load tags if present
		if row.HasTags {
			tags, err := q.database.inner.TagsOfArea(ctx, row.UUID)
			if err != nil {
				return nil, err
			}
			area.Tags = tags
		}

		areas = append(areas, area)
	}

	return areas, nil
}

// First executes the query and returns the first matching area.
func (q *areaQuery) First(ctx context.Context) (*Area, error) {
	areas, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	if len(areas) == 0 {
		return nil, ErrAreaNotFound
	}
	return &areas[0], nil
}

// Count executes the query and returns the count of matching areas.
func (q *areaQuery) Count(ctx context.Context) (int, error) {
	return q.database.inner.CountAreas(ctx, q.filter)
}
