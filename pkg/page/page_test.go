package page

import "testing"

func TestNewNormalizesBounds(t *testing.T) {
	p := New(101, -1)
	if p.Limit != MaxLimit || p.Offset != 0 {
		t.Fatalf("unexpected page: %+v", p)
	}
	if got := New(0, 3); got.Limit != DefaultLimit || got.Offset != 3 {
		t.Fatalf("unexpected defaults: %+v", got)
	}
}
