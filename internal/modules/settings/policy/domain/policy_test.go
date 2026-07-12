package domain

import (
	"testing"
	"time"
)

func TestPublishIncrementsVersion(t *testing.T) {
	p, err := New("Code", "Body", time.Now(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Publish("Code", "Updated", time.Now()); err != nil {
		t.Fatal(err)
	}
	if p.Version != 2 {
		t.Fatalf("version=%d", p.Version)
	}
}
