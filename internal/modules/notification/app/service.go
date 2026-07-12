package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/notification/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/notification/port"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Templates interface {
	Render(t domain.NotifType, data map[string]any) (subject, html string, err error)
}

type Service struct {
	store port.Store
	prefs port.Prefs
	users port.UserEmail
	mail  port.Mailer
	tmpl  Templates
	now   func() time.Time
}

func New(store port.Store, prefs port.Prefs, users port.UserEmail, mail port.Mailer, tmpl Templates) *Service {
	return &Service{
		store: store, prefs: prefs, users: users, mail: mail, tmpl: tmpl,
		now: func() time.Time { return time.Now().UTC() },
	}
}

// Wire registers event subscribers for the notification catalog.
func (s *Service) Wire(bus events.Bus) {
	bus.Subscribe(events.NameComplianceIssueRaised, s.onComplianceRaised)
	bus.Subscribe(events.NameParticipationDecided, s.onParticipationDecided)
	bus.Subscribe(events.NameBadgeUnlocked, s.onBadgeUnlocked)
	bus.Subscribe(events.NameComplianceOverdue, s.onComplianceOverdue)
	bus.Subscribe(events.NamePolicyPublished, s.onPolicyPublished)
}

func (s *Service) deliver(ctx context.Context, userID id.ID, t domain.NotifType, title string, data map[string]any) error {
	channels := s.prefs.Channels(ctx, t)
	for _, ch := range channels {
		switch ch {
		case domain.ChannelInApp:
			n, err := domain.New(userID, t, title, data, s.now())
			if err != nil {
				return err
			}
			if err = s.store.Create(ctx, n); err != nil {
				return err
			}
		case domain.ChannelEmail:
			if s.mail == nil || s.tmpl == nil {
				continue
			}
			name, email, err := s.users.EmailByID(ctx, userID)
			if err != nil || email == "" {
				slog.Warn("skip email notification", "user", userID, "error", err)
				continue
			}
			data["userName"] = name
			subject, html, err := s.tmpl.Render(t, data)
			if err != nil {
				slog.Warn("template render failed", "type", t, "error", err)
				continue
			}
			if err = s.mail.SendHTML(ctx, email, subject, html); err != nil {
				slog.Warn("email send failed", "to", email, "error", err)
			}
		}
	}
	return nil
}

func (s *Service) onComplianceRaised(ctx context.Context, e events.Event) error {
	ev, ok := e.(events.ComplianceIssueRaised)
	if !ok {
		return nil
	}
	return s.deliver(ctx, ev.OwnerID, domain.TypeComplianceRaised,
		fmt.Sprintf("New compliance issue (%s)", ev.Severity),
		map[string]any{"issueId": ev.IssueID, "departmentId": ev.DepartmentID, "severity": ev.Severity})
}

func (s *Service) onParticipationDecided(ctx context.Context, e events.Event) error {
	ev, ok := e.(events.ParticipationDecided)
	if !ok {
		return nil
	}
	title := "Participation rejected"
	if ev.Approved {
		if ev.Kind == "challenge" {
			title = fmt.Sprintf("Challenge approved · +%d XP", ev.XP)
		} else {
			title = fmt.Sprintf("CSR participation approved · +%d pts", ev.Points)
		}
	}
	return s.deliver(ctx, ev.EmployeeID, domain.TypeApprovalDecision, title,
		map[string]any{"kind": ev.Kind, "approved": ev.Approved, "points": ev.Points, "xp": ev.XP})
}

func (s *Service) onBadgeUnlocked(ctx context.Context, e events.Event) error {
	ev, ok := e.(events.BadgeUnlocked)
	if !ok {
		return nil
	}
	return s.deliver(ctx, ev.EmployeeID, domain.TypeBadgeUnlocked, "Badge unlocked",
		map[string]any{"badgeId": ev.BadgeID})
}

func (s *Service) onComplianceOverdue(ctx context.Context, e events.Event) error {
	ev, ok := e.(events.ComplianceOverdue)
	if !ok {
		return nil
	}
	return s.deliver(ctx, ev.OwnerID, domain.TypeComplianceOverdue, "Compliance issue overdue",
		map[string]any{"issueId": ev.IssueID})
}

func (s *Service) onPolicyPublished(ctx context.Context, e events.Event) error {
	// Policy reminders are primarily scheduler-driven; event reserved for future fan-out.
	_ = e
	return nil
}

func (s *Service) List(ctx context.Context, userID id.ID, p page.Page) (page.Result[domain.Notification], error) {
	return s.store.ListForUser(ctx, userID, p)
}

func (s *Service) MarkRead(ctx context.Context, notificationID, userID id.ID) error {
	return s.store.MarkRead(ctx, notificationID, userID)
}

func (s *Service) UnreadCount(ctx context.Context, userID id.ID) (int, error) {
	return s.store.UnreadCount(ctx, userID)
}

// DeliverPolicyReminder is used by the scheduler.
func (s *Service) DeliverPolicyReminder(ctx context.Context, userID id.ID, policyTitle string, policyID id.ID, version int) error {
	return s.deliver(ctx, userID, domain.TypePolicyReminder,
		fmt.Sprintf("Please acknowledge: %s", policyTitle),
		map[string]any{"policyId": policyID, "version": version, "title": policyTitle})
}
