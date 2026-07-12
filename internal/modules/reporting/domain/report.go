package domain

import (
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type ReportType string

const (
	TypeEnvironmental ReportType = "environmental"
	TypeSocial        ReportType = "social"
	TypeGovernance    ReportType = "governance"
	TypeESGSummary    ReportType = "esg_summary"
	TypeCustom        ReportType = "custom"
)

type Filters struct {
	DepartmentID *id.ID    `json:"departmentId,omitempty"`
	From         *time.Time `json:"from,omitempty"`
	To           *time.Time `json:"to,omitempty"`
	Module       *string   `json:"module,omitempty"`
	Employee     *string   `json:"employee,omitempty"`
	Challenge    *string   `json:"challenge,omitempty"`
	Category     *string   `json:"category,omitempty"`
}

type Section struct {
	Title   string              `json:"title"`
	Summary string              `json:"summary,omitempty"`
	Rows    []map[string]string `json:"rows,omitempty"`
	Metrics map[string]any      `json:"metrics,omitempty"`
	AI      bool                `json:"ai,omitempty"`
}

type Report struct {
	ID          id.ID     `json:"id"`
	Type        ReportType `json:"type"`
	Filters     Filters   `json:"filters"`
	Sections    []Section `json:"sections"`
	GeneratedAt time.Time `json:"generatedAt"`
	GeneratedBy *id.ID    `json:"generatedBy,omitempty"`
}
