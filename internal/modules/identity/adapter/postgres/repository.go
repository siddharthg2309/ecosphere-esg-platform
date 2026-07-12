package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/port"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/db/sqlc"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type Repository struct{ queries *sqlc.Queries }

func New(queries *sqlc.Queries) *Repository { return &Repository{queries: queries} }

func (r *Repository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{ID: uuid(user.ID), Name: user.Name, Email: user.Email, PasswordHash: user.PasswordHash, Role: string(user.Role), DepartmentID: nullableUUID(user.DepartmentID)})
	return err
}

func (r *Repository) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.queries.UserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("user_not_found", "User not found")
	}
	if err != nil {
		return nil, err
	}
	return mapUser(row)
}

func (r *Repository) ByID(ctx context.Context, userID id.ID) (*domain.User, error) {
	row, err := r.queries.UserByID(ctx, uuid(userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("user_not_found", "User not found")
	}
	if err != nil {
		return nil, err
	}
	return mapUser(row)
}

func (r *Repository) SaveRefreshToken(ctx context.Context, token port.RefreshToken) error {
	return r.queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{ID: uuid(token.ID), UserID: uuid(token.UserID), TokenHash: token.TokenHash, ExpiresAt: pgtype.Timestamptz{Time: token.ExpiresAt, Valid: true}})
}

func (r *Repository) ActiveRefreshToken(ctx context.Context, hash string) (port.RefreshToken, error) {
	row, err := r.queries.ActiveRefreshTokenByHash(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return port.RefreshToken{}, errs.NotFound("refresh_token_not_found", "Refresh token not found")
	}
	if err != nil {
		return port.RefreshToken{}, err
	}
	return port.RefreshToken{ID: fromUUID(row.ID), UserID: fromUUID(row.UserID), TokenHash: row.TokenHash, ExpiresAt: row.ExpiresAt.Time}, nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenID id.ID) error {
	return r.queries.RevokeRefreshToken(ctx, uuid(tokenID))
}

func mapUser(row sqlc.User) (*domain.User, error) {
	user := &domain.User{ID: fromUUID(row.ID), Name: row.Name, Email: row.Email, PasswordHash: row.PasswordHash, Role: domain.Role(row.Role), CreatedAt: row.CreatedAt.Time}
	if row.DepartmentID.Valid {
		value := fromUUID(row.DepartmentID)
		user.DepartmentID = &value
	}
	if user.ID == "" {
		return nil, fmt.Errorf("user has invalid id")
	}
	return user, nil
}

func uuid(value id.ID) pgtype.UUID {
	var target pgtype.UUID
	_ = target.Scan(value.String())
	return target
}
func nullableUUID(value *id.ID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}
	return uuid(*value)
}
func fromUUID(value pgtype.UUID) id.ID {
	if !value.Valid {
		return ""
	}
	return id.ID(fmt.Sprintf("%x-%x-%x-%x-%x", value.Bytes[0:4], value.Bytes[4:6], value.Bytes[6:8], value.Bytes[8:10], value.Bytes[10:16]))
}
