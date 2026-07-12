package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/notification/domain"
	platformdb "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/db"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Create(ctx context.Context, n *domain.Notification) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO notifications(id,user_id,type,title,payload,read_at,created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7)`,
		n.ID, n.UserID, n.Type, n.Title, []byte(n.Payload), n.ReadAt, n.CreatedAt)
	return platformdb.MapError(err)
}

func (s *Store) ListForUser(ctx context.Context, userID id.ID, p page.Page) (page.Result[domain.Notification], error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id,user_id,type,title,payload,read_at,created_at
		FROM notifications
		WHERE user_id=$1
		ORDER BY read_at NULLS FIRST, created_at DESC
		LIMIT $2 OFFSET $3`, userID, p.Limit, p.Offset)
	if err != nil {
		return page.Result[domain.Notification]{}, platformdb.MapError(err)
	}
	defer rows.Close()
	items := []domain.Notification{}
	for rows.Next() {
		var n domain.Notification
		var payload []byte
		var readAt *time.Time
		if err = rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &payload, &readAt, &n.CreatedAt); err != nil {
			return page.Result[domain.Notification]{}, err
		}
		n.Payload = json.RawMessage(payload)
		n.ReadAt = readAt
		items = append(items, n)
	}
	var total int
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id=$1`, userID).Scan(&total)
	return page.Result[domain.Notification]{Items: items, Total: total}, nil
}

func (s *Store) MarkRead(ctx context.Context, notificationID, userID id.ID) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE notifications SET read_at=now() WHERE id=$1 AND user_id=$2 AND read_at IS NULL`, notificationID, userID)
	if err != nil {
		return platformdb.MapError(err)
	}
	if tag.RowsAffected() == 0 {
		// already read or not found — idempotent success if exists
		var exists bool
		_ = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM notifications WHERE id=$1 AND user_id=$2)`, notificationID, userID).Scan(&exists)
		if !exists {
			return errs.NotFound("notification_not_found", "Notification not found")
		}
	}
	return nil
}

func (s *Store) UnreadCount(ctx context.Context, userID id.ID) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND read_at IS NULL`, userID).Scan(&n)
	return n, platformdb.MapError(err)
}

type Prefs struct{ pool *pgxpool.Pool }

func NewPrefs(pool *pgxpool.Pool) *Prefs { return &Prefs{pool: pool} }

func (p *Prefs) Channels(ctx context.Context, t domain.NotifType) []domain.Channel {
	var inApp, email bool
	err := p.pool.QueryRow(ctx, `
		SELECT in_app_enabled, email_enabled FROM notification_preferences WHERE event_type=$1`, string(t)).Scan(&inApp, &email)
	if err != nil {
		// default: in-app on
		return []domain.Channel{domain.ChannelInApp}
	}
	out := []domain.Channel{}
	if inApp {
		out = append(out, domain.ChannelInApp)
	}
	if email {
		out = append(out, domain.ChannelEmail)
	}
	return out
}

type UserEmail struct{ pool *pgxpool.Pool }

func NewUserEmail(pool *pgxpool.Pool) *UserEmail { return &UserEmail{pool: pool} }

func (u *UserEmail) EmailByID(ctx context.Context, userID id.ID) (string, string, error) {
	var name, email string
	err := u.pool.QueryRow(ctx, `SELECT name, email FROM users WHERE id=$1`, userID).Scan(&name, &email)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", errs.NotFound("user_not_found", "User not found")
	}
	return name, email, platformdb.MapError(err)
}
