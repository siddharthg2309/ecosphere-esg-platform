package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	ack "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/ack/domain"
	audit "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/audit/domain"
	compliance "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/compliance/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/port"
	policy "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/policy/domain"
	platformdb "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/db"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func mapWrite(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return errs.Conflict("already_acknowledged", "Policy already acknowledged for this version")
	}
	return platformdb.MapError(err)
}

func (r *Repository) CreateAudit(ctx context.Context, a *audit.Audit) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audits(id,title,department_id,auditor_id,audit_date,findings,status,created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.Title, a.DepartmentID, a.AuditorID, a.AuditDate, a.Findings, a.Status, a.CreatedAt)
	return mapWrite(err)
}

func (r *Repository) AuditByID(ctx context.Context, auditID id.ID) (*audit.Audit, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT a.id,a.title,a.department_id,a.auditor_id,a.audit_date,a.findings,a.status,a.created_at,
		       d.name,u.name
		FROM audits a
		JOIN departments d ON d.id=a.department_id
		JOIN users u ON u.id=a.auditor_id
		WHERE a.id=$1`, auditID)
	var a audit.Audit
	err := row.Scan(&a.ID, &a.Title, &a.DepartmentID, &a.AuditorID, &a.AuditDate, &a.Findings, &a.Status, &a.CreatedAt, &a.DepartmentName, &a.AuditorName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("audit_not_found", "Audit not found")
	}
	if err != nil {
		return nil, platformdb.MapError(err)
	}
	return &a, nil
}

func (r *Repository) ListAudits(ctx context.Context, p page.Page) (page.Result[audit.Audit], error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id,a.title,a.department_id,a.auditor_id,a.audit_date,a.findings,a.status,a.created_at,
		       d.name,u.name
		FROM audits a
		JOIN departments d ON d.id=a.department_id
		JOIN users u ON u.id=a.auditor_id
		ORDER BY a.audit_date DESC LIMIT $1 OFFSET $2`, p.Limit, p.Offset)
	if err != nil {
		return page.Result[audit.Audit]{}, platformdb.MapError(err)
	}
	defer rows.Close()
	items := []audit.Audit{}
	for rows.Next() {
		var a audit.Audit
		if err = rows.Scan(&a.ID, &a.Title, &a.DepartmentID, &a.AuditorID, &a.AuditDate, &a.Findings, &a.Status, &a.CreatedAt, &a.DepartmentName, &a.AuditorName); err != nil {
			return page.Result[audit.Audit]{}, err
		}
		items = append(items, a)
	}
	var total int
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audits`).Scan(&total)
	return page.Result[audit.Audit]{Items: items, Total: total}, nil
}

func (r *Repository) CreateIssue(ctx context.Context, i *compliance.ComplianceIssue) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO compliance_issues(id,audit_id,department_id,severity,description,owner_id,due_date,status,created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		i.ID, i.AuditID, i.DepartmentID, i.Severity, i.Description, i.OwnerID, i.DueDate, i.Status, i.CreatedAt)
	return mapWrite(err)
}

func (r *Repository) IssueByID(ctx context.Context, issueID id.ID) (*compliance.ComplianceIssue, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT i.id,i.audit_id,i.department_id,i.severity,i.description,i.owner_id,i.due_date,i.status,i.created_at,
		       u.name,d.name,COALESCE(a.title,'')
		FROM compliance_issues i
		JOIN users u ON u.id=i.owner_id
		JOIN departments d ON d.id=i.department_id
		LEFT JOIN audits a ON a.id=i.audit_id
		WHERE i.id=$1`, issueID)
	var i compliance.ComplianceIssue
	err := row.Scan(&i.ID, &i.AuditID, &i.DepartmentID, &i.Severity, &i.Description, &i.OwnerID, &i.DueDate, &i.Status, &i.CreatedAt,
		&i.OwnerName, &i.DepartmentName, &i.AuditTitle)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("issue_not_found", "Compliance issue not found")
	}
	if err != nil {
		return nil, platformdb.MapError(err)
	}
	return &i, nil
}

func (r *Repository) SaveIssue(ctx context.Context, i *compliance.ComplianceIssue) error {
	_, err := r.pool.Exec(ctx, `UPDATE compliance_issues SET status=$2 WHERE id=$1`, i.ID, i.Status)
	return mapWrite(err)
}

func (r *Repository) ListIssues(ctx context.Context, p page.Page, status string, overdue *bool, now time.Time) (page.Result[compliance.ComplianceIssue], error) {
	q := `
		SELECT i.id,i.audit_id,i.department_id,i.severity,i.description,i.owner_id,i.due_date,i.status,i.created_at,
		       u.name,d.name,COALESCE(a.title,'')
		FROM compliance_issues i
		JOIN users u ON u.id=i.owner_id
		JOIN departments d ON d.id=i.department_id
		LEFT JOIN audits a ON a.id=i.audit_id
		WHERE 1=1`
	args := []any{}
	n := 1
	if status != "" {
		q += ` AND i.status=$` + itoa(n)
		args = append(args, status)
		n++
	}
	if overdue != nil && *overdue {
		q += ` AND i.status='open' AND i.due_date < $` + itoa(n) + `::date`
		args = append(args, now.UTC().Truncate(24*time.Hour))
		n++
	}
	q += ` ORDER BY i.due_date ASC LIMIT $` + itoa(n) + ` OFFSET $` + itoa(n+1)
	args = append(args, p.Limit, p.Offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return page.Result[compliance.ComplianceIssue]{}, platformdb.MapError(err)
	}
	defer rows.Close()
	items := []compliance.ComplianceIssue{}
	for rows.Next() {
		var i compliance.ComplianceIssue
		if err = rows.Scan(&i.ID, &i.AuditID, &i.DepartmentID, &i.Severity, &i.Description, &i.OwnerID, &i.DueDate, &i.Status, &i.CreatedAt,
			&i.OwnerName, &i.DepartmentName, &i.AuditTitle); err != nil {
			return page.Result[compliance.ComplianceIssue]{}, err
		}
		items = append(items, i)
	}
	var total int
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM compliance_issues`).Scan(&total)
	return page.Result[compliance.ComplianceIssue]{Items: items, Total: total}, nil
}

func itoa(n int) string {
	const digits = "0123456789"
	if n < 10 {
		return string(digits[n])
	}
	return itoa(n/10) + string(digits[n%10])
}

func (r *Repository) OpenPastDue(ctx context.Context, now time.Time) ([]compliance.ComplianceIssue, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT i.id,i.audit_id,i.department_id,i.severity,i.description,i.owner_id,i.due_date,i.status,i.created_at,
		       u.name,d.name,COALESCE(a.title,'')
		FROM compliance_issues i
		JOIN users u ON u.id=i.owner_id
		JOIN departments d ON d.id=i.department_id
		LEFT JOIN audits a ON a.id=i.audit_id
		WHERE i.status='open' AND i.due_date < $1::date`, now.UTC().Truncate(24*time.Hour))
	if err != nil {
		return nil, platformdb.MapError(err)
	}
	defer rows.Close()
	items := []compliance.ComplianceIssue{}
	for rows.Next() {
		var i compliance.ComplianceIssue
		if err = rows.Scan(&i.ID, &i.AuditID, &i.DepartmentID, &i.Severity, &i.Description, &i.OwnerID, &i.DueDate, &i.Status, &i.CreatedAt,
			&i.OwnerName, &i.DepartmentName, &i.AuditTitle); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (r *Repository) Stats(ctx context.Context, now time.Time) (open, overdue, auditsFY int, err error) {
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM compliance_issues WHERE status IN ('open','in_progress')`).Scan(&open)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM compliance_issues WHERE status='open' AND due_date < $1::date`, now.UTC().Truncate(24*time.Hour)).Scan(&overdue)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audits WHERE EXTRACT(YEAR FROM audit_date)=EXTRACT(YEAR FROM CURRENT_DATE)`).Scan(&auditsFY)
	return open, overdue, auditsFY, nil
}

func (r *Repository) CreateAck(ctx context.Context, a *ack.PolicyAcknowledgement) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO policy_acknowledgements(id,employee_id,policy_id,version,acknowledged_at)
		VALUES($1,$2,$3,$4,$5)`, a.ID, a.EmployeeID, a.PolicyID, a.Version, a.AcknowledgedAt)
	return mapWrite(err)
}

func (r *Repository) ListAcks(ctx context.Context, p page.Page) (page.Result[ack.PolicyAcknowledgement], error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id,a.employee_id,a.policy_id,a.version,a.acknowledged_at,u.name,COALESCE(d.name,''),p.title
		FROM policy_acknowledgements a
		JOIN users u ON u.id=a.employee_id
		LEFT JOIN departments d ON d.id=u.department_id
		JOIN esg_policies p ON p.id=a.policy_id
		ORDER BY a.acknowledged_at DESC LIMIT $1 OFFSET $2`, p.Limit, p.Offset)
	if err != nil {
		return page.Result[ack.PolicyAcknowledgement]{}, platformdb.MapError(err)
	}
	defer rows.Close()
	items := []ack.PolicyAcknowledgement{}
	for rows.Next() {
		var a ack.PolicyAcknowledgement
		if err = rows.Scan(&a.ID, &a.EmployeeID, &a.PolicyID, &a.Version, &a.AcknowledgedAt, &a.EmployeeName, &a.DepartmentName, &a.PolicyTitle); err != nil {
			return page.Result[ack.PolicyAcknowledgement]{}, err
		}
		items = append(items, a)
	}
	var total int
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM policy_acknowledgements`).Scan(&total)
	return page.Result[ack.PolicyAcknowledgement]{Items: items, Total: total}, nil
}

func (r *Repository) UnacknowledgedPolicies(ctx context.Context, employeeID id.ID) ([]policy.Policy, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id,p.title,p.body,p.version,p.effective_date,p.created_at,p.updated_at
		FROM esg_policies p
		WHERE NOT EXISTS (
		  SELECT 1 FROM policy_acknowledgements a
		  WHERE a.employee_id=$1 AND a.policy_id=p.id AND a.version=p.version
		)
		ORDER BY p.effective_date DESC`, employeeID)
	if err != nil {
		return nil, platformdb.MapError(err)
	}
	defer rows.Close()
	items := []policy.Policy{}
	for rows.Next() {
		var p policy.Policy
		if err = rows.Scan(&p.ID, &p.Title, &p.Body, &p.Version, &p.EffectiveDate, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, nil
}

func (r *Repository) AckRate(ctx context.Context, policyID id.ID, version int) (acked, total int, err error) {
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM policy_acknowledgements WHERE policy_id=$1 AND version=$2`, policyID, version).Scan(&acked)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE status='active' AND role IN ('employee','dept_head')`).Scan(&total)
	return acked, total, nil
}

func (r *Repository) ListByDepartment(ctx context.Context, departmentID id.ID) ([]ack.PolicyAcknowledgement, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id,a.employee_id,a.policy_id,a.version,a.acknowledged_at,u.name,COALESCE(d.name,''),p.title
		FROM policy_acknowledgements a
		JOIN users u ON u.id=a.employee_id
		LEFT JOIN departments d ON d.id=u.department_id
		JOIN esg_policies p ON p.id=a.policy_id
		WHERE u.department_id=$1
		ORDER BY a.acknowledged_at DESC`, departmentID)
	if err != nil {
		return nil, platformdb.MapError(err)
	}
	defer rows.Close()
	items := []ack.PolicyAcknowledgement{}
	for rows.Next() {
		var a ack.PolicyAcknowledgement
		if err = rows.Scan(&a.ID, &a.EmployeeID, &a.PolicyID, &a.Version, &a.AcknowledgedAt, &a.EmployeeName, &a.DepartmentName, &a.PolicyTitle); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, nil
}

func (r *Repository) PolicyByID(ctx context.Context, policyID id.ID) (*policy.Policy, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,title,body,version,effective_date,created_at,updated_at FROM esg_policies WHERE id=$1`, policyID)
	var p policy.Policy
	err := row.Scan(&p.ID, &p.Title, &p.Body, &p.Version, &p.EffectiveDate, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("policy_not_found", "Policy not found")
	}
	if err != nil {
		return nil, platformdb.MapError(err)
	}
	return &p, nil
}

func (r *Repository) ListPolicies(ctx context.Context) ([]policy.Policy, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,title,body,version,effective_date,created_at,updated_at FROM esg_policies ORDER BY title`)
	if err != nil {
		return nil, platformdb.MapError(err)
	}
	defer rows.Close()
	items := []policy.Policy{}
	for rows.Next() {
		var p policy.Policy
		if err = rows.Scan(&p.ID, &p.Title, &p.Body, &p.Version, &p.EffectiveDate, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, nil
}

func (r *Repository) DepartmentBundle(ctx context.Context, departmentID id.ID) (port.DepartmentBundle, error) {
	b := port.DepartmentBundle{
		CarbonTransactions: []map[string]any{},
		CSRParticipations:  []map[string]any{},
		Acknowledgements:   []map[string]any{},
		PriorIssues:        []map[string]any{},
		Evidence:           []map[string]any{},
		OperationalRecords: []map[string]any{},
	}
	// Carbon (optional — table may exist from Phase 2)
	if rows, err := r.pool.Query(ctx, `
		SELECT id::text, source, quantity::text, computed_co2::text, status, txn_date::text
		FROM carbon_transactions WHERE department_id=$1 ORDER BY txn_date DESC LIMIT 50`, departmentID); err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, source, qty, co2, status, date string
			if rows.Scan(&id, &source, &qty, &co2, &status, &date) == nil {
				b.CarbonTransactions = append(b.CarbonTransactions, map[string]any{
					"id": id, "source": source, "quantity": qty, "computedCo2": co2, "status": status, "txnDate": date,
				})
			}
		}
	}
	// CSR
	if rows, err := r.pool.Query(ctx, `
		SELECT ep.id::text, u.name, ca.title, ep.approval, ep.proof_url
		FROM employee_participations ep
		JOIN users u ON u.id=ep.employee_id
		JOIN csr_activities ca ON ca.id=ep.activity_id
		WHERE u.department_id=$1
		ORDER BY ep.created_at DESC LIMIT 50`, departmentID); err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, name, title, approval, proof string
			if rows.Scan(&id, &name, &title, &approval, &proof) == nil {
				b.CSRParticipations = append(b.CSRParticipations, map[string]any{
					"id": id, "employee": name, "activity": title, "approval": approval, "proofUrl": proof,
				})
				if proof != "" {
					b.Evidence = append(b.Evidence, map[string]any{"kind": "csr", "url": proof, "label": title})
				}
			}
		}
	}
	// Prior issues
	if rows, err := r.pool.Query(ctx, `
		SELECT i.id::text, i.description, i.severity, i.status, i.due_date::text, u.name
		FROM compliance_issues i JOIN users u ON u.id=i.owner_id
		WHERE i.department_id=$1 ORDER BY i.created_at DESC LIMIT 50`, departmentID); err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, desc, sev, status, due, owner string
			if rows.Scan(&id, &desc, &sev, &status, &due, &owner) == nil {
				b.PriorIssues = append(b.PriorIssues, map[string]any{
					"id": id, "description": desc, "severity": sev, "status": status, "dueDate": due, "owner": owner,
				})
			}
		}
	}
	// Acks for department employees
	if rows, err := r.pool.Query(ctx, `
		SELECT a.id::text, u.name, p.title, a.version, a.acknowledged_at::text
		FROM policy_acknowledgements a
		JOIN users u ON u.id=a.employee_id
		JOIN esg_policies p ON p.id=a.policy_id
		WHERE u.department_id=$1
		ORDER BY a.acknowledged_at DESC LIMIT 50`, departmentID); err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, name, title, at string
			var version int
			if rows.Scan(&id, &name, &title, &version, &at) == nil {
				b.Acknowledgements = append(b.Acknowledgements, map[string]any{
					"id": id, "employee": name, "policy": title, "version": version, "acknowledgedAt": at,
				})
			}
		}
	}
	return b, nil
}

// Wrappers

type AuditRepo struct{ *Repository }

func (a AuditRepo) Create(ctx context.Context, aud *audit.Audit) error {
	return a.Repository.CreateAudit(ctx, aud)
}
func (a AuditRepo) ByID(ctx context.Context, id id.ID) (*audit.Audit, error) {
	return a.Repository.AuditByID(ctx, id)
}
func (a AuditRepo) List(ctx context.Context, p page.Page) (page.Result[audit.Audit], error) {
	return a.Repository.ListAudits(ctx, p)
}

type IssueRepo struct{ *Repository }

func (i IssueRepo) Create(ctx context.Context, issue *compliance.ComplianceIssue) error {
	return i.Repository.CreateIssue(ctx, issue)
}
func (i IssueRepo) ByID(ctx context.Context, id id.ID) (*compliance.ComplianceIssue, error) {
	return i.Repository.IssueByID(ctx, id)
}
func (i IssueRepo) Save(ctx context.Context, issue *compliance.ComplianceIssue) error {
	return i.Repository.SaveIssue(ctx, issue)
}
func (i IssueRepo) List(ctx context.Context, p page.Page, status string, overdue *bool, now time.Time) (page.Result[compliance.ComplianceIssue], error) {
	return i.Repository.ListIssues(ctx, p, status, overdue, now)
}
func (i IssueRepo) OpenPastDue(ctx context.Context, now time.Time) ([]compliance.ComplianceIssue, error) {
	return i.Repository.OpenPastDue(ctx, now)
}
func (i IssueRepo) Stats(ctx context.Context, now time.Time) (int, int, int, error) {
	return i.Repository.Stats(ctx, now)
}

type AckRepo struct{ *Repository }

func (a AckRepo) Create(ctx context.Context, ack *ack.PolicyAcknowledgement) error {
	return a.Repository.CreateAck(ctx, ack)
}
func (a AckRepo) List(ctx context.Context, p page.Page) (page.Result[ack.PolicyAcknowledgement], error) {
	return a.Repository.ListAcks(ctx, p)
}
func (a AckRepo) UnacknowledgedPolicies(ctx context.Context, employeeID id.ID) ([]policy.Policy, error) {
	return a.Repository.UnacknowledgedPolicies(ctx, employeeID)
}
func (a AckRepo) AckRate(ctx context.Context, policyID id.ID, version int) (int, int, error) {
	return a.Repository.AckRate(ctx, policyID, version)
}
func (a AckRepo) ListByDepartment(ctx context.Context, departmentID id.ID) ([]ack.PolicyAcknowledgement, error) {
	return a.Repository.ListByDepartment(ctx, departmentID)
}

type PolicyRepo struct{ *Repository }

func (p PolicyRepo) ByID(ctx context.Context, id id.ID) (*policy.Policy, error) {
	return p.Repository.PolicyByID(ctx, id)
}
func (p PolicyRepo) List(ctx context.Context) ([]policy.Policy, error) {
	return p.Repository.ListPolicies(ctx)
}

type BundleRepo struct{ *Repository }

func (b BundleRepo) DepartmentBundle(ctx context.Context, departmentID id.ID) (port.DepartmentBundle, error) {
	return b.Repository.DepartmentBundle(ctx, departmentID)
}
