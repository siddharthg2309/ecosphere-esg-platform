package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/app"
	challenge "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/challenge/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/auth"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Handler struct{ service *app.Service }

func New(service *app.Service) *Handler { return &Handler{service: service} }

func (h *Handler) ListChallenges(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	result, err := h.service.ListChallenges(r.Context(), page.New(limit, offset))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) StatusCounts(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.StatusCounts(r.Context())
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) CreateChallenge(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title            string  `json:"title"`
		CategoryID       string  `json:"categoryId"`
		Description      string  `json:"description"`
		XP               int     `json:"xp"`
		Difficulty       string  `json:"difficulty"`
		EvidenceRequired bool    `json:"evidenceRequired"`
		Deadline         *string `json:"deadline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	catID, err := parseID(body.CategoryID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	var deadline *time.Time
	if body.Deadline != nil && *body.Deadline != "" {
		t, err := time.Parse("2006-01-02", *body.Deadline)
		if err != nil {
			httpserver.Error(w, errs.Invalid("invalid_date", "Deadline must be YYYY-MM-DD", nil))
			return
		}
		deadline = &t
	}
	result, err := h.service.CreateChallenge(r.Context(), app.CreateChallengeCmd{
		Title: body.Title, CategoryID: catID, Description: body.Description,
		XP: body.XP, Difficulty: body.Difficulty, EvidenceRequired: body.EvidenceRequired, Deadline: deadline,
	})
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, result)
}

func (h *Handler) Transition(w http.ResponseWriter, r *http.Request) {
	cid, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	var body struct {
		To string `json:"to"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	result, err := h.service.Transition(r.Context(), cid, challenge.ChallengeStatus(body.To))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) Participate(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
		return
	}
	cid, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	var body struct {
		Progress int    `json:"progress"`
		ProofURL string `json:"proofUrl"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	if body.Progress == 0 {
		body.Progress = 100
	}
	result, err := h.service.Participate(r.Context(), cid, principal.UserID, body.Progress, body.ProofURL)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, result)
}

func (h *Handler) ListParticipations(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	result, err := h.service.ListParticipations(r.Context(), page.New(limit, offset))
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

func (h *Handler) Balance(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
		return
	}
	result, err := h.service.Balance(r.Context(), principal.UserID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) Leaderboard(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "employee"
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	result, err := h.service.Leaderboard(r.Context(), scope, limit)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": result, "scope": scope})
}

func (h *Handler) ListRewards(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ListRewards(r.Context())
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": result, "total": len(result)})
}

func (h *Handler) Redeem(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
		return
	}
	rid, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, spent, err := h.service.Redeem(r.Context(), rid, principal.UserID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"reward": result, "pointsSpent": spent})
}

func (h *Handler) ListBadges(w http.ResponseWriter, r *http.Request) {
	badges, counts, err := h.service.ListBadges(r.Context())
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	type row struct {
		ID          id.ID              `json:"id"`
		Name        string             `json:"name"`
		Description string             `json:"description"`
		Icon        string             `json:"icon"`
		UnlockRule  any                `json:"unlockRule"`
		EarnedCount int                `json:"earnedCount"`
	}
	items := make([]row, 0, len(badges))
	for _, b := range badges {
		items = append(items, row{
			ID: b.ID, Name: b.Name, Description: b.Description, Icon: b.Icon,
			UnlockRule: b.UnlockRule, EarnedCount: counts[b.ID],
		})
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items)})
}

func (h *Handler) Transitions(w http.ResponseWriter, r *http.Request) {
	httpserver.JSON(w, http.StatusOK, challenge.Transitions)
}

func parseID(value string) (id.ID, error) {
	parsed, err := id.Parse(value)
	if err != nil {
		return "", errs.Invalid("invalid_id", "Identifier is invalid", nil)
	}
	return parsed, nil
}
