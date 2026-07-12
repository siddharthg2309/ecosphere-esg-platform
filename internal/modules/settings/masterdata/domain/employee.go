package domain

import (
	identity "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"time"
)

type Employee struct {
	ID                  id.ID         `json:"id"`
	Name                string        `json:"name"`
	Email               string        `json:"email"`
	Role                identity.Role `json:"role"`
	DepartmentID        *id.ID        `json:"departmentId,omitempty"`
	XP                  int           `json:"xp"`
	Points              int           `json:"points"`
	CompletedChallenges int           `json:"completedChallenges"`
	Status              string        `json:"status"`
	CreatedAt           time.Time     `json:"createdAt"`
}
