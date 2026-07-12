package ai

import (
	"context"
	"strings"
)

// EvidenceReview is an advisory check for CSR/challenge proof images.
// Humans still approve; this never auto-approves.
type EvidenceReview struct {
	LooksValid bool    `json:"looksValid"`
	Confidence float64 `json:"confidence"`
	Notes      string  `json:"notes"`
}

// EvidenceAssistor reviews proof evidence (advisory only).
type EvidenceAssistor interface {
	ReviewEvidence(ctx context.Context, imageURL string) (EvidenceReview, error)
}

// Narrator produces advisory executive prose for reports.
type Narrator interface {
	Summarize(ctx context.Context, figures map[string]any) (string, error)
}

func (f Fixture) ReviewEvidence(_ context.Context, imageURL string) (EvidenceReview, error) {
	u := strings.ToLower(imageURL)
	switch {
	case u == "" || strings.Contains(u, "missing") || strings.Contains(u, "none"):
		return EvidenceReview{LooksValid: false, Confidence: 0.2, Notes: "No proof image URL provided."}, nil
	case strings.Contains(u, "blur") || strings.Contains(u, "unclear"):
		return EvidenceReview{LooksValid: false, Confidence: 0.45, Notes: "Image appears unclear — request a sharper photo."}, nil
	default:
		return EvidenceReview{LooksValid: true, Confidence: 0.88, Notes: "Proof appears consistent with a sustainability activity (fixture AI)."}, nil
	}
}

func (f Fixture) Summarize(_ context.Context, _ map[string]any) (string, error) {
	return "Executive summary (AI-generated, advisory): Overall ESG performance is stable. Environmental progress is driven by emission tracking against goals; Social engagement shows participation and training momentum; Governance remains strong where policy acknowledgements and open-issue counts are healthy. All numeric scores above are system-calculated — this narrative does not replace them.", nil
}

func (a *OpenRouter) ReviewEvidence(ctx context.Context, imageURL string) (EvidenceReview, error) {
	if a.apiKey == "" {
		return Fixture{}.ReviewEvidence(ctx, imageURL)
	}
	schema := map[string]any{
		"name": "evidence_review", "strict": true,
		"schema": map[string]any{
			"type": "object", "additionalProperties": false,
			"required": []string{"looksValid", "confidence", "notes"},
			"properties": map[string]any{
				"looksValid": map[string]any{"type": "boolean"},
				"confidence": map[string]any{"type": "number", "minimum": 0, "maximum": 1},
				"notes":      map[string]any{"type": "string"},
			},
		},
	}
	prompt := "Review this sustainability participation proof image. Decide if it looks like valid evidence of the claimed activity. Do not approve or reject on behalf of a human — advisory only. Image URL: " + imageURL
	content, err := a.chatJSON(ctx, prompt, schema)
	if err != nil {
		return Fixture{}.ReviewEvidence(ctx, imageURL)
	}
	var out EvidenceReview
	if err := jsonUnmarshal(content, &out); err != nil {
		return Fixture{}.ReviewEvidence(ctx, imageURL)
	}
	if out.Confidence < 0 {
		out.Confidence = 0
	}
	if out.Confidence > 1 {
		out.Confidence = 1
	}
	return out, nil
}

func (a *OpenRouter) Summarize(ctx context.Context, figures map[string]any) (string, error) {
	if a.apiKey == "" {
		return Fixture{}.Summarize(ctx, figures)
	}
	raw, _ := jsonMarshal(figures)
	schema := map[string]any{
		"name": "esg_narrative", "strict": true,
		"schema": map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"summary"},
			"properties":           map[string]any{"summary": map[string]any{"type": "string"}},
		},
	}
	prompt := "Write a short executive ESG narrative (3-5 sentences) from these deterministic figures. Do not invent numbers. State that numbers are system-calculated. JSON figures: " + string(raw)
	content, err := a.chatJSON(ctx, prompt, schema)
	if err != nil {
		return Fixture{}.Summarize(ctx, figures)
	}
	var out struct {
		Summary string `json:"summary"`
	}
	if err := jsonUnmarshal(content, &out); err != nil || strings.TrimSpace(out.Summary) == "" {
		return Fixture{}.Summarize(ctx, figures)
	}
	return out.Summary, nil
}

// Gateway selects fixture or live OpenRouter for evidence + narrative.
type Gateway struct {
	live       *OpenRouter
	useFixture bool
}

func NewGateway(apiKey, model string, useFixture bool) *Gateway {
	return &Gateway{live: NewOpenRouter(apiKey, model), useFixture: useFixture || apiKey == ""}
}

func (g *Gateway) Live() *OpenRouter { return g.live }
func (g *Gateway) UseFixture() bool  { return g.useFixture }

func (g *Gateway) ReviewEvidence(ctx context.Context, imageURL string) (EvidenceReview, error) {
	if g.useFixture {
		return Fixture{}.ReviewEvidence(ctx, imageURL)
	}
	return g.live.ReviewEvidence(ctx, imageURL)
}

func (g *Gateway) Summarize(ctx context.Context, figures map[string]any) (string, error) {
	if g.useFixture {
		return Fixture{}.Summarize(ctx, figures)
	}
	return g.live.Summarize(ctx, figures)
}
