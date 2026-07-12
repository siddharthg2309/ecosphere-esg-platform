package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/shopspring/decimal"
	factor "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/emissionfactor/domain"
	badge "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/badge/domain"
	reward "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/reward/domain"
	identity "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/domain"
	category "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/category/domain"
	config "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/esgconfig/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/port"
	policy "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/policy/domain"
	product "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/product/domain"
	platformauth "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/auth"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
	"strings"
	"time"
)

type ESGConfigChanged struct{}

func (ESGConfigChanged) Name() string { return "ESGConfigChanged" }

type Service struct {
	repo  port.Repository
	email port.EmailSender
	bus   events.Bus
	now   func() time.Time
}

func New(repo port.Repository, email port.EmailSender, bus events.Bus) *Service {
	return &Service{repo: repo, email: email, bus: bus, now: func() time.Time { return time.Now().UTC() }}
}
func (s *Service) CreateCategory(ctx context.Context, name string, typ category.Type, status category.Status) (category.Category, error) {
	v, err := category.New(name, typ, status, s.now())
	if err != nil {
		return v, err
	}
	return v, s.repo.CreateCategory(ctx, v)
}
func (s *Service) UpdateCategory(ctx context.Context, v category.Category) (category.Category, error) {
	current, err := s.repo.Category(ctx, v.ID)
	if err != nil {
		return v, err
	}
	current.Name = strings.TrimSpace(v.Name)
	current.Type = v.Type
	current.Status = v.Status
	current.UpdatedAt = s.now()
	if err = current.Validate(); err != nil {
		return v, err
	}
	if err = s.repo.UpdateCategory(ctx, current); err != nil {
		return current, err
	}
	return s.repo.Category(ctx, current.ID)
}
func (s *Service) Category(ctx context.Context, v id.ID) (category.Category, error) {
	return s.repo.Category(ctx, v)
}
func (s *Service) ListCategories(ctx context.Context, t *category.Type, p page.Page) (page.Result[category.Category], error) {
	return s.repo.ListCategories(ctx, t, p)
}
func (s *Service) DeleteCategory(ctx context.Context, v id.ID) error {
	return s.repo.DeleteCategory(ctx, v)
}
func (s *Service) CreateFactor(ctx context.Context, name string, categoryID id.ID, unit, value string, status category.Status) (factor.Factor, error) {
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return factor.Factor{}, errs.Invalid("invalid_emission_factor", "kgCo2PerUnit must be a decimal", map[string]string{"kgCo2PerUnit": "Invalid decimal"})
	}
	v, err := factor.New(name, categoryID, unit, amount, status, s.now())
	if err != nil {
		return v, err
	}
	return v, s.repo.CreateFactor(ctx, v)
}
func (s *Service) UpdateFactor(ctx context.Context, v factor.Factor) (factor.Factor, error) {
	if err := v.Validate(); err != nil {
		return v, err
	}
	v.UpdatedAt = s.now()
	if err := s.repo.UpdateFactor(ctx, v); err != nil {
		return v, err
	}
	return s.repo.Factor(ctx, v.ID)
}
func (s *Service) Factor(ctx context.Context, v id.ID) (factor.Factor, error) {
	return s.repo.Factor(ctx, v)
}
func (s *Service) ListFactors(ctx context.Context, c *id.ID, p page.Page) (page.Result[factor.Factor], error) {
	return s.repo.ListFactors(ctx, c, p)
}
func (s *Service) DeleteFactor(ctx context.Context, v id.ID) error {
	return s.repo.DeleteFactor(ctx, v)
}
func (s *Service) CreateProduct(ctx context.Context, name string, attrs json.RawMessage, factorID *id.ID) (product.Profile, error) {
	v, err := product.New(name, attrs, factorID, s.now())
	if err != nil {
		return v, err
	}
	return v, s.repo.CreateProduct(ctx, v)
}
func (s *Service) UpdateProduct(ctx context.Context, v product.Profile) (product.Profile, error) {
	validated, err := product.New(v.Product, v.Attributes, v.EmissionFactorID, s.now())
	if err != nil {
		return v, err
	}
	validated.ID = v.ID
	validated.CreatedAt = v.CreatedAt
	if err = s.repo.UpdateProduct(ctx, validated); err != nil {
		return validated, err
	}
	return s.repo.Product(ctx, validated.ID)
}
func (s *Service) Product(ctx context.Context, v id.ID) (product.Profile, error) {
	return s.repo.Product(ctx, v)
}
func (s *Service) ListProducts(ctx context.Context, p page.Page) (page.Result[product.Profile], error) {
	return s.repo.ListProducts(ctx, p)
}
func (s *Service) DeleteProduct(ctx context.Context, v id.ID) error {
	return s.repo.DeleteProduct(ctx, v)
}
func (s *Service) CreatePolicy(ctx context.Context, title, body string, effective time.Time) (policy.Policy, error) {
	v, err := policy.New(title, body, effective, s.now())
	if err != nil {
		return v, err
	}
	return v, s.repo.CreatePolicy(ctx, v)
}
func (s *Service) PublishPolicy(ctx context.Context, policyID id.ID, title, body string) (policy.Policy, error) {
	v, err := s.repo.Policy(ctx, policyID)
	if err != nil {
		return v, err
	}
	if err = v.Publish(title, body, s.now()); err != nil {
		return v, err
	}
	if err = s.repo.UpdatePolicy(ctx, v); err != nil {
		return v, err
	}
	return s.repo.Policy(ctx, v.ID)
}
func (s *Service) Policy(ctx context.Context, v id.ID) (policy.Policy, error) {
	return s.repo.Policy(ctx, v)
}
func (s *Service) ListPolicies(ctx context.Context, p page.Page) (page.Result[policy.Policy], error) {
	return s.repo.ListPolicies(ctx, p)
}
func (s *Service) DeletePolicy(ctx context.Context, v id.ID) error {
	return s.repo.DeletePolicy(ctx, v)
}
func (s *Service) CreateBadge(ctx context.Context, name, description, icon string, rule badge.UnlockRule) (badge.Badge, error) {
	v, err := badge.New(name, description, icon, rule, s.now())
	if err != nil {
		return v, err
	}
	return v, s.repo.CreateBadge(ctx, v)
}
func (s *Service) UpdateBadge(ctx context.Context, v badge.Badge) (badge.Badge, error) {
	if strings.TrimSpace(v.Name) == "" {
		return v, errs.Invalid("invalid_badge", "Badge name is required", nil)
	}
	if err := v.UnlockRule.Validate(); err != nil {
		return v, err
	}
	v.UpdatedAt = s.now()
	if err := s.repo.UpdateBadge(ctx, v); err != nil {
		return v, err
	}
	return s.repo.Badge(ctx, v.ID)
}
func (s *Service) Badge(ctx context.Context, v id.ID) (badge.Badge, error) {
	return s.repo.Badge(ctx, v)
}
func (s *Service) ListBadges(ctx context.Context, p page.Page) (page.Result[badge.Badge], error) {
	return s.repo.ListBadges(ctx, p)
}
func (s *Service) DeleteBadge(ctx context.Context, v id.ID) error { return s.repo.DeleteBadge(ctx, v) }
func (s *Service) CreateReward(ctx context.Context, name, description string, points, stock int, status category.Status) (reward.Reward, error) {
	v, err := reward.New(name, description, points, stock, status, s.now())
	if err != nil {
		return v, err
	}
	return v, s.repo.CreateReward(ctx, v)
}
func (s *Service) UpdateReward(ctx context.Context, v reward.Reward) (reward.Reward, error) {
	validated, err := reward.New(v.Name, v.Description, v.PointsRequired, v.Stock, v.Status, s.now())
	if err != nil {
		return v, err
	}
	validated.ID = v.ID
	validated.CreatedAt = v.CreatedAt
	if err = s.repo.UpdateReward(ctx, validated); err != nil {
		return validated, err
	}
	return s.repo.Reward(ctx, validated.ID)
}
func (s *Service) Reward(ctx context.Context, v id.ID) (reward.Reward, error) {
	return s.repo.Reward(ctx, v)
}
func (s *Service) ListRewards(ctx context.Context, p page.Page) (page.Result[reward.Reward], error) {
	return s.repo.ListRewards(ctx, p)
}
func (s *Service) DeleteReward(ctx context.Context, v id.ID) error {
	return s.repo.DeleteReward(ctx, v)
}
func (s *Service) GetConfig(ctx context.Context) (config.Config, error) { return s.repo.GetConfig(ctx) }
func (s *Service) SaveConfig(ctx context.Context, v config.Config) (config.Config, error) {
	if err := v.Validate(); err != nil {
		return v, err
	}
	if err := s.repo.SaveConfig(ctx, v); err != nil {
		return v, err
	}
	if err := s.bus.Publish(ctx, ESGConfigChanged{}); err != nil {
		return v, err
	}
	return v, nil
}
func (s *Service) ListPreferences(ctx context.Context) ([]config.NotificationPreference, error) {
	return s.repo.ListPreferences(ctx)
}
func (s *Service) SavePreferences(ctx context.Context, values []config.NotificationPreference) ([]config.NotificationPreference, error) {
	seen := map[config.EventType]bool{}
	for _, v := range values {
		if !v.EventType.Valid() || seen[v.EventType] {
			return nil, errs.Invalid("invalid_notification_preferences", "Each supported notification event must appear once", nil)
		}
		seen[v.EventType] = true
	}
	if len(seen) != 4 {
		return nil, errs.Invalid("invalid_notification_preferences", "All four notification events are required", nil)
	}
	if err := s.repo.SavePreferences(ctx, values); err != nil {
		return nil, err
	}
	return s.repo.ListPreferences(ctx)
}
func (s *Service) CreateEmployee(ctx context.Context, name, email string, role identity.Role, departmentID *id.ID) (domain.Employee, error) {
	password, err := temporaryPassword()
	if err != nil {
		return domain.Employee{}, err
	}
	hash, err := platformauth.HashPassword(password)
	if err != nil {
		return domain.Employee{}, err
	}
	user, err := identity.NewUser(name, email, hash, role, departmentID, s.now())
	if err != nil {
		return domain.Employee{}, err
	}
	employee := domain.Employee{ID: user.ID, Name: user.Name, Email: user.Email, Role: user.Role, DepartmentID: user.DepartmentID, Status: "active", CreatedAt: user.CreatedAt}
	if err = s.repo.CreateEmployee(ctx, &employee, hash); err != nil {
		return employee, err
	}
	if s.email != nil {
		if err = s.email.SendTemporaryPassword(ctx, employee.Name, employee.Email, password); err != nil {
			return employee, err
		}
	}
	return employee, nil
}
func (s *Service) Employee(ctx context.Context, v id.ID) (domain.Employee, error) {
	return s.repo.Employee(ctx, v)
}
func (s *Service) ListEmployees(ctx context.Context, p page.Page) (page.Result[domain.Employee], error) {
	return s.repo.ListEmployees(ctx, p)
}
func (s *Service) UpdateEmployee(ctx context.Context, v domain.Employee) (domain.Employee, error) {
	if !v.Role.Valid() || strings.TrimSpace(v.Name) == "" || !strings.Contains(v.Email, "@") {
		return v, errs.Invalid("invalid_employee", "Employee details are invalid", nil)
	}
	if err := s.repo.UpdateEmployee(ctx, v); err != nil {
		return v, err
	}
	return s.repo.Employee(ctx, v.ID)
}
func (s *Service) DeactivateEmployee(ctx context.Context, v id.ID) error {
	return s.repo.DeactivateEmployee(ctx, v)
}
func temporaryPassword() (string, error) {
	value := make([]byte, 12)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return "Es!" + base64.RawURLEncoding.EncodeToString(value), nil
}
