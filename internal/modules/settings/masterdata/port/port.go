package port

import (
	"context"
	factor "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/emissionfactor/domain"
	badge "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/badge/domain"
	reward "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/reward/domain"
	category "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/category/domain"
	config "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/esgconfig/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/domain"
	policy "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/policy/domain"
	product "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/product/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Repository interface {
	CreateCategory(context.Context, category.Category) error
	Category(context.Context, id.ID) (category.Category, error)
	ListCategories(context.Context, *category.Type, page.Page) (page.Result[category.Category], error)
	UpdateCategory(context.Context, category.Category) error
	DeleteCategory(context.Context, id.ID) error
	CreateFactor(context.Context, factor.Factor) error
	Factor(context.Context, id.ID) (factor.Factor, error)
	ListFactors(context.Context, *id.ID, page.Page) (page.Result[factor.Factor], error)
	UpdateFactor(context.Context, factor.Factor) error
	DeleteFactor(context.Context, id.ID) error
	CreateProduct(context.Context, product.Profile) error
	Product(context.Context, id.ID) (product.Profile, error)
	ListProducts(context.Context, page.Page) (page.Result[product.Profile], error)
	UpdateProduct(context.Context, product.Profile) error
	DeleteProduct(context.Context, id.ID) error
	CreatePolicy(context.Context, policy.Policy) error
	Policy(context.Context, id.ID) (policy.Policy, error)
	ListPolicies(context.Context, page.Page) (page.Result[policy.Policy], error)
	UpdatePolicy(context.Context, policy.Policy) error
	DeletePolicy(context.Context, id.ID) error
	CreateBadge(context.Context, badge.Badge) error
	Badge(context.Context, id.ID) (badge.Badge, error)
	ListBadges(context.Context, page.Page) (page.Result[badge.Badge], error)
	UpdateBadge(context.Context, badge.Badge) error
	DeleteBadge(context.Context, id.ID) error
	CreateReward(context.Context, reward.Reward) error
	Reward(context.Context, id.ID) (reward.Reward, error)
	ListRewards(context.Context, page.Page) (page.Result[reward.Reward], error)
	UpdateReward(context.Context, reward.Reward) error
	DeleteReward(context.Context, id.ID) error
	GetConfig(context.Context) (config.Config, error)
	SaveConfig(context.Context, config.Config) error
	ListPreferences(context.Context) ([]config.NotificationPreference, error)
	SavePreferences(context.Context, []config.NotificationPreference) error
	CreateEmployee(context.Context, *domain.Employee, string) error
	Employee(context.Context, id.ID) (domain.Employee, error)
	ListEmployees(context.Context, page.Page) (page.Result[domain.Employee], error)
	UpdateEmployee(context.Context, domain.Employee) error
	DeactivateEmployee(context.Context, id.ID) error
}
type EmailSender interface {
	SendTemporaryPassword(context.Context, string, string, string) error
}
