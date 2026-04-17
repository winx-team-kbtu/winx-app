package recommendation

// ListDTO — parameters for querying the recommendations table (future use).
type ListDTO struct {
	UserID int64
	Limit  int
	Offset int
}
