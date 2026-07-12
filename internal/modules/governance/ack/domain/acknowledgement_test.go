package domain

import (
	"testing"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

func TestNewAckRequiresVersion(t *testing.T) {
	_, err := New(id.New(), id.New(), 0, time.Now())
	if err == nil {
		t.Fatal("expected version error")
	}
	a, err := New(id.New(), id.New(), 3, time.Now())
	if err != nil || a.Version != 3 {
		t.Fatalf("%v %+v", err, a)
	}
}
