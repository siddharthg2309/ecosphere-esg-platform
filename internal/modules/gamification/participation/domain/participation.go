package domain

import (
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type ApprovalStatus string

const (
	Pending     ApprovalStatus = "pending"
	Approved    ApprovalStatus = "approved"
	Rejected    ApprovalStatus = "rejected"
	InProgress  ApprovalStatus = "in_progress"
)

type ChallengeParticipation struct {
	ID          id.ID          `json:"id"`
	ChallengeID id.ID          `json:"challengeId"`
	EmployeeID  id.ID          `json:"employeeId"`
	Progress    int            `json:"progress"`
	ProofURL    string         `json:"proofUrl"`
	Approval    ApprovalStatus `json:"approval"`
	XPAwarded   int            `json:"xpAwarded"`
	CreatedAt   time.Time      `json:"createdAt"`
	// join fields
	EmployeeName   string `json:"employeeName,omitempty"`
	ChallengeTitle string `json:"challengeTitle,omitempty"`
	ChallengeXP    int    `json:"challengeXp,omitempty"`
	EvidenceRequired bool `json:"evidenceRequired,omitempty"`
}

func New(challengeID, employeeID id.ID, progress int, proofURL string, now time.Time) (*ChallengeParticipation, error) {
	if challengeID == "" || employeeID == "" {
		return nil, errs.Invalid("invalid_participation", "Challenge and employee are required", nil)
	}
	if progress < 0 || progress > 100 {
		return nil, errs.Invalid("invalid_progress", "Progress must be 0-100", nil)
	}
	status := Pending
	if progress < 100 && proofURL == "" {
		status = InProgress
	}
	return &ChallengeParticipation{
		ID:          id.New(),
		ChallengeID: challengeID,
		EmployeeID:  employeeID,
		Progress:    progress,
		ProofURL:    proofURL,
		Approval:    status,
		CreatedAt:   now.UTC(),
	}, nil
}

func (p *ChallengeParticipation) Approve(xp int, requireEvidence bool) error {
	if p.Approval == Approved {
		return errs.Conflict("already_approved", "Participation is already approved")
	}
	if p.Approval == Rejected {
		return errs.Conflict("already_rejected", "Participation was rejected")
	}
	if requireEvidence && p.ProofURL == "" {
		return errs.Invalid("evidence_required", "proof file required", nil)
	}
	if xp < 0 {
		return errs.Invalid("invalid_xp", "XP cannot be negative", nil)
	}
	p.Approval = Approved
	p.XPAwarded = xp
	p.Progress = 100
	return nil
}

func (p *ChallengeParticipation) Reject() error {
	if p.Approval == Approved {
		return errs.Conflict("already_approved", "Cannot reject an approved participation")
	}
	p.Approval = Rejected
	p.XPAwarded = 0
	return nil
}
