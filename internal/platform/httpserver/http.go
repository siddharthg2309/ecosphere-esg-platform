package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func Error(w http.ResponseWriter, err error) {
	if domainErr, ok := errs.As(err); ok {
		status := map[errs.Kind]int{errs.KindInvalid: 400, errs.KindUnauthorized: 401, errs.KindForbidden: 403, errs.KindNotFound: 404, errs.KindConflict: 409}[domainErr.Kind]
		JSON(w, status, domainErr)
		return
	}
	slog.Error("unhandled request error", "error", err)
	JSON(w, http.StatusInternalServerError, map[string]string{"code": "internal", "message": "Internal server error"})
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
				JSON(w, 500, map[string]string{"code": "internal", "message": "Internal server error"})
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
