package port

import (
	"context"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/notification/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Store interface {
	Create(ctx context.Context, n *domain.Notification) error
	ListForUser(ctx context.Context, userID id.ID, p page.Page) (page.Result[domain.Notification], error)
	MarkRead(ctx context.Context, notificationID, userID id.ID) error
	UnreadCount(ctx context.Context, userID id.ID) (int, error)
}

type Prefs interface {
	Channels(ctx context.Context, t domain.NotifType) []domain.Channel
}

type UserEmail interface {
	EmailByID(ctx context.Context, userID id.ID) (name, email string, err error)
}

type Mailer interface {
	SendHTML(ctx context.Context, to, subject, html string) error
}
