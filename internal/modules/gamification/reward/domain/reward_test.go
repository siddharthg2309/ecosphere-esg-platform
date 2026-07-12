package domain

import (
	"testing"
	"time"

	category "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/category/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

func TestRedeemSuccess(t *testing.T) {
	r, err := New("Gift Card", "₹500", 800, 12, category.StatusActive, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	emp := &Balance{UserID: id.New(), Points: 1000}
	if err = r.Redeem(emp); err != nil {
		t.Fatal(err)
	}
	if emp.Points != 200 || r.Stock != 11 {
		t.Fatalf("points=%d stock=%d", emp.Points, r.Stock)
	}
}

func TestRedeemInsufficientPoints(t *testing.T) {
	r, _ := New("Gift Card", "", 800, 12, category.StatusActive, time.Now())
	emp := &Balance{UserID: id.New(), Points: 100}
	err := r.Redeem(emp)
	if err == nil {
		t.Fatal("expected error")
	}
	if e, ok := errs.As(err); !ok || e.Code != "insufficient_points" {
		t.Fatalf("%v", err)
	}
}

func TestRedeemOutOfStock(t *testing.T) {
	r, _ := New("Plant a Tree", "", 200, 0, category.StatusActive, time.Now())
	emp := &Balance{UserID: id.New(), Points: 500}
	err := r.Redeem(emp)
	if err == nil {
		t.Fatal("expected error")
	}
	if e, ok := errs.As(err); !ok || e.Code != "out_of_stock" {
		t.Fatalf("%v", err)
	}
}
