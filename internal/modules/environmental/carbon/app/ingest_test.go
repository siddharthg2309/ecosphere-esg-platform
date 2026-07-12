package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/carbon/port"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
)

type testFlags bool

func (f testFlags) IsEnabled(context.Context, string) bool { return bool(f) }

type testStorage struct{}

func (testStorage) Put(_ context.Context, key string, _ io.Reader, _ string, _ int64) (string, error) {
	return key, nil
}
func (testStorage) SignedURL(context.Context, string, time.Duration) (string, error) {
	return "https://storage.test/evidence", nil
}

type testAI struct {
	suggestion port.Suggestion
	err        error
}

func (a testAI) Categorize(context.Context, port.DocInput) (port.Suggestion, error) {
	return a.suggestion, a.err
}

func TestIngestFeatureGateAndValidation(t *testing.T) {
	tests := []struct {
		name, mime string
		size       int64
		enabled    bool
		code       string
	}{
		{"disabled", "application/pdf", 10, false, "disabled"},
		{"too large", "application/pdf", MaxEvidenceSize + 1, true, "invalid_file_size"},
		{"wrong type", "text/plain", 10, true, "invalid_file_type"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := NewIngest(testStorage{}, testAI{}, testFlags(tc.enabled), .65)
			_, err := service.Execute(context.Background(), "invoice.pdf", tc.mime, tc.size, bytes.NewReader([]byte("data")))
			if err == nil {
				t.Fatal("expected error")
			}
			domainErr, ok := errs.As(err)
			if !ok || domainErr.Code != tc.code {
				t.Fatalf("error = %v", err)
			}
		})
	}
}
func TestIngestFallsBackToManualAndPreservesEvidence(t *testing.T) {
	for _, ai := range []testAI{{suggestion: port.Suggestion{Source: "fleet", Quantity: 268, Unit: "L", Confidence: .4}}, {err: errors.New("provider unavailable")}} {
		service := NewIngest(testStorage{}, ai, testFlags(true), .65)
		suggestion, err := service.Execute(context.Background(), "fuel.pdf", "application/pdf", 4, bytes.NewReader([]byte("data")))
		if err != nil {
			t.Fatal(err)
		}
		if suggestion.EvidenceURL == "" {
			t.Fatal("evidence key was not preserved")
		}
		if suggestion.Confidence != 0 {
			t.Fatalf("manual fallback confidence = %v", suggestion.Confidence)
		}
	}
}
