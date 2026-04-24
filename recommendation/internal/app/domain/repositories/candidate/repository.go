package candidate

import (
	"context"
	"fmt"

	"winx-recommendation/internal/app/domain/repositories/preferences"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Row is one scored candidate returned by the SQL query.
type Row struct {
	UserID          int64
	Name            string
	City            string
	Country         string
	PhotoURL        string
	SharedInterests int
	HasPhoto        bool
	DistanceMeters  *float64
}

type Repository interface {
	Score(ctx context.Context, userID int64, uc preferences.UserContext) ([]Row, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Score runs the main recommendation scoring query.
//
// Hard filters applied in SQL:
//   - Exclude users already swiped on
//   - Exclude users already matched with
//   - Distance constraint (unless show_me_global)
//
// Scoring formula (higher = better):
//
//	shared_interests × 3
//	+ has_photo × 2
//	- distance_km × 0.2
func (r *repository) Score(ctx context.Context, userID int64, uc preferences.UserContext) ([]Row, error) {
	// Build the location expression once so we can reuse it inline.
	// Using fmt.Sprintf with float literals is safe (no user input).
	const maxCandidates = 200

	maxDistanceMeters := float64(uc.Preferences.MaxDistanceKM) * 1000

	// When the user has no interests yet ANY(ARRAY[]::bigint[]) still works fine.
	interestIDs := uc.InterestIDs
	if len(interestIDs) == 0 {
		interestIDs = []int64{} // keep it an empty slice, not nil
	}

	hasLocation := uc.Latitude != nil && uc.Longitude != nil

	// locationExpr is a PostGIS geography literal for the user's position.
	// Falls back to NULL when location is unknown (show_me_global covers this).
	var locationExpr string
	if hasLocation {
		locationExpr = fmt.Sprintf(
			"ST_SetSRID(ST_MakePoint(%f, %f), 4326)::geography",
			*uc.Longitude, *uc.Latitude,
		)
	} else {
		locationExpr = "NULL::geography"
	}

	// distanceExpr: metres between candidate and user (NULL when no location).
	distanceExpr := fmt.Sprintf(
		"CASE WHEN p.current_location IS NOT NULL AND %s IS NOT NULL THEN ST_Distance(p.current_location, %s) ELSE NULL END",
		locationExpr, locationExpr,
	)

	// withinExpr: TRUE when candidate is within max distance (or global).
	var withinExpr string
	if uc.Preferences.ShowMeGlobal || !hasLocation {
		withinExpr = "TRUE"
	} else {
		withinExpr = fmt.Sprintf(
			"(p.current_location IS NOT NULL AND ST_DWithin(p.current_location, %s, %f))",
			locationExpr, maxDistanceMeters,
		)
	}

	query := fmt.Sprintf(`
		WITH excluded AS (
			SELECT swiped_id AS uid FROM swipes WHERE swiper_id = ?
			UNION ALL
			SELECT user_two_id FROM matches WHERE user_one_id = ?
			UNION ALL
			SELECT user_one_id  FROM matches WHERE user_two_id = ?
		),
		shared AS (
			SELECT ui.user_id, COUNT(*) AS cnt
			FROM   user_interests ui
			WHERE  ui.interest_id = ANY(?)
			GROUP  BY ui.user_id
		)
		SELECT
			p.user_id,
			COALESCE(p.name, '')                          AS name,
			COALESCE(p.city, '')                          AS city,
			COALESCE(p.country, '')                       AS country,
			COALESCE(ph.url, '')                          AS photo_url,
			COALESCE(sh.cnt, 0)                           AS shared_interests,
			(ph.user_id IS NOT NULL)                      AS has_photo,
			%s                                            AS distance_meters
		FROM  profiles p
		LEFT  JOIN shared sh ON sh.user_id = p.user_id
		LEFT  JOIN (
			SELECT DISTINCT ON (user_id) user_id, url
			FROM   profile_photos
			ORDER  BY user_id, id ASC
		) ph ON ph.user_id = p.user_id
		WHERE
			p.user_id != ?
			AND p.user_id NOT IN (SELECT uid FROM excluded)
			AND %s
		ORDER BY
			(COALESCE(sh.cnt, 0) * 3
			 + CASE WHEN ph.user_id IS NOT NULL THEN 2 ELSE 0 END
			 - COALESCE(%s / 5000.0, 0)
			) DESC
		LIMIT ?
	`, distanceExpr, withinExpr, distanceExpr)

	var rows []Row
	err := r.db.WithContext(ctx).Raw(
		query,
		userID, userID, userID, // excluded CTEs
		pq.Array(interestIDs), // ANY(?)
		userID,                // p.user_id != ?
		maxCandidates,         // LIMIT
	).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("score candidates: %w", err)
	}

	return rows, nil
}
