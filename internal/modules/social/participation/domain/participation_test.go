package domain

import (
	"testing"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

func TestApproveEvidenceGate(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	p, err := New(id.New(), id.New(), "", "", now)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Approve(50, true, now)
	if err == nil {
		t.Fatal("expected evidence_required error")
	}
	if e, ok := errs.As(err); !ok || e.Code != "evidence_required" {
		t.Fatalf("want evidence_required, got %v", err)
	}

	p.ProofURL = "https://files.example/photo.jpg"
	if err = p.Approve(50, true, now); err != nil {
		t.Fatal(err)
	}
	if p.Approval != Approved || p.PointsEarned != 50 || p.CompletionDate == nil {
		t.Fatalf("unexpected state: %+v", p)
	}
}

func TestApproveWithoutEvidenceWhenNotRequired(t *testing.T) {
	now := time.Now().UTC()
	p, _ := New(id.New(), id.New(), "", "", now)
	if err := p.Approve(30, false, now); err != nil {
		t.Fatal(err)
	}
	if p.PointsEarned != 30 {
		t.Fatalf("points=%d", p.PointsEarned)
	}
}

func TestApproveTwiceConflict(t *testing.T) {
	now := time.Now().UTC()
	p, _ := New(id.New(), id.New(), "proof.jpg", "", now)
	_ = p.Approve(10, true, now)
	err := p.Approve(10, true, now)
	if err == nil {
		t.Fatal("expected conflict")
	}
}

func TestReject(t *testing.T) {
	now := time.Now().UTC()
	p, _ := New(id.New(), id.New(), "proof.jpg", "", now)
	if err := p.Reject(); err != nil {
		t.Fatal(err)
	}
	if p.Approval != Rejected {
		t.Fatal(p.Approval)
	}
}
