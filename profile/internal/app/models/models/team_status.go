package models

type TeamStatus string

const (
	TeamStatusForming TeamStatus = "forming"
	TeamStatusReady   TeamStatus = "ready"
	TeamStatusFull    TeamStatus = "full"

	readyFromSize = 3
	fullSize      = 5
)

// TeamStatusForSize derives the status from the current member count:
// 1-2 forming, 3-4 ready, 5 (or more, which the server never allows) full.
// It is computed, never stored, so it cannot drift when membership changes.
func TeamStatusForSize(size int64) TeamStatus {
	switch {
	case size >= fullSize:
		return TeamStatusFull
	case size >= readyFromSize:
		return TeamStatusReady
	default:
		return TeamStatusForming
	}
}

func (t Team) Status() TeamStatus {
	return TeamStatusForSize(t.MemberCount)
}

func (r TeamRoster) Status() TeamStatus {
	return TeamStatusForSize(int64(len(r.Members)))
}
