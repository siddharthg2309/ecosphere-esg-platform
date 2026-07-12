package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/reporting/app"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/reporting/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/auth"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type Handler struct{ service *app.Service }

func New(service *app.Service) *Handler { return &Handler{service: service} }

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Type    string          `json:"type"`
		Filters domain.Filters  `json:"filters"`
		From    string          `json:"from"`
		To      string          `json:"to"`
		DepartmentID string     `json:"departmentId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_request", "Request body is invalid", nil))
		return
	}
	filters := body.Filters
	if body.DepartmentID != "" {
		d, err := id.Parse(body.DepartmentID)
		if err == nil {
			filters.DepartmentID = &d
		}
	}
	if body.From != "" {
		if t, err := time.Parse("2006-01-02", body.From); err == nil {
			filters.From = &t
		}
	}
	if body.To != "" {
		if t, err := time.Parse("2006-01-02", body.To); err == nil {
			filters.To = &t
		}
	}
	var by *id.ID
	if p, ok := auth.PrincipalFrom(r.Context()); ok {
		by = &p.UserID
	}
	result, err := h.service.Generate(r.Context(), domain.ReportType(body.Type), filters, by)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, result)
}

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	rid, err := id.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, errs.Invalid("invalid_id", "Identifier is invalid", nil))
		return
	}
	fmtName := r.URL.Query().Get("fmt")
	if fmtName == "" {
		fmtName = "csv"
	}
	data, mime, err := h.service.Export(r.Context(), rid, fmtName)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="ecosphere-report.%s"`, fmtName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	rid, err := id.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpserver.Error(w, errs.Invalid("invalid_id", "Identifier is invalid", nil))
		return
	}
	result, err := h.service.Get(r.Context(), rid)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}
