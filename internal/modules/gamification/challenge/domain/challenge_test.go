package domain

import (
	"testing"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

func TestIllegalTransition(t *testing.T) {
	c, err := New("Commute Green Week", id.New(), "desc", 120, "medium", true, nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	err = c.Transition(StatusCompleted)
	if err == nil {
		t.Fatal("draft → completed should fail")
	}
	if e, ok := errs.As(err); !ok || e.Code != "bad_transition" {
		t.Fatalf("want bad_transition got %v", err)
	}
}

func TestLegalLifecycle(t *testing.T) {
	c, _ := New("Sprint", id.New(), "d", 200, "hard", true, nil, time.Now())
	steps := []ChallengeStatus{StatusActive, StatusUnderReview, StatusCompleted, StatusArchived}
	for _, to := range steps {
		if err := c.Transition(to); err != nil {
			t.Fatalf("%s: %v", to, err)
		}
	}
	if c.Status != StatusArchived {
		t.Fatal(c.Status)
	}
}

func TestAllowedTargets(t *testing.T) {
	got := AllowedTargets(StatusActive)
	if len(got) != 2 {
		t.Fatalf("%v", got)
	}
}
