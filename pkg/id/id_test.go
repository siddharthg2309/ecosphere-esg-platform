package id

import "testing"

func TestNewCreatesValidDistinctIDs(t *testing.T) {
	a, b := New(), New()
	if a == b {
		t.Fatal("expected distinct IDs")
	}
	if _, err := Parse(a.String()); err != nil {
		t.Fatalf("generated invalid ID: %v", err)
	}
}

func TestParseRejectsInvalidID(t *testing.T) {
	if _, err := Parse("not-an-id"); err == nil {
		t.Fatal("expected invalid ID error")
	}
}
