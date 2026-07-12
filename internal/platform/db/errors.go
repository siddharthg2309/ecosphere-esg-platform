package db

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
)

// MapError converts driver / PostgreSQL errors into soft domain errors.
// Domain errors are returned unchanged. Unknown infrastructure failures become
// a generic internal error — never expose SQL text to callers.
func MapError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := errs.As(err); ok {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.NotFound("not_found", "The requested record was not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return errs.Conflict("duplicate_record", "A record with this value already exists")
		case "23503": // foreign_key_violation
			return errs.Invalid("invalid_reference", "A related record is missing or still in use", nil)
		case "23514": // check_violation
			return errs.Invalid("constraint_failed", "The request violates a business rule", nil)
		case "23502": // not_null_violation
			return errs.Invalid("missing_required", "A required value is missing", nil)
		case "22P02", "22001", "22003": // invalid_text_representation / string / numeric
			return errs.Invalid("invalid_input", "One or more values are invalid", nil)
		case "40001", "40P01": // serialization / deadlock
			return errs.Conflict("retry_later", "Please try again in a moment")
		case "42P01", "42703": // undefined_table / undefined_column
			return errs.Internal("data_unavailable", errs.GenericClientMessage)
		default:
			return errs.Internal("database_error", errs.GenericClientMessage)
		}
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "failed to connect"),
		strings.Contains(msg, "dial tcp"),
		strings.Contains(msg, "timeout"),
		strings.Contains(msg, "broken pipe"),
		strings.Contains(msg, "server closed"),
		strings.Contains(msg, "conn closed"),
		strings.Contains(msg, "too many clients"):
		return errs.Internal("database_unavailable", "The service is temporarily unavailable. Please try again.")
	case errs.LooksTechnical(err.Error()):
		return errs.Internal("database_error", errs.GenericClientMessage)
	default:
		// Unknown non-domain error — still never leak raw text to the HTTP layer.
		return errs.Internal("internal", errs.GenericClientMessage)
	}
}
