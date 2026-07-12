package app

import (
	"context"
	"time"

	badge "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/badge/domain"
	challenge "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/challenge/domain"
	participation "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/participation/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/port"
	reward "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/reward/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Service struct {
	challenges     port.ChallengeRepo
	participations port.ChallengeParticipationRepo
	rewards        port.RewardRepo
	badges         port.BadgeRepo
	users          port.UserBalanceRepo
	leaderboard    port.LeaderboardRepo
	flags          port.Flags
	bus            events.Bus
	now            func() time.Time
}

func New(
	challenges port.ChallengeRepo,
	participations port.ChallengeParticipationRepo,
	rewards port.RewardRepo,
	badges port.BadgeRepo,
	users port.UserBalanceRepo,
	leaderboard port.LeaderboardRepo,
	flags port.Flags,
	bus events.Bus,
) *Service {
	s := &Service{
		challenges: challenges, participations: participations, rewards: rewards,
		badges: badges, users: users, leaderboard: leaderboard, flags: flags, bus: bus,
		now: func() time.Time { return time.Now().UTC() },
	}
	if bus != nil {
		bus.Subscribe(events.NameChallengeCompleted, s.onChallengeCompleted)
	}
	return s
}

type CreateChallengeCmd struct {
	Title            string
	CategoryID       id.ID
	Description      string
	XP               int
	Difficulty       string
	EvidenceRequired bool
	Deadline         *time.Time
}

func (s *Service) CreateChallenge(ctx context.Context, cmd CreateChallengeCmd) (*challenge.Challenge, error) {
	c, err := challenge.New(cmd.Title, cmd.CategoryID, cmd.Description, cmd.XP, cmd.Difficulty, cmd.EvidenceRequired, cmd.Deadline, s.now())
	if err != nil {
		return nil, err
	}
	if err = s.challenges.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) ListChallenges(ctx context.Context, p page.Page) (page.Result[challenge.Challenge], error) {
	return s.challenges.List(ctx, p)
}

func (s *Service) StatusCounts(ctx context.Context) (map[challenge.ChallengeStatus]int, error) {
	return s.challenges.StatusCounts(ctx)
}

func (s *Service) Transition(ctx context.Context, challengeID id.ID, to challenge.ChallengeStatus) (*challenge.Challenge, error) {
	c, err := s.challenges.ByID(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	if err = c.Transition(to); err != nil {
		return nil, err
	}
	c.UpdatedAt = s.now()
	if err = s.challenges.Save(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) Participate(ctx context.Context, challengeID, employeeID id.ID, progress int, proofURL string) (*participation.ChallengeParticipation, error) {
	if _, err := s.challenges.ByID(ctx, challengeID); err != nil {
		return nil, err
	}
	p, err := participation.New(challengeID, employeeID, progress, proofURL, s.now())
	if err != nil {
		return nil, err
	}
	if err = s.participations.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) ListParticipations(ctx context.Context, p page.Page) (page.Result[participation.ChallengeParticipation], error) {
	return s.participations.List(ctx, p)
}

func (s *Service) ApproveParticipation(ctx context.Context, participationID, by id.ID) (*participation.ChallengeParticipation, error) {
	_ = by
	cp, err := s.participations.ByID(ctx, participationID)
	if err != nil {
		return nil, err
	}
	ch, err := s.challenges.ByID(ctx, cp.ChallengeID)
	if err != nil {
		return nil, err
	}
	if err = cp.Approve(ch.XP, ch.EvidenceRequired); err != nil {
		return nil, err
	}
	if err = s.participations.Save(ctx, cp); err != nil {
		return nil, err
	}
	if err = s.users.AddXP(ctx, cp.EmployeeID, ch.XP); err != nil {
		return nil, err
	}
	if err = s.users.IncCompleted(ctx, cp.EmployeeID); err != nil {
		return nil, err
	}
	_ = s.bus.Publish(ctx,
		events.ParticipationDecided{Kind: "challenge", EmployeeID: cp.EmployeeID, Approved: true, Points: 0, XP: ch.XP},
		events.ChallengeCompleted{EmployeeID: cp.EmployeeID, ChallengeID: ch.ID, XP: ch.XP},
	)
	return cp, nil
}

func (s *Service) RejectParticipation(ctx context.Context, participationID, by id.ID) (*participation.ChallengeParticipation, error) {
	_ = by
	cp, err := s.participations.ByID(ctx, participationID)
	if err != nil {
		return nil, err
	}
	if err = cp.Reject(); err != nil {
		return nil, err
	}
	if err = s.participations.Save(ctx, cp); err != nil {
		return nil, err
	}
	_ = s.bus.Publish(ctx, events.ParticipationDecided{
		Kind: "challenge", EmployeeID: cp.EmployeeID, Approved: false, Points: 0, XP: 0,
	})
	return cp, nil
}

func (s *Service) ListRewards(ctx context.Context) ([]reward.Reward, error) {
	return s.rewards.List(ctx)
}

func (s *Service) Redeem(ctx context.Context, rewardID, employeeID id.ID) (*reward.Reward, int, error) {
	r, spent, err := s.rewards.RedeemAtomic(ctx, rewardID, employeeID)
	if err != nil {
		return nil, 0, err
	}
	_ = s.bus.Publish(ctx, events.RewardRedeemed{EmployeeID: employeeID, RewardID: rewardID, Points: spent})
	return r, spent, nil
}

func (s *Service) ListBadges(ctx context.Context) ([]badge.Badge, map[id.ID]int, error) {
	all, err := s.badges.All(ctx)
	if err != nil {
		return nil, nil, err
	}
	counts, err := s.badges.EarnerCounts(ctx)
	if err != nil {
		return nil, nil, err
	}
	return all, counts, nil
}

func (s *Service) Balance(ctx context.Context, userID id.ID) (*port.UserBalance, error) {
	return s.users.Get(ctx, userID)
}

func (s *Service) Leaderboard(ctx context.Context, scope string, limit int) ([]port.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	if scope == "department" {
		return s.leaderboard.Departments(ctx, limit)
	}
	return s.leaderboard.Employees(ctx, limit)
}

func (s *Service) onChallengeCompleted(ctx context.Context, e events.Event) error {
	if s.flags != nil && !s.flags.IsEnabled(ctx, "auto_award_badges") {
		return nil
	}
	ev, ok := e.(events.ChallengeCompleted)
	if !ok {
		return nil
	}
	emp, err := s.users.Get(ctx, ev.EmployeeID)
	if err != nil {
		return err
	}
	all, err := s.badges.All(ctx)
	if err != nil {
		return err
	}
	for _, b := range all {
		if !b.UnlockRule.Satisfied(emp.XP, emp.CompletedChallenges) {
			continue
		}
		owns, err := s.badges.Owns(ctx, emp.ID, b.ID)
		if err != nil {
			return err
		}
		if owns {
			continue
		}
		if err = s.badges.Award(ctx, emp.ID, b.ID); err != nil {
			return err
		}
		_ = s.bus.Publish(ctx, events.BadgeUnlocked{EmployeeID: emp.ID, BadgeID: b.ID})
	}
	return nil
}
