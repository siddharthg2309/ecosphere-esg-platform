package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// SendHTML sends a simple multipart-friendly HTML email via SMTP (MailHog in dev).
func (s *Sender) SendHTML(ctx context.Context, to, subject, html string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if s.Address == "" {
		return nil
	}
	from := s.From
	if from == "" {
		from = "noreply@ecosphere.local"
	}
	msg := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		html,
	}, "\r\n")
	return smtp.SendMail(s.Address, nil, from, []string{to}, []byte(msg))
}

// Templates renders notification emails with design-token styling.
type Templates struct{}

func NewTemplates() *Templates { return &Templates{} }

func (t *Templates) Render(notifType string, data map[string]any) (subject, html string, err error) {
	title := fmt.Sprint(data["title"])
	switch notifType {
	case "compliance_raised":
		subject = "EcoSphere · New compliance issue assigned to you"
		html = wrap("New compliance issue", fmt.Sprintf(
			"<p>A <strong>%s</strong> severity compliance issue has been raised and assigned to you.</p><p>Please review it in EcoSphere and take action before the due date.</p>",
			esc(fmt.Sprint(data["severity"]))))
	case "approval_decision":
		subject = "EcoSphere · Participation decision"
		status := "rejected"
		if data["approved"] == true {
			status = "approved"
		}
		html = wrap("Participation "+status, fmt.Sprintf(
			"<p>Your <strong>%s</strong> participation was <strong>%s</strong>.</p>",
			esc(fmt.Sprint(data["kind"])), status))
	case "policy_reminder":
		subject = "EcoSphere · Policy acknowledgement required"
		html = wrap("Policy acknowledgement", fmt.Sprintf(
			"<p>Please acknowledge <strong>%s</strong> to stay compliant.</p>",
			esc(title)))
	case "badge_unlock":
		subject = "EcoSphere · Badge unlocked"
		html = wrap("Badge unlocked", "<p>Congratulations — you unlocked a new sustainability badge.</p>")
	case "compliance_overdue":
		subject = "EcoSphere · Compliance issue overdue"
		html = wrap("Overdue compliance issue", "<p>An open compliance issue assigned to you is now <strong>overdue</strong>. Please resolve it as soon as possible.</p>")
	default:
		subject = "EcoSphere notification"
		html = wrap("Notification", "<p>You have a new notification in EcoSphere.</p>")
	}
	return subject, html, nil
}

func wrap(heading, body string) string {
	return fmt.Sprintf(`<!DOCTYPE html><html><body style="margin:0;padding:0;background:#F8F9FA;font-family:Inter,Segoe UI,sans-serif;color:#212529">
<table role="presentation" width="100%%" cellspacing="0" cellpadding="0"><tr><td align="center" style="padding:24px">
<table role="presentation" width="600" style="max-width:600px;background:#FFFFFF;border:1px solid #E9ECEF;border-radius:10px">
<tr><td style="background:#714B67;color:#fff;padding:16px 24px;font-weight:700;font-size:18px">EcoSphere</td></tr>
<tr><td style="padding:24px"><h1 style="margin:0 0 12px;font-size:20px">%s</h1>%s
<p style="color:#6C757D;font-size:13px;margin-top:24px">This is an automated message from EcoSphere ESG.</p></td></tr>
</table></td></tr></table></body></html>`, heading, body)
}

func esc(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}
