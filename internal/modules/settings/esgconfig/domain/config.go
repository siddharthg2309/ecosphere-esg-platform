package domain

import "github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"

type Config struct {
	AutoEmissionCalc      bool `json:"autoEmissionCalc"`
	RequireCSREvidence    bool `json:"requireCsrEvidence"`
	AutoAwardBadges       bool `json:"autoAwardBadges"`
	NotifyComplianceEmail bool `json:"notifyComplianceEmail"`
	WeightEnv             int  `json:"weightEnv"`
	WeightSocial          int  `json:"weightSocial"`
	WeightGov             int  `json:"weightGov"`
}

func (c Config) Validate() error {
	if c.WeightEnv < 0 || c.WeightSocial < 0 || c.WeightGov < 0 || c.WeightEnv+c.WeightSocial+c.WeightGov != 100 {
		return errs.Invalid("weights_sum", "Environmental, social and governance weights must total 100", map[string]string{"weights": "Must total 100"})
	}
	return nil
}

type EventType string

const (
	EventComplianceRaised EventType = "compliance_raised"
	EventApprovalDecision EventType = "approval_decision"
	EventPolicyReminder   EventType = "policy_reminder"
	EventBadgeUnlock      EventType = "badge_unlock"
)

func (e EventType) Valid() bool {
	return e == EventComplianceRaised || e == EventApprovalDecision || e == EventPolicyReminder || e == EventBadgeUnlock
}

type NotificationPreference struct {
	EventType    EventType `json:"eventType"`
	InAppEnabled bool      `json:"inAppEnabled"`
	EmailEnabled bool      `json:"emailEnabled"`
}
