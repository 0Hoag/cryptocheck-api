package paginator

import "testing"

func TestPaginatorQueryAdjustAppliesDefaultsAndMaximum(t *testing.T) {
	tests := []struct {
		name  string
		query PaginatorQuery
		page  int
		limit int64
	}{
		{name: "defaults invalid values", query: PaginatorQuery{}, page: defaultPage, limit: defaultLimit},
		{name: "keeps valid values", query: PaginatorQuery{Page: 3, Limit: 25}, page: 3, limit: 25},
		{name: "caps oversized pages", query: PaginatorQuery{Page: 1, Limit: maxLimit + 1}, page: 1, limit: maxLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.query.Adjust()
			if tt.query.Page != tt.page || tt.query.Limit != tt.limit {
				t.Fatalf("got page=%d limit=%d, want page=%d limit=%d", tt.query.Page, tt.query.Limit, tt.page, tt.limit)
			}
		})
	}
}
