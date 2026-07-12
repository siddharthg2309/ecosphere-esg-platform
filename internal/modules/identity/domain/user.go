package domain

import (
	"net/mail"
	"strings"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type Role string

const (
	RoleEmployee Role = "employee"
	RoleDeptHead Role = "dept_head"
	RoleAuditor  Role = "auditor"
	RoleAdmin    Role = "admin"
)

func (r Role) Valid() bool {
	switch r {
	case RoleEmployee, RoleDeptHead, RoleAuditor, RoleAdmin:
		return true
	default:
		return false
	}
}

type User struct {
	ID           id.ID     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	DepartmentID *id.ID    `json:"departmentId,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

func NewUser(name, email, passwordHash string, role Role, departmentID *id.ID, now time.Time) (*User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	fields := map[string]string{}
	if name == "" {
		fields["name"] = "Name is required"
	}
	if _, err := mail.ParseAddress(email); err != nil {
		fields["email"] = "A valid email is required"
	}
	if !role.Valid() {
		fields["role"] = "Role is invalid"
	}
	if passwordHash == "" {
		fields["password"] = "Password is required"
	}
	if len(fields) > 0 {
		return nil, errs.Invalid("invalid_user", "User details are invalid", fields)
	}
	return &User{ID: id.New(), Name: name, Email: email, PasswordHash: passwordHash, Role: role, DepartmentID: departmentID, CreatedAt: now.UTC()}, nil
}
