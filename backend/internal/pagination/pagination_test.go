package pagination

import "testing"

func TestBuild_ComputesTotalPages(t *testing.T) {
	p := Build([]int{1, 2, 3}, 25, Query{Page: 1, Limit: 10})
	if p.TotalPages != 3 {
		t.Fatalf("expected 3 total pages, got %d", p.TotalPages)
	}
	if p.Total != 25 || p.Page != 1 || p.Limit != 10 {
		t.Fatalf("unexpected envelope: %+v", p)
	}
}

func TestBuild_MinimumOnePage(t *testing.T) {
	p := Build([]int{}, 0, Query{Page: 1, Limit: 20})
	if p.TotalPages != 1 {
		t.Fatalf("expected a minimum of 1 total page, got %d", p.TotalPages)
	}
}

func TestQuery_OffsetAndLimit(t *testing.T) {
	q := Query{Page: 3, Limit: 20}
	if q.Offset() != 40 {
		t.Fatalf("expected offset 40, got %d", q.Offset())
	}
	if q.LimitInt32() != 20 {
		t.Fatalf("expected limit 20, got %d", q.LimitInt32())
	}
}
