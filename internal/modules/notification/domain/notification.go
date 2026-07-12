package domain

import (
	"encoding/json"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

// NotifType matches Notification Settings event_type values.
type NotifType string

const (
	TypeComplianceRaised  NotifType = "compliance_raised"
	TypeApprovalDecision  NotifType = "approval_decision"
	TypePolicyReminder    NotifType = "policy_reminder"
	TypeBadgeUnlocked     NotifType = "badge_unlock" // settings key is badge_unlock
	TypeComplianceOverdue NotifType = "compliance_overdue"
)

type Channel string

const (
	ChannelInApp Channel = "in_app"
	ChannelEmail Channel = "email"
)

type Notification struct {
	ID        id.ID           `json:"id"`
	UserID    id.ID           `json:"userId"`
	Type      NotifType       `json:"type"`
	Title     string          `json:"title"`
	Payload   json.RawMessage `json:"payload"`
	ReadAt    *time.Time      `json:"readAt,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
}

func New(userID id.ID, t NotifType, title string, payload any, now time.Time) (*Notification, error) {
	if userID == "" {
		return nil, errs.Invalid("invalid_notification", "User is required", nil)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &Notification{
		ID: id.New(), UserID: userID, Type: t, Title: title,
		Payload: raw, CreatedAt: now.UTC(),
	}, nil
}

func (n *Notification) MarkRead(now time.Time) {
	t := now.UTC()
	n.ReadAt = &t
}
