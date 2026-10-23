package models

// DeckCandidate is the academic-only projection of a student shown in a
// course deck. It deliberately carries no contacts, photos or Social data.
type DeckCandidate struct {
	UserID        int64
	Name          string
	Skills        []string
	PreferredRole string
}
