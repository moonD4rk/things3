package things3

import (
	"context"

	idb "github.com/moond4rk/things3/internal/db"
)

// areaQuery provides a fluent interface for building area queries.
type areaQuery struct {
	database     *db
	filter       idb.AreaFilter
	includeItems bool
}

// Areas creates a new areaQuery for querying areas.
func (d *db) Areas() *areaQuery {
	return &areaQuery{
		database: d,
	}
}

// WithUUID filters areas by UUID.
func (q *areaQuery) WithUUID(uuid string) AreaQueryBuilder {
	q.filter.UUID = &uuid
	return q
}

// WithTitle filters areas by title.
func (q *areaQuery) WithTitle(title string) AreaQueryBuilder {
	q.filter.Title = &title
	return q
}

// Visible filters areas by visibility status.
// Pass true to include only visible areas.
// Pass false to include only hidden areas.
func (q *areaQuery) Visible(visible bool) AreaQueryBuilder {
	q.filter.Visible = &visible
	return q
}

// InTag filters areas by a specific tag title.
func (q *areaQuery) InTag(title string) AreaQueryBuilder {
	q.filter.TagTitle = &title
	return q
}

// HasTag filters areas by whether they have any tags.
func (q *areaQuery) HasTag(has bool) AreaQueryBuilder {
	q.filter.HasTag = &has
	return q
}

// IncludeItems includes tasks in each area.
func (q *areaQuery) IncludeItems(include bool) AreaQueryBuilder {
	q.includeItems = include
	return q
}

// All executes the query and returns all matching areas.
func (q *areaQuery) All(ctx context.Context) ([]Area, error) {
	rows, err := q.database.inner.QueryAreas(ctx, q.filter)
	if err != nil {
		return nil, err
	}

	var areas []Area
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

		// Load items if requested
		if q.includeItems {
			items, err := q.database.Tasks().
				InArea(area.UUID).
				ContextTrashed(false).
				IncludeItems(true).
				All(ctx)
			if err != nil {
				return nil, err
			}
			area.Items = items
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
