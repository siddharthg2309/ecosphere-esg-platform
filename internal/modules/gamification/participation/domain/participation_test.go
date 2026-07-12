package domain

import (
	"testing"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

func TestChallengeApproveEvidenceGate(t *testing.T) {
	p, _ := New(id.New(), id.New(), 100, "", time.Now())
	err := p.Approve(120, true)
	if err == nil {
		t.Fatal("expected evidence gate")
	}
	if e, ok := errs.As(err); !ok || e.Code != "evidence_required" {
		t.Fatalf("%v", err)
	}
	p.ProofURL = "proof.jpg"
	if err = p.Approve(120, true); err != nil {
		t.Fatal(err)
	}
	if p.XPAwarded != 120 || p.Approval != Approved {
		t.Fatalf("%+v", p)
	}
}

func TestChallengeAwardsXPNotCSRPointsField(t *testing.T) {
	p, _ := New(id.New(), id.New(), 100, "x.jpg", time.Now())
	_ = p.Approve(80, true)
	if p.XPAwarded != 80 {
		t.Fatal(p.XPAwarded)
	}
}
