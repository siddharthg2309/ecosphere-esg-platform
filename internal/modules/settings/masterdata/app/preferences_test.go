package app

import (
	"context"
	"testing"

	config "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/esgconfig/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/port"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
)

type preferenceRepo struct {
	port.Repository
	values []config.NotificationPreference
}

func (r *preferenceRepo) SavePreferences(_ context.Context, values []config.NotificationPreference) error {
	r.values = append([]config.NotificationPreference(nil), values...)
	return nil
}
func (r *preferenceRepo) ListPreferences(context.Context) ([]config.NotificationPreference, error) {
	return r.values, nil
}

func TestSavePreferencesRequiresFrozenCatalog(t *testing.T) {
	values := []config.NotificationPreference{
		{EventType: config.EventComplianceRaised},
		{EventType: config.EventApprovalDecision},
		{EventType: config.EventPolicyReminder},
		{EventType: config.EventBadgeUnlock},
	}
	service := New(&preferenceRepo{}, nil, events.NewInProcess())
	if _, err := service.SavePreferences(context.Background(), values); err == nil {
		t.Fatal("expected missing compliance_overdue preference to fail")
	}
	values = append(values, config.NotificationPreference{EventType: config.EventComplianceOverdue})
	if _, err := service.SavePreferences(context.Background(), values); err != nil {
		t.Fatalf("expected complete preference catalog: %v", err)
	}
}
