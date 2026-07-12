package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	govapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/app"
	notifapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/notification/app"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

// Runner executes daily overdue + policy reminder jobs with a PG advisory lock.
type Runner struct {
	pool   *pgxpool.Pool
	gov    *govapp.Service
	notif  *notifapp.Service
	bus    events.Bus
	now    func() time.Time
	ticker *time.Ticker
}

func New(pool *pgxpool.Pool, gov *govapp.Service, notif *notifapp.Service, bus events.Bus) *Runner {
	return &Runner{pool: pool, gov: gov, notif: notif, bus: bus, now: func() time.Time { return time.Now().UTC() }}
}

// Start launches a background loop (checks every hour; jobs dedupe via scheduler_runs day).
func (r *Runner) Start(ctx context.Context) {
	r.ticker = time.NewTicker(1 * time.Hour)
	go func() {
		r.tick(ctx)
		for {
			select {
			case <-ctx.Done():
				r.ticker.Stop()
				return
			case <-r.ticker.C:
				r.tick(ctx)
			}
		}
	}()
}

func (r *Runner) tick(ctx context.Context) {
	if !r.tryLock(ctx) {
		return
	}
	defer r.unlock(ctx)
	r.runOverdue(ctx)
	r.runPolicyReminders(ctx)
}

func (r *Runner) tryLock(ctx context.Context) bool {
	var ok bool
	// pg_try_advisory_lock key for ecosphere-scheduler
	err := r.pool.QueryRow(ctx, `SELECT pg_try_advisory_lock(74201904)`).Scan(&ok)
	return err == nil && ok
}

func (r *Runner) unlock(ctx context.Context) {
	_, _ = r.pool.Exec(ctx, `SELECT pg_advisory_unlock(74201904)`)
}

func (r *Runner) runOverdue(ctx context.Context) {
	if !r.shouldRun(ctx, "compliance_overdue") {
		return
	}
	issues, err := r.gov.OpenPastDue(ctx)
	if err != nil {
		slog.Error("scheduler overdue list failed", "error", err)
		return
	}
	for _, issue := range issues {
		_ = r.bus.Publish(ctx, events.ComplianceOverdue{IssueID: issue.ID, OwnerID: issue.OwnerID})
	}
	r.markRun(ctx, "compliance_overdue")
	slog.Info("scheduler compliance overdue", "count", len(issues))
}

func (r *Runner) runPolicyReminders(ctx context.Context) {
	if !r.shouldRun(ctx, "policy_reminder") {
		return
	}
	// Employees with any unacknowledged policy
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, p.id, p.title, p.version
		FROM users u
		CROSS JOIN esg_policies p
		WHERE u.status='active' AND u.role IN ('employee','dept_head')
		  AND NOT EXISTS (
		    SELECT 1 FROM policy_acknowledgements a
		    WHERE a.employee_id=u.id AND a.policy_id=p.id AND a.version=p.version
		  )
		LIMIT 500`)
	if err != nil {
		slog.Error("scheduler policy reminder query failed", "error", err)
		return
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		var userID, policyID id.ID
		var title string
		var version int
		if err = rows.Scan(&userID, &policyID, &title, &version); err != nil {
			continue
		}
		_ = r.notif.DeliverPolicyReminder(ctx, userID, title, policyID, version)
		n++
	}
	r.markRun(ctx, "policy_reminder")
	slog.Info("scheduler policy reminders", "count", n)
}

func (r *Runner) shouldRun(ctx context.Context, job string) bool {
	var last time.Time
	err := r.pool.QueryRow(ctx, `SELECT last_run_at FROM scheduler_runs WHERE job_name=$1`, job).Scan(&last)
	if err == nil && r.now().Sub(last) < 20*time.Hour {
		return false
	}
	return true
}

func (r *Runner) markRun(ctx context.Context, job string) {
	_, _ = r.pool.Exec(ctx, `
		INSERT INTO scheduler_runs(job_name,last_run_at) VALUES($1,now())
		ON CONFLICT (job_name) DO UPDATE SET last_run_at=now()`, job)
}

// RunOnce exposes jobs for tests / manual trigger.
func (r *Runner) RunOnce(ctx context.Context) {
	r.runOverdue(ctx)
	r.runPolicyReminders(ctx)
}
