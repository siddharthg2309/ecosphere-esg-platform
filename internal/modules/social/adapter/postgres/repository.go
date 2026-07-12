package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	activity "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/activity/domain"
	participation "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/participation/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/port"
	training "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/training/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func mapWrite(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return errs.Conflict("duplicate_participation", "Already joined this activity")
		}
		if pgErr.Code == "23503" {
			return errs.Invalid("invalid_reference", "Referenced record does not exist", nil)
		}
	}
	return err
}

func (r *Repository) Create(ctx context.Context, a *activity.CSRActivity) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO csr_activities(id,title,category_id,description,points,evidence_required,status,activity_date,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		a.ID, a.Title, a.CategoryID, a.Description, a.Points, a.EvidenceRequired, a.Status, a.ActivityDate, a.CreatedAt, a.UpdatedAt)
	return mapWrite(err)
}

func (r *Repository) ByID(ctx context.Context, activityID id.ID) (*activity.CSRActivity, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT a.id,a.title,a.category_id,a.description,a.points,a.evidence_required,a.status,a.activity_date,a.created_at,a.updated_at,
		       COALESCE((SELECT COUNT(*) FROM employee_participations ep WHERE ep.activity_id=a.id),0)
		FROM csr_activities a WHERE a.id=$1`, activityID)
	var a activity.CSRActivity
	var activityDate *time.Time
	err := row.Scan(&a.ID, &a.Title, &a.CategoryID, &a.Description, &a.Points, &a.EvidenceRequired, &a.Status, &activityDate, &a.CreatedAt, &a.UpdatedAt, &a.JoinedCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("csr_activity_not_found", "CSR activity not found")
	}
	if err != nil {
		return nil, err
	}
	a.ActivityDate = activityDate
	return &a, nil
}

func (r *Repository) List(ctx context.Context, p page.Page) (page.Result[activity.CSRActivity], error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id,a.title,a.category_id,a.description,a.points,a.evidence_required,a.status,a.activity_date,a.created_at,a.updated_at,
		       COALESCE((SELECT COUNT(*) FROM employee_participations ep WHERE ep.activity_id=a.id),0)
		FROM csr_activities a ORDER BY a.created_at DESC LIMIT $1 OFFSET $2`, p.Limit, p.Offset)
	if err != nil {
		return page.Result[activity.CSRActivity]{}, err
	}
	defer rows.Close()
	items := []activity.CSRActivity{}
	for rows.Next() {
		var a activity.CSRActivity
		var activityDate *time.Time
		if err = rows.Scan(&a.ID, &a.Title, &a.CategoryID, &a.Description, &a.Points, &a.EvidenceRequired, &a.Status, &activityDate, &a.CreatedAt, &a.UpdatedAt, &a.JoinedCount); err != nil {
			return page.Result[activity.CSRActivity]{}, err
		}
		a.ActivityDate = activityDate
		items = append(items, a)
	}
	var total int
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM csr_activities`).Scan(&total)
	return page.Result[activity.CSRActivity]{Items: items, Total: total}, nil
}

func (r *Repository) CreateParticipation(ctx context.Context, p *participation.EmployeeParticipation) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO employee_participations(id,employee_id,activity_id,proof_url,notes,approval,points_earned,completion_date,created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.EmployeeID, p.ActivityID, p.ProofURL, p.Notes, p.Approval, p.PointsEarned, p.CompletionDate, p.CreatedAt)
	return mapWrite(err)
}

func (r *Repository) ParticipationByID(ctx context.Context, pid id.ID) (*participation.EmployeeParticipation, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT p.id,p.employee_id,p.activity_id,p.proof_url,p.notes,p.approval,p.points_earned,p.completion_date,p.created_at,
		       u.name,a.title,a.points,a.evidence_required
		FROM employee_participations p
		JOIN users u ON u.id=p.employee_id
		JOIN csr_activities a ON a.id=p.activity_id
		WHERE p.id=$1`, pid)
	var out participation.EmployeeParticipation
	var completion *time.Time
	err := row.Scan(&out.ID, &out.EmployeeID, &out.ActivityID, &out.ProofURL, &out.Notes, &out.Approval, &out.PointsEarned, &completion, &out.CreatedAt,
		&out.EmployeeName, &out.ActivityTitle, &out.ActivityPoints, &out.EvidenceRequired)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("participation_not_found", "Participation not found")
	}
	if err != nil {
		return nil, err
	}
	out.CompletionDate = completion
	return &out, nil
}

func (r *Repository) SaveParticipation(ctx context.Context, p *participation.EmployeeParticipation) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE employee_participations SET proof_url=$2,notes=$3,approval=$4,points_earned=$5,completion_date=$6 WHERE id=$1`,
		p.ID, p.ProofURL, p.Notes, p.Approval, p.PointsEarned, p.CompletionDate)
	return err
}

func (r *Repository) ListParticipations(ctx context.Context, p page.Page, approval string) (page.Result[participation.EmployeeParticipation], error) {
	q := `
		SELECT p.id,p.employee_id,p.activity_id,p.proof_url,p.notes,p.approval,p.points_earned,p.completion_date,p.created_at,
		       u.name,a.title,a.points,a.evidence_required
		FROM employee_participations p
		JOIN users u ON u.id=p.employee_id
		JOIN csr_activities a ON a.id=p.activity_id`
	args := []any{}
	if approval != "" {
		q += ` WHERE p.approval=$1`
		args = append(args, approval)
		q += ` ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, p.Limit, p.Offset)
	} else {
		q += ` ORDER BY p.created_at DESC LIMIT $1 OFFSET $2`
		args = append(args, p.Limit, p.Offset)
	}
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return page.Result[participation.EmployeeParticipation]{}, err
	}
	defer rows.Close()
	items := []participation.EmployeeParticipation{}
	for rows.Next() {
		var out participation.EmployeeParticipation
		var completion *time.Time
		if err = rows.Scan(&out.ID, &out.EmployeeID, &out.ActivityID, &out.ProofURL, &out.Notes, &out.Approval, &out.PointsEarned, &completion, &out.CreatedAt,
			&out.EmployeeName, &out.ActivityTitle, &out.ActivityPoints, &out.EvidenceRequired); err != nil {
			return page.Result[participation.EmployeeParticipation]{}, err
		}
		out.CompletionDate = completion
		items = append(items, out)
	}
	var total int
	if approval != "" {
		_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM employee_participations WHERE approval=$1`, approval).Scan(&total)
	} else {
		_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM employee_participations`).Scan(&total)
	}
	return page.Result[participation.EmployeeParticipation]{Items: items, Total: total}, nil
}

func (r *Repository) ListByEmployee(ctx context.Context, employeeID id.ID) ([]participation.EmployeeParticipation, error) {
	res, err := r.ListParticipations(ctx, page.New(100, 0), "")
	if err != nil {
		return nil, err
	}
	out := []participation.EmployeeParticipation{}
	for _, item := range res.Items {
		if item.EmployeeID == employeeID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *Repository) AddPoints(ctx context.Context, userID id.ID, points int) error {
	tag, err := r.pool.Exec(ctx, `UPDATE users SET points = points + $2 WHERE id=$1`, userID, points)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("user_not_found", "User not found")
	}
	return nil
}

func (r *Repository) CreateTraining(ctx context.Context, t *training.Training) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO trainings(id,name,assigned_to,status,created_at) VALUES($1,$2,$3,$4,$5)`,
		t.ID, t.Name, t.AssignedTo, t.Status, t.CreatedAt)
	return err
}

func (r *Repository) ListTrainings(ctx context.Context) ([]training.Training, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.id,t.name,t.assigned_to,t.status,t.created_at,
		       COALESCE((SELECT COUNT(*) FROM training_completions c WHERE c.training_id=t.id),0),
		       COALESCE((SELECT COUNT(*) FROM users u WHERE u.status='active'),0)
		FROM trainings t ORDER BY t.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []training.Training{}
	for rows.Next() {
		var t training.Training
		if err = rows.Scan(&t.ID, &t.Name, &t.AssignedTo, &t.Status, &t.CreatedAt, &t.Completed, &t.Total); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, nil
}

func (r *Repository) CompleteTraining(ctx context.Context, c *training.Completion) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO training_completions(id,employee_id,training_id,completed_at) VALUES($1,$2,$3,$4)
		ON CONFLICT (employee_id, training_id) DO NOTHING`, c.ID, c.EmployeeID, c.TrainingID, c.CompletedAt)
	return mapWrite(err)
}

func (r *Repository) Metrics(ctx context.Context) (port.DiversityMetrics, error) {
	var total, women, men, nonBinary, leaders, womenLeaders int
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE status='active'`).Scan(&total)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE status='active' AND gender='woman'`).Scan(&women)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE status='active' AND gender='man'`).Scan(&men)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE status='active' AND gender='non_binary'`).Scan(&nonBinary)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE status='active' AND is_leadership`).Scan(&leaders)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE status='active' AND is_leadership AND gender='woman'`).Scan(&womenLeaders)

	var csrJoined, employees, trainings, completions int
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT employee_id) FROM employee_participations`).Scan(&csrJoined)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role='employee' AND status='active'`).Scan(&employees)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM trainings WHERE status='active'`).Scan(&trainings)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM training_completions`).Scan(&completions)

	pct := func(n, d int) float64 {
		if d == 0 {
			return 0
		}
		return float64(n) * 100 / float64(d)
	}
	m := port.DiversityMetrics{
		GenderWomenPct:        pct(women, total),
		GenderMenPct:          pct(men, total),
		GenderNonBinaryPct:    pct(nonBinary, total),
		LeadershipWomenPct:    pct(womenLeaders, leaders),
		DiverseLeadersPct:     pct(womenLeaders+0, leaders),
		LeadershipTargetPct:   50,
		CSRParticipationPct:   pct(csrJoined, employees),
		TrainingCompletionPct: 0,
	}
	if trainings > 0 && employees > 0 {
		m.TrainingCompletionPct = pct(completions, trainings*employees)
	}
	return m, nil
}

// Adapters implementing port interfaces via thin wrappers.

type ActivityRepo struct{ *Repository }

func (a ActivityRepo) Create(ctx context.Context, act *activity.CSRActivity) error {
	return a.Repository.Create(ctx, act)
}
func (a ActivityRepo) ByID(ctx context.Context, id id.ID) (*activity.CSRActivity, error) {
	return a.Repository.ByID(ctx, id)
}
func (a ActivityRepo) List(ctx context.Context, p page.Page) (page.Result[activity.CSRActivity], error) {
	return a.Repository.List(ctx, p)
}

type ParticipationRepo struct{ *Repository }

func (p ParticipationRepo) Create(ctx context.Context, part *participation.EmployeeParticipation) error {
	return p.Repository.CreateParticipation(ctx, part)
}
func (p ParticipationRepo) ByID(ctx context.Context, id id.ID) (*participation.EmployeeParticipation, error) {
	return p.Repository.ParticipationByID(ctx, id)
}
func (p ParticipationRepo) Save(ctx context.Context, part *participation.EmployeeParticipation) error {
	return p.Repository.SaveParticipation(ctx, part)
}
func (p ParticipationRepo) List(ctx context.Context, pg page.Page, approval string) (page.Result[participation.EmployeeParticipation], error) {
	return p.Repository.ListParticipations(ctx, pg, approval)
}
func (p ParticipationRepo) ListByEmployee(ctx context.Context, employeeID id.ID) ([]participation.EmployeeParticipation, error) {
	return p.Repository.ListByEmployee(ctx, employeeID)
}

type TrainingRepo struct{ *Repository }

func (t TrainingRepo) Create(ctx context.Context, tr *training.Training) error {
	return t.Repository.CreateTraining(ctx, tr)
}
func (t TrainingRepo) List(ctx context.Context) ([]training.Training, error) {
	return t.Repository.ListTrainings(ctx)
}
func (t TrainingRepo) Complete(ctx context.Context, c *training.Completion) error {
	return t.Repository.CompleteTraining(ctx, c)
}
