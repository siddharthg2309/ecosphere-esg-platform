package ai

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
)

// Handler exposes AI assist endpoints (advisory only).
type Handler struct{ gw *Gateway }

func NewHandler(gw *Gateway) *Handler { return &Handler{gw: gw} }

func (h *Handler) ReviewEvidence(w http.ResponseWriter, r *http.Request) {
	// Limit body (~4MB) so uploads are not stored and do not flood memory.
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	var body struct {
		ImageURL     string `json:"imageUrl"`
		ProofURL     string `json:"proofUrl"`
		ImageDataURL string `json:"imageDataUrl"` // data:image/...;base64,... — not persisted
		FileName     string `json:"fileName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid or image too large", nil))
		return
	}
	in := EvidenceInput{
		ImageURL: body.ImageURL,
		DataURL:  body.ImageDataURL,
		FileName: body.FileName,
	}
	if in.ImageURL == "" {
		in.ImageURL = body.ProofURL
	}
	// Reject accidental huge non-data payloads stored as "url"
	if strings.HasPrefix(in.ImageURL, "data:") && in.DataURL == "" {
		in.DataURL = in.ImageURL
		in.ImageURL = ""
	}
	result, err := h.gw.Review(r.Context(), in)
	if err != nil {
		result = EvidenceReview{LooksValid: false, Confidence: 0, Notes: "Evidence assist unavailable"}
	}
	httpserver.JSON(w, http.StatusOK, result)
}
