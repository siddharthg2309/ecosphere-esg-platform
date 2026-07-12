package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/app"
	audit "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/audit/domain"
	compliance "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/compliance/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/auth"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Handler struct{ service *app.Service }

func New(service *app.Service) *Handler { return &Handler{service: service} }

func (h *Handler) ListAudits(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	result, err := h.service.ListAudits(r.Context(), page.New(limit, offset))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) CreateAudit(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title        string `json:"title"`
		DepartmentID string `json:"departmentId"`
		AuditorID    string `json:"auditorId"`
		AuditDate    string `json:"auditDate"`
		Findings     string `json:"findings"`
		Status       string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	dept, err := parseID(body.DepartmentID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	auditor, err := parseID(body.AuditorID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	date, err := time.Parse("2006-01-02", body.AuditDate)
	if err != nil {
		httpserver.Error(w, errs.Invalid("invalid_date", "Audit date must be YYYY-MM-DD", nil))
		return
	}
	status := audit.AuditStatus(body.Status)
	if status == "" {
		status = audit.StatusDraft
	}
	result, err := h.service.CreateAudit(r.Context(), body.Title, dept, auditor, date, body.Findings, status)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, result)
}

func (h *Handler) GetAudit(w http.ResponseWriter, r *http.Request) {
	aid, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.service.GetAudit(r.Context(), aid)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) DepartmentBundle(w http.ResponseWriter, r *http.Request) {
	dept, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.service.DepartmentBundle(r.Context(), dept)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) ListIssues(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	status := r.URL.Query().Get("status")
	var overdue *bool
	if v := r.URL.Query().Get("overdue"); v == "true" {
		t := true
		overdue = &t
	} else if v == "false" {
		f := false
		overdue = &f
	}
	result, err := h.service.ListIssues(r.Context(), page.New(limit, offset), status, overdue)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) RaiseIssue(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DepartmentID string `json:"departmentId"`
		OwnerID      string `json:"ownerId"`
		Severity     string `json:"severity"`
		Description  string `json:"description"`
		DueDate      string `json:"dueDate"`
		AuditID      string `json:"auditId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	dept, err := parseID(body.DepartmentID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	owner, err := parseID(body.OwnerID)
	if err != nil {
		httpserver.Error(w, errs.Invalid("owner_required", "every issue needs an owner", nil))
		return
	}
	due, err := time.Parse("2006-01-02", body.DueDate)
	if err != nil {
		httpserver.Error(w, errs.Invalid("due_required", "every issue needs a due date", nil))
		return
	}
	var auditID *id.ID
	if body.AuditID != "" {
		a, err := parseID(body.AuditID)
		if err != nil {
			httpserver.Error(w, err)
			return
		}
		auditID = &a
	}
	result, err := h.service.RaiseIssue(r.Context(), app.RaiseIssueCmd{
		DepartmentID: dept, OwnerID: owner, Severity: compliance.Severity(body.Severity),
		Description: body.Description, DueDate: due, AuditID: auditID,
	})
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, result)
}

func (h *Handler) UpdateIssue(w http.ResponseWriter, r *http.Request) {
	issueID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	result, err := h.service.UpdateIssue(r.Context(), issueID, compliance.IssueStatus(body.Status))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) GetIssue(w http.ResponseWriter, r *http.Request) {
	issueID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.service.GetIssue(r.Context(), issueID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) Unacknowledged(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
		return
	}
	result, err := h.service.Unacknowledged(r.Context(), principal.UserID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": result, "total": len(result)})
}

func (h *Handler) Acknowledge(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
		return
	}
	policyID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.service.AcknowledgePolicy(r.Context(), principal.UserID, policyID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, result)
}

func (h *Handler) ListAcknowledgements(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	result, err := h.service.ListAcknowledgements(r.Context(), page.New(limit, offset))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) ListGovernancePolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := h.service.ListPolicies(r.Context())
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	type row struct {
		ID            id.ID     `json:"id"`
		Title         string    `json:"title"`
		Body          string    `json:"body"`
		Version       int       `json:"version"`
		EffectiveDate time.Time `json:"effectiveDate"`
		Acked         int       `json:"acked"`
		Total         int       `json:"total"`
		AckRate       float64   `json:"ackRate"`
	}
	items := make([]row, 0, len(policies))
	for _, p := range policies {
		acked, total, _ := h.service.PolicyAckRate(r.Context(), p.ID, p.Version)
		rate := 0.0
		if total > 0 {
			rate = float64(acked) * 100 / float64(total)
		}
		items = append(items, row{
			ID: p.ID, Title: p.Title, Body: p.Body, Version: p.Version, EffectiveDate: p.EffectiveDate,
			Acked: acked, Total: total, AckRate: rate,
		})
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items)})
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	open, overdue, audits, err := h.service.Stats(r.Context())
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{
		"openIssues": open, "overdueIssues": overdue, "auditsFY": audits, "governanceScore": 88,
	})
}

func parseID(value string) (id.ID, error) {
	parsed, err := id.Parse(value)
	if err != nil {
		return "", errs.Invalid("invalid_id", "Identifier is invalid", nil)
	}
	return parsed, nil
}
