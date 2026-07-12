package events

import "testing"

func TestCatalogEventNamesAreUnique(t *testing.T) {
	names := []string{
		(ESGConfigChanged{}).Name(),
		(EmissionRecorded{}).Name(),
		(ParticipationDecided{}).Name(),
		(ChallengeCompleted{}).Name(),
		(BadgeUnlocked{}).Name(),
		(RewardRedeemed{}).Name(),
		(PolicyPublished{}).Name(),
		(ComplianceIssueRaised{}).Name(),
		(ComplianceOverdue{}).Name(),
	}
	seen := map[string]bool{}
	for _, name := range names {
		if name == "" || seen[name] {
			t.Fatalf("event name must be non-empty and unique: %q", name)
		}
		seen[name] = true
	}
}
