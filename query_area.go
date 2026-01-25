package things3

import (
	"context"
)

// areaQuery provides a fluent interface for building area queries.
type areaQuery struct {
	database *db

	uuid         *string
	title        *string
	visible      *bool
	tagTitle     any // string, bool, or nil
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
	q.uuid = &uuid
	return q
}

// WithTitle filters areas by title.
func (q *areaQuery) WithTitle(title string) AreaQueryBuilder {
	q.title = &title
	return q
}

// Visible filters areas by visibility status.
// Pass true to include only visible areas.
// Pass false to include only hidden areas.
func (q *areaQuery) Visible(visible bool) AreaQueryBuilder {
	q.visible = &visible
	return q
}

// InTag filters areas by tag.
func (q *areaQuery) InTag(tag any) AreaQueryBuilder {
	q.tagTitle = tag
	return q
}

// IncludeItems includes tasks in each area.
func (q *areaQuery) IncludeItems(include bool) AreaQueryBuilder {
	q.includeItems = include
	return q
}

// buildWhere builds the WHERE clause for the area query using filterBuilder.
func (q *areaQuery) buildWhere() string {
	fb := newFilterBuilder()

	if q.uuid != nil {
		fb.addEqual("AREA.uuid", *q.uuid)
	}
	if q.title != nil {
		fb.addEqual("AREA.title", *q.title)
	}
	if q.visible != nil {
		fb.addTruthy("AREA.visible", q.visible)
	}
	fb.addEqual("TAG.title", q.tagTitle)

	return fb.sql()
}

// All executes the query and returns all matching areas.
func (q *areaQuery) All(ctx context.Context) ([]Area, error) {
	sql := buildAreasSQL(q.buildWhere())
	rows, err := q.database.executeQuery(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []Area
	for rows.Next() {
		area, err := scanArea(rows)
		if err != nil {
			return nil, err
		}

		// Load tags if present
		if area.Tags != nil {
			tags, err := q.database.getTagsOfArea(ctx, area.UUID)
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

		areas = append(areas, *area)
	}

	return areas, rows.Err()
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
	areaSQL := buildAreasSQL(q.buildWhere())
	countSQL := buildCountSQL(areaSQL)

	var count int
	if err := q.database.executeQueryRow(ctx, countSQL).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
