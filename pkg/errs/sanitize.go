package errs

import (
	"strings"
	"unicode"
)

// Default client-facing copy when a raw driver / infrastructure error is detected.
const GenericClientMessage = "Something went wrong. Please try again."

// LooksTechnical reports whether a message appears to be a database/driver dump
// that must never be shown to end users.
func LooksTechnical(message string) bool {
	if message == "" {
		return true
	}
	lower := strings.ToLower(message)
	needles := []string{
		"sqlstate",
		"pq:",
		"pgx",
		"error:",
		"fatal:",
		"detail:",
		"hint:",
		"duplicate key",
		"violates unique constraint",
		"violates foreign key",
		"violates check constraint",
		"violates not-null",
		"relation \"",
		"column \"",
		"syntax error",
		"connection refused",
		"failed to connect",
		"dial tcp",
		"driver:",
		"could not connect",
		"server closed the connection",
		"ssl is not enabled",
		"password authentication failed",
		"role \"",
		"database \"",
		"context deadline exceeded",
		"i/o timeout",
		"broken pipe",
		"oid ",
		"prepared statement",
	}
	for _, n := range needles {
		if strings.Contains(lower, n) {
			return true
		}
	}
	// Long all-lowercase dump with many underscores often looks like SQL identifiers.
	if len(message) > 160 && strings.Count(message, "_") > 6 {
		return true
	}
	// Control characters / multi-line stack-ish content
	for _, r := range message {
		if r == '\n' || r == '\r' || (!unicode.IsPrint(r) && !unicode.IsSpace(r)) {
			return true
		}
	}
	return false
}

func friendlyForKind(kind Kind) string {
	switch kind {
	case KindInvalid:
		return "Please check your input and try again."
	case KindNotFound:
		return "The requested record was not found."
	case KindConflict:
		return "This action conflicts with the current data. Please refresh and try again."
	case KindForbidden:
		return "You do not have permission to perform this action."
	case KindUnauthorized:
		return "Please sign in to continue."
	default:
		return GenericClientMessage
	}
}

// ClientSafe returns a domain error safe to serialize to the client.
// Technical / driver text is replaced with a soft message; original codes are kept when possible.
func ClientSafe(err error) *Error {
	if err == nil {
		return Internal("internal", GenericClientMessage)
	}
	if e, ok := As(err); ok {
		msg := e.Message
		if LooksTechnical(msg) {
			msg = friendlyForKind(e.Kind)
		}
		out := &Error{Kind: e.Kind, Code: e.Code, Message: msg}
		if e.Kind == KindInvalid && len(e.Fields) > 0 {
			// Drop field details that look technical; keep human form-field messages.
			fields := map[string]string{}
			for k, v := range e.Fields {
				if !LooksTechnical(v) {
					fields[k] = v
				}
			}
			if len(fields) > 0 {
				out.Fields = fields
			}
		}
		if out.Kind == KindInternal && LooksTechnical(out.Message) {
			out.Message = GenericClientMessage
		}
		return out
	}
	if LooksTechnical(err.Error()) {
		return Internal("internal", GenericClientMessage)
	}
	return Internal("internal", GenericClientMessage)
}
