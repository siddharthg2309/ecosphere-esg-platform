package port

import (
	"context"
	"time"

	ack "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/ack/domain"
	audit "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/audit/domain"
	compliance "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/compliance/domain"
	policy "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/policy/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type AuditRepo interface {
	Create(ctx context.Context, a *audit.Audit) error
	ByID(ctx context.Context, id id.ID) (*audit.Audit, error)
	List(ctx context.Context, p page.Page) (page.Result[audit.Audit], error)
}

type IssueRepo interface {
	Create(ctx context.Context, i *compliance.ComplianceIssue) error
	ByID(ctx context.Context, id id.ID) (*compliance.ComplianceIssue, error)
	Save(ctx context.Context, i *compliance.ComplianceIssue) error
	List(ctx context.Context, p page.Page, status string, overdue *bool, now time.Time) (page.Result[compliance.ComplianceIssue], error)
	OpenPastDue(ctx context.Context, now time.Time) ([]compliance.ComplianceIssue, error)
	Stats(ctx context.Context, now time.Time) (open, overdue, auditsFY int, err error)
}

type AckRepo interface {
	Create(ctx context.Context, a *ack.PolicyAcknowledgement) error
	List(ctx context.Context, p page.Page) (page.Result[ack.PolicyAcknowledgement], error)
	UnacknowledgedPolicies(ctx context.Context, employeeID id.ID) ([]policy.Policy, error)
	AckRate(ctx context.Context, policyID id.ID, version int) (acked, total int, err error)
	ListByDepartment(ctx context.Context, departmentID id.ID) ([]ack.PolicyAcknowledgement, error)
}

type PolicyReader interface {
	ByID(ctx context.Context, id id.ID) (*policy.Policy, error)
	List(ctx context.Context) ([]policy.Policy, error)
}

type DepartmentBundle struct {
	CarbonTransactions []map[string]any `json:"carbonTransactions"`
	CSRParticipations  []map[string]any `json:"csr"`
	Acknowledgements   []map[string]any `json:"acknowledgements"`
	PriorIssues        []map[string]any `json:"priorIssues"`
	Evidence           []map[string]any `json:"evidence"`
	OperationalRecords []map[string]any `json:"operationalRecords"`
}

type BundleRepo interface {
	DepartmentBundle(ctx context.Context, departmentID id.ID) (DepartmentBundle, error)
}
