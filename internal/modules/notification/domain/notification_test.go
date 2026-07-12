package domain

import (
	"testing"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

func TestNewNotification(t *testing.T) {
	n, err := New(id.New(), TypeComplianceRaised, "Issue raised", map[string]string{"severity": "high"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if n.ReadAt != nil {
		t.Fatal("should start unread")
	}
	n.MarkRead(time.Now())
	if n.ReadAt == nil {
		t.Fatal("expected read")
	}
}
