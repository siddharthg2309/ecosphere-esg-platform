package http

import (
	"net/http"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/scoring/app"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
)

type Handler struct{ service *app.Service }

func New(service *app.Service) *Handler { return &Handler{service: service} }

func (h *Handler) Departments(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	scores, err := h.service.Departments(r.Context(), period)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": scores, "total": len(scores)})
}

func (h *Handler) Overall(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	overall, scores, weights, err := h.service.Overall(r.Context(), period)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	// Average pillars for display
	var e, s, g int
	if len(scores) > 0 {
		for _, sc := range scores {
			e += sc.Env
			s += sc.Social
			g += sc.Gov
		}
		e /= len(scores)
		s /= len(scores)
		g /= len(scores)
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{
		"overall": overall, "environmental": e, "social": s, "governance": g,
		"weights": weights, "departments": scores,
	})
}

func (h *Handler) Recompute(w http.ResponseWriter, r *http.Request) {
	scores, err := h.service.RecomputeAll(r.Context())
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": scores, "total": len(scores)})
}
