package things3

import (
	"context"
)

// tagQuery provides a fluent interface for building tag queries.
type tagQuery struct {
	database *db

	uuid         *string
	title        *string
	parentUUID   *string
	includeItems bool
}

// Tags creates a new tagQuery for querying tags.
func (d *db) Tags() *tagQuery {
	return &tagQuery{
		database: d,
	}
}

// WithUUID filters tags by UUID.
func (q *tagQuery) WithUUID(uuid string) TagQueryBuilder {
	q.uuid = &uuid
	return q
}

// WithTitle filters tags by title.
func (q *tagQuery) WithTitle(title string) TagQueryBuilder {
	q.title = &title
	return q
}

// WithParent filters tags by parent tag UUID.
// Use this to find child tags of a specific parent tag.
func (q *tagQuery) WithParent(parentUUID string) TagQueryBuilder {
	q.parentUUID = &parentUUID
	return q
}

// IncludeItems includes areas and tasks for each tag.
func (q *tagQuery) IncludeItems(include bool) TagQueryBuilder {
	q.includeItems = include
	return q
}

// buildWhere builds the WHERE clause for the tag query using filterBuilder.
func (q *tagQuery) buildWhere() string {
	fb := newFilterBuilder()

	if q.uuid != nil {
		fb.addEqual("uuid", *q.uuid)
	}
	if q.title != nil {
		fb.addEqual("title", *q.title)
	}
	if q.parentUUID != nil {
		fb.addEqual("parent", *q.parentUUID)
	}

	return fb.sql()
}

// All executes the query and returns all matching tags.
func (q *tagQuery) All(ctx context.Context) ([]Tag, error) {
	sql := buildTagsSQL(q.buildWhere())
	rows, err := q.database.executeQuery(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		tag, err := scanTag(rows)
		if err != nil {
			return nil, err
		}

		// Load items if requested
		if q.includeItems {
			areas, err := q.database.Areas().InTag(tag.Title).All(ctx)
			if err != nil {
				return nil, err
			}
			tasks, err := q.database.Tasks().InTag(tag.Title).ContextTrashed(false).All(ctx)
			if err != nil {
				return nil, err
			}

			items := make([]any, 0, len(areas)+len(tasks))
			for i := range areas {
				items = append(items, &areas[i])
			}
			for i := range tasks {
				items = append(items, &tasks[i])
			}
			tag.Items = items
		}

		tags = append(tags, *tag)
	}

	return tags, rows.Err()
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
