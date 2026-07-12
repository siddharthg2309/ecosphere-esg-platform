package ai

import (
	"encoding/json"
	"net/http"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
)

// Handler exposes AI assist endpoints (advisory only).
type Handler struct{ gw *Gateway }

func NewHandler(gw *Gateway) *Handler { return &Handler{gw: gw} }

func (h *Handler) ReviewEvidence(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ImageURL string `json:"imageUrl"`
		ProofURL string `json:"proofUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	url := body.ImageURL
	if url == "" {
		url = body.ProofURL
	}
	result, err := h.gw.ReviewEvidence(r.Context(), url)
	if err != nil {
		// Soft fail to fixture-like response
		result = EvidenceReview{LooksValid: false, Confidence: 0, Notes: "Evidence assist unavailable"}
	}
	httpserver.JSON(w, http.StatusOK, result)
}
