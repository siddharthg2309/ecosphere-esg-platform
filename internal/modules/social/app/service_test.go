package app

import (
	"context"
	"testing"
	"time"

	activity "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/activity/domain"
	participation "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/participation/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/port"
	training "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/training/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type memActivities struct{ items map[id.ID]*activity.CSRActivity }

func (m *memActivities) Create(_ context.Context, a *activity.CSRActivity) error {
	if m.items == nil {
		m.items = map[id.ID]*activity.CSRActivity{}
	}
	m.items[a.ID] = a
	return nil
}
func (m *memActivities) ByID(_ context.Context, activityID id.ID) (*activity.CSRActivity, error) {
	a, ok := m.items[activityID]
	if !ok {
		return nil, errs.NotFound("not_found", "not found")
	}
	cp := *a
	return &cp, nil
}
func (m *memActivities) List(context.Context, page.Page) (page.Result[activity.CSRActivity], error) {
	return page.Result[activity.CSRActivity]{}, nil
}

type memParts struct{ items map[id.ID]*participation.EmployeeParticipation }

func (m *memParts) Create(_ context.Context, p *participation.EmployeeParticipation) error {
	if m.items == nil {
		m.items = map[id.ID]*participation.EmployeeParticipation{}
	}
	m.items[p.ID] = p
	return nil
}
func (m *memParts) ByID(_ context.Context, pid id.ID) (*participation.EmployeeParticipation, error) {
	p, ok := m.items[pid]
	if !ok {
		return nil, errs.NotFound("not_found", "not found")
	}
	cp := *p
	return &cp, nil
}
func (m *memParts) Save(_ context.Context, p *participation.EmployeeParticipation) error {
	m.items[p.ID] = p
	return nil
}
func (m *memParts) List(context.Context, page.Page, string) (page.Result[participation.EmployeeParticipation], error) {
	return page.Result[participation.EmployeeParticipation]{}, nil
}
func (m *memParts) ListByEmployee(context.Context, id.ID) ([]participation.EmployeeParticipation, error) {
	return nil, nil
}

type memUsers struct{ points map[id.ID]int }

func (m *memUsers) AddPoints(_ context.Context, userID id.ID, points int) error {
	if m.points == nil {
		m.points = map[id.ID]int{}
	}
	m.points[userID] += points
	return nil
}

type memFlags struct{ require bool }

func (m memFlags) IsEnabled(_ context.Context, key string) bool {
	return key == "require_csr_evidence" && m.require
}

type nopTrainings struct{}

func (nopTrainings) Create(context.Context, *training.Training) error { return nil }
func (nopTrainings) List(context.Context) ([]training.Training, error) {
	return nil, nil
}
func (nopTrainings) Complete(context.Context, *training.Completion) error { return nil }

type nopDiversity struct{}

func (nopDiversity) Metrics(context.Context) (port.DiversityMetrics, error) {
	return port.DiversityMetrics{}, nil
}

func TestApproveBlocksWithoutProofWhenFlagOn(t *testing.T) {
	acts := &memActivities{}
	parts := &memParts{}
	users := &memUsers{}
	bus := events.NewInProcess()
	svc := New(acts, parts, nopTrainings{}, users, memFlags{require: true}, nopDiversity{}, bus)
	svc.now = func() time.Time { return time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC) }

	act, err := activity.New("Tree Plantation", id.New(), "plant", 50, false, nil, svc.now())
	if err != nil {
		t.Fatal(err)
	}
	_ = acts.Create(context.Background(), act)
	p, _ := participation.New(id.New(), act.ID, "", "", svc.now())
	_ = parts.Create(context.Background(), p)

	_, err = svc.ApproveParticipation(context.Background(), p.ID, id.New())
	if err == nil {
		t.Fatal("expected evidence_required")
	}
	if e, ok := errs.As(err); !ok || e.Code != "evidence_required" {
		t.Fatalf("%v", err)
	}
}

func TestApproveAwardsPointsNotXP(t *testing.T) {
	acts := &memActivities{}
	parts := &memParts{}
	users := &memUsers{}
	bus := events.NewInProcess()
	var decided events.ParticipationDecided
	bus.Subscribe(events.NameParticipationDecided, func(_ context.Context, e events.Event) error {
		decided = e.(events.ParticipationDecided)
		return nil
	})
	svc := New(acts, parts, nopTrainings{}, users, memFlags{require: false}, nopDiversity{}, bus)
	svc.now = func() time.Time { return time.Now().UTC() }

	act, _ := activity.New("Workshop", id.New(), "d", 30, false, nil, svc.now())
	_ = acts.Create(context.Background(), act)
	emp := id.New()
	p, _ := participation.New(emp, act.ID, "cert.pdf", "", svc.now())
	_ = parts.Create(context.Background(), p)

	out, err := svc.ApproveParticipation(context.Background(), p.ID, id.New())
	if err != nil {
		t.Fatal(err)
	}
	if out.PointsEarned != 30 {
		t.Fatalf("points=%d", out.PointsEarned)
	}
	if users.points[emp] != 30 {
		t.Fatalf("user points=%d", users.points[emp])
	}
	if decided.Kind != "csr" || decided.Points != 30 || decided.XP != 0 {
		t.Fatalf("%+v", decided)
	}
}
