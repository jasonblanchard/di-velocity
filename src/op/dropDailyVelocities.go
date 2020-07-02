package op

import (
	"database/sql"
)

// DropDailyVelocities Drops velocities table
func DropDailyVelocities(db *sql.DB) error {
	rows, err := db.Query("DELETE FROM velocities")
	if rows != nil {
		defer rows.Close()
	}
	return err
}
