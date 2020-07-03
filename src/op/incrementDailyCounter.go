package op

import (
	"database/sql"
	"time"
)

// IncrementDailyCounter Creates or increments a velocity event
func IncrementDailyCounter(db *sql.DB, day time.Time, creatorID string) error {
	rows, err := db.Query("SELECT id, count FROM daily_counts WHERE day = $1 AND creator_id = $2", day, creatorID)
	defer rows.Close()
	if err != nil {
		return err
	}

	if rows.Next() {
		var id int
		var count int
		rows.Scan(&id, &count)
		count = count + 1
		_, err = db.Query("UPDATE daily_counts SET count = $1 WHERE id = $2", count, id)
		if err != nil {
			return err
		}
		return nil
	}

	_, err = db.Query("INSERT INTO daily_counts (day, count, creator_id) VALUES ($1, $2, $3)", day, 1, creatorID)
	if err != nil {
		return err
	}

	return nil
}
