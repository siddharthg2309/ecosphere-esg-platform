package http

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/notification/app"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/auth"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Handler struct{ service *app.Service }

func New(service *app.Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	result, err := h.service.List(r.Context(), principal.UserID, page.New(limit, offset))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	unread, _ := h.service.UnreadCount(r.Context(), principal.UserID)
	httpserver.JSON(w, http.StatusOK, map[string]any{"items": result.Items, "total": result.Total, "unread": unread})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
		return
	}
	nid, err := id.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, errs.Invalid("invalid_id", "Identifier is invalid", nil))
		return
	}
	if err = h.service.MarkRead(r.Context(), nid, principal.UserID); err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]string{"status": "read"})
}
