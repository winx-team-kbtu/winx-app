package models

import "time"

// TeamRoster is the team and its members' academic information from one snapshot.
type TeamRoster struct {
	TeamID      int64
	TeamName    string
	CourseID    int64
	CourseCode  string
	CourseTitle string
	Members     []DeckCandidate
	CreatedBy   *int64
	JoinCode    string
	CreatedAt   time.Time
}

func (r TeamRoster) Team() Team {
	return Team{
		ID: r.TeamID, Name: r.TeamName, CourseID: r.CourseID, CreatedBy: r.CreatedBy,
		JoinCode: r.JoinCode, CreatedAt: r.CreatedAt, MemberCount: int64(len(r.Members)),
	}
}
