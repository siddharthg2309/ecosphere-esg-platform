package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
	factor "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/emissionfactor/domain"
	badge "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/badge/domain"
	reward "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/reward/domain"
	identity "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/domain"
	category "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/category/domain"
	config "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/esgconfig/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/app"
	policy "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/policy/domain"
	product "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/product/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Handler struct{ service *app.Service }

func New(service *app.Service) *Handler { return &Handler{service: service} }

type entityRequest struct {
	Name             string           `json:"name"`
	Type             category.Type    `json:"type"`
	Status           category.Status  `json:"status"`
	CategoryID       string           `json:"categoryId"`
	Unit             string           `json:"unit"`
	KgCO2PerUnit     string           `json:"kgCo2PerUnit"`
	Product          string           `json:"product"`
	Attributes       json.RawMessage  `json:"attributes"`
	EmissionFactorID *string          `json:"emissionFactorId"`
	Title            string           `json:"title"`
	Body             string           `json:"body"`
	EffectiveDate    string           `json:"effectiveDate"`
	Description      string           `json:"description"`
	Icon             string           `json:"icon"`
	UnlockRule       badge.UnlockRule `json:"unlockRule"`
	PointsRequired   int              `json:"pointsRequired"`
	Stock            int              `json:"stock"`
	Email            string           `json:"email"`
	Role             identity.Role    `json:"role"`
	DepartmentID     *string          `json:"departmentId"`
}

func decode(r *http.Request) (entityRequest, error) {
	var v entityRequest
	err := json.NewDecoder(r.Body).Decode(&v)
	return v, err
}
func pagination(r *http.Request) page.Page {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return page.New(limit, offset)
}
func pathID(r *http.Request) (id.ID, error) {
	v, err := id.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return "", errs.Invalid("invalid_id", "ID must be a UUID", map[string]string{"id": "Invalid UUID"})
	}
	return v, nil
}
func optID(value *string) (*id.ID, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	v, err := id.Parse(*value)
	if err != nil {
		return nil, errs.Invalid("invalid_id", "Referenced ID must be a UUID", nil)
	}
	return &v, nil
}

func (h *Handler) List(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var result any
		var err error
		p := pagination(r)
		switch entity {
		case "categories":
			var typ *category.Type
			if raw := r.URL.Query().Get("type"); raw != "" {
				v := category.Type(raw)
				typ = &v
			}
			result, err = h.service.ListCategories(r.Context(), typ, p)
		case "emission-factors":
			var categoryID *id.ID
			if raw := r.URL.Query().Get("category"); raw != "" {
				v, e := id.Parse(raw)
				if e != nil {
					httpserver.Error(w, errs.Invalid("invalid_category", "Category filter must be a UUID", nil))
					return
				}
				categoryID = &v
			}
			result, err = h.service.ListFactors(r.Context(), categoryID, p)
		case "products":
			result, err = h.service.ListProducts(r.Context(), p)
		case "policies":
			result, err = h.service.ListPolicies(r.Context(), p)
		case "badges":
			result, err = h.service.ListBadges(r.Context(), p)
		case "rewards":
			result, err = h.service.ListRewards(r.Context(), p)
		case "employees":
			result, err = h.service.ListEmployees(r.Context(), p)
		}
		if err != nil {
			httpserver.Error(w, err)
			return
		}
		httpserver.JSON(w, 200, result)
	}
}

func (h *Handler) Get(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entityID, err := pathID(r)
		if err != nil {
			httpserver.Error(w, err)
			return
		}
		var result any
		switch entity {
		case "categories":
			result, err = h.service.Category(r.Context(), entityID)
		case "emission-factors":
			result, err = h.service.Factor(r.Context(), entityID)
		case "products":
			result, err = h.service.Product(r.Context(), entityID)
		case "policies":
			result, err = h.service.Policy(r.Context(), entityID)
		case "badges":
			result, err = h.service.Badge(r.Context(), entityID)
		case "rewards":
			result, err = h.service.Reward(r.Context(), entityID)
		case "employees":
			result, err = h.service.Employee(r.Context(), entityID)
		}
		if err != nil {
			httpserver.Error(w, err)
			return
		}
		httpserver.JSON(w, 200, result)
	}
}

func (h *Handler) Create(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := decode(r)
		if err != nil {
			httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
			return
		}
		var result any
		switch entity {
		case "categories":
			result, err = h.service.CreateCategory(r.Context(), body.Name, body.Type, body.Status)
		case "emission-factors":
			var categoryID id.ID
			categoryID, err = id.Parse(body.CategoryID)
			if err == nil {
				result, err = h.service.CreateFactor(r.Context(), body.Name, categoryID, body.Unit, body.KgCO2PerUnit, body.Status)
			} else {
				err = errs.Invalid("invalid_category", "Category ID must be a UUID", nil)
			}
		case "products":
			var factorID *id.ID
			factorID, err = optID(body.EmissionFactorID)
			if err == nil {
				result, err = h.service.CreateProduct(r.Context(), body.Product, body.Attributes, factorID)
			}
		case "policies":
			var effective time.Time
			effective, err = time.Parse("2006-01-02", body.EffectiveDate)
			if err == nil {
				result, err = h.service.CreatePolicy(r.Context(), body.Title, body.Body, effective)
			} else {
				err = errs.Invalid("invalid_effective_date", "Effective date must use YYYY-MM-DD", nil)
			}
		case "badges":
			result, err = h.service.CreateBadge(r.Context(), body.Name, body.Description, body.Icon, body.UnlockRule)
		case "rewards":
			result, err = h.service.CreateReward(r.Context(), body.Name, body.Description, body.PointsRequired, body.Stock, body.Status)
		case "employees":
			var departmentID *id.ID
			departmentID, err = optID(body.DepartmentID)
			if err == nil {
				result, err = h.service.CreateEmployee(r.Context(), body.Name, body.Email, body.Role, departmentID)
			}
		}
		if err != nil {
			httpserver.Error(w, err)
			return
		}
		httpserver.JSON(w, 201, result)
	}
}

func (h *Handler) Update(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entityID, err := pathID(r)
		if err != nil {
			httpserver.Error(w, err)
			return
		}
		body, err := decode(r)
		if err != nil {
			httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
			return
		}
		var result any
		switch entity {
		case "categories":
			result, err = h.service.UpdateCategory(r.Context(), category.Category{ID: entityID, Name: body.Name, Type: body.Type, Status: body.Status})
		case "emission-factors":
			categoryID, e1 := id.Parse(body.CategoryID)
			amount, e2 := decimal.NewFromString(body.KgCO2PerUnit)
			if e1 != nil || e2 != nil {
				err = errs.Invalid("invalid_emission_factor", "Category and factor value are invalid", nil)
			} else {
				result, err = h.service.UpdateFactor(r.Context(), factor.Factor{ID: entityID, Name: body.Name, CategoryID: categoryID, Unit: body.Unit, KgCO2PerUnit: amount, Status: body.Status})
			}
		case "products":
			var factorID *id.ID
			factorID, err = optID(body.EmissionFactorID)
			if err == nil {
				result, err = h.service.UpdateProduct(r.Context(), product.Profile{ID: entityID, Product: body.Product, Attributes: body.Attributes, EmissionFactorID: factorID})
			}
		case "policies":
			result, err = h.service.PublishPolicy(r.Context(), entityID, body.Title, body.Body)
		case "badges":
			result, err = h.service.UpdateBadge(r.Context(), badge.Badge{ID: entityID, Name: body.Name, Description: body.Description, Icon: body.Icon, UnlockRule: body.UnlockRule})
		case "rewards":
			result, err = h.service.UpdateReward(r.Context(), reward.Reward{ID: entityID, Name: body.Name, Description: body.Description, PointsRequired: body.PointsRequired, Stock: body.Stock, Status: body.Status})
		case "employees":
			var departmentID *id.ID
			departmentID, err = optID(body.DepartmentID)
			if err == nil {
				var current interface{}
				_ = current
				employee, e := h.service.Employee(r.Context(), entityID)
				if e != nil {
					err = e
				} else {
					employee.Name = body.Name
					employee.Email = body.Email
					employee.Role = body.Role
					employee.DepartmentID = departmentID
					employee.Status = string(body.Status)
					if employee.Status == "" {
						employee.Status = "active"
					}
					result, err = h.service.UpdateEmployee(r.Context(), employee)
				}
			}
		}
		if err != nil {
			httpserver.Error(w, err)
			return
		}
		httpserver.JSON(w, 200, result)
	}
}

func (h *Handler) Delete(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entityID, err := pathID(r)
		if err != nil {
			httpserver.Error(w, err)
			return
		}
		switch entity {
		case "categories":
			err = h.service.DeleteCategory(r.Context(), entityID)
		case "emission-factors":
			err = h.service.DeleteFactor(r.Context(), entityID)
		case "products":
			err = h.service.DeleteProduct(r.Context(), entityID)
		case "policies":
			err = h.service.DeletePolicy(r.Context(), entityID)
		case "badges":
			err = h.service.DeleteBadge(r.Context(), entityID)
		case "rewards":
			err = h.service.DeleteReward(r.Context(), entityID)
		case "employees":
			err = h.service.DeactivateEmployee(r.Context(), entityID)
		}
		if err != nil {
			httpserver.Error(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.GetConfig(r.Context())
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, 200, v)
}
func (h *Handler) SaveConfig(w http.ResponseWriter, r *http.Request) {
	var v config.Config
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	saved, err := h.service.SaveConfig(r.Context(), v)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, 200, saved)
}
func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	v, err := h.service.ListPreferences(r.Context())
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, 200, v)
}
func (h *Handler) SavePreferences(w http.ResponseWriter, r *http.Request) {
	var v []config.NotificationPreference
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	saved, err := h.service.SavePreferences(r.Context(), v)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, 200, saved)
}

var _ = policy.Policy{}
