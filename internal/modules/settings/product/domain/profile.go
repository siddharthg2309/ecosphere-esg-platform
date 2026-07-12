package domain

import (
	"encoding/json"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"strings"
	"time"
)

type Profile struct {
	ID               id.ID           `json:"id"`
	Product          string          `json:"product"`
	Attributes       json.RawMessage `json:"attributes"`
	EmissionFactorID *id.ID          `json:"emissionFactorId,omitempty"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

func New(product string, attributes json.RawMessage, factorID *id.ID, now time.Time) (Profile, error) {
	if len(attributes) == 0 {
		attributes = json.RawMessage(`{}`)
	}
	p := Profile{ID: id.New(), Product: strings.TrimSpace(product), Attributes: attributes, EmissionFactorID: factorID, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	if p.Product == "" || !json.Valid(p.Attributes) {
		return Profile{}, errs.Invalid("invalid_product_profile", "Product and valid ESG attributes are required", nil)
	}
	return p, nil
}
