package models

// TeamRoster is what a team member may see about their own team: team name,
// course and members' academic fields only. No contacts, photos or choices.
type TeamRoster struct {
	TeamID      int64
	TeamName    string
	CourseID    int64
	CourseCode  string
	CourseTitle string
	Members     []DeckCandidate
}
