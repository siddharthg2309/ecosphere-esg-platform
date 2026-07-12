package domain

import "testing"

func TestUnlockRuleSatisfied(t *testing.T) {
	tests := []struct {
		rule          UnlockRule
		xp, completed int
		want          bool
	}{{UnlockRule{"xp", 100}, 100, 0, true}, {UnlockRule{"xp", 100}, 99, 5, false}, {UnlockRule{"challenges", 3}, 0, 3, true}}
	for _, tt := range tests {
		if got := tt.rule.Satisfied(tt.xp, tt.completed); got != tt.want {
			t.Errorf("got %v want %v", got, tt.want)
		}
	}
}
