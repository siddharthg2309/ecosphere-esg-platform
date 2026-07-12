package domain

import (
	"testing"
	"time"
)

func TestNewNormalizesDepartmentCode(t *testing.T) {
	d, err := New("Logistics", " log ", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if d.Code != "LOG" {
		t.Fatalf("code=%q", d.Code)
	}
}
func TestNewRejectsInvalidDepartmentCode(t *testing.T) {
	if _, err := New("Logistics", "!", time.Now()); err == nil {
		t.Fatal("expected validation error")
	}
}
