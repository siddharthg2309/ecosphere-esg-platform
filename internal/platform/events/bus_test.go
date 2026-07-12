package events

import (
	"context"
	"testing"
)

type testEvent struct{}

func (testEvent) Name() string { return "TestEvent" }

func TestInProcessPublishesToSubscribers(t *testing.T) {
	bus := NewInProcess()
	called := 0
	bus.Subscribe("TestEvent", func(context.Context, Event) error { called++; return nil })
	if err := bus.Publish(context.Background(), testEvent{}); err != nil {
		t.Fatal(err)
	}
	if called != 1 {
		t.Fatalf("handler called %d times", called)
	}
}
