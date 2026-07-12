package domain

import (
	"strings"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type Training struct {
	ID          id.ID     `json:"id"`
	Name        string    `json:"name"`
	AssignedTo  string    `json:"assignedTo"`
	Status      string    `json:"status"`
	Completed   int       `json:"completed"`
	Total       int       `json:"total"`
	CreatedAt   time.Time `json:"createdAt"`
}

func New(name, assignedTo string, now time.Time) (*Training, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errs.Invalid("invalid_training", "Training name is required", nil)
	}
	if assignedTo == "" {
		assignedTo = "All employees"
	}
	return &Training{
		ID:         id.New(),
		Name:       name,
		AssignedTo: assignedTo,
		Status:     "active",
		CreatedAt:  now.UTC(),
	}, nil
}

type Completion struct {
	ID         id.ID     `json:"id"`
	EmployeeID id.ID     `json:"employeeId"`
	TrainingID id.ID     `json:"trainingId"`
	CompletedAt time.Time `json:"completedAt"`
}
