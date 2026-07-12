package port

import (
	"context"

	badge "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/badge/domain"
	challenge "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/challenge/domain"
	participation "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/participation/domain"
	reward "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/reward/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type ChallengeRepo interface {
	Create(ctx context.Context, c *challenge.Challenge) error
	ByID(ctx context.Context, id id.ID) (*challenge.Challenge, error)
	Save(ctx context.Context, c *challenge.Challenge) error
	List(ctx context.Context, p page.Page) (page.Result[challenge.Challenge], error)
	StatusCounts(ctx context.Context) (map[challenge.ChallengeStatus]int, error)
}

type ChallengeParticipationRepo interface {
	Create(ctx context.Context, p *participation.ChallengeParticipation) error
	ByID(ctx context.Context, id id.ID) (*participation.ChallengeParticipation, error)
	Save(ctx context.Context, p *participation.ChallengeParticipation) error
	List(ctx context.Context, p page.Page) (page.Result[participation.ChallengeParticipation], error)
}

type RewardRepo interface {
	ByID(ctx context.Context, id id.ID) (*reward.Reward, error)
	List(ctx context.Context) ([]reward.Reward, error)
	// RedeemAtomic locks rows, applies domain Redeem, inserts redemption record.
	RedeemAtomic(ctx context.Context, rewardID, employeeID id.ID) (*reward.Reward, int, error)
}

type BadgeRepo interface {
	All(ctx context.Context) ([]badge.Badge, error)
	Owns(ctx context.Context, employeeID, badgeID id.ID) (bool, error)
	Award(ctx context.Context, employeeID, badgeID id.ID) error
	EarnerCounts(ctx context.Context) (map[id.ID]int, error)
}

type UserBalance struct {
	ID                  id.ID
	XP                  int
	Points              int
	CompletedChallenges int
	Name                string
	DepartmentID        *id.ID
}

type UserBalanceRepo interface {
	Get(ctx context.Context, userID id.ID) (*UserBalance, error)
	AddXP(ctx context.Context, userID id.ID, xp int) error
	IncCompleted(ctx context.Context, userID id.ID) error
}

type LeaderboardEntry struct {
	Rank         int    `json:"rank"`
	ID           id.ID  `json:"id"`
	Name         string `json:"name"`
	XP           int    `json:"xp"`
	BadgeCount   int    `json:"badgeCount"`
	DepartmentID *id.ID `json:"departmentId,omitempty"`
}

type LeaderboardRepo interface {
	Employees(ctx context.Context, limit int) ([]LeaderboardEntry, error)
	Departments(ctx context.Context, limit int) ([]LeaderboardEntry, error)
}

type Flags interface {
	IsEnabled(ctx context.Context, key string) bool
}
