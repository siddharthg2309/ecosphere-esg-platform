package domain

import (
	"github.com/shopspring/decimal"
	category "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/category/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"testing"
	"time"
)

func TestFactorRejectsNonPositiveValue(t *testing.T) {
	if _, err := New("Diesel", id.New(), "litre", decimal.Zero, category.StatusActive, time.Now()); err == nil {
		t.Fatal("expected validation error")
	}
}
