package app

import (
	"context"
	"time"

	ack "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/ack/domain"
	audit "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/audit/domain"
	compliance "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/compliance/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/port"
	policy "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/policy/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Service struct {
	audits   port.AuditRepo
	issues   port.IssueRepo
	acks     port.AckRepo
	policies port.PolicyReader
	bundle   port.BundleRepo
	bus      events.Bus
	now      func() time.Time
}

func New(audits port.AuditRepo, issues port.IssueRepo, acks port.AckRepo, policies port.PolicyReader, bundle port.BundleRepo, bus events.Bus) *Service {
	return &Service{
		audits: audits, issues: issues, acks: acks, policies: policies, bundle: bundle, bus: bus,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) CreateAudit(ctx context.Context, title string, dept, auditor id.ID, date time.Time, findings string, status audit.AuditStatus) (*audit.Audit, error) {
	a, err := audit.New(title, dept, auditor, date, findings, status, s.now())
	if err != nil {
		return nil, err
	}
	if err = s.audits.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) ListAudits(ctx context.Context, p page.Page) (page.Result[audit.Audit], error) {
	return s.audits.List(ctx, p)
}

func (s *Service) GetAudit(ctx context.Context, auditID id.ID) (*audit.Audit, error) {
	return s.audits.ByID(ctx, auditID)
}

type RaiseIssueCmd struct {
	DepartmentID id.ID
	OwnerID      id.ID
	Severity     compliance.Severity
	Description  string
	DueDate      time.Time
	AuditID      *id.ID
}

func (s *Service) RaiseIssue(ctx context.Context, cmd RaiseIssueCmd) (*compliance.ComplianceIssue, error) {
	issue, err := compliance.NewIssue(cmd.DepartmentID, cmd.OwnerID, cmd.Severity, cmd.Description, cmd.DueDate, cmd.AuditID, s.now())
	if err != nil {
		return nil, err
	}
	if err = s.issues.Create(ctx, issue); err != nil {
		return nil, err
	}
	_ = s.bus.Publish(ctx, events.ComplianceIssueRaised{
		IssueID: issue.ID, OwnerID: issue.OwnerID, DepartmentID: issue.DepartmentID, Severity: string(issue.Severity),
	})
	return issue, nil
}

func (s *Service) ListIssues(ctx context.Context, p page.Page, status string, overdue *bool) (page.Result[compliance.ComplianceIssue], error) {
	res, err := s.issues.List(ctx, p, status, overdue, s.now())
	if err != nil {
		return res, err
	}
	for i := range res.Items {
		res.Items[i].Overdue = res.Items[i].IsOverdue(s.now())
	}
	return res, nil
}

func (s *Service) UpdateIssue(ctx context.Context, issueID id.ID, status compliance.IssueStatus) (*compliance.ComplianceIssue, error) {
	issue, err := s.issues.ByID(ctx, issueID)
	if err != nil {
		return nil, err
	}
	if err = issue.UpdateStatus(status); err != nil {
		return nil, err
	}
	if err = s.issues.Save(ctx, issue); err != nil {
		return nil, err
	}
	issue.Overdue = issue.IsOverdue(s.now())
	return issue, nil
}

func (s *Service) GetIssue(ctx context.Context, issueID id.ID) (*compliance.ComplianceIssue, error) {
	issue, err := s.issues.ByID(ctx, issueID)
	if err != nil {
		return nil, err
	}
	issue.Overdue = issue.IsOverdue(s.now())
	return issue, nil
}

func (s *Service) AcknowledgePolicy(ctx context.Context, employeeID, policyID id.ID) (*ack.PolicyAcknowledgement, error) {
	p, err := s.policies.ByID(ctx, policyID)
	if err != nil {
		return nil, err
	}
	a, err := ack.New(employeeID, policyID, p.Version, s.now())
	if err != nil {
		return nil, err
	}
	if err = s.acks.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) Unacknowledged(ctx context.Context, employeeID id.ID) ([]policy.Policy, error) {
	return s.acks.UnacknowledgedPolicies(ctx, employeeID)
}

func (s *Service) ListAcknowledgements(ctx context.Context, p page.Page) (page.Result[ack.PolicyAcknowledgement], error) {
	return s.acks.List(ctx, p)
}

func (s *Service) ListPolicies(ctx context.Context) ([]policy.Policy, error) {
	return s.policies.List(ctx)
}

func (s *Service) PolicyAckRate(ctx context.Context, policyID id.ID, version int) (acked, total int, err error) {
	return s.acks.AckRate(ctx, policyID, version)
}

func (s *Service) Stats(ctx context.Context) (open, overdue, audits int, err error) {
	return s.issues.Stats(ctx, s.now())
}

func (s *Service) DepartmentBundle(ctx context.Context, departmentID id.ID) (port.DepartmentBundle, error) {
	return s.bundle.DepartmentBundle(ctx, departmentID)
}

func (s *Service) OpenPastDue(ctx context.Context) ([]compliance.ComplianceIssue, error) {
	return s.issues.OpenPastDue(ctx, s.now())
}
