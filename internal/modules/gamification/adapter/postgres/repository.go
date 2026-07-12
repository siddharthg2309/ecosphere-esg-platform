package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	badge "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/badge/domain"
	challenge "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/challenge/domain"
	participation "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/participation/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/port"
	reward "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/reward/domain"
	category "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/category/domain"
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
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return errs.Conflict("duplicate_participation", "Already joined this challenge")
		case "23503":
			return errs.Invalid("invalid_reference", "Referenced record does not exist", nil)
		}
	}
	return platformdb.MapError(err)
}

func (r *Repository) CreateChallenge(ctx context.Context, c *challenge.Challenge) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO challenges(id,title,category_id,description,xp,difficulty,evidence_required,deadline,status,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		c.ID, c.Title, c.CategoryID, c.Description, c.XP, c.Difficulty, c.EvidenceRequired, c.Deadline, c.Status, c.CreatedAt, c.UpdatedAt)
	return mapWrite(err)
}

func (r *Repository) ChallengeByID(ctx context.Context, cid id.ID) (*challenge.Challenge, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id,title,category_id,description,xp,difficulty,evidence_required,deadline,status,created_at,updated_at,
		       COALESCE((SELECT COUNT(*) FROM challenge_participations cp WHERE cp.challenge_id=challenges.id AND cp.approval='pending'),0)
		FROM challenges WHERE id=$1`, cid)
	var c challenge.Challenge
	var deadline *time.Time
	err := row.Scan(&c.ID, &c.Title, &c.CategoryID, &c.Description, &c.XP, &c.Difficulty, &c.EvidenceRequired, &deadline, &c.Status, &c.CreatedAt, &c.UpdatedAt, &c.PendingCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("challenge_not_found", "Challenge not found")
	}
	if err != nil {
		return nil, err
	}
	c.Deadline = deadline
	return &c, nil
}

func (r *Repository) SaveChallenge(ctx context.Context, c *challenge.Challenge) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE challenges SET title=$2,description=$3,xp=$4,difficulty=$5,evidence_required=$6,deadline=$7,status=$8,updated_at=$9 WHERE id=$1`,
		c.ID, c.Title, c.Description, c.XP, c.Difficulty, c.EvidenceRequired, c.Deadline, c.Status, c.UpdatedAt)
	return err
}

func (r *Repository) ListChallenges(ctx context.Context, p page.Page) (page.Result[challenge.Challenge], error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,title,category_id,description,xp,difficulty,evidence_required,deadline,status,created_at,updated_at,
		       COALESCE((SELECT COUNT(*) FROM challenge_participations cp WHERE cp.challenge_id=challenges.id AND cp.approval='pending'),0)
		FROM challenges ORDER BY created_at DESC LIMIT $1 OFFSET $2`, p.Limit, p.Offset)
	if err != nil {
		return page.Result[challenge.Challenge]{}, err
	}
	defer rows.Close()
	items := []challenge.Challenge{}
	for rows.Next() {
		var c challenge.Challenge
		var deadline *time.Time
		if err = rows.Scan(&c.ID, &c.Title, &c.CategoryID, &c.Description, &c.XP, &c.Difficulty, &c.EvidenceRequired, &deadline, &c.Status, &c.CreatedAt, &c.UpdatedAt, &c.PendingCount); err != nil {
			return page.Result[challenge.Challenge]{}, err
		}
		c.Deadline = deadline
		items = append(items, c)
	}
	var total int
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM challenges`).Scan(&total)
	return page.Result[challenge.Challenge]{Items: items, Total: total}, nil
}

func (r *Repository) StatusCounts(ctx context.Context) (map[challenge.ChallengeStatus]int, error) {
	rows, err := r.pool.Query(ctx, `SELECT status, COUNT(*) FROM challenges GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[challenge.ChallengeStatus]int{}
	for rows.Next() {
		var s challenge.ChallengeStatus
		var n int
		if err = rows.Scan(&s, &n); err != nil {
			return nil, err
		}
		out[s] = n
	}
	return out, nil
}

func (r *Repository) CreateParticipation(ctx context.Context, p *participation.ChallengeParticipation) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO challenge_participations(id,challenge_id,employee_id,progress,proof_url,approval,xp_awarded,created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
		p.ID, p.ChallengeID, p.EmployeeID, p.Progress, p.ProofURL, p.Approval, p.XPAwarded, p.CreatedAt)
	return mapWrite(err)
}

func (r *Repository) ParticipationByID(ctx context.Context, pid id.ID) (*participation.ChallengeParticipation, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT p.id,p.challenge_id,p.employee_id,p.progress,p.proof_url,p.approval,p.xp_awarded,p.created_at,
		       u.name,c.title,c.xp,c.evidence_required
		FROM challenge_participations p
		JOIN users u ON u.id=p.employee_id
		JOIN challenges c ON c.id=p.challenge_id
		WHERE p.id=$1`, pid)
	var out participation.ChallengeParticipation
	err := row.Scan(&out.ID, &out.ChallengeID, &out.EmployeeID, &out.Progress, &out.ProofURL, &out.Approval, &out.XPAwarded, &out.CreatedAt,
		&out.EmployeeName, &out.ChallengeTitle, &out.ChallengeXP, &out.EvidenceRequired)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("participation_not_found", "Participation not found")
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *Repository) SaveParticipation(ctx context.Context, p *participation.ChallengeParticipation) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE challenge_participations SET progress=$2,proof_url=$3,approval=$4,xp_awarded=$5 WHERE id=$1`,
		p.ID, p.Progress, p.ProofURL, p.Approval, p.XPAwarded)
	return err
}

func (r *Repository) ListParticipations(ctx context.Context, p page.Page) (page.Result[participation.ChallengeParticipation], error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id,p.challenge_id,p.employee_id,p.progress,p.proof_url,p.approval,p.xp_awarded,p.created_at,
		       u.name,c.title,c.xp,c.evidence_required
		FROM challenge_participations p
		JOIN users u ON u.id=p.employee_id
		JOIN challenges c ON c.id=p.challenge_id
		ORDER BY p.created_at DESC LIMIT $1 OFFSET $2`, p.Limit, p.Offset)
	if err != nil {
		return page.Result[participation.ChallengeParticipation]{}, err
	}
	defer rows.Close()
	items := []participation.ChallengeParticipation{}
	for rows.Next() {
		var out participation.ChallengeParticipation
		if err = rows.Scan(&out.ID, &out.ChallengeID, &out.EmployeeID, &out.Progress, &out.ProofURL, &out.Approval, &out.XPAwarded, &out.CreatedAt,
			&out.EmployeeName, &out.ChallengeTitle, &out.ChallengeXP, &out.EvidenceRequired); err != nil {
			return page.Result[participation.ChallengeParticipation]{}, err
		}
		items = append(items, out)
	}
	var total int
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM challenge_participations`).Scan(&total)
	return page.Result[participation.ChallengeParticipation]{Items: items, Total: total}, nil
}

func (r *Repository) RewardByID(ctx context.Context, rid id.ID) (*reward.Reward, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,name,description,points_required,stock,status,created_at,updated_at FROM rewards WHERE id=$1`, rid)
	var out reward.Reward
	var status string
	err := row.Scan(&out.ID, &out.Name, &out.Description, &out.PointsRequired, &out.Stock, &status, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("reward_not_found", "Reward not found")
	}
	if err != nil {
		return nil, err
	}
	out.Status = category.Status(status)
	return &out, nil
}

func (r *Repository) ListRewards(ctx context.Context) ([]reward.Reward, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,description,points_required,stock,status,created_at,updated_at FROM rewards ORDER BY points_required`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []reward.Reward{}
	for rows.Next() {
		var out reward.Reward
		var status string
		if err = rows.Scan(&out.ID, &out.Name, &out.Description, &out.PointsRequired, &out.Stock, &status, &out.CreatedAt, &out.UpdatedAt); err != nil {
			return nil, err
		}
		out.Status = category.Status(status)
		items = append(items, out)
	}
	return items, nil
}

// RedeemAtomic performs stock/points deduction under row locks.
func (r *Repository) RedeemAtomic(ctx context.Context, rewardID, employeeID id.ID) (*reward.Reward, int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var rew reward.Reward
	var status string
	err = tx.QueryRow(ctx, `
		SELECT id,name,description,points_required,stock,status,created_at,updated_at
		FROM rewards WHERE id=$1 FOR UPDATE`, rewardID).Scan(
		&rew.ID, &rew.Name, &rew.Description, &rew.PointsRequired, &rew.Stock, &status, &rew.CreatedAt, &rew.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, 0, errs.NotFound("reward_not_found", "Reward not found")
	}
	if err != nil {
		return nil, 0, err
	}
	rew.Status = category.Status(status)

	var points int
	err = tx.QueryRow(ctx, `SELECT points FROM users WHERE id=$1 FOR UPDATE`, employeeID).Scan(&points)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, 0, errs.NotFound("user_not_found", "User not found")
	}
	if err != nil {
		return nil, 0, err
	}
	bal := &reward.Balance{UserID: employeeID, Points: points}
	if err = rew.Redeem(bal); err != nil {
		return nil, 0, err
	}
	tag, err := tx.Exec(ctx, `UPDATE rewards SET stock = stock - 1, updated_at=now() WHERE id=$1 AND stock > 0`, rewardID)
	if err != nil {
		return nil, 0, err
	}
	if tag.RowsAffected() == 0 {
		return nil, 0, errs.Conflict("out_of_stock", "reward unavailable")
	}
	tag, err = tx.Exec(ctx, `UPDATE users SET points = points - $2 WHERE id=$1 AND points >= $2`, employeeID, rew.PointsRequired)
	if err != nil {
		return nil, 0, err
	}
	if tag.RowsAffected() == 0 {
		return nil, 0, errs.Invalid("insufficient_points", "not enough points", nil)
	}
	redemptionID := id.New()
	_, err = tx.Exec(ctx, `INSERT INTO reward_redemptions(id,employee_id,reward_id,points_spent) VALUES($1,$2,$3,$4)`,
		redemptionID, employeeID, rewardID, rew.PointsRequired)
	if err != nil {
		return nil, 0, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, 0, err
	}
	return &rew, rew.PointsRequired, nil
}

func (r *Repository) AllBadges(ctx context.Context) ([]badge.Badge, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,description,icon,unlock_rule,created_at,updated_at FROM badges ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []badge.Badge{}
	for rows.Next() {
		var b badge.Badge
		var rule []byte
		if err = rows.Scan(&b.ID, &b.Name, &b.Description, &b.Icon, &rule, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(rule, &b.UnlockRule)
		items = append(items, b)
	}
	return items, nil
}

func (r *Repository) OwnsBadge(ctx context.Context, employeeID, badgeID id.ID) (bool, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM employee_badges WHERE employee_id=$1 AND badge_id=$2`, employeeID, badgeID).Scan(&n)
	return n > 0, err
}

func (r *Repository) AwardBadge(ctx context.Context, employeeID, badgeID id.ID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO employee_badges(id,employee_id,badge_id) VALUES($1,$2,$3)
		ON CONFLICT (employee_id, badge_id) DO NOTHING`, id.New(), employeeID, badgeID)
	return err
}

func (r *Repository) EarnerCounts(ctx context.Context) (map[id.ID]int, error) {
	rows, err := r.pool.Query(ctx, `SELECT badge_id, COUNT(*) FROM employee_badges GROUP BY badge_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[id.ID]int{}
	for rows.Next() {
		var bid id.ID
		var n int
		if err = rows.Scan(&bid, &n); err != nil {
			return nil, err
		}
		out[bid] = n
	}
	return out, nil
}

func (r *Repository) GetUser(ctx context.Context, userID id.ID) (*port.UserBalance, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,name,xp,points,completed_challenges,department_id FROM users WHERE id=$1`, userID)
	var u port.UserBalance
	var dept *id.ID
	err := row.Scan(&u.ID, &u.Name, &u.XP, &u.Points, &u.CompletedChallenges, &dept)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.NotFound("user_not_found", "User not found")
	}
	if err != nil {
		return nil, err
	}
	u.DepartmentID = dept
	return &u, nil
}

func (r *Repository) AddXP(ctx context.Context, userID id.ID, xp int) error {
	tag, err := r.pool.Exec(ctx, `UPDATE users SET xp = xp + $2 WHERE id=$1`, userID, xp)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("user_not_found", "User not found")
	}
	return nil
}

func (r *Repository) IncCompleted(ctx context.Context, userID id.ID) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET completed_challenges = completed_challenges + 1 WHERE id=$1`, userID)
	return err
}

func (r *Repository) LeaderboardEmployees(ctx context.Context, limit int) ([]port.LeaderboardEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.name, u.xp,
		       COALESCE((SELECT COUNT(*) FROM employee_badges eb WHERE eb.employee_id=u.id),0) AS badges,
		       RANK() OVER (ORDER BY u.xp DESC) AS rk
		FROM users u
		WHERE u.role='employee' AND u.status='active'
		ORDER BY u.xp DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []port.LeaderboardEntry{}
	for rows.Next() {
		var e port.LeaderboardEntry
		if err = rows.Scan(&e.ID, &e.Name, &e.XP, &e.BadgeCount, &e.Rank); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, nil
}

func (r *Repository) LeaderboardDepartments(ctx context.Context, limit int) ([]port.LeaderboardEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT d.id, d.name, COALESCE(SUM(u.xp),0) AS xp,
		       COALESCE((SELECT COUNT(*) FROM employee_badges eb JOIN users uu ON uu.id=eb.employee_id WHERE uu.department_id=d.id),0) AS badges,
		       RANK() OVER (ORDER BY COALESCE(SUM(u.xp),0) DESC) AS rk
		FROM departments d
		LEFT JOIN users u ON u.department_id=d.id AND u.status='active'
		WHERE d.status='active'
		GROUP BY d.id, d.name
		ORDER BY xp DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []port.LeaderboardEntry{}
	for rows.Next() {
		var e port.LeaderboardEntry
		if err = rows.Scan(&e.ID, &e.Name, &e.XP, &e.BadgeCount, &e.Rank); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, nil
}

// Interface wrappers

type ChallengeRepo struct{ *Repository }

func (c ChallengeRepo) Create(ctx context.Context, ch *challenge.Challenge) error {
	return c.Repository.CreateChallenge(ctx, ch)
}
func (c ChallengeRepo) ByID(ctx context.Context, id id.ID) (*challenge.Challenge, error) {
	return c.Repository.ChallengeByID(ctx, id)
}
func (c ChallengeRepo) Save(ctx context.Context, ch *challenge.Challenge) error {
	return c.Repository.SaveChallenge(ctx, ch)
}
func (c ChallengeRepo) List(ctx context.Context, p page.Page) (page.Result[challenge.Challenge], error) {
	return c.Repository.ListChallenges(ctx, p)
}
func (c ChallengeRepo) StatusCounts(ctx context.Context) (map[challenge.ChallengeStatus]int, error) {
	return c.Repository.StatusCounts(ctx)
}

type ChallengeParticipationRepo struct{ *Repository }

func (c ChallengeParticipationRepo) Create(ctx context.Context, p *participation.ChallengeParticipation) error {
	return c.Repository.CreateParticipation(ctx, p)
}
func (c ChallengeParticipationRepo) ByID(ctx context.Context, id id.ID) (*participation.ChallengeParticipation, error) {
	return c.Repository.ParticipationByID(ctx, id)
}
func (c ChallengeParticipationRepo) Save(ctx context.Context, p *participation.ChallengeParticipation) error {
	return c.Repository.SaveParticipation(ctx, p)
}
func (c ChallengeParticipationRepo) List(ctx context.Context, p page.Page) (page.Result[participation.ChallengeParticipation], error) {
	return c.Repository.ListParticipations(ctx, p)
}

type RewardRepo struct{ *Repository }

func (r RewardRepo) ByID(ctx context.Context, id id.ID) (*reward.Reward, error) {
	return r.Repository.RewardByID(ctx, id)
}
func (r RewardRepo) List(ctx context.Context) ([]reward.Reward, error) {
	return r.Repository.ListRewards(ctx)
}
func (r RewardRepo) RedeemAtomic(ctx context.Context, rewardID, employeeID id.ID) (*reward.Reward, int, error) {
	return r.Repository.RedeemAtomic(ctx, rewardID, employeeID)
}

type BadgeRepo struct{ *Repository }

func (b BadgeRepo) All(ctx context.Context) ([]badge.Badge, error) { return b.Repository.AllBadges(ctx) }
func (b BadgeRepo) Owns(ctx context.Context, employeeID, badgeID id.ID) (bool, error) {
	return b.Repository.OwnsBadge(ctx, employeeID, badgeID)
}
func (b BadgeRepo) Award(ctx context.Context, employeeID, badgeID id.ID) error {
	return b.Repository.AwardBadge(ctx, employeeID, badgeID)
}
func (b BadgeRepo) EarnerCounts(ctx context.Context) (map[id.ID]int, error) {
	return b.Repository.EarnerCounts(ctx)
}

type UserBalanceRepo struct{ *Repository }

func (u UserBalanceRepo) Get(ctx context.Context, userID id.ID) (*port.UserBalance, error) {
	return u.Repository.GetUser(ctx, userID)
}
func (u UserBalanceRepo) AddXP(ctx context.Context, userID id.ID, xp int) error {
	return u.Repository.AddXP(ctx, userID, xp)
}
func (u UserBalanceRepo) IncCompleted(ctx context.Context, userID id.ID) error {
	return u.Repository.IncCompleted(ctx, userID)
}

type LeaderboardRepo struct{ *Repository }

func (l LeaderboardRepo) Employees(ctx context.Context, limit int) ([]port.LeaderboardEntry, error) {
	return l.Repository.LeaderboardEmployees(ctx, limit)
}
func (l LeaderboardRepo) Departments(ctx context.Context, limit int) ([]port.LeaderboardEntry, error) {
	return l.Repository.LeaderboardDepartments(ctx, limit)
}
