package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	factor "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/emissionfactor/domain"
	badge "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/badge/domain"
	reward "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/reward/domain"
	identity "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/domain"
	category "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/category/domain"
	config "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/esgconfig/domain"
	master "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/domain"
	policy "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/policy/domain"
	product "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/product/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/db/sqlc"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
	"time"
)

type Repository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool, q: sqlc.New(pool)} }
func mapWrite(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return errs.Conflict("duplicate_master_data", "A record with this unique value already exists")
		}
		if pgErr.Code == "23503" {
			return errs.Conflict("record_in_use", "This record is referenced and cannot be deleted")
		}
		if pgErr.Code == "23514" {
			return errs.Invalid("constraint_failed", "The record violates a business rule", nil)
		}
	}
	return err
}
func absent(err error, code, name string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.NotFound(code, name+" not found")
	}
	return err
}
func uuid(v id.ID) pgtype.UUID { var out pgtype.UUID; _ = out.Scan(v.String()); return out }
func nullableUUID(v *id.ID) pgtype.UUID {
	if v == nil {
		return pgtype.UUID{}
	}
	return uuid(*v)
}
func fromUUID(v pgtype.UUID) id.ID {
	if !v.Valid {
		return ""
	}
	return id.ID(fmt.Sprintf("%x-%x-%x-%x-%x", v.Bytes[0:4], v.Bytes[4:6], v.Bytes[6:8], v.Bytes[8:10], v.Bytes[10:16]))
}
func numeric(v decimal.Decimal) pgtype.Numeric {
	var out pgtype.Numeric
	_ = out.Scan(v.String())
	return out
}
func fromNumeric(v pgtype.Numeric) decimal.Decimal {
	raw, _ := v.Value()
	out, _ := decimal.NewFromString(fmt.Sprint(raw))
	return out
}

func (r *Repository) CreateCategory(ctx context.Context, v category.Category) error {
	_, err := r.q.CreateCategory(ctx, sqlc.CreateCategoryParams{ID: uuid(v.ID), Name: v.Name, Type: string(v.Type), Status: string(v.Status)})
	return mapWrite(err)
}
func (r *Repository) Category(ctx context.Context, v id.ID) (category.Category, error) {
	row, err := r.q.CategoryByID(ctx, uuid(v))
	if err != nil {
		return category.Category{}, absent(err, "category_not_found", "Category")
	}
	return mapCategory(row), nil
}
func (r *Repository) ListCategories(ctx context.Context, t *category.Type, p page.Page) (page.Result[category.Category], error) {
	filter := ""
	if t != nil {
		filter = string(*t)
	}
	rows, err := r.q.ListCategories(ctx, sqlc.ListCategoriesParams{Column1: filter, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return page.Result[category.Category]{}, err
	}
	total, err := r.q.CountCategories(ctx, filter)
	if err != nil {
		return page.Result[category.Category]{}, err
	}
	items := make([]category.Category, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapCategory(row))
	}
	return page.Result[category.Category]{Items: items, Total: int(total)}, nil
}
func (r *Repository) UpdateCategory(ctx context.Context, v category.Category) error {
	_, err := r.q.UpdateCategory(ctx, sqlc.UpdateCategoryParams{ID: uuid(v.ID), Name: v.Name, Type: string(v.Type), Status: string(v.Status)})
	return mapWrite(absent(err, "category_not_found", "Category"))
}
func (r *Repository) DeleteCategory(ctx context.Context, v id.ID) error {
	n, err := r.q.DeleteCategory(ctx, uuid(v))
	if err != nil {
		return mapWrite(err)
	}
	if n == 0 {
		return errs.NotFound("category_not_found", "Category not found")
	}
	return nil
}
func mapCategory(v sqlc.Category) category.Category {
	return category.Category{ID: fromUUID(v.ID), Name: v.Name, Type: category.Type(v.Type), Status: category.Status(v.Status), CreatedAt: v.CreatedAt.Time, UpdatedAt: v.UpdatedAt.Time}
}

func (r *Repository) CreateFactor(ctx context.Context, v factor.Factor) error {
	_, err := r.q.CreateEmissionFactor(ctx, sqlc.CreateEmissionFactorParams{ID: uuid(v.ID), Name: v.Name, CategoryID: uuid(v.CategoryID), Unit: v.Unit, Kgco2PerUnit: numeric(v.KgCO2PerUnit), Status: string(v.Status)})
	return mapWrite(err)
}
func (r *Repository) Factor(ctx context.Context, v id.ID) (factor.Factor, error) {
	row, err := r.q.EmissionFactorByID(ctx, uuid(v))
	if err != nil {
		return factor.Factor{}, absent(err, "factor_not_found", "Emission factor")
	}
	return mapFactor(row), nil
}
func (r *Repository) ListFactors(ctx context.Context, c *id.ID, p page.Page) (page.Result[factor.Factor], error) {
	filter := nullableUUID(c)
	rows, err := r.q.ListEmissionFactors(ctx, sqlc.ListEmissionFactorsParams{Column1: filter, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return page.Result[factor.Factor]{}, err
	}
	total, err := r.q.CountEmissionFactors(ctx, filter)
	if err != nil {
		return page.Result[factor.Factor]{}, err
	}
	items := make([]factor.Factor, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapFactor(row))
	}
	return page.Result[factor.Factor]{Items: items, Total: int(total)}, nil
}
func (r *Repository) UpdateFactor(ctx context.Context, v factor.Factor) error {
	_, err := r.q.UpdateEmissionFactor(ctx, sqlc.UpdateEmissionFactorParams{ID: uuid(v.ID), Name: v.Name, CategoryID: uuid(v.CategoryID), Unit: v.Unit, Kgco2PerUnit: numeric(v.KgCO2PerUnit), Status: string(v.Status)})
	return mapWrite(absent(err, "factor_not_found", "Emission factor"))
}
func (r *Repository) DeleteFactor(ctx context.Context, v id.ID) error {
	n, err := r.q.DeleteEmissionFactor(ctx, uuid(v))
	if err != nil {
		return mapWrite(err)
	}
	if n == 0 {
		return errs.NotFound("factor_not_found", "Emission factor not found")
	}
	return nil
}
func mapFactor(v sqlc.EmissionFactor) factor.Factor {
	return factor.Factor{ID: fromUUID(v.ID), Name: v.Name, CategoryID: fromUUID(v.CategoryID), Unit: v.Unit, KgCO2PerUnit: fromNumeric(v.Kgco2PerUnit), Status: category.Status(v.Status), CreatedAt: v.CreatedAt.Time, UpdatedAt: v.UpdatedAt.Time}
}

func (r *Repository) CreateProduct(ctx context.Context, v product.Profile) error {
	_, err := r.q.CreateProductProfile(ctx, sqlc.CreateProductProfileParams{ID: uuid(v.ID), Product: v.Product, Attributes: v.Attributes, EmissionFactorID: nullableUUID(v.EmissionFactorID)})
	return mapWrite(err)
}
func (r *Repository) Product(ctx context.Context, v id.ID) (product.Profile, error) {
	row, err := r.q.ProductProfileByID(ctx, uuid(v))
	if err != nil {
		return product.Profile{}, absent(err, "product_not_found", "Product profile")
	}
	return mapProduct(row), nil
}
func (r *Repository) ListProducts(ctx context.Context, p page.Page) (page.Result[product.Profile], error) {
	rows, err := r.q.ListProductProfiles(ctx, sqlc.ListProductProfilesParams{Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return page.Result[product.Profile]{}, err
	}
	total, err := r.q.CountProductProfiles(ctx)
	if err != nil {
		return page.Result[product.Profile]{}, err
	}
	items := make([]product.Profile, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapProduct(v))
	}
	return page.Result[product.Profile]{Items: items, Total: int(total)}, nil
}
func (r *Repository) UpdateProduct(ctx context.Context, v product.Profile) error {
	_, err := r.q.UpdateProductProfile(ctx, sqlc.UpdateProductProfileParams{ID: uuid(v.ID), Product: v.Product, Attributes: v.Attributes, EmissionFactorID: nullableUUID(v.EmissionFactorID)})
	return mapWrite(absent(err, "product_not_found", "Product profile"))
}
func (r *Repository) DeleteProduct(ctx context.Context, v id.ID) error {
	n, err := r.q.DeleteProductProfile(ctx, uuid(v))
	if err != nil {
		return mapWrite(err)
	}
	if n == 0 {
		return errs.NotFound("product_not_found", "Product profile not found")
	}
	return nil
}
func mapProduct(v sqlc.ProductEsgProfile) product.Profile {
	p := product.Profile{ID: fromUUID(v.ID), Product: v.Product, Attributes: json.RawMessage(v.Attributes), CreatedAt: v.CreatedAt.Time, UpdatedAt: v.UpdatedAt.Time}
	if v.EmissionFactorID.Valid {
		x := fromUUID(v.EmissionFactorID)
		p.EmissionFactorID = &x
	}
	return p
}

func (r *Repository) CreatePolicy(ctx context.Context, v policy.Policy) error {
	_, err := r.q.CreatePolicy(ctx, sqlc.CreatePolicyParams{ID: uuid(v.ID), Title: v.Title, Body: v.Body, Version: int32(v.Version), EffectiveDate: pgtype.Date{Time: v.EffectiveDate, Valid: true}})
	return mapWrite(err)
}
func (r *Repository) Policy(ctx context.Context, v id.ID) (policy.Policy, error) {
	row, err := r.q.PolicyByID(ctx, uuid(v))
	if err != nil {
		return policy.Policy{}, absent(err, "policy_not_found", "Policy")
	}
	return mapPolicy(row), nil
}
func (r *Repository) ListPolicies(ctx context.Context, p page.Page) (page.Result[policy.Policy], error) {
	rows, err := r.q.ListPolicies(ctx, sqlc.ListPoliciesParams{Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return page.Result[policy.Policy]{}, err
	}
	total, err := r.q.CountPolicies(ctx)
	if err != nil {
		return page.Result[policy.Policy]{}, err
	}
	items := make([]policy.Policy, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapPolicy(v))
	}
	return page.Result[policy.Policy]{Items: items, Total: int(total)}, nil
}
func (r *Repository) UpdatePolicy(ctx context.Context, v policy.Policy) error {
	_, err := r.q.UpdatePolicy(ctx, sqlc.UpdatePolicyParams{ID: uuid(v.ID), Title: v.Title, Body: v.Body, Version: int32(v.Version), EffectiveDate: pgtype.Date{Time: v.EffectiveDate, Valid: true}})
	return mapWrite(absent(err, "policy_not_found", "Policy"))
}
func (r *Repository) DeletePolicy(ctx context.Context, v id.ID) error {
	n, err := r.q.DeletePolicy(ctx, uuid(v))
	if err != nil {
		return mapWrite(err)
	}
	if n == 0 {
		return errs.NotFound("policy_not_found", "Policy not found")
	}
	return nil
}
func mapPolicy(v sqlc.EsgPolicy) policy.Policy {
	return policy.Policy{ID: fromUUID(v.ID), Title: v.Title, Body: v.Body, Version: int(v.Version), EffectiveDate: v.EffectiveDate.Time, CreatedAt: v.CreatedAt.Time, UpdatedAt: v.UpdatedAt.Time}
}

func (r *Repository) CreateBadge(ctx context.Context, v badge.Badge) error {
	rule, _ := json.Marshal(v.UnlockRule)
	_, err := r.q.CreateBadge(ctx, sqlc.CreateBadgeParams{ID: uuid(v.ID), Name: v.Name, Description: v.Description, Icon: v.Icon, UnlockRule: rule})
	return mapWrite(err)
}
func (r *Repository) Badge(ctx context.Context, v id.ID) (badge.Badge, error) {
	row, err := r.q.BadgeByID(ctx, uuid(v))
	if err != nil {
		return badge.Badge{}, absent(err, "badge_not_found", "Badge")
	}
	return mapBadge(row), nil
}
func (r *Repository) ListBadges(ctx context.Context, p page.Page) (page.Result[badge.Badge], error) {
	rows, err := r.q.ListBadges(ctx, sqlc.ListBadgesParams{Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return page.Result[badge.Badge]{}, err
	}
	total, err := r.q.CountBadges(ctx)
	if err != nil {
		return page.Result[badge.Badge]{}, err
	}
	items := make([]badge.Badge, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapBadge(v))
	}
	return page.Result[badge.Badge]{Items: items, Total: int(total)}, nil
}
func (r *Repository) UpdateBadge(ctx context.Context, v badge.Badge) error {
	rule, _ := json.Marshal(v.UnlockRule)
	_, err := r.q.UpdateBadge(ctx, sqlc.UpdateBadgeParams{ID: uuid(v.ID), Name: v.Name, Description: v.Description, Icon: v.Icon, UnlockRule: rule})
	return mapWrite(absent(err, "badge_not_found", "Badge"))
}
func (r *Repository) DeleteBadge(ctx context.Context, v id.ID) error {
	n, err := r.q.DeleteBadge(ctx, uuid(v))
	if err != nil {
		return mapWrite(err)
	}
	if n == 0 {
		return errs.NotFound("badge_not_found", "Badge not found")
	}
	return nil
}
func mapBadge(v sqlc.Badge) badge.Badge {
	var rule badge.UnlockRule
	_ = json.Unmarshal(v.UnlockRule, &rule)
	return badge.Badge{ID: fromUUID(v.ID), Name: v.Name, Description: v.Description, Icon: v.Icon, UnlockRule: rule, CreatedAt: v.CreatedAt.Time, UpdatedAt: v.UpdatedAt.Time}
}

func (r *Repository) CreateReward(ctx context.Context, v reward.Reward) error {
	_, err := r.q.CreateReward(ctx, sqlc.CreateRewardParams{ID: uuid(v.ID), Name: v.Name, Description: v.Description, PointsRequired: int32(v.PointsRequired), Stock: int32(v.Stock), Status: string(v.Status)})
	return mapWrite(err)
}
func (r *Repository) Reward(ctx context.Context, v id.ID) (reward.Reward, error) {
	row, err := r.q.RewardByID(ctx, uuid(v))
	if err != nil {
		return reward.Reward{}, absent(err, "reward_not_found", "Reward")
	}
	return mapReward(row), nil
}
func (r *Repository) ListRewards(ctx context.Context, p page.Page) (page.Result[reward.Reward], error) {
	rows, err := r.q.ListRewards(ctx, sqlc.ListRewardsParams{Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return page.Result[reward.Reward]{}, err
	}
	total, err := r.q.CountRewards(ctx)
	if err != nil {
		return page.Result[reward.Reward]{}, err
	}
	items := make([]reward.Reward, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapReward(v))
	}
	return page.Result[reward.Reward]{Items: items, Total: int(total)}, nil
}
func (r *Repository) UpdateReward(ctx context.Context, v reward.Reward) error {
	_, err := r.q.UpdateReward(ctx, sqlc.UpdateRewardParams{ID: uuid(v.ID), Name: v.Name, Description: v.Description, PointsRequired: int32(v.PointsRequired), Stock: int32(v.Stock), Status: string(v.Status)})
	return mapWrite(absent(err, "reward_not_found", "Reward"))
}
func (r *Repository) DeleteReward(ctx context.Context, v id.ID) error {
	n, err := r.q.DeleteReward(ctx, uuid(v))
	if err != nil {
		return mapWrite(err)
	}
	if n == 0 {
		return errs.NotFound("reward_not_found", "Reward not found")
	}
	return nil
}
func mapReward(v sqlc.Reward) reward.Reward {
	return reward.Reward{ID: fromUUID(v.ID), Name: v.Name, Description: v.Description, PointsRequired: int(v.PointsRequired), Stock: int(v.Stock), Status: category.Status(v.Status), CreatedAt: v.CreatedAt.Time, UpdatedAt: v.UpdatedAt.Time}
}

func (r *Repository) GetConfig(ctx context.Context) (config.Config, error) {
	v, err := r.q.GetESGConfig(ctx)
	if err != nil {
		return config.Config{}, err
	}
	return config.Config{AutoEmissionCalc: v.AutoEmissionCalc, RequireCSREvidence: v.RequireCsrEvidence, AutoAwardBadges: v.AutoAwardBadges, NotifyComplianceEmail: v.NotifyComplianceEmail, WeightEnv: int(v.WeightEnv), WeightSocial: int(v.WeightSocial), WeightGov: int(v.WeightGov)}, nil
}
func (r *Repository) SaveConfig(ctx context.Context, v config.Config) error {
	_, err := r.q.UpdateESGConfig(ctx, sqlc.UpdateESGConfigParams{AutoEmissionCalc: v.AutoEmissionCalc, RequireCsrEvidence: v.RequireCSREvidence, AutoAwardBadges: v.AutoAwardBadges, NotifyComplianceEmail: v.NotifyComplianceEmail, WeightEnv: int32(v.WeightEnv), WeightSocial: int32(v.WeightSocial), WeightGov: int32(v.WeightGov)})
	return mapWrite(err)
}
func (r *Repository) ListPreferences(ctx context.Context) ([]config.NotificationPreference, error) {
	rows, err := r.q.ListNotificationPreferences(ctx)
	if err != nil {
		return nil, err
	}
	values := make([]config.NotificationPreference, 0, len(rows))
	for _, v := range rows {
		values = append(values, config.NotificationPreference{EventType: config.EventType(v.EventType), InAppEnabled: v.InAppEnabled, EmailEnabled: v.EmailEnabled})
	}
	return values, nil
}
func (r *Repository) SavePreferences(ctx context.Context, values []config.NotificationPreference) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)
	for _, v := range values {
		if _, err = q.UpsertNotificationPreference(ctx, sqlc.UpsertNotificationPreferenceParams{EventType: string(v.EventType), InAppEnabled: v.InAppEnabled, EmailEnabled: v.EmailEnabled}); err != nil {
			return mapWrite(err)
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) CreateEmployee(ctx context.Context, v *master.Employee, passwordHash string) error {
	row, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{ID: uuid(v.ID), Name: v.Name, Email: v.Email, PasswordHash: passwordHash, Role: string(v.Role), DepartmentID: nullableUUID(v.DepartmentID)})
	if err != nil {
		return mapWrite(err)
	}
	*v = mapEmployee(row)
	return nil
}
func (r *Repository) Employee(ctx context.Context, v id.ID) (master.Employee, error) {
	row, err := r.q.UserByID(ctx, uuid(v))
	if err != nil {
		return master.Employee{}, absent(err, "employee_not_found", "Employee")
	}
	return mapEmployee(row), nil
}
func (r *Repository) ListEmployees(ctx context.Context, p page.Page) (page.Result[master.Employee], error) {
	rows, err := r.q.ListEmployees(ctx, sqlc.ListEmployeesParams{Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return page.Result[master.Employee]{}, err
	}
	total, err := r.q.CountEmployees(ctx)
	if err != nil {
		return page.Result[master.Employee]{}, err
	}
	items := make([]master.Employee, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapEmployee(v))
	}
	return page.Result[master.Employee]{Items: items, Total: int(total)}, nil
}
func (r *Repository) UpdateEmployee(ctx context.Context, v master.Employee) error {
	_, err := r.q.UpdateEmployee(ctx, sqlc.UpdateEmployeeParams{ID: uuid(v.ID), Name: v.Name, Email: v.Email, Role: string(v.Role), DepartmentID: nullableUUID(v.DepartmentID), Status: v.Status})
	return mapWrite(absent(err, "employee_not_found", "Employee"))
}
func (r *Repository) DeactivateEmployee(ctx context.Context, v id.ID) error {
	n, err := r.q.DeactivateEmployee(ctx, uuid(v))
	if err != nil {
		return err
	}
	if n == 0 {
		return errs.NotFound("employee_not_found", "Employee not found")
	}
	return nil
}
func mapEmployee(v sqlc.User) master.Employee {
	e := master.Employee{ID: fromUUID(v.ID), Name: v.Name, Email: v.Email, Role: identity.Role(v.Role), XP: int(v.Xp), Points: int(v.Points), CompletedChallenges: int(v.CompletedChallenges), Status: v.Status, CreatedAt: v.CreatedAt.Time}
	if v.DepartmentID.Valid {
		x := fromUUID(v.DepartmentID)
		e.DepartmentID = &x
	}
	return e
}

var _ = time.Time{}
