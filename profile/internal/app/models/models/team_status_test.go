package models

import "testing"

func TestTeamStatusForSize(t *testing.T) {
	cases := []struct {
		size int64
		want TeamStatus
	}{
		{0, TeamStatusForming},
		{1, TeamStatusForming},
		{2, TeamStatusForming},
		{3, TeamStatusReady},
		{4, TeamStatusReady},
		{5, TeamStatusFull},
		{6, TeamStatusFull},
	}

	for _, tc := range cases {
		if got := TeamStatusForSize(tc.size); got != tc.want {
			t.Errorf("size %d: expected %s, got %s", tc.size, tc.want, got)
		}
	}
}

func TestTeamAndRosterUseTheSameRule(t *testing.T) {
	if got := (Team{MemberCount: 3}).Status(); got != TeamStatusReady {
		t.Fatalf("team of 3 must be ready, got %s", got)
	}

	roster := TeamRoster{Members: make([]DeckCandidate, 5)}
	if got := roster.Status(); got != TeamStatusFull {
		t.Fatalf("roster of 5 must be full, got %s", got)
	}
}
