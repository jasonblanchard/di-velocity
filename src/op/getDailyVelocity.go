package op

import (
	"database/sql"
	"time"

	"github.com/jasonblanchard/di-velocity/src/domain"
)

// TODO: GetDailyVelocity => GetDailyVelocityScores

// GetDailyVelocity returns velocity score for each day between start and end (inclusive)
func GetDailyVelocity(db *sql.DB, start time.Time, end time.Time) (domain.DailyVelocities, error) {
	// TODO: Include date range
	rows, err := db.Query("SELECT count, day, creator_id FROM daily_counts WHERE creator_id = $1", "1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var velocities = make([]domain.DailyVelocity, 0)

	for rows.Next() {
		dailyVelocity := domain.DailyVelocity{}
		// TODO: Compute score based on counter
		// ...or do this outside
		if err := rows.Scan(&dailyVelocity.Score, &dailyVelocity.Day, &dailyVelocity.CreatorID); err != nil {
			return nil, err
		}
		velocities = append(velocities, dailyVelocity)
	}

	return velocities, nil
}
