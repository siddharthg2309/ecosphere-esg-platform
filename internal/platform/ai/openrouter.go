package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	carbonport "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/carbon/port"
)

type OpenRouter struct {
	apiKey, model string
	client        *http.Client
}

func NewOpenRouter(apiKey, model string) *OpenRouter {
	return &OpenRouter{apiKey: apiKey, model: model, client: &http.Client{Timeout: 25 * time.Second}}
}

func (a *OpenRouter) chatJSON(ctx context.Context, userPrompt string, schema map[string]any) (string, error) {
	if a.apiKey == "" {
		return "", errors.New("OpenRouter API key is not configured")
	}
	body := map[string]any{
		"model": a.model, "max_tokens": 500,
		"response_format": map[string]any{"type": "json_schema", "json_schema": schema},
		"messages": []map[string]any{
			{"role": "system", "content": "You are an ESG assistant for EcoSphere. Be precise. Never invent numeric ESG scores or CO2 values. Outputs are advisory only."},
			{"role": "user", "content": userPrompt},
		},
	}
	raw, _ := json.Marshal(body)
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer "+a.apiKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := a.client.Do(req)
		if err != nil {
			last = err
		} else {
			var out struct {
				Choices []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				} `json:"choices"`
			}
			decodeErr := json.NewDecoder(resp.Body).Decode(&out)
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 && decodeErr == nil && len(out.Choices) > 0 {
				return out.Choices[0].Message.Content, nil
			}
			last = fmt.Errorf("OpenRouter status %d", resp.StatusCode)
		}
		if attempt < 2 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 250 * time.Millisecond):
			}
		}
	}
	return "", last
}

func (a *OpenRouter) Categorize(ctx context.Context, in carbonport.DocInput) (carbonport.Suggestion, error) {
	if a.apiKey == "" {
		return carbonport.Suggestion{}, errors.New("OpenRouter API key is not configured")
	}
	schema := map[string]any{"name": "environmental_suggestion", "strict": true, "schema": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"source", "categoryId", "quantity", "unit", "confidence"}, "properties": map[string]any{"source": map[string]any{"type": "string", "enum": []string{"purchase", "manufacturing", "expense", "fleet"}}, "categoryId": map[string]any{"type": []string{"string", "null"}}, "quantity": map[string]any{"type": "number", "minimum": 0}, "unit": map[string]any{"type": "string"}, "confidence": map[string]any{"type": "number", "minimum": 0, "maximum": 1}}}}
	prompt := fmt.Sprintf("Categorize the operational document. Extract only values visible in the document. Never calculate carbon emissions.\nDocument URL: %s\nMIME: %s\nHint: %s", in.FileURL, in.MimeType, in.Hint)
	content, err := a.chatJSON(ctx, prompt, schema)
	if err != nil {
		return carbonport.Suggestion{}, err
	}
	var suggestion carbonport.Suggestion
	if err = json.Unmarshal([]byte(content), &suggestion); err != nil || !valid(suggestion) {
		return carbonport.Suggestion{}, errors.New("OpenRouter returned an invalid suggestion")
	}
	return suggestion, nil
}

func valid(s carbonport.Suggestion) bool {
	return (s.Source == "purchase" || s.Source == "manufacturing" || s.Source == "expense" || s.Source == "fleet") && s.Quantity >= 0 && strings.TrimSpace(s.Unit) != "" && s.Confidence >= 0 && s.Confidence <= 1
}

type Fixture struct{}

func (Fixture) Categorize(_ context.Context, in carbonport.DocInput) (carbonport.Suggestion, error) {
	name := strings.ToLower(filepath.Base(in.Hint))
	if name == "" {
		name = strings.ToLower(in.FileURL)
	}
	switch {
	case strings.Contains(name, "fuel") || strings.Contains(name, "diesel") || strings.Contains(name, "fleet"):
		return carbonport.Suggestion{Source: "fleet", Quantity: 268, Unit: "litre", Confidence: .96}, nil
	case strings.Contains(name, "electric") || strings.Contains(name, "kwh"):
		return carbonport.Suggestion{Source: "purchase", Quantity: 1200, Unit: "kWh", Confidence: .93}, nil
	default:
		return carbonport.Suggestion{Source: "expense", Quantity: 1, Unit: "unit", Confidence: .55}, nil
	}
}

func jsonUnmarshal(s string, v any) error { return json.Unmarshal([]byte(s), v) }
func jsonMarshal(v any) ([]byte, error)  { return json.Marshal(v) }
