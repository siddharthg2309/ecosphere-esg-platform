package app

import (
	"context"
	"sync"
	"testing"
	"time"

	badge "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/badge/domain"
	challenge "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/challenge/domain"
	participation "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/participation/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/port"
	reward "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/reward/domain"
	category "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/category/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type memChallenges struct{ items map[id.ID]*challenge.Challenge }

func (m *memChallenges) Create(_ context.Context, c *challenge.Challenge) error {
	if m.items == nil {
		m.items = map[id.ID]*challenge.Challenge{}
	}
	m.items[c.ID] = c
	return nil
}
func (m *memChallenges) ByID(_ context.Context, cid id.ID) (*challenge.Challenge, error) {
	c, ok := m.items[cid]
	if !ok {
		return nil, errs.NotFound("not_found", "not found")
	}
	cp := *c
	return &cp, nil
}
func (m *memChallenges) Save(_ context.Context, c *challenge.Challenge) error {
	m.items[c.ID] = c
	return nil
}
func (m *memChallenges) List(context.Context, page.Page) (page.Result[challenge.Challenge], error) {
	return page.Result[challenge.Challenge]{}, nil
}
func (m *memChallenges) StatusCounts(context.Context) (map[challenge.ChallengeStatus]int, error) {
	return map[challenge.ChallengeStatus]int{}, nil
}

type memCP struct{ items map[id.ID]*participation.ChallengeParticipation }

func (m *memCP) Create(_ context.Context, p *participation.ChallengeParticipation) error {
	if m.items == nil {
		m.items = map[id.ID]*participation.ChallengeParticipation{}
	}
	m.items[p.ID] = p
	return nil
}
func (m *memCP) ByID(_ context.Context, pid id.ID) (*participation.ChallengeParticipation, error) {
	p, ok := m.items[pid]
	if !ok {
		return nil, errs.NotFound("not_found", "not found")
	}
	cp := *p
	return &cp, nil
}
func (m *memCP) Save(_ context.Context, p *participation.ChallengeParticipation) error {
	m.items[p.ID] = p
	return nil
}
func (m *memCP) List(context.Context, page.Page) (page.Result[participation.ChallengeParticipation], error) {
	return page.Result[participation.ChallengeParticipation]{}, nil
}

type memUsers struct {
	mu    sync.Mutex
	items map[id.ID]*port.UserBalance
}

func (m *memUsers) Get(_ context.Context, userID id.ID) (*port.UserBalance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.items[userID]
	if !ok {
		return nil, errs.NotFound("not_found", "not found")
	}
	cp := *u
	return &cp, nil
}
func (m *memUsers) AddXP(_ context.Context, userID id.ID, xp int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[userID].XP += xp
	return nil
}
func (m *memUsers) IncCompleted(_ context.Context, userID id.ID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[userID].CompletedChallenges++
	return nil
}

type memBadges struct {
	all   []badge.Badge
	owned map[string]bool
}

func (m *memBadges) All(context.Context) ([]badge.Badge, error) { return m.all, nil }
func (m *memBadges) Owns(_ context.Context, employeeID, badgeID id.ID) (bool, error) {
	return m.owned[string(employeeID)+":"+string(badgeID)], nil
}
func (m *memBadges) Award(_ context.Context, employeeID, badgeID id.ID) error {
	if m.owned == nil {
		m.owned = map[string]bool{}
	}
	key := string(employeeID) + ":" + string(badgeID)
	if m.owned[key] {
		return nil
	}
	m.owned[key] = true
	return nil
}
func (m *memBadges) EarnerCounts(context.Context) (map[id.ID]int, error) {
	return map[id.ID]int{}, nil
}

type memRewards struct{}

func (memRewards) ByID(context.Context, id.ID) (*reward.Reward, error) { return nil, nil }
func (memRewards) List(context.Context) ([]reward.Reward, error)       { return nil, nil }
func (memRewards) RedeemAtomic(context.Context, id.ID, id.ID) (*reward.Reward, int, error) {
	return nil, 0, nil
}

type memLB struct{}

func (memLB) Employees(context.Context, int) ([]port.LeaderboardEntry, error) { return nil, nil }
func (memLB) Departments(context.Context, int) ([]port.LeaderboardEntry, error) {
	return nil, nil
}

type autoFlags struct{ on bool }

func (f autoFlags) IsEnabled(_ context.Context, key string) bool {
	return key == "auto_award_badges" && f.on
}

func TestApproveChallengeAwardsXPAndBadge(t *testing.T) {
	chRepo := &memChallenges{}
	cpRepo := &memCP{}
	empID := id.New()
	users := &memUsers{items: map[id.ID]*port.UserBalance{
		empID: {ID: empID, XP: 0, CompletedChallenges: 0},
	}}
	b, _ := badge.New("Green Beginner", "100 XP", "award", badge.UnlockRule{Type: "xp", Value: 100}, time.Now())
	badges := &memBadges{all: []badge.Badge{b}}
	bus := events.NewInProcess()
	var unlocked int
	bus.Subscribe(events.NameBadgeUnlocked, func(context.Context, events.Event) error {
		unlocked++
		return nil
	})
	svc := New(chRepo, cpRepo, memRewards{}, badges, users, memLB{}, autoFlags{on: true}, bus)
	svc.now = func() time.Time { return time.Now().UTC() }

	ch, _ := challenge.New("Commute Green Week", id.New(), "cycle", 120, "medium", true, nil, svc.now())
	_ = ch.Transition(challenge.StatusActive)
	_ = chRepo.Create(context.Background(), ch)
	p, _ := participation.New(ch.ID, empID, 100, "proof.jpg", svc.now())
	_ = cpRepo.Create(context.Background(), p)

	out, err := svc.ApproveParticipation(context.Background(), p.ID, id.New())
	if err != nil {
		t.Fatal(err)
	}
	if out.XPAwarded != 120 {
		t.Fatalf("xp=%d", out.XPAwarded)
	}
	bal, _ := users.Get(context.Background(), empID)
	if bal.XP != 120 || bal.CompletedChallenges != 1 {
		t.Fatalf("%+v", bal)
	}
	if unlocked != 1 {
		t.Fatalf("unlocked=%d", unlocked)
	}
	// idempotent second complete event
	_ = svc.onChallengeCompleted(context.Background(), events.ChallengeCompleted{EmployeeID: empID, ChallengeID: ch.ID, XP: 120})
	if unlocked != 1 {
		t.Fatalf("badge awarded twice: %d", unlocked)
	}
}

func TestAutoAwardOffNoBadge(t *testing.T) {
	chRepo := &memChallenges{}
	cpRepo := &memCP{}
	empID := id.New()
	users := &memUsers{items: map[id.ID]*port.UserBalance{empID: {ID: empID}}}
	b, _ := badge.New("Green Beginner", "", "a", badge.UnlockRule{Type: "xp", Value: 1}, time.Now())
	badges := &memBadges{all: []badge.Badge{b}}
	bus := events.NewInProcess()
	svc := New(chRepo, cpRepo, memRewards{}, badges, users, memLB{}, autoFlags{on: false}, bus)
	ch, _ := challenge.New("C", id.New(), "d", 50, "easy", false, nil, time.Now())
	_ = chRepo.Create(context.Background(), ch)
	p, _ := participation.New(ch.ID, empID, 100, "", time.Now())
	_ = cpRepo.Create(context.Background(), p)
	_, err := svc.ApproveParticipation(context.Background(), p.ID, id.New())
	if err != nil {
		t.Fatal(err)
	}
	if len(badges.owned) != 0 {
		t.Fatal("badge should not award when flag off")
	}
}

func TestRedeemDomainGuards(t *testing.T) {
	r, _ := reward.New("Card", "", 800, 1, category.StatusActive, time.Now())
	emp := &reward.Balance{UserID: id.New(), Points: 100}
	if err := r.Redeem(emp); err == nil {
		t.Fatal("expected insufficient")
	}
}
