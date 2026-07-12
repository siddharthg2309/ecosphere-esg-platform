package app

import (
	"context"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/port"
	platformauth "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/auth"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type Tokens struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	User         *domain.User `json:"user,omitempty"`
}

type Service struct {
	repo                  port.Repository
	secret                []byte
	accessTTL, refreshTTL time.Duration
	now                   func() time.Time
}

func New(repo port.Repository, secret []byte, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{repo: repo, secret: secret, accessTTL: accessTTL, refreshTTL: refreshTTL, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Login(ctx context.Context, email, password string) (Tokens, error) {
	user, err := s.repo.ByEmail(ctx, email)
	if err != nil || !platformauth.CheckPassword(password, user.PasswordHash) {
		return Tokens{}, errs.Unauthorized("invalid_credentials", "Email or password is incorrect")
	}
	return s.issue(ctx, user)
}

func (s *Service) Refresh(ctx context.Context, raw string) (Tokens, error) {
	stored, err := s.repo.ActiveRefreshToken(ctx, platformauth.HashRefreshToken(raw))
	if err != nil {
		return Tokens{}, errs.Unauthorized("invalid_refresh_token", "Refresh token is invalid or expired")
	}
	user, err := s.repo.ByID(ctx, stored.UserID)
	if err != nil {
		return Tokens{}, errs.Unauthorized("invalid_refresh_token", "Refresh token is invalid or expired")
	}
	if err := s.repo.RevokeRefreshToken(ctx, stored.ID); err != nil {
		return Tokens{}, err
	}
	return s.issue(ctx, user)
}

func (s *Service) Me(ctx context.Context, userID id.ID) (*domain.User, error) {
	return s.repo.ByID(ctx, userID)
}

func (s *Service) issue(ctx context.Context, user *domain.User) (Tokens, error) {
	now := s.now()
	access, err := platformauth.IssueAccess(user, s.accessTTL, s.secret, now)
	if err != nil {
		return Tokens{}, err
	}
	raw, hash, err := platformauth.NewRefreshToken()
	if err != nil {
		return Tokens{}, err
	}
	if err := s.repo.SaveRefreshToken(ctx, port.RefreshToken{ID: id.New(), UserID: user.ID, TokenHash: hash, ExpiresAt: now.Add(s.refreshTTL)}); err != nil {
		return Tokens{}, err
	}
	return Tokens{AccessToken: access, RefreshToken: raw, User: user}, nil
}
