package mcpserver

import "testing"

// TestPaginate covers the envelope math plus the page-zero, out-of-range, and
// integer-overflow guards.
func TestPaginate(t *testing.T) {
	items := make([]int, 25)
	for i := range items {
		items[i] = i
	}
	cases := []struct {
		name                      string
		page, limit               int
		wantLen, wantTotal        int
		wantPageOut, wantPagesOut int
	}{
		{"first page default", 1, 10, 10, 25, 1, 3},
		{"last partial page", 3, 10, 5, 25, 3, 3},
		{"page zero clamps to one", 0, 10, 10, 25, 1, 3},
		{"beyond last page is empty", 4, 10, 0, 25, 4, 3},
		{"limit over max clamps to all", 1, 1000, 25, 25, 1, 1},
		{"extreme page does not overflow", 92233720368547760, 100, 0, 25, 92233720368547760, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			slice, total, page, pages := paginate(items, tc.page, tc.limit)
			if len(slice) != tc.wantLen || total != tc.wantTotal || page != tc.wantPageOut || pages != tc.wantPagesOut {
				t.Errorf("paginate(page=%d, limit=%d) = len %d, total %d, page %d, pages %d; want len %d, total %d, page %d, pages %d",
					tc.page, tc.limit, len(slice), total, page, pages, tc.wantLen, tc.wantTotal, tc.wantPageOut, tc.wantPagesOut)
			}
		})
	}
}

// TestPageResultNeverNil proves the envelope always carries a JSON array, even for
// an empty result or an out-of-range page.
func TestPageResultNeverNil(t *testing.T) {
	if got := pageResult([]int{}, 1, 20); got.Items == nil {
		t.Errorf("empty result Items must not be nil")
	}
	if got := pageResult([]int{1, 2, 3}, 9, 20); got.Items == nil {
		t.Errorf("out-of-range page Items must not be nil")
	}
}
