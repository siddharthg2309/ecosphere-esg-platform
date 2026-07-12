package domain

import (
	"testing"
	"time"
)

func TestRoleValidation(t *testing.T) {
	tests := []struct {
		role  Role
		valid bool
	}{{RoleEmployee, true}, {RoleDeptHead, true}, {RoleAuditor, true}, {RoleAdmin, true}, {Role("owner"), false}}
	for _, tt := range tests {
		if tt.role.Valid() != tt.valid {
			t.Errorf("role %q valid=%v", tt.role, tt.valid)
		}
	}
}

func TestNewUserValidatesEmail(t *testing.T) {
	if _, err := NewUser("Alex", "not-email", "hash", RoleEmployee, nil, time.Now()); err == nil {
		t.Fatal("expected validation error")
	}
}
