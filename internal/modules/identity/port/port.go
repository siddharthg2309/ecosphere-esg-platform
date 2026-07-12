package port

import (
	"context"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type RefreshToken struct {
	ID        id.ID
	UserID    id.ID
	TokenHash string
	ExpiresAt time.Time
}

type Repository interface {
	Create(context.Context, *domain.User) error
	ByEmail(context.Context, string) (*domain.User, error)
	ByID(context.Context, id.ID) (*domain.User, error)
	SaveRefreshToken(context.Context, RefreshToken) error
	ActiveRefreshToken(context.Context, string) (RefreshToken, error)
	RevokeRefreshToken(context.Context, id.ID) error
}
