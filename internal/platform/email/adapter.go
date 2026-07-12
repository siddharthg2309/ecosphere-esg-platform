package email

import (
	"context"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/notification/domain"
)

// TemplateAdapter adapts Templates to notification.Templates interface.
type TemplateAdapter struct{ *Templates }

func (t TemplateAdapter) Render(typ domain.NotifType, data map[string]any) (string, string, error) {
	return t.Templates.Render(string(typ), data)
}

// MailAdapter adapts Sender to notification Mailer.
type MailAdapter struct{ *Sender }

func (m MailAdapter) SendHTML(ctx context.Context, to, subject, html string) error {
	return m.Sender.SendHTML(ctx, to, subject, html)
}
