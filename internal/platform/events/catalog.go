package events

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

const (
	NameESGConfigChanged      = "settings.esg_config_changed"
	NameEmissionRecorded      = "environmental.emission_recorded"
	NameParticipationDecided  = "engagement.participation_decided"
	NameChallengeCompleted    = "gamification.challenge_completed"
	NameBadgeUnlocked         = "gamification.badge_unlocked"
	NameRewardRedeemed        = "gamification.reward_redeemed"
	NamePolicyPublished       = "governance.policy_published"
	NameComplianceIssueRaised = "governance.compliance_issue_raised"
	NameComplianceOverdue     = "governance.compliance_overdue"
)

type ESGConfigChanged struct {
	ChangedAt time.Time `json:"changedAt"`
}

func (ESGConfigChanged) Name() string { return NameESGConfigChanged }

type EmissionRecorded struct {
	DepartmentID id.ID           `json:"departmentId"`
	Source       string          `json:"source"`
	CO2          decimal.Decimal `json:"co2"`
	At           time.Time       `json:"at"`
}

func (EmissionRecorded) Name() string { return NameEmissionRecorded }

type ParticipationDecided struct {
	Kind       string `json:"kind"`
	EmployeeID id.ID  `json:"employeeId"`
	Approved   bool   `json:"approved"`
	Points     int    `json:"points"`
	XP         int    `json:"xp"`
}

func (ParticipationDecided) Name() string { return NameParticipationDecided }

type ChallengeCompleted struct {
	EmployeeID  id.ID `json:"employeeId"`
	ChallengeID id.ID `json:"challengeId"`
	XP          int   `json:"xp"`
}

func (ChallengeCompleted) Name() string { return NameChallengeCompleted }

type BadgeUnlocked struct {
	EmployeeID id.ID `json:"employeeId"`
	BadgeID    id.ID `json:"badgeId"`
}

func (BadgeUnlocked) Name() string { return NameBadgeUnlocked }

type RewardRedeemed struct {
	EmployeeID id.ID `json:"employeeId"`
	RewardID   id.ID `json:"rewardId"`
	Points     int   `json:"points"`
}

func (RewardRedeemed) Name() string { return NameRewardRedeemed }

type PolicyPublished struct {
	PolicyID id.ID `json:"policyId"`
	Version  int   `json:"version"`
}

func (PolicyPublished) Name() string { return NamePolicyPublished }

type ComplianceIssueRaised struct {
	IssueID      id.ID  `json:"issueId"`
	OwnerID      id.ID  `json:"ownerId"`
	DepartmentID id.ID  `json:"departmentId"`
	Severity     string `json:"severity"`
}

func (ComplianceIssueRaised) Name() string { return NameComplianceIssueRaised }

type ComplianceOverdue struct {
	IssueID id.ID `json:"issueId"`
	OwnerID id.ID `json:"ownerId"`
}

func (ComplianceOverdue) Name() string { return NameComplianceOverdue }
