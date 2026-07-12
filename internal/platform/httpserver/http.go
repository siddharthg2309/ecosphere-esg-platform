package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/db"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// Error maps infrastructure/database failures to soft domain errors, logs the
// original cause server-side, and never writes raw driver text to the response.
func Error(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	// Log the real error before sanitization (includes SQL / driver detail for ops).
	if _, ok := errs.As(err); !ok {
		slog.Error("request error", "error", err)
	} else if e, ok := errs.As(err); ok && errs.LooksTechnical(e.Message) {
		slog.Error("request domain error with technical message", "code", e.Code, "error", err)
	}

	mapped := db.MapError(err)
	safe := errs.ClientSafe(mapped)

	statusByKind := map[errs.Kind]int{
		errs.KindInvalid:      http.StatusBadRequest,
		errs.KindUnauthorized: http.StatusUnauthorized,
		errs.KindForbidden:    http.StatusForbidden,
		errs.KindNotFound:     http.StatusNotFound,
		errs.KindConflict:     http.StatusConflict,
		errs.KindInternal:     http.StatusInternalServerError,
	}
	status := statusByKind[safe.Kind]
	if status == 0 {
		status = http.StatusInternalServerError
	}
	JSON(w, status, safe)
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, requestID)))
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic", "value", recovered, "stack", string(debug.Stack()))
				JSON(w, http.StatusInternalServerError, errs.Internal("internal", errs.GenericClientMessage))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started), "requestId", r.Context().Value(requestIDKey))
	})
}

func CORS(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
