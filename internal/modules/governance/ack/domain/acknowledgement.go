package domain

import (
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type PolicyAcknowledgement struct {
	ID             id.ID     `json:"id"`
	EmployeeID     id.ID     `json:"employeeId"`
	PolicyID       id.ID     `json:"policyId"`
	Version        int       `json:"version"`
	AcknowledgedAt time.Time `json:"acknowledgedAt"`
	// joins
	EmployeeName   string `json:"employeeName,omitempty"`
	DepartmentName string `json:"departmentName,omitempty"`
	PolicyTitle    string `json:"policyTitle,omitempty"`
}

func New(employeeID, policyID id.ID, version int, now time.Time) (*PolicyAcknowledgement, error) {
	if employeeID == "" || policyID == "" {
		return nil, errs.Invalid("invalid_acknowledgement", "Employee and policy are required", nil)
	}
	if version <= 0 {
		return nil, errs.Invalid("invalid_version", "Policy version must be positive", nil)
	}
	return &PolicyAcknowledgement{
		ID: id.New(), EmployeeID: employeeID, PolicyID: policyID,
		Version: version, AcknowledgedAt: now.UTC(),
	}, nil
}
