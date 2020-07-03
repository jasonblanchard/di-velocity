package op

import (
	"database/sql"
	"time"
)

// IncrementDailyCounter Creates or increments a velocity event
func IncrementDailyCounter(db *sql.DB, day time.Time, creatorID string) error {
	rows, err := db.Query("SELECT count FROM daily_counts WHERE day = $1 AND creator_id = $2", day, creatorID)
	defer rows.Close()
	if err != nil {
		return err
	}

	if rows.Next() {
		// get the score
		// increment
		// update
		return nil
	}

	// write row
	// TODO: Change "score" to "counter" to express what it really is? "score" is what's derived on the way out.
	_, err = db.Query("INSERT INTO daily_counts (day, count, creator_id) VALUES ($1, $2, $3)", day, 1, creatorID)
	if err != nil {
		return err
	}

	return nil
}
