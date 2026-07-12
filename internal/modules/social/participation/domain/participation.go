package domain

import (
	"strings"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type ApprovalStatus string

const (
	Pending  ApprovalStatus = "pending"
	Approved ApprovalStatus = "approved"
	Rejected ApprovalStatus = "rejected"
)

type EmployeeParticipation struct {
	ID             id.ID          `json:"id"`
	EmployeeID     id.ID          `json:"employeeId"`
	ActivityID     id.ID          `json:"activityId"`
	ProofURL       string         `json:"proofUrl"`
	Notes          string         `json:"notes,omitempty"`
	Approval       ApprovalStatus `json:"approval"`
	PointsEarned   int            `json:"pointsEarned"`
	CompletionDate *time.Time     `json:"completionDate,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	// join fields for queue
	EmployeeName string `json:"employeeName,omitempty"`
	ActivityTitle string `json:"activityTitle,omitempty"`
	ActivityPoints int  `json:"activityPoints,omitempty"`
	EvidenceRequired bool `json:"evidenceRequired,omitempty"`
}

func New(employeeID, activityID id.ID, proofURL, notes string, now time.Time) (*EmployeeParticipation, error) {
	if employeeID == "" || activityID == "" {
		return nil, errs.Invalid("invalid_participation", "Employee and activity are required", nil)
	}
	return &EmployeeParticipation{
		ID:         id.New(),
		EmployeeID: employeeID,
		ActivityID: activityID,
		ProofURL:   strings.TrimSpace(proofURL),
		Notes:      strings.TrimSpace(notes),
		Approval:   Pending,
		CreatedAt:  now.UTC(),
	}, nil
}

// Approve awards points from the CSR activity. requireEvidence blocks empty proof.
func (p *EmployeeParticipation) Approve(pts int, requireEvidence bool, now time.Time) error {
	if p.Approval == Approved {
		return errs.Conflict("already_approved", "Participation is already approved")
	}
	if p.Approval == Rejected {
		return errs.Conflict("already_rejected", "Participation was rejected")
	}
	if requireEvidence && p.ProofURL == "" {
		return errs.Invalid("evidence_required", "proof file required", nil)
	}
	if pts < 0 {
		return errs.Invalid("invalid_points", "Points cannot be negative", nil)
	}
	p.Approval = Approved
	p.PointsEarned = pts
	t := now.UTC()
	p.CompletionDate = &t
	return nil
}

func (p *EmployeeParticipation) Reject() error {
	if p.Approval == Approved {
		return errs.Conflict("already_approved", "Cannot reject an approved participation")
	}
	if p.Approval == Rejected {
		return errs.Conflict("already_rejected", "Participation is already rejected")
	}
	p.Approval = Rejected
	p.PointsEarned = 0
	p.CompletionDate = nil
	return nil
}
