package op

import (
	"database/sql"
	"time"

	"github.com/jasonblanchard/di-velocity/src/domain"
)

// GetDailyCounts returns velocity score for each day between start and end (inclusive)
func GetDailyCounts(db *sql.DB, start time.Time, end time.Time) (domain.DailyCounts, error) {
	rows, err := db.Query(`
SELECT count, day, creator_id
FROM daily_counts
WHERE creator_id = $1
AND day >= $2
AND day <= $3
`, "1", start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dailyCounts = make(domain.DailyCounts, 0)

	for rows.Next() {
		dailyCount := domain.DailyCount{}
		if err := rows.Scan(&dailyCount.Count, &dailyCount.Day, &dailyCount.CreatorID); err != nil {
			return nil, err
		}
		dailyCounts = append(dailyCounts, dailyCount)
	}

	return dailyCounts, nil
}
