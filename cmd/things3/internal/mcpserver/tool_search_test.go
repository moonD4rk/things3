package mcpserver

import (
	"context"
	"testing"
)

func search(t *testing.T, srv *Server, in SearchInput) PageResult[Item] {
	t.Helper()
	_, page, err := srv.handleSearch(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("search %+v: %v", in, err)
	}
	return page
}

func TestSearchEmpty(t *testing.T) {
	srv := newTestServer(t, Config{})
	page := search(t, srv, SearchInput{Query: "zzznomatchzzz"})
	if page.Total != 0 || len(page.Items) != 0 {
		t.Errorf("no-match search should be empty, got total=%d items=%d", page.Total, len(page.Items))
	}
	if page.Items == nil {
		t.Errorf("items must be a non-nil empty array")
	}
}

func TestSearchCrossType(t *testing.T) {
	srv := newTestServer(t, Config{})
	page := search(t, srv, SearchInput{Query: "Project in Today", Limit: 100})
	hasProject := false
	for i := range page.Items {
		if page.Items[i].Type == typeProject {
			hasProject = true
		}
	}
	if !hasProject {
		t.Errorf("cross-type search should surface a project:\n%+v", page.Items)
	}
}

func TestSearchTypeFilter(t *testing.T) {
	srv := newTestServer(t, Config{})

	t.Run("todo only", func(t *testing.T) {
		page := search(t, srv, SearchInput{Query: "Project in Today", Type: "todo", Limit: 100})
		for i := range page.Items {
			if page.Items[i].Type != typeTodo {
				t.Errorf("type=todo leaked a %q", page.Items[i].Type)
			}
		}
	})

	t.Run("project only", func(t *testing.T) {
		page := search(t, srv, SearchInput{Query: "Project in", Type: "project", Limit: 100})
		if len(page.Items) == 0 {
			t.Fatalf("expected project matches for 'Project in'")
		}
		for i := range page.Items {
			if page.Items[i].Type != typeProject {
				t.Errorf("type=project leaked a %q", page.Items[i].Type)
			}
		}
	})
}
