package domain

import (
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"strings"
	"time"
)

type UnlockRule struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

func (r UnlockRule) Validate() error {
	if (r.Type != "xp" && r.Type != "challenges") || r.Value <= 0 {
		return errs.Invalid("invalid_unlock_rule", "Unlock rule requires xp or challenges and a positive value", nil)
	}
	return nil
}
func (r UnlockRule) Satisfied(xp, completed int) bool {
	if r.Type == "xp" {
		return xp >= r.Value
	}
	if r.Type == "challenges" {
		return completed >= r.Value
	}
	return false
}

type Badge struct {
	ID          id.ID      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	UnlockRule  UnlockRule `json:"unlockRule"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func New(name, description, icon string, rule UnlockRule, now time.Time) (Badge, error) {
	if strings.TrimSpace(name) == "" {
		return Badge{}, errs.Invalid("invalid_badge", "Badge name is required", nil)
	}
	if err := rule.Validate(); err != nil {
		return Badge{}, err
	}
	return Badge{ID: id.New(), Name: strings.TrimSpace(name), Description: description, Icon: icon, UnlockRule: rule, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}, nil
}
