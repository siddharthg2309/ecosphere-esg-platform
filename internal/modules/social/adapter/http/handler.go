package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/app"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/auth"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Handler struct{ service *app.Service }

func New(service *app.Service) *Handler { return &Handler{service: service} }

func (h *Handler) ListActivities(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	result, err := h.service.ListActivities(r.Context(), page.New(limit, offset))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title            string  `json:"title"`
		CategoryID       string  `json:"categoryId"`
		Description      string  `json:"description"`
		Points           int     `json:"points"`
		EvidenceRequired bool    `json:"evidenceRequired"`
		ActivityDate     *string `json:"activityDate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	catID, err := id.Parse(body.CategoryID)
	if err != nil {
		httpserver.Error(w, errs.Invalid("invalid_category", "Category id is invalid", nil))
		return
	}
	var activityDate *time.Time
	if body.ActivityDate != nil && *body.ActivityDate != "" {
		t, err := time.Parse("2006-01-02", *body.ActivityDate)
		if err != nil {
			httpserver.Error(w, errs.Invalid("invalid_date", "Activity date must be YYYY-MM-DD", nil))
			return
		}
		activityDate = &t
	}
	result, err := h.service.CreateActivity(r.Context(), app.CreateActivityCmd{
		Title: body.Title, CategoryID: catID, Description: body.Description,
		Points: body.Points, EvidenceRequired: body.EvidenceRequired, ActivityDate: activityDate,
	})
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, result)
}

func (h *Handler) GetActivity(w http.ResponseWriter, r *http.Request) {
	activityID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.service.GetActivity(r.Context(), activityID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) JoinActivity(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
		return
	}
	var body struct {
		ActivityID string `json:"activityId"`
		ProofURL   string `json:"proofUrl"`
		Notes      string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	activityID, err := parseID(body.ActivityID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.service.JoinActivity(r.Context(), app.JoinActivityCmd{
		EmployeeID: principal.UserID, ActivityID: activityID, ProofURL: body.ProofURL, Notes: body.Notes,
	})
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, result)
}

func (h *Handler) ListParticipations(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	approval := r.URL.Query().Get("approval")
	result, err := h.service.ListParticipations(r.Context(), page.New(limit, offset), approval)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) ApproveParticipation(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
		return
	}
	pid, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.service.ApproveParticipation(r.Context(), pid, principal.UserID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) RejectParticipation(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
		return
	}
	pid, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.service.RejectParticipation(r.Context(), pid, principal.UserID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) Diversity(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Diversity(r.Context())
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) ListTrainings(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ListTrainings(r.Context())
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": result, "total": len(result)})
}

func (h *Handler) CreateTraining(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name       string `json:"name"`
		AssignedTo string `json:"assignedTo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	result, err := h.service.CreateTraining(r.Context(), body.Name, body.AssignedTo)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, result)
}

func (h *Handler) CompleteTraining(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
		return
	}
	tid, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	if err = h.service.CompleteTraining(r.Context(), tid, principal.UserID); err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func parseID(value string) (id.ID, error) {
	parsed, err := id.Parse(value)
	if err != nil {
		return "", errs.Invalid("invalid_id", "Identifier is invalid", nil)
	}
	return parsed, nil
}
