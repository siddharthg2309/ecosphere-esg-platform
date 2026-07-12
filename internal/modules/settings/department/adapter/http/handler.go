package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/department/app"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/department/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Handler struct{ service *app.Service }

func New(service *app.Service) *Handler { return &Handler{service: service} }

type request struct {
	Name          string        `json:"name"`
	Code          string        `json:"code"`
	HeadID        *string       `json:"headId"`
	ParentID      *string       `json:"parentId"`
	EmployeeCount int           `json:"employeeCount"`
	Status        domain.Status `json:"status"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body request
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	headID, err := optionalID(body.HeadID, "headId")
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	parentID, err := optionalID(body.ParentID, "parentId")
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.service.Create(r.Context(), app.CreateCommand{Name: body.Name, Code: body.Code, HeadID: headID, ParentID: parentID, EmployeeCount: body.EmployeeCount})
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, result)
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	result, err := h.service.List(r.Context(), page.New(limit, offset))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	departmentID, err := requiredID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.service.ByID(r.Context(), departmentID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	departmentID, err := requiredID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	var body request
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	headID, err := optionalID(body.HeadID, "headId")
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	parentID, err := optionalID(body.ParentID, "parentId")
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	if body.Status == "" {
		body.Status = domain.StatusActive
	}
	result, err := h.service.Update(r.Context(), app.UpdateCommand{ID: departmentID, Name: body.Name, Code: body.Code, HeadID: headID, ParentID: parentID, EmployeeCount: body.EmployeeCount, Status: body.Status})
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}
func (h *Handler) Deactivate(w http.ResponseWriter, r *http.Request) {
	departmentID, err := requiredID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.service.Deactivate(r.Context(), departmentID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}
func requiredID(value string) (id.ID, error) {
	parsed, err := id.Parse(value)
	if err != nil {
		return "", errs.Invalid("invalid_id", "ID must be a UUID", map[string]string{"id": "Invalid UUID"})
	}
	return parsed, nil
}
func optionalID(value *string, field string) (*id.ID, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	parsed, err := id.Parse(*value)
	if err != nil {
		return nil, errs.Invalid("invalid_id", "ID must be a UUID", map[string]string{field: "Invalid UUID"})
	}
	return &parsed, nil
}
