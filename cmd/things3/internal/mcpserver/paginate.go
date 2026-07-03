package mcpserver

// Pagination bounds. Unlike the CLI there is deliberately no unlimited mode:
// MCP output lands in a model context window. The default is larger than the
// CLI's because a tool round trip costs more than a terminal keystroke.
const (
	// DefaultLimit is the page size when limit is unset or non-positive.
	DefaultLimit = 20
	// MaxLimit caps the page size a caller may request.
	MaxLimit = 100
)

// clampLimit resolves a requested limit into the range [1, MaxLimit], mapping a
// non-positive request to DefaultLimit.
func clampLimit(n int) int {
	switch {
	case n <= 0:
		return DefaultLimit
	case n > MaxLimit:
		return MaxLimit
	default:
		return n
	}
}

// paginate returns the 1-based page slice of items and its metadata. An
// out-of-range page yields an empty slice with intact total/page/pages.
func paginate[T any](items []T, page, limit int) (slice []T, total, pageOut, pages int) {
	limit = clampLimit(limit)
	if page < 1 {
		page = 1
	}
	total = len(items)
	pages = max((total+limit-1)/limit, 1)
	// Guard on pages before multiplying: an extreme page would overflow
	// (page-1)*limit into a negative offset and panic the slice.
	if page > pages {
		return []T{}, total, page, pages
	}
	offset := (page - 1) * limit
	return items[offset:min(offset+limit, total)], total, page, pages
}

// pageResult paginates a full result slice into a success envelope, guaranteeing
// a non-nil Items array.
func pageResult[T any](items []T, page, limit int) PageResult[T] {
	slice, total, pageOut, pages := paginate(items, page, limit)
	if slice == nil {
		slice = []T{}
	}
	return PageResult[T]{Success: true, Items: slice, Total: total, Page: pageOut, Pages: pages}
}

// pageError builds a failure envelope carrying a structured error and an empty
// Items array so the shape validates against the tool's output schema.
func pageError[T any](te *ToolError) PageResult[T] {
	return PageResult[T]{Success: false, Error: te, Items: []T{}, Page: 1, Pages: 1}
}
