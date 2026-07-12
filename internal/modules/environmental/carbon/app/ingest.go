package app

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/carbon/port"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

const MaxEvidenceSize int64 = 10 << 20

var allowedMIME = map[string]bool{"application/pdf": true, "image/jpeg": true, "image/png": true}

type IngestService struct {
	storage    port.Storage
	ai         port.AIGateway
	flags      port.Flags
	confidence float64
}

func NewIngest(storage port.Storage, ai port.AIGateway, flags port.Flags, confidence float64) *IngestService {
	return &IngestService{storage: storage, ai: ai, flags: flags, confidence: confidence}
}
func (s *IngestService) Execute(ctx context.Context, filename, mime string, size int64, reader io.Reader) (port.Suggestion, error) {
	if !s.flags.IsEnabled(ctx, "auto_emission_calc") {
		return port.Suggestion{}, errs.Conflict("disabled", "Automatic emission calculation is disabled")
	}
	if size <= 0 || size > MaxEvidenceSize {
		return port.Suggestion{}, errs.Invalid("invalid_file_size", "Evidence must be 10 MB or smaller", map[string]string{"file": "Maximum size is 10 MB"})
	}
	if !allowedMIME[mime] {
		return port.Suggestion{}, errs.Invalid("invalid_file_type", "Evidence must be PDF, JPEG, or PNG", map[string]string{"file": "Unsupported file type"})
	}
	ext := strings.ToLower(filepath.Ext(filename))
	key := fmt.Sprintf("environmental/%s%s", id.New(), ext)
	stored, err := s.storage.Put(ctx, key, reader, mime, size)
	if err != nil {
		return port.Suggestion{}, err
	}
	url, err := s.storage.SignedURL(ctx, stored, 15*time.Minute)
	if err != nil {
		return port.Suggestion{}, err
	}
	suggestion, aiErr := s.ai.Categorize(ctx, port.DocInput{FileURL: url, MimeType: mime, Hint: filename})
	suggestion.EvidenceURL = stored
	if aiErr != nil || suggestion.Confidence < s.confidence {
		return port.Suggestion{EvidenceURL: stored}, nil
	}
	return suggestion, nil
}
