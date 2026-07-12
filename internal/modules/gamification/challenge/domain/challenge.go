package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type ChallengeStatus string

const (
	StatusDraft       ChallengeStatus = "draft"
	StatusActive      ChallengeStatus = "active"
	StatusUnderReview ChallengeStatus = "under_review"
	StatusCompleted   ChallengeStatus = "completed"
	StatusArchived    ChallengeStatus = "archived"
)

// Transitions mirrors the domain state machine — keep TS CHALLENGE_TRANSITIONS in sync.
var Transitions = map[ChallengeStatus][]ChallengeStatus{
	StatusDraft:       {StatusActive, StatusArchived},
	StatusActive:      {StatusUnderReview, StatusArchived},
	StatusUnderReview: {StatusCompleted, StatusActive, StatusArchived},
	StatusCompleted:   {StatusArchived},
}

type Challenge struct {
	ID               id.ID           `json:"id"`
	Title            string          `json:"title"`
	CategoryID       id.ID           `json:"categoryId"`
	Description      string          `json:"description"`
	XP               int             `json:"xp"`
	Difficulty       string          `json:"difficulty"`
	EvidenceRequired bool            `json:"evidenceRequired"`
	Deadline         *time.Time      `json:"deadline,omitempty"`
	Status           ChallengeStatus `json:"status"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
	// read helpers
	PendingCount int `json:"pendingCount,omitempty"`
}

func New(title string, categoryID id.ID, description string, xp int, difficulty string, evidenceRequired bool, deadline *time.Time, now time.Time) (*Challenge, error) {
	c := &Challenge{
		ID:               id.New(),
		Title:            strings.TrimSpace(title),
		CategoryID:       categoryID,
		Description:      strings.TrimSpace(description),
		XP:               xp,
		Difficulty:       strings.TrimSpace(difficulty),
		EvidenceRequired: evidenceRequired,
		Deadline:         deadline,
		Status:           StatusDraft,
		CreatedAt:        now.UTC(),
		UpdatedAt:        now.UTC(),
	}
	if c.Difficulty == "" {
		c.Difficulty = "medium"
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Challenge) Validate() error {
	fields := map[string]string{}
	if c.Title == "" || len(c.Title) > 200 {
		fields["title"] = "Title is required and must be at most 200 characters"
	}
	if c.CategoryID == "" {
		fields["categoryId"] = "Category is required"
	}
	if c.XP < 0 {
		fields["xp"] = "XP cannot be negative"
	}
	switch strings.ToLower(c.Difficulty) {
	case "easy", "medium", "hard":
		c.Difficulty = strings.ToLower(c.Difficulty)
	default:
		fields["difficulty"] = "Difficulty must be easy, medium, or hard"
	}
	if len(fields) > 0 {
		return errs.Invalid("invalid_challenge", "Challenge details are invalid", fields)
	}
	return nil
}

func (c *Challenge) Transition(to ChallengeStatus) error {
	allowed := Transitions[c.Status]
	for _, ok := range allowed {
		if ok == to {
			c.Status = to
			return nil
		}
	}
	return errs.Invalid("bad_transition", fmt.Sprintf("%s → %s not allowed", c.Status, to), nil)
}

func AllowedTargets(from ChallengeStatus) []ChallengeStatus {
	return append([]ChallengeStatus(nil), Transitions[from]...)
}
