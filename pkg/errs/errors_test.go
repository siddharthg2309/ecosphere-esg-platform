package errs

import (
	"fmt"
	"testing"
)

func TestAsFindsWrappedDomainError(t *testing.T) {
	want := Invalid("bad_value", "bad value", map[string]string{"name": "required"})
	got, ok := As(fmt.Errorf("wrapped: %w", want))
	if !ok || got != want {
		t.Fatalf("expected wrapped domain error")
	}
}
