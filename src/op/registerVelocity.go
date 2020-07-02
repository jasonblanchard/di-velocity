package op

import (
	"database/sql"
	"time"
)

// RegisterVelocity Creates or increments a velocity event
func RegisterVelocity(db *sql.DB, day time.Time, creatorID string) error {
	rows, err := db.Query("SELECT score FROM velocities WHERE day = $1 AND creator_id = $2", day, creatorID)
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
	_, err = db.Query("INSERT INTO velocities (day, score, creator_id) VALUES ($1, $2, $3)", day, 1, creatorID)
	if err != nil {
		return err
	}

	return nil
}
