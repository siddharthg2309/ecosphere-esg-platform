package db

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
)

func TestMapErrorUniqueViolation(t *testing.T) {
	err := MapError(&pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint \"x\""})
	e, ok := errs.As(err)
	if !ok || e.Kind != errs.KindConflict {
		t.Fatalf("got %#v", err)
	}
	if errs.LooksTechnical(e.Message) {
		t.Fatalf("technical message leaked: %q", e.Message)
	}
}

func TestMapErrorNoRows(t *testing.T) {
	err := MapError(pgx.ErrNoRows)
	e, ok := errs.As(err)
	if !ok || e.Kind != errs.KindNotFound {
		t.Fatalf("got %#v", err)
	}
}

func TestMapErrorKeepsDomain(t *testing.T) {
	want := errs.Invalid("evidence_required", "proof file required", nil)
	got := MapError(want)
	if got != want {
		t.Fatalf("domain error was rewritten: %#v", got)
	}
}

func TestMapErrorConnection(t *testing.T) {
	err := MapError(errors.New("failed to connect to `host=localhost`: dial tcp 127.0.0.1:5432: connection refused"))
	e, ok := errs.As(err)
	if !ok || e.Kind != errs.KindInternal {
		t.Fatalf("got %#v", err)
	}
	if errs.LooksTechnical(e.Message) {
		t.Fatalf("technical: %q", e.Message)
	}
}
