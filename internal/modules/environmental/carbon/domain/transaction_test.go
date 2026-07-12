package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

func TestVerifyComputesDeterministicallyAndIsImmutable(t *testing.T) {
	now := time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC)
	txn, err := New(id.New(), SourceFleet, decimal.RequireFromString("268"), id.New(), decimal.RequireFromString("2.6800"), now, "evidence/invoice.pdf", now)
	if err != nil {
		t.Fatal(err)
	}
	verifier := id.New()
	if err = txn.Verify(verifier, now); err != nil {
		t.Fatal(err)
	}
	if got := txn.ComputedCO2.StringFixed(3); got != "718.240" {
		t.Fatalf("computed CO2 = %s", got)
	}
	if err = txn.Verify(verifier, now); err == nil {
		t.Fatal("expected repeated verification to fail")
	} else if e, ok := errs.As(err); !ok || e.Code != "already_verified" {
		t.Fatalf("unexpected error: %v", err)
	}
}
