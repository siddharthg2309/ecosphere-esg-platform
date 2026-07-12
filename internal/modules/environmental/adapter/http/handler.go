package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
	carbonapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/carbon/app"
	carbondomain "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/carbon/domain"
	carbonport "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/carbon/port"
	goalapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/goal/app"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/domain"
	platformauth "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/auth"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type Handler struct {
	carbon *carbonapp.Service
	goals  *goalapp.Service
	ingest *carbonapp.IngestService
}

func New(carbon *carbonapp.Service, goals *goalapp.Service, ingest *carbonapp.IngestService) *Handler {
	return &Handler{carbon: carbon, goals: goals, ingest: ingest}
}

func (h *Handler) Ingest(w http.ResponseWriter, r *http.Request) {
	principal, _ := platformauth.PrincipalFrom(r.Context())
	if principal.Role == domain.RoleAuditor {
		httpserver.Error(w, errs.Forbidden("role_forbidden", "Auditors cannot ingest carbon evidence"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, carbonapp.MaxEvidenceSize+(1<<20))
	if err := r.ParseMultipartForm(carbonapp.MaxEvidenceSize); err != nil {
		httpserver.Error(w, errs.Invalid("invalid_upload", "Upload is invalid or too large", nil))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpserver.Error(w, errs.Invalid("file_required", "A document is required", map[string]string{"file": "Select a file"}))
		return
	}
	defer file.Close()
	peek := make([]byte, 512)
	n, readErr := io.ReadFull(file, peek)
	if readErr != nil && readErr != io.ErrUnexpectedEOF {
		httpserver.Error(w, errs.Invalid("invalid_upload", "Could not read the document", nil))
		return
	}
	peek = peek[:n]
	mime := http.DetectContentType(peek)
	result, err := h.ingest.Execute(r.Context(), header.Filename, mime, header.Size, io.MultiReader(bytes.NewReader(peek), file))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

type transactionRequest struct {
	DepartmentID     string              `json:"departmentId"`
	Source           carbondomain.Source `json:"source"`
	Quantity         string              `json:"quantity"`
	EmissionFactorID string              `json:"emissionFactorId"`
	Unit             string              `json:"unit"`
	TxnDate          string              `json:"txnDate"`
	EvidenceURL      string              `json:"evidenceUrl"`
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var body transactionRequest
	if err := decode(r, &body); err != nil {
		httpserver.Error(w, err)
		return
	}
	principal, _ := platformauth.PrincipalFrom(r.Context())
	departmentID, err := scopedDepartment(principal, body.DepartmentID, false)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	factorID, err := parseID(body.EmissionFactorID, "emissionFactorId")
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	quantity, err := decimal.NewFromString(body.Quantity)
	if err != nil {
		httpserver.Error(w, errs.Invalid("invalid_quantity", "Quantity must be a decimal", map[string]string{"quantity": "Invalid decimal"}))
		return
	}
	txnDate, err := time.Parse("2006-01-02", body.TxnDate)
	if err != nil {
		httpserver.Error(w, errs.Invalid("invalid_date", "Transaction date must use YYYY-MM-DD", map[string]string{"txnDate": "Invalid date"}))
		return
	}
	result, err := h.carbon.Record(r.Context(), carbonapp.RecordCommand{DepartmentID: departmentID, Source: body.Source, Quantity: quantity, EmissionFactorID: factorID, Unit: body.Unit, TxnDate: txnDate, EvidenceURL: body.EvidenceURL})
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, result)
}
func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	principal, _ := platformauth.PrincipalFrom(r.Context())
	transactionID, err := parseID(chi.URLParam(r, "id"), "id")
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.carbon.Verify(r.Context(), transactionID, principal.UserID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}
func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	principal, _ := platformauth.PrincipalFrom(r.Context())
	departmentID, err := queryDepartment(principal, r.URL.Query().Get("dept"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	filter := carbonport.Filter{DepartmentID: departmentID, Page: pagination(r)}
	if v := r.URL.Query().Get("from"); v != "" {
		x, e := time.Parse("2006-01-02", v)
		if e != nil {
			httpserver.Error(w, errs.Invalid("invalid_date_range", "From date is invalid", nil))
			return
		}
		filter.From = &x
	}
	if v := r.URL.Query().Get("to"); v != "" {
		x, e := time.Parse("2006-01-02", v)
		if e != nil {
			httpserver.Error(w, errs.Invalid("invalid_date_range", "To date is invalid", nil))
			return
		}
		filter.To = &x
	}
	if v := r.URL.Query().Get("source"); v != "" {
		x := carbondomain.Source(v)
		if !x.Valid() {
			httpserver.Error(w, errs.Invalid("invalid_source", "Source is invalid", nil))
			return
		}
		filter.Source = &x
	}
	if v := r.URL.Query().Get("status"); v != "" {
		x := carbondomain.Status(v)
		if x != carbondomain.StatusDraft && x != carbondomain.StatusVerified {
			httpserver.Error(w, errs.Invalid("invalid_status", "Status is invalid", nil))
			return
		}
		filter.Status = &x
	}
	result, err := h.carbon.List(r.Context(), filter)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	principal, _ := platformauth.PrincipalFrom(r.Context())
	departmentID, err := queryDepartment(principal, r.URL.Query().Get("dept"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	now := time.Now().UTC()
	from := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	to := now
	if v := r.URL.Query().Get("from"); v != "" {
		from, err = time.Parse("2006-01-02", v)
	}
	if err == nil {
		if v := r.URL.Query().Get("to"); v != "" {
			to, err = time.Parse("2006-01-02", v)
		}
	}
	if err != nil || from.After(to) {
		httpserver.Error(w, errs.Invalid("invalid_date_range", "Date range is invalid", nil))
		return
	}
	result, err := h.carbon.Summary(r.Context(), departmentID, from, to)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

type goalRequest struct {
	Name         string `json:"name"`
	DepartmentID string `json:"departmentId"`
	TargetCO2    string `json:"targetCo2"`
	CurrentCO2   string `json:"currentCo2"`
	Deadline     string `json:"deadline"`
}

func (h *Handler) ListGoals(w http.ResponseWriter, r *http.Request) {
	principal, _ := platformauth.PrincipalFrom(r.Context())
	departmentID, err := queryDepartment(principal, r.URL.Query().Get("dept"))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.goals.List(r.Context(), departmentID, pagination(r))
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}
func (h *Handler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	var body goalRequest
	if err := decode(r, &body); err != nil {
		httpserver.Error(w, err)
		return
	}
	principal, _ := platformauth.PrincipalFrom(r.Context())
	departmentID, err := scopedDepartment(principal, body.DepartmentID, true)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	if principal.Role == domain.RoleDeptHead {
		ok, e := h.carbon.IsDepartmentHead(r.Context(), principal.UserID, departmentID)
		if e != nil {
			httpserver.Error(w, e)
			return
		}
		if !ok {
			httpserver.Error(w, errs.Forbidden("not_dept_head", "You can manage goals only for your department"))
			return
		}
	}
	target, current, deadline, err := goalValues(body)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.goals.Create(r.Context(), body.Name, departmentID, target, current, deadline)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, result)
}
func (h *Handler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	goalID, err := parseID(chi.URLParam(r, "id"), "id")
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	existing, err := h.goals.ByID(r.Context(), goalID)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	principal, _ := platformauth.PrincipalFrom(r.Context())
	if principal.Role == domain.RoleDeptHead {
		ok, e := h.carbon.IsDepartmentHead(r.Context(), principal.UserID, existing.DepartmentID)
		if e != nil {
			httpserver.Error(w, e)
			return
		}
		if !ok {
			httpserver.Error(w, errs.Forbidden("not_dept_head", "You can manage goals only for your department"))
			return
		}
	}
	var body goalRequest
	if err = decode(r, &body); err != nil {
		httpserver.Error(w, err)
		return
	}
	target, _, deadline, err := goalValues(body)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	result, err := h.goals.Update(r.Context(), goalID, body.Name, target, deadline)
	if err != nil {
		httpserver.Error(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, result)
}

func decode(r *http.Request, v any) error {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return errs.Invalid("invalid_request", "Request body is invalid", nil)
	}
	return nil
}
func parseID(v, field string) (id.ID, error) {
	x, err := id.Parse(v)
	if err != nil {
		return "", errs.Invalid("invalid_id", "ID must be a UUID", map[string]string{field: "Invalid UUID"})
	}
	return x, nil
}
func pagination(r *http.Request) page.Page {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return page.New(limit, offset)
}
func scopedDepartment(p platformauth.Principal, requested string, manage bool) (id.ID, error) {
	if manage && p.Role != domain.RoleAdmin && p.Role != domain.RoleDeptHead {
		return "", errs.Forbidden("role_forbidden", "Only department heads and admins can manage goals")
	}
	if p.Role != domain.RoleAdmin {
		if p.DepartmentID == nil {
			return "", errs.Forbidden("department_required", "Your account is not assigned to a department")
		}
		if requested != "" && requested != p.DepartmentID.String() {
			return "", errs.Forbidden("department_forbidden", "You cannot access another department")
		}
		return *p.DepartmentID, nil
	}
	return parseID(requested, "departmentId")
}
func queryDepartment(p platformauth.Principal, requested string) (*id.ID, error) {
	if p.Role == domain.RoleEmployee || p.Role == domain.RoleDeptHead {
		v, err := scopedDepartment(p, requested, false)
		return &v, err
	}
	if requested == "" {
		return nil, nil
	}
	v, err := parseID(requested, "dept")
	return &v, err
}
func goalValues(body goalRequest) (decimal.Decimal, decimal.Decimal, time.Time, error) {
	target, e := decimal.NewFromString(body.TargetCO2)
	if e != nil {
		return decimal.Zero, decimal.Zero, time.Time{}, errs.Invalid("invalid_target", "Target CO2 must be a decimal", nil)
	}
	current := decimal.Zero
	if body.CurrentCO2 != "" {
		current, e = decimal.NewFromString(body.CurrentCO2)
	}
	if e != nil {
		return decimal.Zero, decimal.Zero, time.Time{}, errs.Invalid("invalid_current", "Current CO2 must be a decimal", nil)
	}
	deadline, e := time.Parse("2006-01-02", body.Deadline)
	if e != nil {
		return decimal.Zero, decimal.Zero, time.Time{}, errs.Invalid("invalid_deadline", "Deadline must use YYYY-MM-DD", nil)
	}
	return target, current, deadline, nil
}
