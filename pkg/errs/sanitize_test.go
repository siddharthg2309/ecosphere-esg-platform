package errs

import (
	"errors"
	"testing"
)

func TestLooksTechnicalDetectsPostgresDump(t *testing.T) {
	cases := []string{
		`ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)`,
		`pq: relation "csr_activities" does not exist`,
		`failed to connect to ` + `postgres: dial tcp 127.0.0.1:5432: connection refused`,
		"password authentication failed for user \"ecosphere\"",
	}
	for _, c := range cases {
		if !LooksTechnical(c) {
			t.Fatalf("expected technical: %q", c)
		}
	}
}

func TestLooksTechnicalAllowsHumanMessages(t *testing.T) {
	cases := []string{
		"proof file required",
		"Department code is already in use",
		"You do not have permission to perform this action",
		"Please check your input and try again.",
	}
	for _, c := range cases {
		if LooksTechnical(c) {
			t.Fatalf("expected human-safe: %q", c)
		}
	}
}

func TestClientSafeSoftensDomainTechnicalMessage(t *testing.T) {
	raw := Invalid("bad", `ERROR: column "foo" does not exist`, nil)
	safe := ClientSafe(raw)
	if LooksTechnical(safe.Message) {
		t.Fatalf("still technical: %q", safe.Message)
	}
	if safe.Code != "bad" {
		t.Fatalf("code=%s", safe.Code)
	}
}

func TestClientSafeMapsUnknownToInternal(t *testing.T) {
	safe := ClientSafe(errors.New("pq: SSL is not enabled on the server"))
	if safe.Kind != KindInternal {
		t.Fatalf("kind=%s", safe.Kind)
	}
	if safe.Message != GenericClientMessage {
		t.Fatalf("message=%q", safe.Message)
	}
}

func TestClientSafeKeepsFriendlyDomain(t *testing.T) {
	raw := Conflict("duplicate_participation", "Already joined this activity")
	safe := ClientSafe(raw)
	if safe.Message != "Already joined this activity" {
		t.Fatalf("message=%q", safe.Message)
	}
}
