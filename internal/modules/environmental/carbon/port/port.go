package port

import (
	"context"
	"io"
	"time"

	"github.com/shopspring/decimal"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/carbon/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Filter struct {
	DepartmentID *id.ID
	From, To     *time.Time
	Source       *domain.Source
	Status       *domain.Status
	Page         page.Page
}

type Summary struct {
	Total    decimal.Decimal                   `json:"total"`
	BySource map[domain.Source]decimal.Decimal `json:"bySource"`
}

type Repository interface {
	Create(context.Context, *domain.Transaction) error
	ByID(context.Context, id.ID) (*domain.Transaction, error)
	SaveVerified(context.Context, *domain.Transaction) error
	List(context.Context, Filter) (page.Result[domain.Transaction], error)
	Summary(context.Context, *id.ID, time.Time, time.Time) (Summary, error)
	Factor(context.Context, id.ID) (unit string, value decimal.Decimal, active bool, err error)
	DepartmentExists(context.Context, id.ID) (bool, error)
	IsDepartmentHead(context.Context, id.ID, id.ID) (bool, error)
}

type Flags interface {
	IsEnabled(context.Context, string) bool
}

type DocInput struct{ FileURL, MimeType, Hint string }
type Suggestion struct {
	Source      string  `json:"source"`
	CategoryID  *string `json:"categoryId"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	Confidence  float64 `json:"confidence"`
	EvidenceURL string  `json:"evidenceUrl"`
}
type AIGateway interface {
	Categorize(context.Context, DocInput) (Suggestion, error)
}
type Storage interface {
	Put(context.Context, string, io.Reader, string, int64) (string, error)
	SignedURL(context.Context, string, time.Duration) (string, error)
}
