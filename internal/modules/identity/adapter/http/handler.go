package http

import (
	"encoding/json"
	"net/http"

	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/app"
	platformauth "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/auth"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
)

type Handler struct{ service *app.Service }

func New(service *app.Service) *Handler { return &Handler{service: service} }

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Email == "" || request.Password == "" {
		httpserver.Error(w, errs.Invalid("invalid_request", "Email and password are required", map[string]string{"email": "Required", "password": "Required"}))
		return
	}
	result, err := h.service.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.RefreshToken == "" {
		httpserver.Error(w, errs.Invalid("invalid_request", "Refresh token is required", map[string]string{"refreshToken": "Required"}))
		return
	}
	result, err := h.service.Refresh(r.Context(), request.RefreshToken)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result.User = nil
	httpserver.JSON(w, http.StatusOK, result)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := platformauth.PrincipalFrom(r.Context())
	if !ok {
		httpserver.Error(w, errs.Unauthorized("authentication_required", "Authentication is required"))
		return
	}
	user, err := h.service.Me(r.Context(), principal.UserID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, user)
}
