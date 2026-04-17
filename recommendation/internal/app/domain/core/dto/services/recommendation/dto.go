package recommendation

// ListDTO — parameters passed from handler to service.
type ListDTO struct {
	UserID int64
	Limit  int
	Offset int
}

// ScoredCandidate is what the service returns after running the full algorithm.
type ScoredCandidate struct {
	UserID          int64
	Score           float64
	SharedInterests int
	HasPhoto        bool
	DistanceKM      *float64 // nil when location is unknown
	City            string
	Country         string
	// Enriched from MongoDB
	Gender     string
	Age        int
	LookingFor string
}
